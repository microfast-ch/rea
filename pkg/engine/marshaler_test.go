package engine

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
	err = xmltree.Walk(tree, func(node *xmltree.Node, depth uint) error {
		nodePath = append(nodePath, node)
		return nil
	})

	if err != nil {
		t.Error(t)
	}

	// Verify dataset
	if diff := cmp.Diff(wantXML, serializeNodePathCustom(t, nodePath)); diff != "" {
		t.Errorf("nodePath as XML mismatch (-want +got):\n%s", diff)
	}
}

// This tests for marshaling and unmarshaling differences, especially with the namespaces as in
// - https://github.com/golang/go/issues/9519
// - https://github.com/golang/go/issues/13400#issuecomment-168334855
func TestCustomEncoderLarge(t *testing.T) {
	// Build tree
	tree, err := xmltree.Parse([]byte(testDocLarge))
	if err != nil {
		t.Error(t)
	}

	// Flatten path
	nodePath := []*xmltree.Node{}
	err = xmltree.Walk(tree, func(node *xmltree.Node, depth uint) error {
		nodePath = append(nodePath, node)
		return nil
	})

	if err != nil {
		t.Error(t)
	}

	// Verify dataset
	if diff := cmp.Diff(testDocLargeWant, serializeNodePathCustom(t, nodePath)); diff != "" {
		t.Errorf("nodePath as XML mismatch (-want +got):\n%s", diff)
	}
}

func serializeNodePathCustom(t *testing.T, nodePath []*xmltree.Node) string {
	var buf strings.Builder
	enc := xml.NewEncoder(&buf)

	for i := range nodePath {
		if err := EncodeToken(enc, &buf, nodePath[i].Token); err != nil {
			t.Errorf("encoding token %d: %s", i, err)
		}
	}

	enc.Flush()

	return buf.String()
}

// testDocLarge is a real OpenDocument content.xml.
const testDocLarge = `<?xml version="1.0" encoding="UTF-8"?>
  <office:body>
    <office:text>
      <text:sequence-decls>
        <text:sequence-decl text:display-outline-level="0" text:name="Illustration"/>
        <text:sequence-decl text:display-outline-level="0" text:name="Table"/>
        <text:sequence-decl text:display-outline-level="0" text:name="Text"/>
        <text:sequence-decl text:display-outline-level="0" text:name="Drawing"/>
        <text:sequence-decl text:display-outline-level="0" text:name="Figure"/>
      </text:sequence-decls>
      <text:h text:style-name="Heading_20_1" text:outline-level="1">My Title</text:h>
      <text:p text:style-name="P1">First text</text:p>
      <text:h text:style-name="Heading_20_2" text:outline-level="2">My Subtitle</text:h>
      <text:p text:style-name="P1">Second Text</text:p>
      <text:h text:style-name="Heading_20_2" text:outline-level="2">My 2nd Subtitle</text:h>
      <text:p text:style-name="P1"/>
      <text:p text:style-name="P1">Text with <text:span text:style-name="T2">different font</text:span> here</text:p>
      <text:p text:style-name="P1"/>
      <text:p text:style-name="P1">And a table:</text:p>
      <table:table table:name="Table1" table:style-name="Table1">
        <table:table-column table:style-name="Table1.A" table:number-columns-repeated="3"></table:table-column>
        <table:table-row table:style-name="TableLine94786515912960">
          <table:table-cell table:style-name="Table1.A1" office:value-type="string">
            <text:p text:style-name="P2">a1,1</text:p>
          </table:table-cell>
          <table:table-cell table:style-name="Table1.A1" office:value-type="string">
            <text:p text:style-name="P2">A1,2</text:p>
          </table:table-cell>
          <table:table-cell table:style-name="Table1.C1" office:value-type="string">
            <text:p text:style-name="P2">A1,3</text:p>
          </table:table-cell>
        </table:table-row>
        <table:table-row table:style-name="TableLine94786515912960">
          <table:table-cell table:style-name="Table1.A2" office:value-type="string">
            <text:p text:style-name="P2">A2,1</text:p>
          </table:table-cell>
          <table:table-cell table:style-name="Table1.A2" office:value-type="string">
            <text:p text:style-name="P2">A2,2</text:p>
          </table:table-cell>
          <table:table-cell table:style-name="Table1.C2" office:value-type="string">
            <text:p text:style-name="P2">A2,3</text:p>
          </table:table-cell>
        </table:table-row>
        <table:table-row table:style-name="TableLine94786515912960">
          <table:table-cell table:style-name="Table1.A2" office:value-type="string">
            <text:p text:style-name="P2">A3,1</text:p>
          </table:table-cell>
          <table:table-cell table:style-name="Table1.A2" office:value-type="string">
            <text:p text:style-name="P3">A3,2</text:p>
          </table:table-cell>
          <table:table-cell table:style-name="Table1.C2" office:value-type="string">
            <text:p text:style-name="P2">A3,3</text:p>
          </table:table-cell>
        </table:table-row>
        <table:table-row table:style-name="TableLine94786515912960">
          <table:table-cell table:style-name="Table1.A2" office:value-type="string">
            <text:p text:style-name="P2">A4,1</text:p>
          </table:table-cell>
          <table:table-cell table:style-name="Table1.A2" office:value-type="string">
            <text:p text:style-name="P2">A4,2</text:p>
          </table:table-cell>
          <table:table-cell table:style-name="Table1.C2" office:value-type="string">
            <text:p text:style-name="P2">A4,3</text:p>
          </table:table-cell>
        </table:table-row>
      </table:table>
      <text:p text:style-name="P1"/>
    </office:text>
  </office:body>`

// testDocLargeWant is a the testDocLarge but with self closing tags expaned,
// as Go serializes them into two tokens.
const testDocLargeWant = `<?xml version="1.0" encoding="UTF-8"?>
  <office:body>
    <office:text>
      <text:sequence-decls>
        <text:sequence-decl text:display-outline-level="0" text:name="Illustration"></text:sequence-decl>
        <text:sequence-decl text:display-outline-level="0" text:name="Table"></text:sequence-decl>
        <text:sequence-decl text:display-outline-level="0" text:name="Text"></text:sequence-decl>
        <text:sequence-decl text:display-outline-level="0" text:name="Drawing"></text:sequence-decl>
        <text:sequence-decl text:display-outline-level="0" text:name="Figure"></text:sequence-decl>
      </text:sequence-decls>
      <text:h text:style-name="Heading_20_1" text:outline-level="1">My Title</text:h>
      <text:p text:style-name="P1">First text</text:p>
      <text:h text:style-name="Heading_20_2" text:outline-level="2">My Subtitle</text:h>
      <text:p text:style-name="P1">Second Text</text:p>
      <text:h text:style-name="Heading_20_2" text:outline-level="2">My 2nd Subtitle</text:h>
      <text:p text:style-name="P1"></text:p>
      <text:p text:style-name="P1">Text with <text:span text:style-name="T2">different font</text:span> here</text:p>
      <text:p text:style-name="P1"></text:p>
      <text:p text:style-name="P1">And a table:</text:p>
      <table:table table:name="Table1" table:style-name="Table1">
        <table:table-column table:style-name="Table1.A" table:number-columns-repeated="3"></table:table-column>
        <table:table-row table:style-name="TableLine94786515912960">
          <table:table-cell table:style-name="Table1.A1" office:value-type="string">
            <text:p text:style-name="P2">a1,1</text:p>
          </table:table-cell>
          <table:table-cell table:style-name="Table1.A1" office:value-type="string">
            <text:p text:style-name="P2">A1,2</text:p>
          </table:table-cell>
          <table:table-cell table:style-name="Table1.C1" office:value-type="string">
            <text:p text:style-name="P2">A1,3</text:p>
          </table:table-cell>
        </table:table-row>
        <table:table-row table:style-name="TableLine94786515912960">
          <table:table-cell table:style-name="Table1.A2" office:value-type="string">
            <text:p text:style-name="P2">A2,1</text:p>
          </table:table-cell>
          <table:table-cell table:style-name="Table1.A2" office:value-type="string">
            <text:p text:style-name="P2">A2,2</text:p>
          </table:table-cell>
          <table:table-cell table:style-name="Table1.C2" office:value-type="string">
            <text:p text:style-name="P2">A2,3</text:p>
          </table:table-cell>
        </table:table-row>
        <table:table-row table:style-name="TableLine94786515912960">
          <table:table-cell table:style-name="Table1.A2" office:value-type="string">
            <text:p text:style-name="P2">A3,1</text:p>
          </table:table-cell>
          <table:table-cell table:style-name="Table1.A2" office:value-type="string">
            <text:p text:style-name="P3">A3,2</text:p>
          </table:table-cell>
          <table:table-cell table:style-name="Table1.C2" office:value-type="string">
            <text:p text:style-name="P2">A3,3</text:p>
          </table:table-cell>
        </table:table-row>
        <table:table-row table:style-name="TableLine94786515912960">
          <table:table-cell table:style-name="Table1.A2" office:value-type="string">
            <text:p text:style-name="P2">A4,1</text:p>
          </table:table-cell>
          <table:table-cell table:style-name="Table1.A2" office:value-type="string">
            <text:p text:style-name="P2">A4,2</text:p>
          </table:table-cell>
          <table:table-cell table:style-name="Table1.C2" office:value-type="string">
            <text:p text:style-name="P2">A4,3</text:p>
          </table:table-cell>
        </table:table-row>
      </table:table>
      <text:p text:style-name="P1"></text:p>
    </office:text>
  </office:body>`
