package odf

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	_, err := New(bytes.NewReader([]byte("")), 10)
	require.Error(t, err)

	testdata, err := ioutil.ReadFile("testdata/Basic1.ott")
	require.Nil(t, err)

	doc, err := New(bytes.NewReader(testdata), int64(len(testdata)))
	require.Nil(t, err)

	require.Equal(t, "application/vnd.oasis.opendocument.text-template", doc.MIMEType())

	err = doc.Close()
	require.Nil(t, err)

	// Test invalid file
	testdata, err = ioutil.ReadFile("testdata/not_odf.docx")
	require.Nil(t, err)

	doc, err = New(bytes.NewReader(testdata), int64(len(testdata)))
	require.Error(t, err)
}

func TestNewFromFile(t *testing.T) {
	doc, err := NewFromFile("testdata/Basic1.ott")
	require.Nil(t, err)

	require.Equal(t, "application/vnd.oasis.opendocument.text-template", doc.MIMEType())

	err = doc.Close()
	require.Nil(t, err)

	// Test invalid file
	doc, err = NewFromFile("testdata/not_odf.docx")
	require.Error(t, err)
}
