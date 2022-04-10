package writer

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/djboris9/rea/internal/utils"
	"github.com/djboris9/rea/pkg/document"
	"github.com/djboris9/rea/pkg/engine"
	"github.com/djboris9/rea/pkg/xmltree"
)

// Opens, reads and return the content.xml as XML node.
func getContentFromTemplateAsXML(tmpl document.PackagedDocument) (*xmltree.Node, error) {
	// Check for text or text-template mimetype
	if tmpl.MIMEType() != "application/vnd.oasis.opendocument.text-template" &&
		tmpl.MIMEType() != "application/vnd.oasis.opendocument.text" {
		return nil, utils.FormatError(document.ErrMimetype, fmt.Sprintf("Unsupported mimetype: %s", tmpl.MIMEType()))
	}

	tmplContentXML, err := tmpl.Open("content.xml")
	if err != nil {
		return nil, fmt.Errorf("loading content.xml from template: %w", err)
	}

	tmplContent, err := ioutil.ReadAll(tmplContentXML)
	if err != nil {
		return nil, fmt.Errorf("reading content.xml from template: %w", err)
	}

	tree, err := xmltree.Parse(tmplContent)
	if err != nil {
		return nil, fmt.Errorf("parsing content.xml as tree: %w", err)
	}

	return tree, nil
}

// Run engine. TODO: Pass data from cli input.
func processTemplateWithLuaEnginge(xmlTree *xmltree.Node, templateData *TemplateProcessingData) error {
	templateData.TemplateXMLTree = xmlTree
	luaTree, err := engine.NewLuaTree(xmlTree)

	if err != nil {
		return fmt.Errorf("creating lua tree from content.xml: %w", err)
	}

	templateData.TemplateLuaProg = luaTree.LuaProg
	templateData.TemplateLuaNodeList = luaTree.NodeList

	luaEngine := engine.NewLuaEngine(luaTree, nil)
	err = luaEngine.Exec()

	if err != nil {
		return fmt.Errorf("executing lua engine: %w", err)
	}

	templateData.LuaNodePathStr = luaEngine.GetNodePathString()

	var buf strings.Builder
	err = luaEngine.WriteXML(&buf)

	if err != nil {
		return fmt.Errorf("writing executed template: %w", err)
	}

	content := buf.String()
	templateData.ContentXML = content

	return nil
}

// Write takes a text or text-template file, processed it with the given
// configuration and writes the result to the writer.
func Write(tmpl document.PackagedDocument, config *TemplateConfig, out io.Writer) (*TemplateProcessingData, error) {
	templateData := &TemplateProcessingData{
		TemplateMimeType: tmpl.MIMEType(),
	}

	xmlTree, err := getContentFromTemplateAsXML(tmpl)
	if err != nil {
		return templateData, err
	}

	err = processTemplateWithLuaEnginge(xmlTree, templateData)
	if err != nil {
		return templateData, err
	}

	// Write file, overriding mimetype and content.xml
	// TODO: Override/Delete thumbnail and remove it from the manifest.xml
	ov := document.Overrides{
		"mimetype": document.Override{
			Data: []byte("application/vnd.oasis.opendocument.text"),
		},
		"content.xml": document.Override{
			Data: []byte(templateData.ContentXML),
		},
	}
	err = tmpl.Write(out, ov)

	if err != nil {
		return templateData, fmt.Errorf("writing rendered template: %w", err)
	}

	return templateData, nil
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
