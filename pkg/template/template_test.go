package template

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/djboris9/rea/pkg/odf"
	"github.com/stretchr/testify/require"
)

func TestTemplateODT(t *testing.T) {
	// Load template without any template
	testdata, err := ioutil.ReadFile("../../testdata/Basic1.ott")
	require.Nil(t, err)

	tmpl, err := odf.New(bytes.NewReader(testdata), int64(len(testdata)))
	require.Nil(t, err)

	out := bytes.NewBuffer([]byte(""))
	_, err = TemplateODT(tmpl, &TemplateConfig{}, out)
	require.Nil(t, err)

	// Readout content.xml and new mimetype
	doc, err := odf.New(bytes.NewReader(out.Bytes()), int64(out.Len()))
	require.Equal(t, "application/vnd.oasis.opendocument.text", doc.MIMEType()) // `text-template` is now `text`

	contentFD, err := doc.Open("content.xml")
	require.Nil(t, err)

	content, err := ioutil.ReadAll(contentFD)
	require.Nil(t, err)

	require.Greater(t, len(content), 10000)
	contentFD.Close()
}
