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
