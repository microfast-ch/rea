// Package bundle implements a file writer that packages all relevant processing
// data into a tar archive for debugging or further processing.
package bundle

import (
	"archive/tar"
	"fmt"
	"io"
	"strings"

	"github.com/djboris9/rea/pkg/xmltree"
)

type BundleWriter struct {
	tw    *tar.Writer
	debug bool
}

// New creates a new writer. If debug is enabled, it will also persist objects that
// are only for debugging needed.
func New(w io.Writer, debug bool) *BundleWriter {
	tw := tar.NewWriter(w)

	return &BundleWriter{
		tw:    tw,
		debug: debug,
	}
}

func (b *BundleWriter) AddTemplateMimeType(mt string) error {
	return b.writeFile("template/mimetype", mt)
}

func (b *BundleWriter) AddLuaProg(luaProg string) error {
	return b.writeFile("template/luaprog.lua", luaProg)
}

func (b *BundleWriter) AddLuaNodeList(nodeList []*xmltree.Node) error {
	buf := &strings.Builder{}
	for i := range nodeList {
		fmt.Fprintf(buf, "%d: %v\n", i, nodeList[i].Token)
	}

	return b.writeFile("template/luaprog.nodelist", buf.String())
}

func (b *BundleWriter) AddTemplateXMLTree(tree *xmltree.Node) error {
	return b.writeFile("template/content.xmltree", tree.Dump())
}

func (b *BundleWriter) AddContentXML(doc string) error {
	return b.writeFile("processed/content.xml", doc)
}

func (b *BundleWriter) writeFile(fname, content string) error {
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

func (b *BundleWriter) Close() error {
	err := b.tw.Close()
	if err != nil {
		return fmt.Errorf("closing tar writer: %w", err)
	}

	return nil
}
