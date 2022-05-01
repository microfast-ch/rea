package ooxml

// Overrides defines an override identified by the file path as map key.
type Overrides map[string]Override

// Override represents a content override for a file.
type Override struct {
	Data   []byte // File contents to write
	Delete bool   // Do not write file with the given path to the package
}
