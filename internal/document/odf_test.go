package document

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/microfast-ch/rea/internal/odf"
	"github.com/stretchr/testify/require"
)

func TestTemplateODT(t *testing.T) {
	// Load template without any template
	tmpl, err := NewFromFile("../../testdata/Basic1.ott")
	require.Nil(t, err)

	out := bytes.NewBuffer([]byte(""))
	_, err = tmpl.Write(&Model{}, out)
	require.Nil(t, err)

	// Readout content.xml and new mimetype
	doc, err := odf.New(bytes.NewReader(out.Bytes()), int64(out.Len()))
	require.Nil(t, err)
	require.Equal(t, "application/vnd.oasis.opendocument.text", doc.MIMEType()) // `text-template` is now `text`

	contentFD, err := doc.Open("content.xml")
	require.Nil(t, err)

	content, err := ioutil.ReadAll(contentFD)
	require.Nil(t, err)

	require.Greater(t, len(content), 10000)
	contentFD.Close()
}

// TODO TestTemplateOOXML
