package engine

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/djboris9/xmltree"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

	e := NewLuaEngine(lt, nil)
	err = e.Exec("")

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

	wantXML := testdata

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

	if diff := cmp.Diff(wantXML, serializeNodePath(t, e.nodePath)); diff != "" {
		t.Errorf("nodePath as XML mismatch (-want +got):\n%s", diff)
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

	wantXML := xml.Header + `
<p>
  <ul>
    <li>ABC</li>
    <li></li>
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
		"EndNode(14)",   // </li>
		"SetToken(15)",  // Spaces
		"StartNode(16)", // <li>
		"SetToken(17)",  // HIJ
		"EndNode(18)",   // </li>
		"SetToken(19)",  // Spaces
		"EndNode(20)",   // </ul>
		"SetToken(21)",  // Spaces
		"EndNode(22)",   // </p1>
	}
	if diff := cmp.Diff(want, e.nodePathStr); diff != "" {
		t.Errorf("nodePathStr mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}

	if diff := cmp.Diff(wantXML, serializeNodePath(t, e.nodePath)); diff != "" {
		t.Errorf("nodePath as XML mismatch (-want +got):\n%s", diff)
	}
}

func TestRenderLoopBlock(t *testing.T) {
	testdata := xml.Header + `
<p>
  <ul>
    <li>ABC</li>
    <li>[[ for i=1,3 do ]]X[# i #]Y[[ end ]]</li>
    <li>HIJ</li>
  </ul>
</p>`

	wantXML := xml.Header + `
<p>
  <ul>
    <li>ABC</li>
    <li>X1YX2YX3Y</li>
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
		"CharData(13)",  // "X"
		"Print(???)",    // "1"
		"CharData(14)",  // "Y"
		"CharData(13)",  // "X"
		"Print(???)",    // "2"
		"CharData(14)",  // "Y"
		"CharData(13)",  // "X"
		"Print(???)",    // "3"
		"CharData(14)",  // "Y"
		"EndNode(15)",   // </li>
		"SetToken(16)",  // Spaces
		"StartNode(17)", // <li>
		"SetToken(18)",  // HIJ
		"EndNode(19)",   // </li>
		"SetToken(20)",  // Spaces
		"EndNode(21)",   // </ul>
		"SetToken(22)",  // Spaces
		"EndNode(23)",   // </p1>
	}
	if diff := cmp.Diff(want, e.nodePathStr); diff != "" {
		t.Errorf("nodePathStr mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}

	if diff := cmp.Diff(wantXML, serializeNodePath(t, e.nodePath)); diff != "" {
		t.Errorf("nodePath as XML mismatch (-want +got):\n%s", diff)
	}
}

func TestRenderIfBlockSpanned(t *testing.T) {
	testdata := xml.Header + `
<article>
  <p1>ABC</p1>
  <p2>DFG[[ if False then ]]HIJ</p2>
  <p3>KLM[[ end ]]NOP</p3>
</article>`

	wantXML := xml.Header + `
<article>
  <p1>ABC</p1>
  <p2>DFG</p2><p3>NOP</p3>
</article>`

	e, err := prepareLua(t, testdata)
	if err != nil {
		t.Error(err)
	}

	want := []string{
		"SetToken(1)",              // XML Header
		"SetToken(2)",              // Spaces
		"StartNode(3)",             // <article>
		"SetToken(4)",              // Spaces
		"StartNode(5)",             // <p1>
		"SetToken(6)",              // "ABC"
		"EndNode(7)",               // </p1>
		"SetToken(8)",              // Spaces
		"StartNode(9)",             // <p2>
		"CharData(11)",             // "DFG"
		"EndNode(p2) - balanced",   // </p2>
		"StartNode(p3) - balanced", // <p3>
		"CharData(18)",             // "NOP"
		"EndNode(19)",              // </p3>
		"SetToken(20)",             // Spaces
		"EndNode(21)",              // </article>
	}

	if diff := cmp.Diff(want, e.nodePathStr); diff != "" {
		t.Errorf("nodePathStr mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}

	if diff := cmp.Diff(wantXML, serializeNodePath(t, e.nodePath)); diff != "" {
		t.Errorf("nodePath as XML mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}
}

func TestRenderLoopSpanned(t *testing.T) {
	testdata := xml.Header + `
<p>
  <ul>
    <li>ABC[[ for i=1,3 do ]]DEF</li>
    <li>X[# i #]Y</li>
    <li>GHJ[[ end ]]JKL</li>
  </ul>
</p>`

	wantXML := xml.Header + `
<p>
  <ul>
    <li>ABCDEF</li>
    <li>X1Y</li>
    <li>GHJ</li><li>DEF</li>
    <li>X2Y</li>
    <li>GHJ</li><li>DEF</li>
    <li>X3Y</li>
    <li>GHJJKL</li>
  </ul>
</p>`

	e, err := prepareLua(t, testdata)
	if err != nil {
		t.Error(err)
	}

	want := []string{
		"SetToken(1)",              // XML Header
		"SetToken(2)",              // Spaces
		"StartNode(3)",             // <p1>
		"SetToken(4)",              // Spaces
		"StartNode(5)",             // <ul>
		"SetToken(6)",              // Spaces
		"StartNode(7)",             // <li>
		"CharData(9)",              // ABC
		"CharData(10)",             // DEF
		"EndNode(11)",              // </li>
		"SetToken(12)",             // Spaces
		"StartNode(13)",            // <li>
		"CharData(15)",             // "X"
		"Print(???)",               // "1"
		"CharData(16)",             // "Y"
		"EndNode(17)",              // </li>
		"SetToken(18)",             // Spaces
		"StartNode(19)",            // <li>
		"CharData(21)",             // GHJ
		"EndNode(li) - balanced",   // </li>
		"StartNode(li) - balanced", // <li>
		"CharData(10)",             // DEF
		"EndNode(11)",              // </li>
		"SetToken(12)",             // Spaces
		"StartNode(13)",            // <li>
		"CharData(15)",             // "X"
		"Print(???)",               // "2"
		"CharData(16)",             // "Y"
		"EndNode(17)",              // </li>
		"SetToken(18)",             // Spaces
		"StartNode(19)",            // <li>
		"CharData(21)",             // GHJ
		"EndNode(li) - balanced",   // </li>
		"StartNode(li) - balanced", // <li>
		"CharData(10)",             // DEF
		"EndNode(11)",              // </li>
		"SetToken(12)",             // Spaces
		"StartNode(13)",            // <li>
		"CharData(15)",             // "X"
		"Print(???)",               // "3"
		"CharData(16)",             // "Y"
		"EndNode(17)",              // </li>
		"SetToken(18)",             // Spaces
		"StartNode(19)",            // <li>
		"CharData(21)",             // GHJ
		"CharData(22)",             // JKL
		"EndNode(23)",              // </li>
		"SetToken(24)",             // Spaces
		"EndNode(25)",              // </ul>
		"SetToken(26)",             // Spaces
		"EndNode(27)",              // </p>
	}

	if diff := cmp.Diff(want, e.nodePathStr); diff != "" {
		t.Errorf("nodePathStr mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}

	if diff := cmp.Diff(wantXML, serializeNodePath(t, e.nodePath)); diff != "" {
		t.Errorf("nodePath as XML mismatch (-want +got):\n%s", diff)
	}
}

func serializeNodePath(t *testing.T, nodePath []*xmltree.Node) string {
	// TODO: Change to e.WriteXML
	var buf strings.Builder
	enc := xml.NewEncoder(&buf)

	for i := range nodePath {
		if err := enc.EncodeToken(nodePath[i].Token); err != nil {
			t.Errorf("encoding token %d: %s", i, err)
		}
	}

	enc.Flush()

	return buf.String()
}

func TestGetCommonPaths(t *testing.T) {
	nodeA := &xmltree.Node{
		Token: xml.CharData("nodeA"),
	}
	nodeB := &xmltree.Node{
		Parent: nodeA,
		Token:  xml.CharData("nodeB"),
	}
	nodeC := &xmltree.Node{
		Parent: nodeB,
		Token:  xml.CharData("nodeC"),
	}
	nodeD := &xmltree.Node{
		Parent: nodeC,
		Token:  xml.CharData("nodeD"),
	}
	nodeE := &xmltree.Node{
		Parent: nodeD,
		Token:  xml.CharData("nodeE"),
	}
	nodeX := &xmltree.Node{
		Parent: nodeB,
		Token:  xml.CharData("nodeX"),
	}
	nodeY := &xmltree.Node{
		Parent: nodeX,
		Token:  xml.CharData("nodeY"),
	}
	nodeZ := &xmltree.Node{
		Parent: nodeY,
		Token:  xml.CharData("nodeZ"),
	}

	lT, cP, rT := getCommonPaths(nodeA, nil)

	if diff := cmp.Diff([]*xmltree.Node{}, lT); diff != "" {
		t.Errorf("getCommonPaths(nodeA, nil).lT mismatch (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff((*xmltree.Node)(nil), cP); diff != "" {
		t.Errorf("getCommonPaths(nodeA, nil).cP mismatch (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff([]*xmltree.Node{}, rT); diff != "" {
		t.Errorf("getCommonPaths(nodeA, nil).rT mismatch (-want +got):\n%s", diff)
	}

	opt := cmpopts.IgnoreFields(xmltree.Node{}, "Parent")

	stack := []*xmltree.Node{nodeA, nodeB, nodeC, nodeD, nodeE}
	lT, cP, rT = getCommonPaths(nodeZ, stack)

	if diff := cmp.Diff([]*xmltree.Node{nodeX, nodeY}, lT, opt); diff != "" {
		t.Errorf("getCommonPaths(nodeZ, stack).lT mismatch (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(nodeB, cP); diff != "" {
		t.Errorf("getCommonPaths(nodeZ, stack).cP mismatch (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff([]*xmltree.Node{nodeC, nodeD, nodeE}, rT, opt); diff != "" {
		t.Errorf("getCommonPaths(nodeZ, stack).rT mismatch (-want +got):\n%s", diff)
	}
}

func TestRenderFragmentInCodeBlock(t *testing.T) {
	testdata := xml.Header + `
<article>
  <p1>[[ if false <span>then </span>]]No Print[[ end ]]</p1>
  <p2>[[ if true <span>then </span>]]Print[[ end ]]</p2>
</article>`

	wantXML := xml.Header + `
<article>
  <p1></p1>
  <p2><span></span>Print</p2>
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
		"EndNode(12)",   // </p1>
		"SetToken(13)",  // Spaces
		"StartNode(14)", // <p2>
		"StartNode(16)", // <span>
		"EndNode(18)",   // </span>
		"CharData(20)",  // "Print"
		"EndNode(21)",   // </p2>
		"SetToken(22)",  // Spaces
		"EndNode(23)",   // </article>
	}

	if diff := cmp.Diff(want, e.nodePathStr); diff != "" {
		t.Errorf("nodePathStr mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}

	if diff := cmp.Diff(wantXML, serializeNodePath(t, e.nodePath)); diff != "" {
		t.Errorf("nodePath as XML mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}
}

func TestRenderFragmentInCodeDirective(t *testing.T) {
	testdata := xml.Header + `
<article>
  <p1>[[ if false th<span>en </span>]]No Print[[ end ]]</p1>
  <p2>[[ if true th<span>en </span>]]Print[[ end ]]</p2>
</article>`

	wantXML := xml.Header + `
<article>
  <p1></p1>
  <p2><span></span>Print</p2>
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
		"EndNode(12)",   // </p1>
		"SetToken(13)",  // Spaces
		"StartNode(14)", // <p2>
		"StartNode(16)", // <span>
		"EndNode(18)",   // </span>
		"CharData(20)",  // "Print"
		"EndNode(21)",   // </p2>
		"SetToken(22)",  // Spaces
		"EndNode(23)",   // </article>
	}

	if diff := cmp.Diff(want, e.nodePathStr); diff != "" {
		t.Errorf("nodePathStr mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}

	if diff := cmp.Diff(wantXML, serializeNodePath(t, e.nodePath)); diff != "" {
		t.Errorf("nodePath as XML mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}
}

func TestRenderUnbalancedParentStacks(t *testing.T) {
	testdata := xml.Header + `
<text>
  <p>[[ for i=1,<span>2</span> do ]]</p>
  <i>[# i #][[ end ]]</i>
</text>`

	wantXML := xml.Header + `
<text>
  <p><span></span></p>
  <i>1</i><p><span></span></p>
  <i>2</i>
</text>`

	e, err := prepareLua(t, testdata)
	if err != nil {
		t.Error(err)
	}

	want := []string{
		"SetToken(1)",             // XML Header
		"SetToken(2)",             // Spaces
		"StartNode(3)",            // <text>
		"SetToken(4)",             // Spaces
		"StartNode(5)",            // <p>
		"StartNode(7)",            // <span>
		"EndNode(9)",              // </span>
		"EndNode(11)",             // </p>
		"SetToken(12)",            // Spaces
		"StartNode(13)",           // <i>
		"Print(???)",              // "1"
		"EndNode(i) - balanced",   // </i>
		"StartNode(p) - balanced", // <p>
		"StartNode(7)",            // <span>
		"EndNode(9)",              // </span>
		"EndNode(11)",             // </p>
		"SetToken(12)",            // Spaces
		"StartNode(13)",           // <i>
		"Print(???)",              // "2"
		"EndNode(15)",             // </i>
		"SetToken(16)",            // Spaces
		"EndNode(17)",             // </text>
	}

	if diff := cmp.Diff(want, e.nodePathStr); diff != "" {
		t.Errorf("nodePathStr mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}

	if diff := cmp.Diff(wantXML, serializeNodePath(t, e.nodePath)); diff != "" {
		t.Errorf("nodePath as XML mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}
}

func TestRenderUnbalancedMultipleLevels(t *testing.T) {
	testdata := xml.Header + `
<body>
  <p>[[ for i=1,2 do ]]</p>
  <list>
    <span>[# i #][[ end ]]</span>
  </list>
</body>`

	wantXML := xml.Header + `
<body>
  <p></p>
  <list>
    <span>1</span></list><p></p>
  <list>
    <span>2</span>
  </list>
</body>`

	e, err := prepareLua(t, testdata)
	if err != nil {
		t.Error(err)
	}

	want := []string{
		"SetToken(1)",              // XML Header
		"SetToken(2)",              // Spaces
		"StartNode(3)",             // <body>
		"SetToken(4)",              // Spaces
		"StartNode(5)",             // <p>
		"EndNode(7)",               // </p>
		"SetToken(8)",              // Spaces
		"StartNode(9)",             // <list>
		"SetToken(10)",             // Spaces
		"StartNode(11)",            // <span>
		"Print(???)",               // "1"
		"EndNode(span) - balanced", // </span>
		"EndNode(list) - balanced", // </list>
		"StartNode(p) - balanced",  // <p>
		"EndNode(7)",               // </p>
		"SetToken(8)",              // Spaces
		"StartNode(9)",             // <list>
		"SetToken(10)",             // Spaces
		"StartNode(11)",            // <span>
		"Print(???)",               // "2"
		"EndNode(13)",              // </span>
		"SetToken(14)",             // Spaces
		"EndNode(15)",              // </list>
		"SetToken(16)",             // Spaces
		"EndNode(17)",              // </body>
	}

	if diff := cmp.Diff(want, e.nodePathStr); diff != "" {
		t.Errorf("nodePathStr mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}

	if diff := cmp.Diff(wantXML, serializeNodePath(t, e.nodePath)); diff != "" {
		t.Errorf("nodePath as XML mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}
}

func TestHandleExcessiveTokens(t *testing.T) {
	testdata := xml.Header + `
<body>
  <p>[[ if false then ]]</p>
  <p>Hello</p>
  <p>[[ end ]]</p>
</body>`

	wantXML := xml.Header + `
<body>
  <p></p>
</body>`

	// Possible mechanism:
	// If the nextToken in balancing matches the EndNode to be balanced, ignore it
	want := []string{
		"SetToken(1)",  // XML Header
		"SetToken(2)",  // Spaces
		"StartNode(3)", // <body>
		"SetToken(4)",  // Spaces
		"StartNode(5)", // <p>
		// "EndNode(p) - balanced", // These tokens shouldn't be rendered
		// "StartNode(p) - balanced", // These tokens shouldn't be rendered
		"EndNode(15)",  // </p>
		"SetToken(16)", // Spaces
		"EndNode(17)",  // </body>
	}

	e, err := prepareLua(t, testdata)
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(want, e.nodePathStr); diff != "" {
		t.Errorf("nodePathStr mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}

	if diff := cmp.Diff(wantXML, serializeNodePath(t, e.nodePath)); diff != "" {
		t.Errorf("nodePath as XML mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}
}

func TestIterationNodes(t *testing.T) {
	testdata := xml.Header + `
<body>[[ SetIterationNodes({"li"}) ]]
  <li>
    <p>Pre</p>
    <p>[[ for i=1,2 do ]]Loop [# i #][[ end ]]</p>
    <p>Post</p>
  </li>
</body>`

	wantXML := xml.Header + `
<body>
  <li>
    <p>Pre</p>
    <p>Loop 1</p></li><li><p>Loop 2</p>
    <p>Post</p>
  </li>
</body>`

	want := []string{
		"SetToken(1)",              // XML Header
		"SetToken(2)",              // Spaces
		"StartNode(3)",             // <body>
		"CharData(5)",              // Spaces
		"StartNode(6)",             // <li>
		"SetToken(7)",              // Spaces
		"StartNode(8)",             // <p>
		"SetToken(9)",              // "Pre"
		"EndNode(10)",              // </p>
		"SetToken(11)",             // Spaces
		"StartNode(12)",            // <p>
		"CharData(14)",             // "Loop "
		"Print(???)",               // i=1
		"EndNode(p) - balanced",    // </p>
		"EndNode(li) - balanced",   // </li>
		"StartNode(li) - balanced", // <li>
		"StartNode(p) - balanced",  // <p>
		"CharData(14)",             // "Loop "
		"Print(???)",               // i=2
		"EndNode(15)",              // </p>
		"SetToken(16)",             // Spaces
		"StartNode(17)",            // <p>
		"SetToken(18)",             // "Post"
		"EndNode(19)",              // </p>
		"SetToken(20)",             // Spaces
		"EndNode(21)",              // </li>
		"SetToken(22)",             // Spaces
		"EndNode(23)",              // </body>
	}

	e, err := prepareLua(t, testdata)
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(want, e.nodePathStr); diff != "" {
		t.Errorf("nodePathStr mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}

	if diff := cmp.Diff(wantXML, serializeNodePath(t, e.nodePath)); diff != "" {
		t.Errorf("nodePath as XML mismatch (-want +got):\n%s", diff)
		t.Log(e.lt.LuaProg)
	}
}
