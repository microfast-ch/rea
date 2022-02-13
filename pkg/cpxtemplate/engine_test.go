package cpxtemplate

import "testing"

func TestValidate(t *testing.T) {
	Validate()
}

func TestXMLNodes(t *testing.T) {
	Validate()
	// TODO: Try to replace "different font" with macro
}

const testDoc = `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:officeooo="http://openoffice.org/2009/office" xmlns:css3t="http://www.w3.org/TR/css3-text/" xmlns:grddl="http://www.w3.org/2003/g/data-view#" xmlns:xhtml="http://www.w3.org/1999/xhtml" xmlns:formx="urn:openoffice:names:experimental:ooxml-odf-interop:xmlns:form:1.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:rpt="http://openoffice.org/2005/report" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:chart="urn:oasis:names:tc:opendocument:xmlns:chart:1.0" xmlns:svg="urn:oasis:names:tc:opendocument:xmlns:svg-compatible:1.0" xmlns:draw="urn:oasis:names:tc:opendocument:xmlns:drawing:1.0" xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0" xmlns:oooc="http://openoffice.org/2004/calc" xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0" xmlns:ooow="http://openoffice.org/2004/writer" xmlns:meta="urn:oasis:names:tc:opendocument:xmlns:meta:1.0" xmlns:xlink="http://www.w3.org/1999/xlink" xmlns:fo="urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0" xmlns:ooo="http://openoffice.org/2004/office" xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0" xmlns:dr3d="urn:oasis:names:tc:opendocument:xmlns:dr3d:1.0" xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0" xmlns:number="urn:oasis:names:tc:opendocument:xmlns:datastyle:1.0" xmlns:of="urn:oasis:names:tc:opendocument:xmlns:of:1.2" xmlns:calcext="urn:org:documentfoundation:names:experimental:calc:xmlns:calcext:1.0" xmlns:tableooo="http://openoffice.org/2009/table" xmlns:drawooo="http://openoffice.org/2010/draw" xmlns:loext="urn:org:documentfoundation:names:experimental:office:xmlns:loext:1.0" xmlns:dom="http://www.w3.org/2001/xml-events" xmlns:field="urn:openoffice:names:experimental:ooo-ms-interop:xmlns:field:1.0" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:math="http://www.w3.org/1998/Math/MathML" xmlns:form="urn:oasis:names:tc:opendocument:xmlns:form:1.0" xmlns:script="urn:oasis:names:tc:opendocument:xmlns:script:1.0" xmlns:xforms="http://www.w3.org/2002/xforms" office:version="1.3">
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
  </office:body>
</office:document-content>`
