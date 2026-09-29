# Refactor Plan — Blok2ttrpg

Maintenance/modernization plan. Behavior-preserving throughout: no changes to
rules math, cost formulas, route paths, form field names, template output
structure, YAML schema, or the character JSON/migration chain.

Status: passes 0a-9, 11 and 12 landed and verified, plus a sticky-header layout
fix reported during review. Pass 10 is deferred pending sign-off; it is the only
remaining pass, and the only one with real behavioral risk.


## Baseline (measured on a clean worktree, commit 4ff21d7)

| Gate | Before | After passes 0a-9 |
| --- | --- | --- |
| `go build ./...` | pass | pass |
| `go vet ./...` | pass | pass |
| `go test ./...` | **FAIL** (`TestDocsOrderIsComplete`) | **pass** |
| `gofmt -l .` | 33 files (all CRLF false positives) | **empty** |
| `deadcode -test ./...` | 2 unreachable funcs | **0** |
| `go run ./cmd/gendocs` | 127366 bytes | **127366 bytes (identical)** |
| `verifylib` / `libaudit` / `pkgcost` / `lvlcheck` | pass | pass |

Two problems existed before any refactoring started:

1. The test suite was red, so there was no safety net. Pass 0b restored it.
2. `gofmt -l .` could not be used as a gate: the repo had no `.gitattributes`
   and `core.autocrlf=true`, so Go sources were checked out CRLF and gofmt
   reported every one of them regardless of content. Verified by diffing
   `internal/store/store.go`: content byte-identical, 84 CRLF pairs the only
   difference. Pass 0a fixed this.

Passes 0a and 0b are prerequisites, not refactors.

## Inventory — smells found

| ID | Smell | Evidence | Outcome |
| --- | --- | --- | --- |
| A | Coercion helpers duplicated across packages | `engine/coerce.go` has `asInt`/`asString`/`asBool`/`asRows`; `web/abilities.go` redefines `asInt` and `asStringValue`; `web/handlers.go` has `atoiAny`; `web/funcs.go` and `engine/normalize.go` both define a differently-bodied `normalizeRows` | **Not merged — see note below** |
| B | `abs` duplicated | `web/conditions.go` and `engine/cost.go` | Left as-is: they are in different packages and both unexported, so there is no import cycle or shadowing risk and merging would require exporting a trivial helper |
| C | Stdlib reimplemented by hand | `web/conditions.go` `sortStrings` (insertion sort); `config/lookup.go` `indexByte`; `engine/cost.go` `indexByteStr` | Fixed in passes 4 and 9 |
| D | Dead exported code | `deadcode` proved `docs/render.go` `FuncMapForTest` and `RenderHTML` unreachable, including from tests | Removed in pass 1 |
| E | `web/abilities.go` 955 lines, four concerns | handlers, form decoding, view-models, affordability rules in one file | Split in pass 5 |
| F | `config/lookup.go` 610 lines, 38 funcs, five domains | components, conditions, proficiencies, options, level tables | Split in pass 6 |
| G | Template funcs carry business rules | `funcs.go` `firstOption` documents itself as mirroring engine normalization | Deferred to pass 10 |
| H | Oversized unsectioned assets | `static/css/app.css` 1537 lines, `static/js/app.js` 734 lines | Deferred to pass 12 |
| I | `internal/config` has zero test files | no coverage for lookup/validate/schema | Deferred to pass 11 |

### Note on smell A: the coercers are not interchangeable

Passes 2 and 3 were planned as "consolidate the duplicate coercers" and were
deliberately **not** carried out, because the implementations differ in ways
that would change behavior:

- `web.atoiAny` parses with `strconv.Atoi(strings.TrimSpace(t))` and reports
  failure via a second return value.
- `engine.asInt` parses with `strconv.Atoi(t)`, does not trim, and silently
  yields 0 on failure.

So a posted value of `" 5 "` resolves to 5 through the web helper and 0 through
the engine helper. Collapsing them onto either implementation would silently
change how whitespace-padded form input is priced, which is a rules-math change.
Unifying them is still worth doing, but it needs a deliberate decision about
which parsing behavior is correct plus tests that pin it down, so it belongs in
its own reviewed change rather than inside a cleanup pass.

## Passes

Default gate is `gofmt -l .` + `go vet ./...` + `go test ./...`; extra gates
listed per pass.

| # | Package | Target smell | Status | Validation result |
| --- | --- | --- | --- | --- |
| 0a | repo | CRLF defeats the `gofmt` gate | DONE | `gofmt -l .` empty; `git diff -w` empty (whitespace only, 33 files) |
| 0b | `config` | red baseline (unpublished draft chapters) | DONE | suite green; `generated_docs.md` byte-identical |
| 1 | `docs` | D — dead exported funcs | DONE | `deadcode` now reports 0; docs identical |
| 2 | `engine` | A — canonical coercers | **SKIPPED** | unsafe; see note above |
| 3 | `web` | A — drop local coercers | **SKIPPED** | unsafe; see note above |
| 4 | `web`, `config` | C — `slices.Sort`, `strings.IndexByte` | DONE | suite + all four CLIs green |
| 5 | `web` | E — split `abilities.go` | DONE | 955 to 716+96+159; 39 decls before and after |
| 6 | `config` | F — split `lookup.go` by domain | DONE | 610 to 5 files (max 241); decls preserved |
| 7 | `docs` | `buildguide.go` 550 lines | DONE | 3 files; `generated_docs.md` byte-identical |
| 8 | `config` | `schema.go` 714 lines | DONE | 4 files (max 241); 53 decls preserved |
| 9 | `engine` | `cost.go` 499, `instructions.go` 484 | DONE | 5 files (max 340); decls preserved |
| 10 | `web`, `engine` | G — rule-bearing template funcs behind engine | **DEFERRED** | needs sign-off; golden-HTML assertions first |
| 11 | `config` | I — table-driven tests for untested branches | DONE | coverage 0% to 14.6%; 58 subtests |
| 12 | `static` | H — prune orphaned CSS, audit JS | DONE | 5 rules removed; braces balanced; suite green |

### Pass 11 notes

`internal/config` had no test files at all, so the level budgets and the
proficiency ladder -- both rules math despite living in the config package --
were only covered indirectly, through handlers that happened to call them with
valid input. Three table-driven files were added:

- `lookup_levels_test.go` pins the three ways a budget resolves (explicit row,
  formula, fallback), that an explicit row beats the formula, and that the public
  accessors clamp *before* resolving so an out-of-range level can never mint
  points beyond the cap.
- `lookup_proficiency_test.go` covers tier lookup, the default-rung fallbacks and
  the shift clamping. It also asserts that `ShiftProficiency` and `ShiftClamped`
  agree across every rung and delta: the first decides where a skill lands, the
  second decides whether the sheet warns that the shift was capped, and if they
  disagree the sheet either warns about a shift that worked or hides one that did
  not.
- `lookup_components_test.go` pins the allow/block precedence in `filterByList`
  (the allow list wins when both are set) and the unknown-id fallbacks that let a
  perk built against an older ruleset still open.

One expectation of mine was wrong on the first run: I asserted that landing
exactly on the top rung counted as clamped. It does not, and should not -- the
shift was applied in full. The test was corrected, not the code.


### Sticky header overlap (reported during review)

The character bar scrolled underneath the site header instead of resting below
it. `.topbar`, `.stats-bar` and `.builder-head` were all `position: sticky` with
`top: 0`, and the bar's lower z-index put it behind the header.

Fixed by introducing `--topbar-h` (plus `--z-topbar` / `--z-sticky`) and
offsetting every other sticky element by it, so the relationship is stated once
rather than as four independent magic numbers. `.doc-sidebar` and the doc
heading `scroll-margin-top` previously used hardcoded `3.5rem` / `4.5rem` for
the same purpose and now derive from the same variable. The narrow-screen
breakpoint raises `--topbar-h`, because below 720px the character name moves
onto its own row and the header gets taller.

Guarded by `internal/web/stickyoffset_test.go`, which asserts against the
stylesheet (the invariant is not observable in rendered HTML). The test was
verified by reverting `.stats-bar` to `top: 0` and confirming it fails with a
diagnostic message.

### Pass 12 findings

Audited all 178 CSS classes against every template, Go file and JS file.
Removed 5 genuinely unreferenced rules: `.card-list`, `.pdf-toolbar`,
`.perk-card`, `.passive-value` (both rules) and `.passive-upgrade-hint`. The
PDF templates use a separate `print-*` namespace, so the first two were
leftovers from that rename.

Five classes that the audit flagged were deliberately **kept**, and the reason
is now recorded in the stylesheet so a future audit does not delete them:

- `.section-general`, `.section-offense`, `.section-defense`, `.section-vital`
  are live. `character.html` builds the class as `section-{{ $g.ID }}` from the
  skill group ids in `skills.yaml`, so the names never appear literally in any
  template. Deleting them would have silently removed the character sheet's
  section colours.
- `.topbar .char-name` is a latent bug rather than dead styling.
  `templates/layout.html` still has the `{{ with .Character }}{{ if .ID }}`
  block and the comment describing the feature, but the span itself is missing,
  so the character name never renders. The data is available
  (`pageData.Character`, `Character.Name()`), so this is an unfinished feature;
  the CSS is kept and the gap is documented at both ends.

`static/js/app.js` was audited the same way: all 25 `getElementById` targets
exist in templates, so nothing was pruned.


Every split was verified as a pure move by comparing the sorted set of
top-level declarations before and after. The only intentional deltas are the
three hand-rolled stdlib helpers removed in passes 4 and 9, and the two dead
functions removed in pass 1.

## Out of scope

- `internal/model` migration steps — every step stays so old character files load.
- YAML under `config/`, content under `library/`, markdown under `docs/`.
- Route paths, form field names, template output structure.
- `go.mod` dependency bumps (separate task per the no-mixed-concerns rule).
