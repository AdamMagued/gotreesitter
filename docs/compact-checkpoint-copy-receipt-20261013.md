# Compact checkpoint copies: 2026-10-13 receipt

The retained candidate reduces owned YAML allocation pressure by reusing a snapshot's private scanner-state copy for byte-identical checkpoints. It saves **30 allocations per parse (141 → 111, −21.28%)** and **768 allocated bytes (15,144 → 14,376, −5.07%)** on the pinned R4 YAML sample. It has no measured throughput improvement. Ordinary YAML routing and compact graduation remain unchanged.

## Ownership and authentication

`Core.PrivateCheckpointCopyOwned` validates the active scheduler owner and the complete retained checkpoint span. It compares the supplied private bytes with the independently owned checkpoint interner bytes. Equal private bytes can represent the snapshot's DFA payload, after checkpoint, and before checkpoint. Unequal bytes receive a fresh private allocation. The method returns the interner's authenticated length and SHA-256 metadata in the same operation.

The interner's storage is never returned. Snapshot construction still clones live DFA state. Snapshot cloning still gives another snapshot independent mutable storage. Destination, scanner contract, reset generation, and checkpoint-identity checks remain. Publication and restore still compare private checkpoint copies with the authoritative interner bytes. The memory accounting remains conservative and charges all retained slice capacities, including shared private views.

The new regressions mutate private state, forge its local receipts, expire scheduler ownership, and corrupt a retained span. They require rejection or an independent exact fallback, and confirm that neither snapshot nor clone mutation changes the interner.

## Measurements

Baseline: `f76188e3d5e86287b18c24563e2f327eb4a3a3cc`, merged to main in #1434 as `70f56382512a4aedafd57cd00cc3e1e6cc750efc`. The complete Go/module source set is byte-identical across those commits. Candidate engine code is limited to the internal checkpoint-copy operation and the existing root snapshot adapter. YAML input: 1,998 bytes, SHA-256 `5aa2c70bccfe73254f13eca40476c3f88f522c6a47aea499f9d672dbc5aae8c4`.

Runs use Docker, 6 GiB memory, one CPU pinned to CPU 6, `GOMAXPROCS=1`, Go 1.25.14, `GOMEMLIMIT=5GiB`, `GOFLAGS=-p=1 -buildvcs=false`, 20 shuffled alternating baseline/candidate pairs, one process per seed, `-count=1`, `-benchtime=750ms`, and `-benchmem`. Every comparison uses `scripts/run_randomized_benchmarks.sh`.

| Experiment | Time before → after | Bytes before → after | Allocations before → after | Decision |
| --- | --- | --- | --- | --- |
| Private-copy reuse with repeated authentication | 921.1 → 922.4 µs, p=.925 | 15,144 → 14,376 | 141 → 111 | Allocation win; superseded by the consolidated operation |
| Consolidated authenticated checkpoint copy | 882.4 → 902.3 µs, p=.799 | 15,144 → 14,376 | 141 → 111 | Memory-efficiency candidate; timing inconclusive, median +2.26% |
| Exact packing of private DFA byte fields | 661.6 → 658.8 µs, p=.925 | 15,144 → 15,144 | 141 → 141 | Discarded; no measured benefit |
| Snapshot boxing with inline private bytes | 782.6 → 777.5 µs, p=.640 | 15,144 → 15,720 | 141 → 126 | Discarded; higher bytes and no time win |
| Fused ordinary graph-edge/node publication, Go | 1.855 → 1.802 ms, −2.88%, p=.040 | 1,696 → 1,696 | 25 → 25 | First-run signal; not retained |
| Same Go publication experiment, independent seeds 21–40 | 1.473 → 1.403 ms, p=.091 | 1,696 → 1,696 | 25 → 25 | Independent repeat inconclusive |

The combined Go publication comparison across all 40 pairs is also inconclusive (1.657 → 1.690 ms, p=.190). Its paired median ratio is .9672, with 29 faster and 11 slower pairs. Variation between the two runs prevents a reliable pooled throughput claim. No publication shortcut is included in the checkpoint candidate.

The two checkpoint experiments independently reproduce the same allocation reduction. Their separate timing medians are inconclusive. The retained candidate's directional timing increase is recorded; it remains below the 10% timing ratchet. This receipt does not claim that the Go or ordinary YAML throughput gaps are closed.

## Deterministic gates

All **412 ordinary work/reuse rows are byte-identical** to the baseline. Owned YAML full and first-edit node counters remain **484 → 484** each. Those values include the owner's earlier approved correctness trade; this change introduces no new ledger exception.

The focused root snapshot/scanner/ownership/relex tests, complete native core and phase-zero suites, YAML's ordinary and owned 72-step edit sessions, and YAML's ownership/EOF/document-stream locked-C tests pass. Every session reparses without edits with zero allocations. The exact private-copy regression passes.

All **206** edit sessions and **618** C fixture checks pass, one grammar per process. The strict real-file census is unchanged: **159 PASS / 45 existing FAIL**, with **120 C-exact compact acceptances**. Scoped core/root race checks and the exact phase-zero CI selection pass. Three trials each of the tiny-budget adversarial and wide-GLR memory gates pass. The **1,049,572-byte / 524-document YAML stream** accepts on compact with no fallback, full coverage, no error, and exact locked-C structure; its combined Go/C process peaks at **156,832 KiB RSS (153.01 bytes/source byte)**.

Twenty shuffled Go pairs preserve all routes and allocation counts. Full DFA parsing changes **8.465 → 8.343 ms (p=.583)**; one-byte edits **178.4 → 176.6 µs (p=.314)**; reparsing without edits **6.371 → 6.298 ns (p=.547)** with **0 bytes and 0 allocations**. Go R4 legacy changes **1.479 → 1.395 ms (p=.314)** and compact **1.800 → 1.711 ms (p=.277)**. These directions are inconclusive; there is no claimed Go speed win.

The clean layout/API guards pass: **756 root Go files**, **12 API tag sets**, **454 language-name uses**, and **101 environment reads**, all unchanged. Canonical-source ownership regressions pass after the final test placement. The measured candidate engine files match their recorded hashes.

## Evidence and reproduction

Raw evidence lives at `/home/draco/.local/state/gotreesitter-evidence/20261013-compact-go-yaml-architecture`. `candidate-source.json` pins all 3,471 baseline Go/module files and the two candidate engine-file hashes. The raw before/after files and `*-benchstat.txt` reports keep every seed, route count, allocation, and byte measurement.

For owned YAML timing, select the single language and benchmark within the same Docker limits, using a clean baseline checkout at the pinned SHA:

```sh
GTS_ADMISSION_REAL_CORPUS_LANGS=yaml \
  bash scripts/run_randomized_benchmarks.sh \
  --output /evidence/yaml-after.txt \
  --baseline-root /baseline --baseline-output /evidence/yaml-before.txt \
  --bench-regex '^BenchmarkAdmissionR4OwnedExternalLexer$' \
  --require-benchmarks BenchmarkAdmissionR4OwnedExternalLexer \
  --runs 20 --benchtime 750ms
```

The benchmark's diagnostic ownership activation is explicit and separate from ordinary routing. Run invariant and C suites one grammar per process. Keep correctness checks separate from timing.
