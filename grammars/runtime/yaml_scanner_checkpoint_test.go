//go:build !grammar_subset || grammar_subset_yaml

package grammarruntime

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"
)

func TestYAMLCheckpointRestorePreservesIndependentHistories(t *testing.T) {
	scanner := YamlExternalScanner{}
	payload := scanner.Create().(*yamlPayload)
	deep := make([]byte, 10+4*17)
	for i := 0; i < len(deep); i += 2 {
		binary.LittleEndian.PutUint16(deep[i:], uint16(300+i))
	}
	// Negative cursor coordinates and indentation lengths are valid scanner
	// state. Keep their signed representation through every restore.
	binary.LittleEndian.PutUint16(deep[4:], 0xffff)
	binary.LittleEndian.PutUint16(deep[12:], 0xfffe)
	shallow := append([]byte(nil), deep[:14]...)
	header := append([]byte(nil), deep[:10]...)
	for _, checkpoint := range [][]byte{deep, shallow, nil, header, deep, deep[:9], shallow, append(shallow, 1, 2, 3)} {
		scanner.Deserialize(payload, checkpoint)
		fresh := scanner.Create().(*yamlPayload)
		scanner.Deserialize(fresh, checkpoint)
		if !reflect.DeepEqual(payload.state, fresh.state) {
			t.Fatalf("restore after another history differs for %d-byte checkpoint: got %+v want %+v", len(checkpoint), payload.state, fresh.state)
		}
		if payload.state.indTypStk[0] != yIndRot || payload.state.indLenStk[0] != -1 {
			t.Fatal("restore lost the root indentation sentinel")
		}
	}
	checkpoint := append([]byte(nil), deep...)
	scanner.Deserialize(payload, checkpoint)
	clear(checkpoint)
	buf := make([]byte, len(deep))
	if n := scanner.Serialize(payload, buf); n != len(deep) || !bytes.Equal(buf, deep) {
		t.Fatal("restored state aliases its serialized checkpoint")
	}
}

func TestYAMLCheckpointRestoreReusesIndentStorage(t *testing.T) {
	scanner := YamlExternalScanner{}
	payload := scanner.Create()
	deep := make([]byte, 10+4*17)
	scanner.Deserialize(payload, deep)
	if got := testing.AllocsPerRun(100, func() {
		scanner.Deserialize(payload, deep[:14])
		scanner.Deserialize(payload, nil)
		scanner.Deserialize(payload, deep)
	}); got != 0 {
		t.Fatalf("warmed checkpoint restores allocate %g times, want zero", got)
	}
}
