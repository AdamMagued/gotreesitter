# Compact proof and checkpoint validation: 2026-10-13 receipt

This change preserves a newly asserted blended-lineage marker when alternative-set membership stays unchanged, and validates a private scanner checkpoint against one retained record. The two mechanisms have separate regressions and measurements.

## Blended-lineage correctness

`recordNodeLineageSet` previously returned when set union added no members, before persisting a newly asserted blended marker. That could leave historical lineage claiming to be a non-blended containment witness. The fix treats a changed blended marker as a real mutation even when membership stays identical. It journals the prior state, preserves the marker after commit, restores it after rollback, and invalidates a stale reused-lineage proof.

The baseline regression fails with `unchanged membership discarded the new blended marker`. The fixed regression checks both rollback and commit, unchanged membership, and permanent marker retention after later unblended input. No coverage proof is weakened.

## Exact checkpoint validation

The old validation path obtained length/digest receipt metadata and then performed a separate retained-record lookup to compare bytes. `Core.MatchCheckpointPrivateCopy` checks length, digest, complete retained span, and exact byte equality with one lookup. The root adapter passes its saved receipt by reference. The method neither retains the digest pointer nor exposes interner storage.

Snapshot destination, reset-generation, scanner-contract, checkpoint-identity, and private-byte ownership checks remain. Empty checkpoints retain the existing semantics, and unavailable metadata and changed bytes retain their separate error classes. Native regressions cover changed bytes, wrong lengths, forged/missing digests, unknown IDs, empty states, malformed retained spans, and nil cores.

## Isolated measurement

Baseline: `f95c04fab71ccaadf542576a2c59e807a8cf8735`, merged in #1446 as main commit `210d191b489e8e12b554f76c22b537d0d5b07389`. The Go/module source sets are byte-identical across those commits. YAML R4 sample: 1,998 bytes, SHA-256 `5aa2c70bccfe73254f13eca40476c3f88f522c6a47aea499f9d672dbc5aae8c4`.

Every comparison uses `scripts/run_randomized_benchmarks.sh`, Docker with 6 GiB memory, one CPU pinned to CPU 6, `GOMAXPROCS=1`, `GOMEMLIMIT=5GiB`, Go 1.25.14, one process per explicit shuffle seed, `-count=1`, `-benchtime=750ms`, and `-benchmem`.

| Validator-only comparison | Time before → after | Bytes before → after | Allocations before → after |
| --- | --- | --- | --- |
| Seeds 1–20 | 765.3 → 678.1 µs, p=.134; inconclusive | 14,376 → 14,376 | 111 → 111 |
| Independent seeds 21–40 | 788.0 → 738.6 µs, **−6.28%, p=.021** | 14,376 → 14,376 | 111 → 111 |

Every measured owned parse stays on compact: one route and zero fallbacks. These timings describe the validator alone. The combined-source comparison is reported separately before retention. Ordinary YAML routing and compact graduation remain unchanged.

## Combined correctness gates

All **412 ordinary ledger rows** and diagnostic owned YAML full/first-edit nodes **484 → 484** are unchanged. No new ledger exception is required.

The complete native core and phase-zero suites pass. All **206 edit sessions** and **618 C fixture checks** pass, one grammar per process. Every session has zero allocations when reparsed without edits. The strict real-file census stays **159 PASS / 45 existing FAIL**, including **120 C-exact compact acceptances**.

Scoped native/root race tests, the exact phase-zero CI selection, and three trials each of the tiny-memory adversarial and wide-GLR gates pass. The **1,049,572-byte / 524-document YAML stream** accepts on compact without fallback, covers the full input, has no error, and matches locked C. The combined Go/C process peaks at **157,116 KiB RSS (153.29 bytes/source byte)**.

## Remaining measurements and evidence

Twenty combined-source owned YAML pairs show **670.5 → 617.0 µs, −7.98%, p=.001**, with **14,376 B/op**, **111 allocs/op**, one compact route and zero fallbacks unchanged.

Twenty Go pairs show full DFA **8.560 → 7.699 ms (p=.149)**, R4 compact **1.754 → 1.693 ms (p=.698)**, and R4 legacy **1.364 → 1.236 ms (p=.174)**, all inconclusive. One-byte edits change **180.9 → 167.7 µs, −7.32%, p=.043**. Because the legacy timing controls also move downward and the changes do not introduce a Go incremental algorithm, the edit result is recorded without attributing a Go speed win. No-edit time **5.745 → 5.637 ns (p=.947)** is inconclusive and keeps **0 bytes / 0 allocations**. Allocation counts stay unchanged for all five Go benchmarks.

The clean layout/API guards pass: **756 root Go files**, **12 API tag sets**, **454 language-name uses**, and **101 environment reads**, all unchanged. The three engine-file hashes match the measured overlays. Canonical-source regressions are checked before publication. This receipt does not claim that the Go gap is closed or that compact can graduate.

Evidence is pinned under `/home/draco/.local/state/gotreesitter-evidence/20261013-fused-checkpoint-validation`. `candidate-source.json` records all 3,473 baseline Go/module file hashes and the three candidate engine-file hashes. Raw before/after benchmark files, benchstat reports, strict census results, and Docker metadata preserve the measurements and limits.

The Go copy-attribution audit finds the largest `runtime.duffcopy` callers in materialization views and election, followed by canonicalization and DFA snapshots. Dispatch headers remain 88 bytes; the E-A2 target of at most 32 bytes is still open. Existing point caches, reduction-plan caches, small alternative sets, and owned-snapshot transfer are already implemented.

## Discarded Go election experiment

A separate overlay borrowed headers during election instead of copying them for state lookup. Full diagnostic receipts kept their existing copy. Focused root tests, Go/YAML edit sessions and C witnesses, and all 412 ordinary counters passed. Twenty shuffled alternating pairs gave compact Go **1.365 → 1.352 ms (p=.355)** and legacy control **1.031 → 1.028 ms (p=1.000)**. Bytes and allocations were unchanged. The runtime change is discarded; it establishes no throughput win and does not complete E-A2. Raw evidence lives at `/home/draco/.local/state/gotreesitter-evidence/20261013-go-election-borrow`.
