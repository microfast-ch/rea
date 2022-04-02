package ooxml

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

const mainDocumentContentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"
const openxmlNamespace = "http://schemas.openxmlformats.org/package/2006/content-types"

var ErrUnexpectedTokenType = errors.New("unexpected token type")
var ErrContentTypeValidation = errors.New("[Content_Types].xml of ooxml in invalid")

// validateManifest parses and validates a manifest.xml and sets a new mimetype.
// If the mimetype couldn't be set, this function returns an error.
func validateManifest(b []byte) error {
	// Decode XML
	d := xml.NewDecoder(bytes.NewReader(b))
	hasTypesElement := false
	hasDocumentElement := false

	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading xml token: %w", err)
		}

		// Rewrite mimetype
		switch e := tok.(type) {
		case xml.StartElement:
			if hasTypesElement && hasDocumentElement {
				return nil
			}
			if e.Name.Local == "Types" && hasCorrextXmlns(e) {
				hasTypesElement = true
				continue
			}
			if e.Name.Local == "Override" && partNameIsDocument(e) {
				hasDocumentElement = true
				continue
			}
		}
	}

	return fmt.Errorf("no document in meta data found: %w", ErrContentTypeValidation)

}

func hasCorrextXmlns(e xml.StartElement) bool {
	for _, a := range e.Attr {
		if a.Name.Local == "xmlns" && a.Value == openxmlNamespace {
			return true
		}
	}
	return false
}

// Checks if StartElement of manifest is for the document root.
func partNameIsDocument(e xml.StartElement) bool {
	var hasDocPart = false
	var hasCorrectMimeType = false

	for _, a := range e.Attr {
		if a.Name.Local == "PartName" && a.Value == "/word/document.xml" {
			hasDocPart = true
		}
		if a.Name.Local == "ContentType" && a.Value == mainDocumentContentType {
			hasCorrectMimeType = true
		}
	}

	return hasDocPart && hasCorrectMimeType
}
