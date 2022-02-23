package cpxtemplate

import (
	"encoding/xml"
	"testing"

	"github.com/djboris9/rea/pkg/xmltree"
	"github.com/google/go-cmp/cmp"
)

// prepareLua loads the testdata as XML and executes the LuaEngine.
// It returns the LuaEngine object.
func prepareLua(t *testing.T, testdata string) (*LuaEngine, error) {
	tree, err := xmltree.Parse([]byte(testdata))
	if err != nil {
		t.Fatalf("parsing tree: %v", err)
	}

	lt, err := NewLuaTree(tree)
	if err != nil {
		return nil, err
	}

	e := NewLuaEngine(lt)
	err = e.Exec()
	if err != nil {
		t.Errorf("executing lua engine: %s", err)
	}

	return e, nil
}

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

	_, err := prepareLua(t, testdata)
	if err != nil {
		t.Error(err)
	}
}

func TestRenderEqual(t *testing.T) {
	testdata := xml.Header + `
<p>
  <ul>
    <li>ABC</li>
    <li>DFG</li>
    <li>HIJ</li>
  </ul>
</p>`

	e, err := prepareLua(t, testdata)
	if err != nil {
		t.Error(err)
	}

	want := []string{
		"SetToken(1)",   // XML Header
		"SetToken(2)",   // Spaces
		"StartNode(3)",  // <p1>
		"SetToken(4)",   // Spaces
		"StartNode(5)",  // <ul>
		"SetToken(6)",   // Spaces
		"StartNode(7)",  // <li>
		"SetToken(8)",   // ABC
		"EndNode(9)",    // </li>
		"SetToken(10)",  // Spaces
		"StartNode(11)", // <li>
		"SetToken(12)",  // DFG
		"EndNode(13)",   // </li>
		"SetToken(14)",  // Spaces
		"StartNode(15)", // <li>
		"SetToken(16)",  // HIJ
		"EndNode(17)",   // </li>
		"SetToken(18)",  // Spaces
		"EndNode(19)",   // </ul>
		"SetToken(20)",  // Spaces
		"EndNode(21)",   // </p1>
	}

	if diff := cmp.Diff(want, e.nodePathStr); diff != "" {
		t.Errorf("nodePathStr mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}

}

func TestRenderIfBlock(t *testing.T) {
	testdata := xml.Header + `
<p>
  <ul>
    <li>ABC</li>
    <li>[[ if False then ]]DFG[[ end ]]</li>
    <li>HIJ</li>
  </ul>
</p>`

	e, err := prepareLua(t, testdata)
	if err != nil {
		t.Error(err)
	}

	want := []string{
		"SetToken(1)",   // XML Header
		"SetToken(2)",   // Spaces
		"StartNode(3)",  // <p1>
		"SetToken(4)",   // Spaces
		"StartNode(5)",  // <ul>
		"SetToken(6)",   // Spaces
		"StartNode(7)",  // <li>
		"SetToken(8)",   // ABC
		"EndNode(9)",    // </li>
		"SetToken(10)",  // Spaces
		"StartNode(11)", // <li>
		// "SetToken(12)" --> "DFG" is not printed
		"EndNode(13)",   // </li>
		"SetToken(14)",  // Spaces
		"StartNode(15)", // <li>
		"SetToken(16)",  // HIJ
		"EndNode(17)",   // </li>
		"SetToken(18)",  // Spaces
		"EndNode(19)",   // </ul>
		"SetToken(20)",  // Spaces
		"EndNode(21)",   // </p1>
	}
	if diff := cmp.Diff(want, e.nodePathStr); diff != "" {
		t.Errorf("nodePathStr mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}
}

func TestRenderLoopBlock(t *testing.T) {
	testdata := xml.Header + `
<p>
  <ul>
    <li>ABC</li>
    <li>[[ for i=1,3 do ]]X[# i #]]Y[[ end ]]</li>
    <li>HIJ</li>
  </ul>
</p>`
	e, err := prepareLua(t, testdata)
	if err != nil {
		t.Error(err)
	}

	want := []string{
		"SetToken(1)",   // XML Header
		"SetToken(2)",   // Spaces
		"StartNode(3)",  // <p1>
		"SetToken(4)",   // Spaces
		"StartNode(5)",  // <ul>
		"SetToken(6)",   // Spaces
		"StartNode(7)",  // <li>
		"SetToken(8)",   // ABC
		"EndNode(9)",    // </li>
		"SetToken(10)",  // Spaces
		"StartNode(11)", // <li>
		"CharData(12)",  // "X"
		"Print(???)",    // "1"
		"CharData(12)",  // "Y"
		"CharData(12)",  // "X"
		"Print(???)",    // "2"
		"CharData(12)",  // "Y"
		"CharData(12)",  // "X"
		"Print(???)",    // "3"
		"CharData(12)",  // "Y"
		"EndNode(13)",   // </li>
		"SetToken(14)",  // Spaces
		"StartNode(15)", // <li>
		"SetToken(16)",  // HIJ
		"EndNode(17)",   // </li>
		"SetToken(18)",  // Spaces
		"EndNode(19)",   // </ul>
		"SetToken(20)",  // Spaces
		"EndNode(21)",   // </p1>
	}
	if diff := cmp.Diff(want, e.nodePathStr); diff != "" {
		t.Errorf("nodePathStr mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}

}

func TestRenderIfBlockSpanned(t *testing.T) {
	testdata := xml.Header + `
<article>
  <p1>ABC</p1>
  <p2>DFG[[ if False then ]]HIJ</p2>
  <p3>KLM[[ end ]]NOP</p3>
</article>`

	e, err := prepareLua(t, testdata)
	if err != nil {
		t.Error(err)
	}

	want := []string{
		"SetToken(1)",   // XML Header
		"SetToken(2)",   // Spaces
		"StartNode(3)",  // <article>
		"SetToken(4)",   // Spaces
		"StartNode(5)",  // <p1>
		"SetToken(6)",   // ABC
		"EndNode(7)",    // </p1>
		"SetToken(8)",   // Spaces
		"StartNode(9)",  // <p2>
		"CharData(10)",  // DFG
		"EndNode(11)",   // </p2>
		"SetToken(12)",  // Spaces
		"StartNode(13)", // <p3>
		"SetToken(14)",  // NOP
		"EndNode(15)",   // </p3>
		"SetToken(16)",  // Spaces
		"EndNode(17)",   // </article>
	}
	if diff := cmp.Diff(want, e.nodePathStr); diff != "" {
		t.Errorf("nodePathStr mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
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
