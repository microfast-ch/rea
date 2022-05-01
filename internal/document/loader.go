package document

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/microfast-ch/rea/internal/odf"
	"github.com/microfast-ch/rea/internal/ooxml"
)

var errUnsupportedFile = errors.New("unsupported file")

// PackageDocument represents a templateable document.
type PackagedDocument struct {
	doc Format
}

// Format needs to be implemented by templateable documents.
type Format interface {
	MIMEType() string                  // Returns the mimetype of the document
	Open(name string) (fs.File, error) // Opens a file inside the package
	InitScript() string                // Returns the initialization script for the engine
}

// NewFromFile returns a new packaged document instance for the given file path.
// TODO: make this factory function smarter and not only check
// the file extension, but also the MIME type of the file.
func NewFromFile(path string) (*PackagedDocument, error) {
	switch ext := filepath.Ext(path); ext {
	case ".odt", ".ott":
		doc, err := odf.NewFromFile(path)
		return &PackagedDocument{doc: doc}, err
	case ".docx":
		doc, err := ooxml.NewFromFile(path)
		return &PackagedDocument{doc: doc}, err
	default:
		return nil, fmt.Errorf("%w : %s", errUnsupportedFile, ext)
	}
}
