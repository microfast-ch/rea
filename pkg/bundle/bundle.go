// Package bundle implements a file writer that packages all relevant processing
// data into a tar archive for debugging or further processing.
package bundle

import "fmt"

type BundleWriter struct {
	w io.Writer
}

// New creates a new writer. If debug is enabled, it will also persist objects that
// are only for debugging needed.
func New(w io.Writer, debug bool) *BundleWriter {
	return &BundleWriter{
		w: w,
	}
}

func (b *BundleWriter) AddLuaProg(xx any) {
	// TODO
}
