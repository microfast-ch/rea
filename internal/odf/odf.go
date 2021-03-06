package odf

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"strings"

	"github.com/microfast-ch/rea/internal/utils"
	"golang.org/x/exp/slices"
)

// TODO: Improve text.
var ErrMimetype = errors.New("mimetypeErr")
var ErrOverride = errors.New("overrideErr")
var ErrArchive = errors.New("archiveErr")

type Odf struct {
	zipFD       *zip.Reader
	zipFDCloser io.Closer
	mimetype    string
}

// NewFromFile returns a new ODF instance for the given document file path.
// The file is validated to be a valid ODF package but no content or structure is processed.
func NewFromFile(path string) (*Odf, error) {
	rc, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file %s: %w", path, err)
	}

	odf := &Odf{
		zipFD:       &rc.Reader,
		zipFDCloser: rc,
		mimetype:    "",
	}

	err = odf.ValidateAndSetMIMEType()

	return odf, err
}

// NewTemplate returns an ODF instance for the given document with the given size.
// The file is validated to be a valid ODF package but no content or structure is processed.
func New(doc io.ReaderAt, size int64) (*Odf, error) {
	rdr, err := zip.NewReader(doc, size)
	if err != nil {
		return nil, fmt.Errorf("creating ODF reader: %w", err)
	}

	odf := &Odf{
		zipFD:    rdr,
		mimetype: "",
	}

	err = odf.ValidateAndSetMIMEType()

	return odf, err
}

func (o *Odf) MIMEType() string {
	return o.mimetype
}

func (o *Odf) Open(name string) (fs.File, error) {
	file, err := o.zipFD.Open(name)

	if file == nil || err != nil {
		return nil, fmt.Errorf("error opening %s: %w", name, err)
	}

	return file, nil
}

func (o *Odf) Close() error {
	if o.zipFDCloser != nil {
		return o.zipFDCloser.Close()
	}

	return nil
}

// https://docs.oasis-open.org/office/OpenDocument/v1.3/os/part2-packages/OpenDocument-v1.3-os-part2-packages.pdf
// The file is validated to be a valid ODF package and the MIME type is set accordingly.
// No content or structure is processed.
func (o *Odf) ValidateAndSetMIMEType() error {
	// 3.2 Validate META-INF/manifest.xml
	fd, err := o.Open("META-INF/manifest.xml")
	if err != nil {
		return fmt.Errorf("opening manifest.xml: %w", err)
	}
	defer fd.Close()

	manifestBytes, err := ioutil.ReadAll(fd)
	if err != nil {
		return fmt.Errorf("reading manifest.xml: %w", err)
	}

	_, err = retypeManifest(manifestBytes, []byte("dummy"))
	if err != nil {
		return fmt.Errorf("validating manifest.xml: %w", err)
	}

	// 3.3 Validate MIME type
	fd, err = o.Open("mimetype")
	if err != nil {
		return err
	}
	defer fd.Close()

	mimetypeBytes, err := ioutil.ReadAll(fd)
	if err != nil {
		return fmt.Errorf("reading mimetype: %w", err)
	}

	mimetype := string(mimetypeBytes)

	// https://www.iana.org/assignments/media-types/media-types.xhtml
	if !strings.HasPrefix(mimetype, "application/vnd.oasis.opendocument.") {
		return utils.FormatError(ErrMimetype, fmt.Sprintf("%s not an OpenDocument file", mimetype))
	}

	o.mimetype = mimetype

	return nil
}

func (o *Odf) InitScript() string {
	// Configures iteration nodes for list and table rows
	return "-- ODF Init Script\nSetIterationNodes({\"list-item\", \"table-row\"})"
}

// Writes an ODF package to the given writer. It will use the loaded ODF contents
// as base and incorporate the overrides. It handles the mimetype and manifest.xml.
func (o *Odf) Write(w io.Writer, ov Overrides) error {
	if ov == nil {
		ov = Overrides{}
	}

	zipWriter := zip.NewWriter(w)

	writtenFiles, err := o.writeMimetype(ov, zipWriter)

	if err != nil {
		return fmt.Errorf("error writing MIMEtype from override: %w", err)
	}

	writtenFiles, err = o.writeOverrides(writtenFiles, ov, zipWriter)
	if err != nil {
		return fmt.Errorf("error writing file overrides: %w", err)
	}

	err = o.writeUntouched(writtenFiles, zipWriter)
	if err != nil {
		return fmt.Errorf("error writing untouched files: %w", err)
	}

	// Finish archive
	err = zipWriter.Close()
	if err != nil {
		return utils.FormatError(err, "error finishing archive")
	}

	return nil
}

func (o *Odf) writeMimetype(ov Overrides, zipWriter *zip.Writer) ([]string, error) {
	// Write mimetype file
	var mimetype []byte

	if mimeTypeOverride, ok := ov["mimetype"]; ok {
		if mimeTypeOverride.Delete {
			return nil, utils.FormatError(ErrMimetype, "unable to delete mimetype for final archive")
		}

		mimetype = mimeTypeOverride.Data
	} else {
		mimetype = []byte(o.MIMEType())
	}

	f, err := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store, // The first file in an ODF package needs to be the mimetype file and uncompressed
	})
	if err != nil {
		return nil, utils.FormatError(ErrMimetype, "unable to create mimetype file in archive")
	}

	_, err = f.Write(mimetype)
	if err != nil {
		return nil, utils.FormatError(ErrMimetype, "unable to create mimetype file in archive")
	}

	// Retype manifest.xml
	var manifestBytes []byte
	if v, ok := ov["META-INF/manifest.xml"]; ok && !v.Delete {
		manifestBytes = v.Data
	} else {
		fd, err := o.Open("META-INF/manifest.xml")
		if err != nil {
			return nil, err
		}
		manifestBytes, err = ioutil.ReadAll(fd)
		fd.Close()

		if err != nil {
			return nil, utils.FormatError(ErrMimetype, "unable to read manifest.xml")
		}
	}

	manifest, err := retypeManifest(manifestBytes, mimetype)
	if err != nil {
		return nil, utils.FormatError(ErrMimetype, "unable to retype manifest.xml")
	}

	ov["META-INF/manifest.xml"] = Override{
		Data: manifest,
	}

	return []string{"mimetype"}, nil
}

func (o *Odf) writeOverrides(writtenFiles []string, ov Overrides, zipWriter *zip.Writer) ([]string, error) {
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
		f, err := zipWriter.Create(fname)
		if err != nil {
			return nil, utils.FormatError(ErrOverride,
				fmt.Sprintf("unable to create file %s fom override in final archive: %q", fname, err))
		}

		_, err = f.Write(fdata.Data)
		if err != nil {
			return nil, utils.FormatError(ErrOverride,
				fmt.Sprintf("unable to write file %s fom override in final archive: %q", fname, err))
		}
	}

	return writtenFiles, nil
}

// Write untouched files contained in the template package which were not processed by any override.
func (o *Odf) writeUntouched(writtenFiles []string, zipWriter *zip.Writer) error {
	for _, v := range o.zipFD.File {
		// Skip already written files (or just skipped/deleted files)
		if slices.Contains(writtenFiles, v.Name) {
			continue
		}

		// Write file from loaded ODF to new package
		f, err := zipWriter.Create(v.Name)
		if err != nil {
			return utils.FormatError(ErrArchive, fmt.Sprintf("unable to recreate file %s from template: %q", v.Name, err))
		}

		data, err := v.Open()
		if err != nil {
			return utils.FormatError(ErrArchive, fmt.Sprintf("unable to open file %s from template: %q", v.Name, err))
		}

		dataBytes, err := ioutil.ReadAll(data)
		data.Close()

		if err != nil {
			return utils.FormatError(ErrArchive, fmt.Sprintf("unable to read file %s from template: %q", v.Name, err))
		}

		_, err = f.Write(dataBytes)
		if err != nil {
			return utils.FormatError(ErrArchive, fmt.Sprintf("unable to write file %s to archive: %q", v.Name, err))
		}
	}

	return nil
}
