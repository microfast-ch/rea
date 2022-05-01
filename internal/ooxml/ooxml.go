package ooxml

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"

	"github.com/djboris9/rea/internal/document"
	"github.com/djboris9/rea/internal/utils"
	"golang.org/x/exp/slices"
)

// ODF defines an OpenDocument file that is concurrently accessible.
type OOXML struct {
	template *document.Template
}

// New returns an ooxml instance for the given document with the given size.
// The file is validated to be a valid ooxml package but no content or structure is processed.
func New(doc io.ReaderAt, size int64) (*OOXML, error) {
	template, err := document.NewTemplate(doc, size)
	if err != nil {
		return nil, err
	}

	odf := &OOXML{template: template}
	err = odf.ValidateAndSetMIMEType()

	return odf, err
}

// NewFromFile returns an OOXML instance for the given document file path.
// The file is validated to be a valid OOXML package but no content or structure is processed.
func NewFromFile(path string) (*OOXML, error) {
	template, err := document.NewFromFile(path)
	if err != nil {
		return nil, err
	}

	odf := &OOXML{template: template}
	err = odf.ValidateAndSetMIMEType()

	return odf, err
}

func (o *OOXML) MIMEType() string {
	return o.template.MIMEType()
}

func (o *OOXML) InitScript() string {
	// Configures iteration nodes for table rows
	// TODO: Lists
	return "-- OOXML Init Script\nSetIterationNodes({\"tr\"})"
}

// Opens the given file as fs.File.
func (o *OOXML) Open(name string) (fs.File, error) {
	return o.template.Open(name)
}

// Close needs to be called at the end of the document processing to release any
// allocated resources.
func (o *OOXML) Close() error {
	return o.template.Close()
}

// Writes an OOXML package to the given writer. It will use the loaded OOXML contents
// as base and incorporate the overrides.
func (o *OOXML) Write(w io.Writer, ov document.Overrides) error {
	if ov == nil {
		ov = document.Overrides{}
	}

	zipWriter := zip.NewWriter(w)

	// Write files from overrides
	writtenFiles := []string{}

	writtenFiles, err := o.writeOverrides(writtenFiles, ov, zipWriter)
	if err != nil {
		return err
	}

	err = o.writeUntouched(writtenFiles, zipWriter)
	if err != nil {
		return fmt.Errorf("error writing untouched files: %w", err)
	}

	// Finish archive
	err = zipWriter.Close()
	if err != nil {
		return utils.FormatError(document.ErrArchive, fmt.Sprintf("unable to close the archive: %q", err))
	}

	return nil
}

func (o *OOXML) writeUntouched(writtenFiles []string, zipWriter *zip.Writer) error {
	// Write other files from loaded ODF that are not already defined in writtenFiles to skip
	for _, v := range o.template.GetZipFiles() {
		// Skip already written files (or just skipped/deleted files)
		if slices.Contains(writtenFiles, v.Name) {
			continue
		}

		// Write file from loaded ODF to new package
		f, err := zipWriter.Create(v.Name)
		if err != nil {
			return utils.FormatError(document.ErrArchive, fmt.Sprintf("unable to recreate file %s from template: %q", v.Name, err))
		}

		data, err := v.Open()
		if err != nil {
			return utils.FormatError(document.ErrArchive, fmt.Sprintf("unable to open file %s from template: %q", v.Name, err))
		}

		dataBytes, err := ioutil.ReadAll(data)
		data.Close()

		if err != nil {
			return utils.FormatError(document.ErrArchive, fmt.Sprintf("unable to read file %s from template: %q", v.Name, err))
		}

		_, err = f.Write(dataBytes)
		if err != nil {
			return utils.FormatError(document.ErrArchive, fmt.Sprintf("unable to write file %s to archive: %q", v.Name, err))
		}
	}

	return nil
}

func (o *OOXML) writeOverrides(writtenFiles []string, ov document.Overrides, zipWriter *zip.Writer) ([]string, error) {
	for fname, fdata := range ov {
		writtenFiles = append(writtenFiles, fname)

		// Skip writing file if we delete it from archive
		if fdata.Delete {
			continue
		}

		// Write file
		f, err := zipWriter.Create(fname)
		if err != nil {
			return nil, utils.FormatError(document.ErrOverride,
				fmt.Sprintf("unable to create file %s fom override in final archive: %q", fname, err))
		}

		_, err = f.Write(fdata.Data)
		if err != nil {
			return nil, utils.FormatError(document.ErrOverride, fmt.Sprintf("unable to write file %s fom override in final archive: %q", fname, err))
		}
	}

	return writtenFiles, nil
}

// http://officeopenxml.com/anatomyofOOXML.php
func (o *OOXML) ValidateAndSetMIMEType() error {
	// Every package must have a [Content_Types].xml, found at the root of the package.
	// This file contains a list of all of the content types of the parts in the package.
	// Every part and its type must be listed in [Content_Types].xml.
	fd, err := o.Open("[Content_Types].xml")
	if err != nil {
		return fmt.Errorf("opening [Content_Types].xml: %w", err)
	}
	defer fd.Close()

	// Main Document: Contains the body of the document.
	fd, err = o.Open("word/document.xml")
	if err != nil {
		return fmt.Errorf("Main Document: %w", err)
	}
	defer fd.Close()

	o.template.SetMIMEType(MainDocumentContentType)

	return nil
}
