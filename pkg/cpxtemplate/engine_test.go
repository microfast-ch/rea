package cpxtemplate

import (
	"encoding/xml"
	"testing"

	"github.com/djboris9/rea/pkg/xmltree"
)

func TestExec(t *testing.T) {
	testdata := xml.Header + `
<p1>
  <p2 no="1">Inside P2</p2>
  <p2 no="2">Inside P2 again</p2>
  <p2 no="3"><p3>Inside P3</p3></p2>
  <p2 no="4" be="5">Before P3 <p3>Inside P3</p3> after P3</p2>
  <!-- my comment :) -->
  <p2 no="5">[[ if (A) then ]]Hallo [# A #]</p2>
  <p2 no="6">[[ end ]]</p2>
</p1>`

	tree, err := xmltree.Parse([]byte(testdata))
	if err != nil {
		t.Fatalf("parsing tree: %v", err)
	}

	lt, err := NewLuaTree(tree)
	if err != nil {
		t.Error(err)
	}

	e := NewLuaEngine(lt)
	err = e.Exec()
	if err != nil {
		t.Errorf("executing lua engine: %s", err)
	}
}

const testDoc = `<?xml version="1.0" encoding="UTF-8"?>
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
        <table:table-column table:style-name="Table1.A" table:number-columns-repeated="3"/>
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
