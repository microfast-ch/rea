package writer

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/microfast-ch/rea/internal/document"
	"github.com/microfast-ch/rea/internal/engine"
	"github.com/microfast-ch/rea/internal/utils"
	"github.com/djboris9/xmltree"
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
func processTemplateWithLuaEngine(xmlTree *xmltree.Node, templateData *TemplateProcessingData, model *Model, initScript string) error {
	templateData.TemplateXMLTree = xmlTree
	luaTree, err := engine.NewLuaTree(xmlTree)

	if err != nil {
		return fmt.Errorf("creating lua tree from content.xml: %w", err)
	}

	templateData.TemplateLuaProg = luaTree.LuaProg
	templateData.TemplateInitScript = initScript
	templateData.TemplateLuaNodeList = luaTree.NodeList

	engineData := &engine.TemplateData{
		Data:     model.Data,
		Metadata: model.Metadata,
	}
	luaEngine := engine.NewLuaEngine(luaTree, engineData)
	err = luaEngine.Exec(initScript)

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

// Passed data must be a primitive or a map.
type Model struct {
	Data     map[string]any
	Metadata map[string]string
}

// Write takes a text or text-template file, processed it with the given
// configuration and writes the result to the writer.
func Write(tmpl document.PackagedDocument, model *Model, out io.Writer) (*TemplateProcessingData, error) {
	templateData := &TemplateProcessingData{
		TemplateMimeType: tmpl.MIMEType(),
	}

	xmlTree, err := getContentFromTemplateAsXML(tmpl)
	if err != nil {
		return templateData, err
	}

	err = processTemplateWithLuaEngine(xmlTree, templateData, model, tmpl.InitScript())
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

type TemplateProcessingData struct {
	// Data of template
	TemplateMimeType    string
	TemplateXMLTree     *xmltree.Node
	TemplateLuaProg     string
	TemplateInitScript  string
	TemplateLuaNodeList []*xmltree.Node

	// Processed data
	LuaNodePathStr []string
	ContentXML     string
}
