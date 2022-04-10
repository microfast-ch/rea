package engine

// BlockToken expresses a 2 char wide token that can be embedded inside CharData
type BlockToken string

const (
	BlockTokenStartCode  BlockToken = "[[" // Starts a code block
	BlockTokenEndCode    BlockToken = "]]" // Ends a code block
	BlockTokenStartPrint BlockToken = "[#" // Starts a printing block
	BlockTokenEndPrint   BlockToken = "#]" // Ends a printing block
)

// isToken checks if the given token is a non char data token.
func isToken(s string) bool {
	if len(s) != 2 {
		return false
	}

	tokens := []BlockToken{BlockTokenStartCode, BlockTokenEndCode, BlockTokenStartPrint, BlockTokenEndPrint}
	for i := range tokens {
		if s == string(tokens[i]) {
			return true
		}
	}

	return false
}

// codeBlockTokenizer splits the string d into strings with the tokens inside.
// Tokens are expected to be 2 chars long. The resulting slice contains at least one element.
// TODO: Add fuzzer
func codeBlockTokenizer(d string) []string {
	ret := []string{}
	lastToken := 0

	for idx := range d {
		if idx+1 >= len(d) {
			break
		}

		if lastToken > idx {
			// Skip round if we are still inside a token
			continue
		}

		curPos := d[idx : idx+2]
		if isToken(curPos) {
			ret = append(ret, d[lastToken:idx])
			ret = append(ret, curPos)
			lastToken = idx + 2
		}
	}

	ret = append(ret, d[lastToken:])

	return ret
}
