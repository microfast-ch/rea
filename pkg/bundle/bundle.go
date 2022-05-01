// Package bundle implements a file writer that packages all relevant processing
// data into a tar archive for debugging or further processing.
package bundle

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/djboris9/xmltree"
)

type Writer struct {
	tw    *tar.Writer
	debug bool
}

// New creates a new writer. If debug is enabled, it will also persist objects that
// are only for debugging needed.
func New(w io.Writer, debug bool) *Writer {
	tw := tar.NewWriter(w)

	return &Writer{
		tw:    tw,
		debug: debug,
	}
}

func (b *Writer) AddTemplateMimeType(mt string) {
	err := b.writeFile("template/mimetype", mt)
	if err != nil {
		log.Fatalf("error: unable to write mimetype: %s", err)
	}
}

func (b *Writer) AddLuaProg(luaProg string) {
	err := b.writeFile("template/luaprog.lua", luaProg)
	if err != nil {
		log.Fatalf("error: unable to write luaprog: %s", err)
	}
}

func (b *Writer) AddInitScript(script string) {
	err := b.writeFile("template/init.lua", script)
	if err != nil {
		log.Fatalf("error: unable to write init script: %s", err)
	}
}

func (b *Writer) AddLuaNodeList(nodeList []*xmltree.Node) {
	buf := &strings.Builder{}
	for i := range nodeList {
		fmt.Fprintf(buf, "%d: %v\n", i, nodeList[i].Token)
	}

	err := b.writeFile("template/luaprog.nodelist", buf.String())
	if err != nil {
		log.Fatalf("error: unable to write LuaNodeList: %s", err)
	}
}

func (b *Writer) AddTemplateXMLTree(tree *xmltree.Node) {
	err := b.writeFile("template/content.xmltree", tree.Dump())
	if err != nil {
		log.Fatalf("error: unable to write LuaNodeList: %s", err)
	}
}

func (b *Writer) AddXMLResult(doc string) {
	err := b.writeFile("processed/content.xml", doc)
	if err != nil {
		log.Fatalf("error: unable to write content.xml: %s", err)
	}
}

func (b *Writer) AddLuaExecTrace(nodePath []string) {
	err := b.writeFile("processed/exec_trace.lua", strings.Join(nodePath, "\n"))
	if err != nil {
		log.Fatalf("error: unable to write exec_trace.lua: %s", err)
	}
}

func (b *Writer) writeFile(fname, content string) error {
	hdr := &tar.Header{
		Name: fname,
		Mode: 0600,
		Size: int64(len(content)),
	}

	if err := b.tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("writing header for %s: %w", fname, err)
	}

	if _, err := b.tw.Write([]byte(content)); err != nil {
		return fmt.Errorf("writing content in %s: %w", fname, err)
	}

	return nil
}

func (b *Writer) Close() error {
	err := b.tw.Close()
	if err != nil {
		return fmt.Errorf("closing tar writer: %w", err)
	}

	return nil
}
