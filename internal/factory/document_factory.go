package factory

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/djboris9/rea/internal/odf"
	"github.com/djboris9/rea/internal/ooxml"
	"github.com/djboris9/rea/pkg/document"
)

var errUnsupportedFileExtension = errors.New("unsupported file extension")

// NewFromFile returns a new packaged document instance for the given file path.
// TODO: make this factory function smarter and not only check
// the file extension, but also the MIME type of the file.
func NewFromFile(path string) (document.PackagedDocument, error) {
	switch ext := filepath.Ext(path); ext {
	case ".odt", ".ott":
		return odf.NewFromFile(path)
	case ".docx":
		return ooxml.NewFromFile(path)
	default:
		return nil, fmt.Errorf("%w : %s", errUnsupportedFileExtension, ext)
	}
}
