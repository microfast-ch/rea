package document

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
)

type PackagedDocument interface {
	MIMEType() string
	Open(name string) (fs.File, error)
	ValidateArchive() (string, error)
	Write(w io.Writer, ov Overrides) error
}

type Template struct {
	zipFD       *zip.Reader
	zipFDCloser io.Closer
	mimetype    string
}

// NewFromFile returns a Template instance for the given document file path.
func NewFromFile(path string) (*Template, error) {
	rc, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", path, err)
	}

	// mimetype, err := validateArchive(&rc.Reader)
	// if err != nil {
	// 	return nil, fmt.Errorf("validating ODF document: %w", err)
	// }

	return &Template{
		zipFD:       &rc.Reader,
		zipFDCloser: rc,
		mimetype:    "",
	}, nil
}

// NewTemplate returns an ODF instance for the given document with the given size.
// The file is validated to be a valid ODF package but no content or structure is processed.
func NewTemplate(doc io.ReaderAt, size int64) (*Template, error) {
	rdr, err := zip.NewReader(doc, size)
	if err != nil {
		return nil, fmt.Errorf("creating ODF reader: %w", err)
	}

	// mimetype, err := validateArchive(rdr)
	// if err != nil {
	// 	return nil, fmt.Errorf("validating ODF document: %w", err)
	// }

	return &Template{
		zipFD:    rdr,
		mimetype: "",
	}, nil
}

// Close needs to be called at the end of the document processing to release any
// allocated resources.
func (o *Template) Close() error {
	if o.zipFDCloser != nil {
		return o.zipFDCloser.Close()
	}

	return nil
}

// Opens the given file as fs.File
func (o *Template) GetZipFiles() []*zip.File {
	return o.zipFD.File
}

func (o *Template) MIMEType() string {
	return o.mimetype
}

// Opens the given file as fs.File
func (o *Template) Open(name string) (fs.File, error) {
	return o.zipFD.Open(name)
}
