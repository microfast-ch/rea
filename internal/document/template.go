package document

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/djboris9/xmltree"
	"github.com/microfast-ch/rea/internal/engine"
	"github.com/microfast-ch/rea/internal/odf"
)

// TODO: Better error description and use it in Go style.
var ErrUnknownType = errors.New("unknownTypeErr")
var ErrMimetype = errors.New("mimetypeErr")
var ErrOverride = errors.New("overrideErr")
var ErrArchive = errors.New("archiveErr")

// Model defines the data that is passed to the engine for templating.
// Passed data must be a primitive or a map.
type Model struct {
	Data     map[string]any
	Metadata map[string]string
}

// Write runs the packaged document through the templating engine using the given model and
// writes a new packaged document on the writer.
func (p *PackagedDocument) Write(model *Model, out io.Writer) (*ProcessingData, error) {
	switch p.doc.(type) {
	case *odf.Odf:
		return p.processOdf(model, out)
	default:
		return nil, ErrUnknownType
	}
}

type ProcessingData struct {
	// Data of template
	TemplateMimeType    string
	TemplateXMLTree     *xmltree.Node
	TemplateLuaProg     string
	TemplateInitScript  string
	TemplateLuaNodeList []*xmltree.Node

	// Processed data
	LuaExecTrace []string
	XMLResult    string
}

// runLuaEngine takes a XML tree and runs the engine on it. `templateData` is updated
// with execution informations that can be used for post processing or error analysis.
func runEngine(xmlTree *xmltree.Node, templateData *ProcessingData, model *Model, initScript string) error {
	// Convert xmlTree to luaTree
	templateData.TemplateXMLTree = xmlTree

	luaTree, err := engine.NewLuaTree(xmlTree)
	if err != nil {
		return fmt.Errorf("creating lua tree from xml tree: %w", err)
	}

	// Register informations for further processing or debugging
	templateData.TemplateLuaProg = luaTree.LuaProg
	templateData.TemplateInitScript = initScript
	templateData.TemplateLuaNodeList = luaTree.NodeList

	// Prepare data for passing to the engine
	engineData := &engine.TemplateData{
		Data:     model.Data,
		Metadata: model.Metadata,
	}

	// Execute the engine
	luaEngine := engine.NewLuaEngine(luaTree, engineData)

	err = luaEngine.Exec(initScript)
	if err != nil {
		return fmt.Errorf("executing lua engine: %w", err)
	}

	// Register the nodePath information which is available after execution
	templateData.LuaExecTrace = luaEngine.GetNodePathString()

	// We serialize the resulting data and return it
	var buf strings.Builder

	err = luaEngine.WriteXML(&buf)
	if err != nil {
		return fmt.Errorf("writing executed template: %w", err)
	}

	content := buf.String()
	templateData.XMLResult = content

	return nil
}
