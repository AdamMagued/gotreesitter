# Compact owned lexer metadata receipt — 2026-10-13

Owned lexer snapshots now share immutable scanner-contract and checkpoint-identity metadata. Each cursor, private checkpoint copy, Language binding, and Core reset generation remains independently owned. A changed contract replaces the frozen metadata holder and never changes previously published snapshots. Destination authentication still checks the current scanner and grammar before publication and restoration. The memory budget conservatively charges both the retained scheduler holder and one holder per snapshot.

The experiment uses baseline `562944a85bdf77d30207c3f91a404fe514ac41b9`; its Go/module sources equal the source of PR #1447, including main sync `f85b39fb4780b39c6d34fc033165dfdf7192905c`. Six overlay source hashes are pinned in `candidate-source.json`. PR #1447 merged as `7aec3727aac81ea7337da5116f9db6349e4315eb`; this branch then merged that main revision without changing the measured runtime. The frozen holder comment was clarified to state that pointer/interface members must themselves refer to immutable data; its implementation is unchanged. Each comparison used the repository randomized script, alternating baseline/candidate order, one process per seed, GOMAXPROCS=1, 750ms, count=1 and benchmem, in a 6GiB single-CPU container pinned to CPU 6.

## Measured memory improvement

Owned YAML allocations fall from **14,376 to 12,936 B/op (−10.02%)**. Allocation count stays **111**, compact routes stay **1**, and fallbacks stay **0** in every sample. Twenty timing pairs give **441.3→459.0 µs, p=.174**; independent seeds 21–40 give **523.0→536.8 µs, p=.495**. The directional timing increases are recorded as inconclusive. Pooled 40 pairs give **480.6→486.1 µs, p=.264**. This is a memory improvement; it establishes no throughput win.

Twenty Go pairs give compact R4 **1.391→1.340 ms, p=.883**, legacy control **1.0491→.9959 ms, p=.904**, full DFA **6.429→6.465 ms, p=.341**, one-byte edit **145.0→148.9 µs, p=.883**, and no-edit **4.488→4.534 ns, p=.779**. All are inconclusive. Allocation counts stay unchanged and no-edit retains **0 B/op / 0 allocs/op**.

## Correctness and work

All **412 ordinary work/reuse rows are identical** before and after. Diagnostic-owned YAML full/first-edit nodes stay **484→484 / 484→484**; prior owner approval remains confined to those diagnostic rows. Focused snapshot/ownership tests, metadata replacement and stale-generation regression, YAML 72-step incremental-equals-fresh session, and locked-C document-stream/ownership/EOF witnesses pass. All **206 invariant sessions** and **618 C-fixture checks** pass, one language per process. The strict real-corpus census remains **159 PASS / 45 known FAIL**, including **120 actual compact/C-exact results**, with no status or decline change. Core and correctly tagged phase0A checks, scoped core/root races, the exact phase0 CI selection, and three trials each of tiny adversarial and wide-GLR budget checks pass.

The **1,049,572-byte / 524-document YAML stream** accepts on compact, covers the input, has no error or fallback, and matches locked C. Peak Go/C process RSS is **153,096 KiB (149.37 bytes/source byte)**. Layout/API guards retain **756 root Go files / 12 API tag sets**; R6 retains **454 language-name uses / 101 environment reads**.

The first fleet container exited with code 255 after the complete invariant/parity/census work; Docker reported no OOM. A second interruption occurred after core race checks. The remaining gates passed in a detached Docker run with exit 0, no OOM, and no timeout. Completed source-authenticated fleet work was reused, and no interrupted race check was counted as passing.

Ordinary YAML routing, allowlists, and the global default remain unchanged. E-A2's header target and E-A7's throughput exit gate are still open. Raw evidence is retained at `/home/draco/.local/state/gotreesitter-evidence/20261013-snapshot-metadata`.


The snapshot metadata change merged after all CI passed as PR #1448, main revision `86e738c0a397acdf6d9fb6eec264cab7805118b4`. The following experiments used that same runtime source. Their receipts retain the rejected measurements without adding a header runtime change.

## Discarded graph shape predicate

A separate baseline-only overlay placed the valid ordinary-link predicate in an inline wrapper and retained the existing diagnostic validator for recovery and malformed records. The expanded equivalence test covered every flag byte, every higher flag bit, and representative payload/score/order extrema, preserving accept/reject results and exact error messages. Core tests, Go/YAML 72-step invariants, C witnesses, and all 412 ordinary work/reuse rows passed.

Twenty shuffled pairs showed no measurable win: compact Go **1.434→1.531 ms, p=.989**, legacy control **1.110→1.267 ms, p=.341**, and owned YAML **495.1→488.0 µs, p=.678**. Compact bytes/allocations remained **1,696 / 25** for Go and **14,376 / 111** for YAML; all Go benchmark allocation counts stayed unchanged. No throughput improvement is established, so the runtime overlay is discarded. Evidence remains at `/home/draco/.local/state/gotreesitter-evidence/20261013-link-shape-fastpath`. The next architectural target is E-A2's at-most-32-byte dispatch header with rollback-safe indexed cold state.


## Rejected naive hot-header allocation experiment

A type-guided prototype reduced the dispatch header from 88 to **32 bytes** by moving lineage, producer-frontier, persisted-lineage cache, and optional version state into immutable cold records. Focused layout/ownership checks, Go/YAML 72-step invariants, C witnesses, and all 412 ordinary rows passed. An audit then found three multi-field cold writes requiring one publication to preserve both values; those are consolidated in the separately gated indexed arena replacement. The naive prototype is not correctness-qualified.

Twenty shuffled alternating pairs reject its allocation design: compact Go **2.059→2.630 ms (+27.75%, p<.001)**, **1,696→about 179,445 B/op**, and **25→3,199 allocations**. Owned YAML **915.7→1,088.8 µs (+18.90%, p<.001)**, **12,936→about 72,554 B/op**, and **111→1,144 allocations**. The runtime remains an external discarded overlay. The replacement uses indexed arena storage, consolidates paired mutations, and charges retained capacity; it has not yet qualified as a performance win. Raw source hashes and measurements are at `/home/draco/.local/state/gotreesitter-evidence/20261013-hot-header`.


## Rejected append-only indexed cold arena

The corrected indexed arena preserves the **32-byte header**, consolidates all three paired lineage writes, and passes growth/fork isolation, focused ownership/layout/zero-allocation canonicalization tests, all **412 ordinary counters**, Go/YAML **72-step invariants**, and C witnesses. Twenty shuffled pairs reject its retention cost: compact Go **1.999→2.707 ms (+35.43%, p<.001)**, **1,696→about 399,707 B/op**, and **25→54 allocations**; owned YAML **872.3→1,055.7 µs (+21.03%, p<.001)**, **12,936→about 126,380 B/op**, and **111→142 allocations**. Legacy Go control **1.613→1.628 ms (p=.547)** is inconclusive.

The runtime overlay is rejected. Consolidating per-update allocations into an append-only arena retained too much history, especially the per-node last-persisted cache. A separate candidate removes those three cache fields and checks the core's journaled node lineage directly, with rollback replay and blended-only regressions; its validation and performance are pending. Raw source-pinned evidence is at `/home/draco/.local/state/gotreesitter-evidence/20261013-hot-header-arena`.


## Rejected cold arena with journaled-node publication

Removing the three persisted-cache fields and testing the compact core's journaled lineage directly passes rollback replay, blended-only update, invalid-owner checks, focused root ownership/layout checks, all **412 ordinary rows**, Go/YAML **72-step invariants**, and C witnesses. It reduces cold history but still fails twenty shuffled pairs: compact Go **1.622→1.966 ms (+21.15%, p=.006)**, **1,696→about 76,562 B/op**, **25→48 allocations**; owned YAML **620.6→706.2 µs (+13.79%, p<.001)**, **12,936→about 57,194 B/op**, **111→140 allocations**. Legacy Go **1.360→1.413 ms (p=.947)** and all Go timing controls are inconclusive.

This runtime overlay is rejected. The next gated derivative suppresses unchanged scalar writes and unions carried alternative sets in a local value before publishing only changed cold state. That derivative has no established performance result. Evidence is at `/home/draco/.local/state/gotreesitter-evidence/20261013-hot-header-node-proof`.


## Rejected unchanged-cold-write suppression

Suppressing unchanged boolean writes, grouping cold-field publications, and unioning alternative sets in a local value before publishing only changed records preserves the targeted rollback/blended/owner regressions, focused root ownership/layout checks, **412 ordinary rows**, Go/YAML **72-step invariants**, and C witnesses. Twenty shuffled pairs still reject this version: compact Go **1.364→1.488 ms (+9.04%, p=.001)**, **1,696→15,344 B/op**, **25→55 allocations**; owned YAML **603.2→658.5 µs (+9.17%, p<.001)**, **12,936→about 35,817 B/op**, **111→135 allocations**. Legacy Go **1.036→1.077 ms (p=.369)** and the other Go timing controls are inconclusive.

This ends the immutable append-only header prototypes. No header runtime change is retained. The measurements identify per-update cold publication and history retention as the costs to remove; the next header architecture must bind indexed cold storage to version lifetime and transaction rollback, with explicit fork ownership and unchanged memory ratchets. Source-pinned evidence is at `/home/draco/.local/state/gotreesitter-evidence/20261013-hot-header-cold-writes`.


## Unqualified version-owned cold storage

A 32-byte header with mutable cold slots in an exclusively owned fresh summary session passes explicit fork/rollback, S5 saved-header, observer-scope, and retained-scanner accounting regressions. Full receipts and observed sessions retain immutable values. All **412 ordinary rows** and diagnostic-owned YAML **484/484 full/edit nodes** stay unchanged. Go/YAML and owned-YAML **72-step invariants**, actual compact Go/C comparison, YAML C witnesses, wide footprint and adversarial memory checks pass. Timing started only after these targeted gates; full graduation coverage is not claimed.

Twenty shuffled pairs show no throughput win: small compact Go **1.772→1.832 ms (p=.698)**, 20-KiB `query_compile` **19.16→19.00 ms (p=.678)**, 235-KiB `grammargen_lr` **207.4→205.7 ms (p=.883)**, and owned YAML **747.0→752.1 µs (p=.529)**. Small Go remains **1,696 B/op / 25 allocations**; the 20-KiB fixture keeps about **141,124 B/op / 204 allocations**. The larger Go fixture rises **3,295,712→about 5,400,222 B/op (+63.85%) / 3,091→3,443 allocations (+11.39%)**, and YAML rises **12,936→15,096 B/op (+16.70%) / 111→118 allocations**. All benchmark work counters and compact completion/fallback counts remain unchanged.

This version is not retained in runtime code. Its fixed chunk-count retention limit discarded the larger Go pool between parses. Separate candidates use a byte cap tied to the next request's budget and embed each immutable diagnostic record in one holder, eliminating retained slice histories. They require their own correctness and timing evidence. An earlier partial comparison was stopped after the dormant-scanner accounting audit and is not used as evidence. Complete raw evidence is at `/home/draco/.local/state/gotreesitter-evidence/20261013-hot-header-owned-accounting`.

## Rejected bounded version-owned cold storage

The next 32-byte-header candidate replaces the fixed chunk limit with an **8-MiB idle retention cap**, clips retained storage to **one sixteenth of the next request's memory budget**, and embeds each immutable diagnostic record in a separate holder. Explicit owned forks, rollback and S5 snapshots remain independent; full receipts and observed sessions retain immutable records. Dormant scanner snapshots remain charged to the budget. The focused ownership, retention, footprint, memory, repeated canonical lifecycle and C tests pass, including Go/YAML and diagnostic-owned YAML 72-step invariants. All **412 ordinary work/reuse rows** and owned YAML **484/484 full/edit node rows** remain identical. Full fleet qualification is withheld because the performance experiment fails.

Each comparison completed **20 shuffled alternating pairs**, at 750ms/count=1/benchmem with GOMAXPROCS=1 in the isolated 6-GiB, CPU-6 container. Small compact Go is **989.5→1,064.4 µs (p=.369)**; 20-KiB `query_compile` is **17.17→16.75 ms (p=.414)**; 235-KiB `grammargen_lr` is **171.4→184.6 ms (p=.114)**; owned YAML is **649.0→659.4 µs (p=.341)**. No timing result establishes a win. Legacy Go and the full/edit/no-edit DFA controls are also inconclusive; no-edit remains **0 B/op / 0 allocations**. Every compact route, fallback and benchmark work counter remains unchanged.

The larger Go allocation regression is removed: `grammargen_lr` retains **3,091 allocations** and **3,295,712→3,295,768 B/op**; `query_compile` retains **204 allocations** and **141,120→141,121 B/op**. Small compact Go remains **1,696 B/op / 25 allocations**. However, owned YAML still rises **12,936→15,096 B/op (+16.70%)** and **111→118 allocations (+6.31%)**. This rejects the runtime change despite reaching the header-size target. Ordinary YAML routing and graduation allowlists remain unchanged. Source hashes, gate logs and complete comparisons are at `/home/draco/.local/state/gotreesitter-evidence/20261013-hot-header-bounded-retention`; the container exits 0 without OOM or timeout.
