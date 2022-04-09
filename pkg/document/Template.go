package document

import (
	"archive/zip"
	"io"
)

type TemplateProvider interface {
	MIMEType() *Template
}

type Template struct {
	zipFD       *zip.Reader
	zipFDCloser io.Closer
	mimetype    string
}

func (o *Template) MIMEType() string {
	return o.mimetype
}
