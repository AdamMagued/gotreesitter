# Swift directive contribution review

Reviewed [PR #1442](https://github.com/odvcencio/gotreesitter/pull/1442),
AdamMagued's commit `3d2f8a9a671c0ae698d04d66d00f981b5e3928fd`, against
the compact branch at `25b72122b1f1a32f38ddbfbf0738695201a92307` and main at
`4f6c2429bc940ccad1c8e94ae31a514757326897` on 2026-10-11.

The five issue #1399 witnesses already parse cleanly on both current routes.
Commit `e54d35f46` (#1416) removed the scanner's directive-over-semicolon
preference on October 2. The proposed directive mask passes these witnesses,
but adds no demonstrated fix to today's scanner. Incorporate the contributor's
regression tests, with attribution, and retain the current scanner.

The additional locked-C test covers 15 cases with seven consecutive source
versions on both routes: 210 fresh and 210 incremental comparisons. It checks
symbols, fields, spans, points, flags and deep digests. Cases include comments,
explicit separators, nested directives, type and function bodies, macros,
raw strings, CRLF and EOF trivia. Both the current scanner and the proposed
mask pass all 420 comparisons. The focused race test and Swift's 72-step R4
invariant session pass, including zero allocations for a reparse without edits.

The wider review found three stale Go digest assertions after `c99e9085c`'s
scanner-padding correction. The locked C lexer counts skipped bytes as progress
even when the visible external token is empty. Restricting that rule to EOF or
scanner-state changes breaks valid Swift comment/directive input. The original
rule remains; no runtime code changes to that rule are part of this review.

For `let x = 1 /* comment */\n#warning("w")\n`, main produces an error and
the compact branch matches locked C exactly, with digest
`a439c90d8076a92f8cbf1ceb2dc8e8cfa138c432cf51e16d33540576a405b64d`.
Its node work falls from 81 on main to 35 on the existing compact branch.

Refresh only the Go expectations for the already tracked malformed witnesses:

| Witness | Previous Go digest | Current Go digest | Locked C digest |
| --- | --- | --- | --- |
| 104,681-byte unsafe witness | `a40f6fe0…` | `414e25c9…` | `dd81933b…`, unchanged |
| 16,384-byte prefix | `a6357cc0…` | `afad13be…` | `b91fecb1…`, unchanged |

Both sides still report errors and Go covers the input with an accepted stop.
The full witness still has one Go root child against C's 353; now assert that
exact known divergence as well as both full digests. The prefix still has a Go
`source_file` against C's `ERROR`. Existing node caps and recovery checks remain.
These witnesses remain correctness gaps; refreshing the receipts does not
certify their C parity. The correction predates this contribution review.

After integrating current main, the Swift grammar regressions, full Swift C
suite and TOML unfinished-value scanner-padding matrix pass. Pruning-certificate
tests, internal facts tests, root facts/outline tests and all 16 editor-gate
tests also pass. No graduation or default-route change is included.

The contribution creates a new test package under `grammars/swift`. Add it to
the grammar-support race lane and select the directive oracle in the existing
locked-C job. The package passes its container race test; the CI planner now
assigns every test package to an execution lane.

Reproducible logs, exact contribution and diagnostic overlays are in
`/home/draco/.local/state/gotreesitter-evidence/20261011-swift-contribution-review`.
Rejected runtime experiments are retained there only as review evidence.
