package document

import (
	"fmt"
	"io"

	"github.com/djboris9/xmltree"
	"github.com/microfast-ch/rea/internal/odf"
	"github.com/microfast-ch/rea/internal/ooxml"
	"github.com/microfast-ch/rea/internal/utils"
)

// processOdf processes the ODF specific entities for current PackagedDocument.
// The PackagedDocument must be of type *odf.Odf.
func (p *PackagedDocument) processOoxml(model *Model, out io.Writer) (*ProcessingData, error) {
	tmpl, ok := p.doc.(*ooxml.OOXML)
	if !ok {
		return nil, fmt.Errorf("%w: processOdf called on non ODF document of type %T", ErrUnknownType, p.doc)
	}

	templateData := &ProcessingData{
		TemplateMimeType: tmpl.MIMEType(),
	}

	xmlTree, err := getOOXMLContent(tmpl)
	if err != nil {
		return templateData, err
	}

	err = runEngine(xmlTree, templateData, model, tmpl.InitScript())
	if err != nil {
		return templateData, err
	}

	// Write file, overriding mimetype and content.xml
	// TODO: Override/Delete thumbnail and remove it from the manifest.xml
	ov := ooxml.Overrides{
		"word/document.xml": ooxml.Override{
			Data: []byte(templateData.XMLResult),
		},
	}

	err = tmpl.Write(out, ov)
	if err != nil {
		return templateData, fmt.Errorf("writing rendered template: %w", err)
	}

	return templateData, nil
}

// getODFContent returns the ODF specific content.xml as XMLTree.
func getOOXMLContent(tmpl *ooxml.OOXML) (*xmltree.Node, error) {
	// Check for text or text-template mimetype
	if tmpl.MIMEType() != ooxml.MainDocumentContentType {
		return nil, utils.FormatError(ErrMimetype, fmt.Sprintf("Unsupported mimetype: %s", tmpl.MIMEType()))
	}

	tmplContentXML, err := tmpl.Open("word/document.xml")
	if err != nil {
		return nil, fmt.Errorf("loading content.xml from template: %w", err)
	}

	tmplContent, err := io.ReadAll(tmplContentXML)
	if err != nil {
		return nil, fmt.Errorf("reading content.xml from template: %w", err)
	}

	tree, err := xmltree.Parse(tmplContent)
	if err != nil {
		return nil, fmt.Errorf("parsing content.xml as tree: %w", err)
	}

	return tree, nil
}

// getODFContent returns the ODF specific content.xml as XMLTree.
func getODFContent(tmpl *odf.Odf) (*xmltree.Node, error) {
	// Check for text or text-template mimetype
	if tmpl.MIMEType() != "application/vnd.oasis.opendocument.text-template" &&
		tmpl.MIMEType() != "application/vnd.oasis.opendocument.text" {
		return nil, utils.FormatError(ErrMimetype, fmt.Sprintf("Unsupported mimetype: %s", tmpl.MIMEType()))
	}

	tmplContentXML, err := tmpl.Open("content.xml")
	if err != nil {
		return nil, fmt.Errorf("loading content.xml from template: %w", err)
	}

	tmplContent, err := io.ReadAll(tmplContentXML)
	if err != nil {
		return nil, fmt.Errorf("reading content.xml from template: %w", err)
	}

	tree, err := xmltree.Parse(tmplContent)
	if err != nil {
		return nil, fmt.Errorf("parsing content.xml as tree: %w", err)
	}

	return tree, nil
}

// processOdf processes the ODF specific entities for current PackagedDocument.
// The PackagedDocument must be of type *odf.Odf.
func (p *PackagedDocument) processOdf(model *Model, out io.Writer) (*ProcessingData, error) {
	tmpl, ok := p.doc.(*odf.Odf)
	if !ok {
		return nil, fmt.Errorf("%w: processOdf called on non ODF document of type %T", ErrUnknownType, p.doc)
	}

	templateData := &ProcessingData{
		TemplateMimeType: tmpl.MIMEType(),
	}

	xmlTree, err := getODFContent(tmpl)
	if err != nil {
		return templateData, err
	}

	err = runEngine(xmlTree, templateData, model, tmpl.InitScript())
	if err != nil {
		return templateData, err
	}

	// Write file, overriding mimetype and content.xml
	// TODO: Override/Delete thumbnail and remove it from the manifest.xml
	ov := odf.Overrides{
		"mimetype": odf.Override{
			Data: []byte("application/vnd.oasis.opendocument.text"),
		},
		"content.xml": odf.Override{
			Data: []byte(templateData.XMLResult),
		},
	}

	err = tmpl.Write(out, ov)
	if err != nil {
		return templateData, fmt.Errorf("writing rendered template: %w", err)
	}

	return templateData, nil
}
