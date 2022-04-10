package odf

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
)

var ErrUpdateRootMediaType = errors.New("couldn't find and update root media-type")

// retypeManifest parses and validates a manifest.xml and sets a new mimetype.
// If the mimetype couldn't be set, this function returns an error.
func retypeManifest(b []byte, newType []byte) ([]byte, error) {
	// Decode XML
	d := xml.NewDecoder(bytes.NewReader(b))
	nodes := []xml.Token{}
	updated := false

	for {
		tokenInternal, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading xml token: %w", err)
		}

		// Internal bytes are only valid for the current scan
		tok := xml.CopyToken(tokenInternal)

		// Rewrite mimetype
		switch v := tok.(type) {
		case xml.StartElement:
			if v.Name.Local == "file-entry" && manifestAttrIsRoot(v) {
				updated = updated || manifestAttrSetMediaType(v, newType)
			}

			nodes = append(nodes, tok)
		default:
			nodes = append(nodes, tok)
		}
	}

	// Check if rewriting occurred
	if !updated {
		return nil, ErrUpdateRootMediaType
	}

	// Encode XML
	buf := bytes.NewBuffer([]byte(""))
	enc := xml.NewEncoder(buf)

	for i := range nodes {
		err := enc.EncodeToken(nodes[i])
		if err != nil {
			return nil, fmt.Errorf("encoding xml token: %w", err)
		}
	}

	err := enc.Flush()
	if err != nil {
		return nil, fmt.Errorf("flushing xml encoder: %w", err)
	}

	return buf.Bytes(), nil
}

// Sets the media-type of the given StartElement if attribute already exists
func manifestAttrSetMediaType(e xml.StartElement, t []byte) bool {
	for i := range e.Attr {
		if e.Attr[i].Name.Local == "media-type" {
			e.Attr[i].Value = string(t)
			return true
		}
	}

	return false
}

// Checks if StartElement of manifest is for the document root.
func manifestAttrIsRoot(e xml.StartElement) bool {
	for _, a := range e.Attr {
		if a.Name.Local == "full-path" && a.Value == "/" {
			return true
		}
	}

	return false
}
