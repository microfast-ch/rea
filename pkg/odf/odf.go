package odf

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"strings"

	"golang.org/x/exp/slices"
)

// ODF defines an OpenDocument file that is concurrently accessible.
type ODF struct {
	zipFD       *zip.Reader
	zipFDCloser io.Closer
	mimetype    string
}

// New returns an ODF instance for the given document with the given size.
// The file is validated to be a valid ODF package but no content or structure is processed.
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

// NewFromFile returns an ODF instance for the given document file path.
// The file is validated to be a valid ODF package but no content or structure is processed.
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
		return "", fmt.Errorf("opening manifest.xml: %w", err)
	}
	defer fd.Close()

	manifestBytes, err := ioutil.ReadAll(fd)
	if err != nil {
		return "", fmt.Errorf("reading manifest.xml: %w", err)
	}

	_, err = retypeManifest(manifestBytes, []byte("dummy"))
	if err != nil {
		return "", fmt.Errorf("validating manifest.xml: %w", err)
	}

	// 3.3 Validate MIME type
	fd, err = rdr.Open("mimetype")
	if err != nil {
		return "", fmt.Errorf("opening mimetype: %w", err)
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

// MIMEType returns the mimetype of the loaded document.
func (o *ODF) MIMEType() string {
	return o.mimetype
}

// Overrides defines an override identified by the file path as map key.
type Overrides map[string]Override

// Override represents a content override for a file.
type Override struct {
	Data   []byte // File contents to write
	Delete bool   // Do not write file with the given path to the package
}

// Writes an ODF package to the given writer. It will use the loaded ODF contents
// as base and incorporate the overrides. It handles the mimetype and manifest.xml.
func (o *ODF) Write(w io.Writer, ov Overrides) error {
	if ov == nil {
		ov = Overrides{}
	}

	fd := zip.NewWriter(w)

	// Write mimetype file
	var mimetype []byte
	if mimeTypeOverride, ok := ov["mimetype"]; ok {
		if mimeTypeOverride.Delete {
			return errors.New("mimetype file cannot be deleted in overrides")
		}

		mimetype = mimeTypeOverride.Data
	} else {
		mimetype = []byte(o.MIMEType())
	}

	f, err := fd.CreateHeader(&zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store, // The first file in an ODF package needs to be the mimetype file and uncompressed
	})
	if err != nil {
		return fmt.Errorf("creating mimetype file in archive: %q", err)
	}
	_, err = f.Write(mimetype)
	if err != nil {
		return fmt.Errorf("writing mimetype file to archive: %q", err)
	}

	// Retype manifest.xml
	var manifestBytes []byte
	if v, ok := ov["META-INF/manifest.xml"]; ok && !v.Delete {
		manifestBytes = v.Data
	} else {
		fd, err := o.zipFD.Open("META-INF/manifest.xml")
		if err != nil {
			return fmt.Errorf("opening manifest.xml: %q", err)
		}
		manifestBytes, err = ioutil.ReadAll(fd)
		fd.Close()

		if err != nil {
			return fmt.Errorf("reading manifest.xml: %q", err)
		}
	}

	manifest, err := retypeManifest(manifestBytes, mimetype)
	if err != nil {
		return fmt.Errorf("retyping manifest.xml: %q", err)
	}

	ov["META-INF/manifest.xml"] = Override{
		Data: manifest,
	}

	// Write files from overrides
	writtenFiles := []string{"mimetype"} // Hold track of files that doesn't need to be written again

	for fname, fdata := range ov {
		writtenFiles = append(writtenFiles, fname)

		// Do not write mimetype a second time, if overridden
		if fname == "mimetype" {
			continue
		}

		// Skip writing file if we delete it from archive
		if fdata.Delete {
			continue
		}

		// Write file
		f, err := fd.Create(fname)
		if err != nil {
			return fmt.Errorf("creating file %q from override in archive: %q", fname, err)
		}

		_, err = f.Write(fdata.Data)
		if err != nil {
			return fmt.Errorf("writing file %q from override to archive: %q", fname, err)
		}
	}

	// Write other files from loaded ODF that are not already defined in writtenFiles to skip
	for _, v := range o.zipFD.File {
		// Skip already written files (or just skipped/deleted files)
		if slices.Contains(writtenFiles, v.Name) {
			continue
		}

		// Write file from loaded ODF to new package
		f, err := fd.Create(v.Name)
		if err != nil {
			return fmt.Errorf("creating file %q in archive: %q", v.Name, err)
		}

		data, err := v.Open()
		if err != nil {
			return fmt.Errorf("opening file %q from source archive: %q", v.Name, err)
		}

		dataBytes, err := ioutil.ReadAll(data)
		data.Close()
		if err != nil {
			return fmt.Errorf("reading file %q from source archive: %q", v.Name, err)
		}

		_, err = f.Write(dataBytes)
		if err != nil {
			return fmt.Errorf("writing file %q to archive: %q", v.Name, err)
		}
	}

	// Finish archive
	err = fd.Close()
	if err != nil {
		return fmt.Errorf("finishing archive: %q", err)
	}

	return nil
}

// Opens the given file as fs.File
func (o *ODF) Open(name string) (fs.File, error) {
	return o.zipFD.Open(name)
}

// Close needs to be called at the end of the document processing to release any
// allocated resources.
func (o *ODF) Close() error {
	if o.zipFDCloser != nil {
		return o.zipFDCloser.Close()
	}

	return nil
}
