package engine

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"io"

	"github.com/djboris9/rea/internal/utils"
)

var ErrXMLMarshalling = errors.New("xmlMarshallingErr")

// XML unmarshaling and marshaling in Go is currently (1.18) not canonical.
// We need to handle xml StartNode and EndNode separately
// - https://github.com/golang/go/issues/9519
// - https://github.com/golang/go/issues/13400#issuecomment-168334855
//
// Usual go marshaling:
//    <p xmlns="text" xmlns:text="text" text:style-name="P2">A4,3</p>
// Marshaling as required for OpenDocuments:
//    <text:p text:style-name="P2">A4,2</text:p>
//
// Therefor we need to embed the xml.Name.Space as a prefix for the xml.StartElement and xml.EndElement

// This one is dangerous, as it doesn't check that the tags are opened and closed in balance.

func EncodeToken(e *xml.Encoder, buf io.Writer, t xml.Token) error {
	switch v := t.(type) {
	case nil:
		return nil
	case xml.StartElement:
		e.Flush() // Flush data from encoder, to write to the right location
		// We cannot call here e.EncodeToken in any case, as it will complain about unbalanced tags
		w := bufio.NewWriter(buf)
		err := writeStartElement(w, v)

		if err != nil {
			return err
		}

		w.Flush()

		return nil
	case xml.EndElement:
		e.Flush() // Flush data from encoder, to write to the right location

		if v.Name.Space == "" {
			fmt.Fprintf(buf, "</%s>", v.Name.Local)
		} else {
			fmt.Fprintf(buf, "</%s:%s>", v.Name.Space, v.Name.Local)
		}

		return nil
	default:
		return e.EncodeToken(t)
	}
}

func writeAttrType(start xml.StartElement, buf *bufio.Writer) error {
	// Based on https://cs.opensource.google/go/go/+/refs/tags/go1.17.7:src/encoding/xml/marshal.go;l=711;drc=refs%2Ftags%2Fgo1.17.7
	for _, attr := range start.Attr {
		name := attr.Name
		if name.Local == "" {
			continue
		}

		err := buf.WriteByte(' ')
		if err != nil {
			return utils.FormatError(ErrXMLMarshalling, fmt.Sprintf("unable to write to buffer: %v", err))
		}

		if name.Space != "" {
			_, err = buf.WriteString(name.Space)
			if err != nil {
				return utils.FormatError(ErrXMLMarshalling, fmt.Sprintf("unable to write to buffer: %v", err))
			}

			err = buf.WriteByte(':')
			if err != nil {
				return utils.FormatError(ErrXMLMarshalling, fmt.Sprintf("unable to write to buffer: %v", err))
			}
		}

		_, err = buf.WriteString(name.Local)
		if err != nil {
			return utils.FormatError(ErrXMLMarshalling, fmt.Sprintf("unable to write to buffer: %v", err))
		}

		_, err = buf.WriteString(`="`)
		if err != nil {
			return utils.FormatError(ErrXMLMarshalling, fmt.Sprintf("unable to write to buffer: %v", err))
		}

		// p.EscapeString(attr.Value) // TODO: Is the lower line equivalent to the current line?
		err = xml.EscapeText(buf, []byte(attr.Value))
		if err != nil {
			return utils.FormatError(ErrXMLMarshalling, fmt.Sprintf("unable to write to buffer: %v", err))
		}

		err = buf.WriteByte('"')
		if err != nil {
			return utils.FormatError(ErrXMLMarshalling, fmt.Sprintf("unable to write to buffer: %v", err))
		}
	}

	return nil
}

func writeStartElement(buf *bufio.Writer, start xml.StartElement) error {
	err := buf.WriteByte('<')
	if err != nil {
		return utils.FormatError(ErrXMLMarshalling, fmt.Sprintf("unable to write to buffer: %v", err))
	}

	if start.Name.Space != "" {
		_, err = buf.WriteString(start.Name.Space)
		if err != nil {
			return utils.FormatError(ErrXMLMarshalling, fmt.Sprintf("unable to write to buffer: %v", err))
		}

		err = buf.WriteByte(':')
		if err != nil {
			return utils.FormatError(ErrXMLMarshalling, fmt.Sprintf("unable to write to buffer: %v", err))
		}
	}

	_, err = buf.WriteString(start.Name.Local)
	if err != nil {
		return utils.FormatError(ErrXMLMarshalling, fmt.Sprintf("unable to write to buffer: %v", err))
	}

	err = writeAttrType(start, buf)
	if err != nil {
		return err
	}

	err = buf.WriteByte('>')
	if err != nil {
		return utils.FormatError(ErrXMLMarshalling, fmt.Sprintf("unable to write to buffer: %v", err))
	}

	return nil
}
