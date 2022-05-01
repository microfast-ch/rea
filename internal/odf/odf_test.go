package odf

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/djboris9/rea/internal/document"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	_, err := New(bytes.NewReader([]byte("")), 10)
	require.Error(t, err)

	testdata, err := ioutil.ReadFile("../../testdata/Basic1.ott")
	require.Nil(t, err)

	doc, err := New(bytes.NewReader(testdata), int64(len(testdata)))
	require.Nil(t, err)

	require.Equal(t, "application/vnd.oasis.opendocument.text-template", doc.MIMEType())

	err = doc.Close()
	require.Nil(t, err)

	// Test invalid file
	testdata, err = ioutil.ReadFile("../../testdata/Basic1.docx")
	require.Nil(t, err)

	doc, err = New(bytes.NewReader(testdata), int64(len(testdata)))
	require.Error(t, err)
}

func TestNewFromFile(t *testing.T) {
	doc, err := NewFromFile("../../testdata/Basic1.ott")
	require.Nil(t, err)

	require.Equal(t, "application/vnd.oasis.opendocument.text-template", doc.MIMEType())

	err = doc.Close()
	require.Nil(t, err)

	// Test invalid file
	doc, err = NewFromFile("../../testdata/Basic1.docx")
	require.Error(t, err)
}

func TestWrite(t *testing.T) {
	// Load valid file
	doc, err := NewFromFile("../../testdata/Basic1.ott")
	require.Nil(t, err)

	// Serialize same file, no overrides
	buf := new(bytes.Buffer)
	err = doc.Write(buf, nil)
	require.Nil(t, err)
	require.Greater(t, buf.Len(), 500)
	doc.Close()

	// Load reserialized file
	doc, err = New(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.Nil(t, err)
	require.Equal(t, "application/vnd.oasis.opendocument.text-template", doc.MIMEType())
	contentFD, err := doc.Open("content.xml")
	require.Nil(t, err)
	contentFD.Close()

	// Override mimetype file, delete content.xml and add new file
	buf = new(bytes.Buffer)
	ov := document.Overrides{
		"mimetype": document.Override{
			Data: []byte("application/vnd.oasis.opendocument.text"),
		},
		"content.xml": document.Override{
			Delete: true,
		},
		"extra.txt": document.Override{
			Data: []byte("my-extra-file"),
		},
	}
	_ = doc.Write(buf, ov)
	doc.Close()

	// Read document again and check overrides
	doc, err = New(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.Nil(t, err)
	require.Equal(t, "application/vnd.oasis.opendocument.text", doc.MIMEType())

	_, err = doc.Open("content.xml")
	require.Error(t, err)

	extra, err := doc.Open("extra.txt")
	require.Nil(t, err)
	extraData, err := ioutil.ReadAll(extra)
	require.Nil(t, err)
	require.Equal(t, []byte("my-extra-file"), extraData)
}
