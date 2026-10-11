package benchfixtures

import (
	"testing"

	"github.com/odvcencio/gotreesitter/internal/incr"
)

func TestW5WorkExemptions(t *testing.T) {
	entries, err := W5TokenExemptions()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(approvedW5TokenReceipts) || len(entries) > 6 {
		t.Fatal("remove retired receipts from the membership ratchet too; additions are forbidden")
	}
	newGate := func(t *testing.T) *W5WorkExemptions {
		t.Helper()
		g, err := NewW5WorkExemptions(entries, incr.PruningCertificateVersion)
		if err != nil {
			t.Fatal(err)
		}
		return g
	}
	t.Run("exact failing receipts", func(t *testing.T) {
		g := newGate(t)
		for _, e := range entries {
			if err := g.Check(e.FixtureSHA256, e.Cell, e.Class, e.Value, e.Ceiling); err != nil {
				t.Fatal(err)
			}
		}
		if err := g.Complete(); err != nil {
			t.Fatal(err)
		}
	})
	for _, version := range []int{2, 3} {
		t.Run("expired", func(t *testing.T) {
			if _, err := NewW5WorkExemptions(entries, version); err == nil && len(entries) != 0 {
				t.Fatal("expired receipts accepted")
			}
		})
	}
	if len(entries) == 0 {
		return
	}
	e := entries[0]
	for name, change := range map[string]func(*W5TokenExemption){
		"changed digest":  func(e *W5TokenExemption) { e.FixtureSHA256 = "changed" },
		"changed cell":    func(e *W5TokenExemption) { e.Cell += "/extra" },
		"changed class":   func(e *W5TokenExemption) { e.Class = "insert-transient-error" },
		"changed receipt": func(e *W5TokenExemption) { e.Value++ },
		"raised ceiling":  func(e *W5TokenExemption) { e.Ceiling++ },
		"extended expiry": func(e *W5TokenExemption) { e.ExpiresAtCertificate++ },
	} {
		t.Run(name, func(t *testing.T) {
			changed := append([]W5TokenExemption(nil), entries...)
			change(&changed[0])
			if _, err := NewW5WorkExemptions(changed, 1); err == nil {
				t.Fatal("changed exemption accepted")
			}
		})
	}
	t.Run("unmatched addition", func(t *testing.T) {
		g := newGate(t)
		for _, other := range entries[1:] {
			if err := g.Check(other.FixtureSHA256, other.Cell, other.Class, other.Value, other.Ceiling); err != nil {
				t.Fatal(err)
			}
		}
		if err := g.Complete(); err == nil {
			t.Fatal("entry without a matching failing cell accepted")
		}
	})
	for _, value := range []uint64{e.Value - 1, e.Value + 1, e.Ceiling, e.Ceiling - 1} {
		t.Run("measured value changed", func(t *testing.T) {
			if err := newGate(t).Check(e.FixtureSHA256, e.Cell, e.Class, value, e.Ceiling); err == nil {
				t.Fatal("changed or passing receipt accepted")
			}
		})
	}
	t.Run("changed fixture cannot match", func(t *testing.T) {
		if err := newGate(t).Check("changed", e.Cell, e.Class, e.Value, e.Ceiling); err == nil {
			t.Fatal("unmatched failure accepted")
		}
	})
	t.Run("duplicate", func(t *testing.T) {
		if _, err := NewW5WorkExemptions(append(entries, e), 1); err == nil {
			t.Fatal("duplicate accepted")
		}
	})
	t.Run("removal", func(t *testing.T) {
		if _, err := NewW5WorkExemptions(entries[1:], 1); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("ordinary passing cell", func(t *testing.T) {
		if err := newGate(t).Check("other", "other", "insert", 10, 20); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("ordinary failure", func(t *testing.T) {
		if err := newGate(t).Check("other", "other", "insert", 21, 20); err == nil {
			t.Fatal("unexempted failure accepted")
		}
	})
}
