package parsercorephase0

import (
	"testing"
)

type fingerprintBinding struct {
	symbols [112]uint16
	remap   []int
}

func TestScannerFingerprintTracksValueBindings(t *testing.T) {
	var value fingerprintBinding
	value.remap = make([]int, len(value.symbols))
	for i := range value.symbols {
		value.symbols[i] = uint16(i + 1)
		value.remap[i] = i
	}
	want := ScannerFingerprint(value)
	copy := value
	copy.remap = append([]int(nil), value.remap...)
	if got := ScannerFingerprint(copy); got != want {
		t.Fatal("equivalent independently allocated bindings have different identities")
	}
	value.remap[71]++
	if ScannerFingerprint(value) == want {
		t.Fatal("an in-place mapping change retained its identity")
	}
	value.remap[71]--
	value.symbols[37]++
	if ScannerFingerprint(value) == want {
		t.Fatal("a changed symbol binding retained its identity")
	}
	value.symbols[37]--
	// Language holds an existing interface value; binding it is setup work.
	var scanner any = value
	if got := testing.AllocsPerRun(100, func() { ScannerFingerprint(scanner) }); got != 0 {
		t.Fatalf("binding identity allocated %g objects", got)
	}
}

func TestScannerFingerprintPointerIdentityExcludesMutablePayload(t *testing.T) {
	first, second := &fingerprintBinding{}, &fingerprintBinding{}
	want := ScannerFingerprint(first)
	first.remap = []int{17}
	first.symbols[3] = 19
	if ScannerFingerprint(first) != want {
		t.Fatal("mutating pointer scanner payload changed its identity")
	}
	if ScannerFingerprint(second) == want {
		t.Fatal("replacing a pointer scanner retained its identity")
	}
}

func TestScannerFingerprintPreservesValueBoundaries(t *testing.T) {
	pairs := [][2]any{
		{struct{ A, B string }{"ab", "c"}, struct{ A, B string }{"a", "bc"}},
		{[]int(nil), []int{}},
		{[]int{1}, []int{1, 0}},
		{struct{ A, B bool }{true, false}, struct{ A, B bool }{false, true}},
		{complex(1.0, 2.0), complex(2.0, 1.0)},
		{1.25, 1.5},
	}
	for i, pair := range pairs {
		if ScannerFingerprint(pair[0]) == ScannerFingerprint(pair[1]) {
			t.Fatalf("different values in pair %d retained the same identity", i)
		}
	}
	if ScannerFingerprint(nil) != ([32]byte{}) {
		t.Fatal("nil scanner identity changed")
	}
}

func TestScannerFingerprintUnusualValuesRemainDeterministic(t *testing.T) {
	value := struct{ Values map[string]int }{map[string]int{"z": 1, "a": 2}}
	want := ScannerFingerprint(value)
	for range 20 {
		if ScannerFingerprint(value) != want {
			t.Fatal("map iteration changed scanner identity")
		}
	}
	value.Values["a"]++
	if ScannerFingerprint(value) == want {
		t.Fatal("changing an unusual value retained its identity")
	}
}
