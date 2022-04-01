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
	err = TemplateODT(tmpl, &TemplateConfig{}, out)
	require.Nil(t, err)
	require.Equal(t, "TODO", out)
}
