package parsercorephase0

import (
	"bytes"
	"testing"
)

func TestPrivateCheckpointCopyOwnedAuthenticatesAndKeepsStoragePrivate(t *testing.T) {
	compact, err := New(&fakeTable{}, Limits{MaxCheckpoints: 4, MaxCheckpointBytes: 16})
	if err != nil {
		t.Fatal(err)
	}
	want := []byte{4, 5, 6}
	id := mustInternCheckpoint(t, compact, want)
	var owner SchedulerTransactionToken
	for _, check := range []CheckpointID{0, id} {
		if _, ok := compact.PrivateCheckpointCopyOwned(owner, check, want); ok {
			t.Fatal("unowned checkpoint accepted")
		}
	}
	if err := compact.ApplySchedulerAtomic(func(token SchedulerTransactionToken) error {
		owner = token
		candidate := append([]byte(nil), want...)
		copy, ok := compact.PrivateCheckpointCopyOwned(owner, id, candidate)
		if !ok || copy.Length != uint32(len(want)) || !bytes.Equal(copy.Bytes, want) {
			t.Fatal("matching private checkpoint unavailable")
		}
		copy.Bytes[0] ^= 1
		if !compact.CheckpointMatches(id, want) {
			t.Fatal("private checkpoint exposed retained interner bytes")
		}
		fallback, ok := compact.PrivateCheckpointCopyOwned(owner, id, candidate)
		if !ok || !bytes.Equal(fallback.Bytes, want) {
			t.Fatal("mismatching private candidate prevented exact fallback")
		}
		candidate[1] ^= 1
		if !bytes.Equal(fallback.Bytes, want) {
			t.Fatal("fallback aliases mismatching candidate")
		}
		if _, ok := compact.PrivateCheckpointCopyOwned(owner, CheckpointID(99), want); ok {
			t.Fatal("unknown checkpoint accepted")
		}
		// A malformed retained span must fail even if the caller supplied bytes
		// with the record's declared length.
		saved := compact.checkpoints.records[id-1]
		compact.checkpoints.records[id-1].offset = ^uint32(0)
		_, valid := compact.PrivateCheckpointCopyOwned(owner, id, want)
		compact.checkpoints.records[id-1] = saved
		if valid {
			t.Fatal("out-of-bounds checkpoint span accepted")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, ok := compact.PrivateCheckpointCopyOwned(owner, id, want); ok {
		t.Fatal("expired scheduler owner accepted")
	}
}
