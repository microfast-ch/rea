package odf

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

type ODF struct {
	zipFD       *zip.Reader
	zipFDCloser io.Closer
	mimetype    string
}

func New(doc io.ReaderAt, size int64) (*ODF, error) {
	rdr, err := zip.NewReader(doc, size)
	if err != nil {
		return nil, fmt.Errorf("creating ODF reader: %w", err)
	}

	mimetype, err := validateArchive(rdr)
	if err != nil {
		return nil, fmt.Errorf("validating ODF document: %w", err)
	}

	return &ODF{
		zipFD:    rdr,
		mimetype: mimetype,
	}, nil
}

func NewFromFile(name string) (*ODF, error) {
	rc, err := zip.OpenReader(name)
	if err != nil {
		return nil, fmt.Errorf("creating ODF reader: %w", err)
	}

	mimetype, err := validateArchive(&rc.Reader)
	if err != nil {
		return nil, fmt.Errorf("validating ODF document: %w", err)
	}

	return &ODF{
		zipFD:       &rc.Reader,
		zipFDCloser: rc,
		mimetype:    mimetype,
	}, nil
}

// https://docs.oasis-open.org/office/OpenDocument/v1.3/os/part2-packages/OpenDocument-v1.3-os-part2-packages.pdf
func validateArchive(rdr *zip.Reader) (string, error) {
	// 3.2 Validate META-INF/manifest.xml
	fd, err := rdr.Open("META-INF/manifest.xml")
	if err != nil {
		return "", fmt.Errorf("validating manifest.xml: %w", err)
	}
	fd.Close()

	// 3.3 Validate MIME type
	fd, err = rdr.Open("mimetype")
	if err != nil {
		return "", fmt.Errorf("validating mimetype: %w", err)
	}
	defer fd.Close()

	mimetypeBytes, err := ioutil.ReadAll(fd)
	if err != nil {
		return "", fmt.Errorf("reading mimetype: %w", err)
	}

	mimetype := string(mimetypeBytes)

	// https://www.iana.org/assignments/media-types/media-types.xhtml
	if !strings.HasPrefix(mimetype, "application/vnd.oasis.opendocument.") {
		return "", fmt.Errorf("mimetype %q not known to be a OpenDocument file", mimetype)
	}

	return mimetype, nil
}

func (o *ODF) MIMEType() string {
	return o.mimetype
}

func (o *ODF) Close() error {
	if o.zipFDCloser != nil {
		return o.zipFDCloser.Close()
	}

	return nil
}
