package grammarruntime

import (
	"testing"

	gotreesitter "github.com/odvcencio/gotreesitter"
)

func TestSwiftDirectiveMaskAtSeparatorBoundary(t *testing.T) {
	scanner := SwiftExternalScanner{}
	state := scanner.Create()

	// When both implicit and explicit semicolons are valid, the scanner
	// operates at a statement/declaration boundary. Directives must be masked
	// internally so the separator is consumed first.
	validSymbols := make([]bool, swtTokenCount)
	validSymbols[swtTokImplicitSemi] = true
	validSymbols[swtTokExplicitSemi] = true
	validSymbols[swtTokDirectiveIf] = true
	validSymbols[swtTokDirectiveElseif] = true
	validSymbols[swtTokDirectiveElse] = true
	validSymbols[swtTokDirectiveEndif] = true

	lexer := &gotreesitter.ExternalLexer{}
	_ = scanner.Scan(state, lexer, validSymbols)

	// Caller's slice must not be modified in place.
	if !validSymbols[swtTokDirectiveIf] || !validSymbols[swtTokDirectiveElseif] ||
		!validSymbols[swtTokDirectiveElse] || !validSymbols[swtTokDirectiveEndif] {
		t.Fatal("Scan mutated the caller's validSymbols slice")
	}
}

func TestSwiftDirectiveNotMaskedWhenNotSeparator(t *testing.T) {
	scanner := SwiftExternalScanner{}
	state := scanner.Create()

	validSymbols := make([]bool, swtTokenCount)
	validSymbols[swtTokExplicitSemi] = true
	validSymbols[swtTokDirectiveIf] = true

	lexer := &gotreesitter.ExternalLexer{}
	_ = scanner.Scan(state, lexer, validSymbols)

	if !validSymbols[swtTokDirectiveIf] {
		t.Fatal("Scan mutated caller's validSymbols")
	}
}
