# TSX generic-call correctness and runtime experiments

Evidence date: 2026-10-12, America/Los_Angeles. Baseline:
`b9705970a208f0fc83ee676dc667669d715e3776`.

## Retained correctness change

[Issue #1443](https://github.com/odvcencio/gotreesitter/issues/1443) reports that
appending a typed function changes an earlier generic call into a binary
expression. The 119-byte prefix parses correctly; the 196-byte extended source
has twelve structural differences from locked C. Both fresh parsing and the
append edit reproduce the bug with the candidate route on and off. Raising the
stack cap from 8 to 64 still fails; the observed frontier never exceeds four
versions. Acceptance tracing shows that only the binary interpretation reaches
final selection after the append.

The exact built-in TSX artifact now uses the existing C conflict-action priority
rule: the last conflict action retains the source version's priority, and every
alternative remains eligible. The artifact SHA-256 is
`419002e62f0bbaf9a7f7bdc0fbe58799bf6ca94038550ccde33695c16859d984`.
Independent probes show that the narrower conflict rule fixes the witness.
TSX's packed-reduction certification stays false. Same-name custom and stale
artifacts receive no certification.

The committed regression checks the entire C tree for the prefix, extended
source, and append edit through both routes. It checks types, fields, byte and
point spans, named/extra/missing flags, and propagated errors. All six cases
pass. The TSX R4 72-step edit session passes with zero no-edit allocations, as
do its three C fixture suites and real-file structural oracle. Existing JSX
`=` cases and the documented bare-ampersand scanner divergence remain green.

All 412 ordinary ledger rows are byte-identical, including fixture/session
hashes, routes, and diagnostics. The TSX counters are:

| Route and operation | Tokens before → after | New nodes before → after | Reused bytes before → after | Block splices before → after |
| --- | ---: | ---: | ---: | ---: |
| Candidate full | 275 → 275 | 503 → 503 | 0 → 0 | 0 → 0 |
| Candidate first edit | 60 → 60 | 214 → 214 | 1,162 → 1,162 | 123 → 123 |
| Default full | 275 → 275 | 476 → 476 | 0 → 0 | 0 → 0 |
| Default first edit | 60 → 60 | 214 → 214 | 1,162 → 1,162 | 123 → 123 |

Each row keeps one maximum live version and zero multi-version token share.
The real TSX sample remains a C-exact compact acceptance. Shared engine code,
scanner code, and grammar blobs have no changes in this slice; its correctness
gates target TSX, with the complete ordinary counter ledger rerun.

## Throughput and surrounding-context validation

A surrounding-context probe checks eight expression forms against eight
following-declaration contexts through both candidate settings: 128 locked-C
comparisons. It improves **124/128 → 128/128** C-exact results. The four repaired
comparisons include the original qualified generic call and a function-type
generic argument, each followed by the typed function. Every probe preserves
its node, token, and maximum-stack counts.

The TSX edit session also passes under the race detector, with zero no-edit
allocations. The artifact identity checks pass under race instrumentation.
Both existing adversarial and wide-GLR memory-budget tests pass. The guard
snapshot runs in the clean detached checkout because the shared directory's
ignored nested user checkouts cause unrelated guard findings. Clean L0/R6
snapshots pass unchanged: 756 root files, 12 API tag sets, 454 language-name
uses, and 101 environment reads.

Twenty shuffled alternating R4 pairs measured:

| Metric | Legacy before → after | Compact before → after |
| --- | ---: | ---: |
| Median parse time | 528.6 → 539.8 µs (+2.13%, p = 0.583) | 631.5 → 656.2 µs (+3.91%, p = 0.221) |
| Allocated bytes | 1,192 → 1,192 | 400 → 400 |
| Allocations | 7 → 7 | 7 → 7 |
| Compact acceptance / fallback | 0 / 0 → 0 / 0 | 1 / 0 → 1 / 0 |

Both time increases are inconclusive and below the 10% ratchet. This is a
correctness improvement with no timing win. TSX's median compact/legacy ratio
is 1.195 → 1.216, above the E-A ceiling of 1.15; it does not graduate.

## Discarded performance experiments

A Go prototype moved native reduction-boundary publication into cold helpers.
It reduced the hot function's stack frame from 1,064 to 744 bytes and passed
focused native/root/C gates, Go's edit invariants, and all 412 unchanged ledger
rows. Twenty shuffled alternating pairs measured compact 1.649 → 1.628 ms
(p = 0.718) and legacy control 1.352 → 1.375 ms (p = 0.698), with unchanged
bytes and allocations. This is inconclusive; the prototype is discarded.

A diagnostic-owned YAML prototype interned the external checkpoint from the
already serialized post-token snapshot. Ownership authentication and state
publication were preserved. Focused snapshot/C/memory gates, the owned
72-step session, all 412 ordinary rows, and the owned 484-node full/edit guards
passed. Twenty pairs measured 511.9 → 541.1 µs (p = 0.314), with unchanged
15,144 bytes and 141 allocations. The slower direction is inconclusive;
the prototype is discarded.

A YAML request probe found zero exact LR-state reuse opportunities among 86
owned requests. Ten requests shared the complete lex mode, base snapshot, and
a positive-width external token, but their LR states differed. Those states
can affect external-token eligibility, extra classification, and scanner retry
behavior. No broader token-sharing optimization was implemented.

All timing uses `scripts/run_randomized_benchmarks.sh`, twenty alternating
shuffle seeds, `GOMAXPROCS=1`, one process per seed, `-count=1`, `750ms`, and
`-benchmem`, on CPU 6 in a 6 GiB container. Correctness and timing run separately.
Raw scripts, overlays, logs, source hashes, and benchmark receipts are in
`/home/draco/.local/state/gotreesitter-evidence/20261013-linear-go-yaml-election`.

Compact graduation remains blocked by ordinary YAML routing and Go throughput;
this TSX correctness change supplies no Go or YAML speed claim.
