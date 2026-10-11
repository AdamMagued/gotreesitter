package benchfixtures

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed w5_work_exemptions.json
var w5WorkExemptionsJSON []byte

// W5TokenExemption permits one exact failing token receipt until certificate
// v2 (universal-core PR 1.2). It never changes the work ceiling.
type W5TokenExemption struct {
	FixtureSHA256        string `json:"fixture_sha256"`
	Cell                 string `json:"cell"`
	Class                string `json:"class"`
	Value                uint64 `json:"value"`
	Ceiling              uint64 `json:"ceiling"`
	ExpiresAtCertificate int    `json:"expires_at_certificate"`
}

type w5ExemptionKey struct{ fixture, cell, class string }

func (e W5TokenExemption) key() w5ExemptionKey {
	return w5ExemptionKey{e.FixtureSHA256, e.Cell, e.Class}
}

// The six initial receipts are the membership ratchet. Remove a receipt here
// and in the JSON together when it no longer needs an exemption. Additions,
// substitutions, changed values and later expiry versions are forbidden.
var approvedW5TokenReceipts = map[w5ExemptionKey]uint64{
	{w5TypeScriptFixtureSHA256, "typescript/1MB/start", "delete-transient-error"}:   285159,
	{w5TypeScriptFixtureSHA256, "typescript/1MB/middle", "delete-transient-error"}:  285160,
	{w5TypeScriptFixtureSHA256, "typescript/1MB/end", "delete-transient-error"}:     285160,
	{w5TypeScriptFixtureSHA256, "typescript/1MB/start", "replace-transient-error"}:  285161,
	{w5TypeScriptFixtureSHA256, "typescript/1MB/middle", "replace-transient-error"}: 285162,
	{w5TypeScriptFixtureSHA256, "typescript/1MB/end", "replace-transient-error"}:    285162,
}

const w5TypeScriptFixtureSHA256 = "4733779015d7c5d5e4c3cdcc588e84d726a936a96a07ba2ab00fd4920c44e507"

func W5TokenExemptions() ([]W5TokenExemption, error) {
	var entries []W5TokenExemption
	if err := json.Unmarshal(w5WorkExemptionsJSON, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

type W5WorkExemptions struct {
	entries map[w5ExemptionKey]W5TokenExemption
	seen    map[w5ExemptionKey]bool
}

func NewW5WorkExemptions(entries []W5TokenExemption, certificateVersion int) (*W5WorkExemptions, error) {
	g := &W5WorkExemptions{make(map[w5ExemptionKey]W5TokenExemption), make(map[w5ExemptionKey]bool)}
	for _, e := range entries {
		key := e.key()
		value, approved := approvedW5TokenReceipts[key]
		if !approved || value != e.Value || e.Ceiling != 262400 || e.ExpiresAtCertificate != 2 {
			return nil, fmt.Errorf("unapproved W5 token exemption: %s/%s", e.Cell, e.Class)
		}
		if certificateVersion >= e.ExpiresAtCertificate {
			return nil, fmt.Errorf("expired W5 token exemption: %s/%s (certificate v%d)", e.Cell, e.Class, certificateVersion)
		}
		if _, duplicate := g.entries[key]; duplicate {
			return nil, fmt.Errorf("duplicate W5 token exemption: %s/%s", e.Cell, e.Class)
		}
		g.entries[key] = e
	}
	return g, nil
}

// Check also observes passing cells so an exemption cannot conceal an
// improvement or a different failing value. Remove stale entries instead.
func (g *W5WorkExemptions) Check(fixture, cell, class string, value, ceiling uint64) error {
	key := w5ExemptionKey{fixture, cell, class}
	if e, ok := g.entries[key]; ok {
		if value != e.Value || ceiling != e.Ceiling || value <= ceiling {
			return fmt.Errorf("stale W5 token exemption: %s/%s got=%d ceiling=%d receipt=%d/%d", cell, class, value, ceiling, e.Value, e.Ceiling)
		}
		g.seen[key] = true
		return nil
	}
	if value > ceiling {
		return fmt.Errorf("%s/%s: tokensConsumed=%d exceeds deterministic linear-work ceiling %d", cell, class, value, ceiling)
	}
	return nil
}

// Complete requires a failing observation for every entry in a full sweep.
func (g *W5WorkExemptions) Complete() error {
	for key := range g.entries {
		if !g.seen[key] {
			return fmt.Errorf("W5 token exemption has no matching failing cell: %s/%s", key.cell, key.class)
		}
	}
	return nil
}
