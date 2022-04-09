package writer

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/djboris9/rea/pkg/document"
	"github.com/djboris9/rea/pkg/engine"
	"github.com/djboris9/rea/pkg/xmltree"
)

// Write takes a text or text-template file, processed it with the given
// configuration and writes the result to the writer.
func Write(tmpl document.PackagedDocument, config *TemplateConfig, out io.Writer) (*TemplateProcessingData, error) {
	tpd := &TemplateProcessingData{
		TemplateMimeType: tmpl.MIMEType(),
	}

	// Check for text or text-template mimetype
	if tmpl.MIMEType() != "application/vnd.oasis.opendocument.text-template" &&
		tmpl.MIMEType() != "application/vnd.oasis.opendocument.text" {
		return tpd, fmt.Errorf("Unsupported mimetype: %s", tmpl.MIMEType())
	}

	// Get content.xml
	tmplContentXML, err := tmpl.Open("content.xml")
	if err != nil {
		return tpd, fmt.Errorf("loading content.xml from template: %w", err)
	}

	tmplContent, err := ioutil.ReadAll(tmplContentXML)
	if err != nil {
		return tpd, fmt.Errorf("reading content.xml from template: %w", err)
	}

	// Run engine. TODO: With passed data
	tree, err := xmltree.Parse(tmplContent)
	if err != nil {
		return tpd, fmt.Errorf("parsing content.xml as tree: %w", err)
	}
	tpd.TemplateXMLTree = tree

	lt, err := engine.NewLuaTree(tree)
	if err != nil {
		return tpd, fmt.Errorf("creating lua tree from content.xml: %w", err)
	}
	tpd.TemplateLuaProg = lt.LuaProg
	tpd.TemplateLuaNodeList = lt.NodeList

	e := engine.NewLuaEngine(lt, nil)
	err = e.Exec()
	if err != nil {
		return tpd, fmt.Errorf("executing lua engine: %w", err)
	}
	tpd.LuaNodePathStr = e.GetNodePathString()

	var buf strings.Builder
	err = e.WriteXML(&buf)
	if err != nil {
		return tpd, fmt.Errorf("writing executed template: %w", err)
	}

	content := buf.String()
	tpd.ContentXML = content

	// Write file, overriding mimetype and content.xml
	// TODO: Override/Delete thumbnail and remove it from the manifest.xml
	ov := document.Overrides{
		"mimetype": document.Override{
			Data: []byte("application/vnd.oasis.opendocument.text"),
		},
		"content.xml": document.Override{
			Data: []byte(content),
		},
	}
	err = tmpl.Write(out, ov)
	if err != nil {
		return tpd, fmt.Errorf("writing rendered template: %w", err)
	}

	return tpd, nil
}

type TemplateConfig struct {
	UserData any
	// MetaData struct { Author etc. }
	// Style overrides etc.
}

type TemplateProcessingData struct {
	// Data of template
	TemplateMimeType    string
	TemplateXMLTree     *xmltree.Node
	TemplateLuaProg     string
	TemplateLuaNodeList []*xmltree.Node

	// Processed data
	LuaNodePathStr []string
	ContentXML     string
}
