package cpxtemplate

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/djboris9/rea/pkg/xmltree"
	"github.com/google/go-cmp/cmp"
)

func TestCustomEncoder(t *testing.T) {
	// Dataset
	xmlData := `<text:p text:style-name="P2">A4,2</text:p>`
	wantXML := xmlData

	// Build tree
	tree, err := xmltree.Parse([]byte(xmlData))
	if err != nil {
		t.Error(t)
	}

	// Flatten path
	nodePath := []*xmltree.Node{}
	xmltree.Walk(tree, func(node *xmltree.Node, depth uint) error {
		nodePath = append(nodePath, node)
		return nil
	})

	// Verify dataset
	if diff := cmp.Diff(wantXML, serializeNodePathCustom(t, nodePath)); diff != "" {
		t.Errorf("nodePath as XML mismatch (-want +got):\n%s", diff)
	}
}

func serializeNodePathCustom(t *testing.T, nodePath []*xmltree.Node) string {
	var buf strings.Builder
	enc := xml.NewEncoder(&buf)
	for i := range nodePath {
		tt := nodePath[i].Token
		t.Logf("%T %v", tt, tt)
		if err := EncodeToken(enc, &buf, nodePath[i].Token); err != nil {
			t.Errorf("encoding token %d: %s", i, err)
		}
	}
	enc.Flush()
	return buf.String()
}
