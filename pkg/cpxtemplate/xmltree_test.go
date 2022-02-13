package cpxtemplate

import (
	"encoding/xml"
	"testing"
)

func TestParse(t *testing.T) {
	t.Skip()
	data := xml.Header + `
<p1>
  <p2 no="1">Inside P2</p2>
  <p2 no="2">Inside P2 again</p2>
  <p2 no="3"><p3>Inside P3</p3></p2>
  <p2 no="4">Before P3 <p3>Inside P3</p3> after P3</p2>
</p1>
	`

	Parse([]byte(data))
}

func TestParse2(t *testing.T) {
	data := xml.Header + `
<p1>
  <p2 no="1">Inside P2</p2>
  <p2 no="2">Inside P2 again</p2>
  <p2 no="3"><p3>Inside P3</p3></p2>
  <p2 no="4">Before P3 <p3>Inside P3</p3> after P3</p2>
</p1>
	`

	Parse2([]byte(data))
}
