# ViBrowsing Roadmap — From Blank Pages to Real Rendering

## Current State Assessment

ViBrowsing has ~20K lines of Go implementing a full HTML→CSS→Layout→Render pipeline. The HTML5 tokenizer/parser is solid, CSS cascade handles 140+ properties, layout covers block/inline/flex/float/positioned, and the renderer does software rasterization to PNG via freetype.

**The core problem: real websites render blank (or nearly blank).**

### Root Causes (ranked by impact)

1. **External CSS is never loaded.** `<link rel="stylesheet">` tags are parsed but their CSS is never fetched. Since ~95% of real websites use external stylesheets, the cascade runs with zero author rules. Everything falls back to user-agent defaults → invisible/wrong layout.

2. **`<style>` extraction happens in main.go, not the engine.** CSS from `<style>` tags works, but stylesheet discovery and application is ad-hoc, not integrated into the rendering pipeline.

3. **No `display: inline-block` support.** Many common patterns (nav bars, button groups, badges) rely on `inline-block`. Currently falls through to default.

4. **Table layout is structural only.** Tables parse but cells don't size to content or distribute column widths. Wikipedia, email clients, many older sites are table-heavy.

5. **No image rendering.** `<img>` shows a placeholder box, never fetches or decodes actual images. Pages look broken even when layout is correct.

6. **`<script>` tags leak into rendered output.** Script content should be suppressed from rendering even without JS execution.

7. **CSS shorthand expansion is incomplete.** Properties like `border`, `margin`, `padding`, `background`, `font` work partially but edge cases cause silent failures.

8. **Style inheritance is flat.** `ComputeStyle` doesn't walk up the DOM for inherited properties (`color`, `font-*`, `line-height`). Children default to user-agent values instead of inheriting from parents.

---

## Milestone Plan

### M1: External CSS Loading (Critical Path)
**Goal:** Fetch and apply `<link rel="stylesheet">` CSS. This single change will transform rendering for most websites.

#### Tasks
- [ ] **M1.1** Add `FetchStylesheet(url, baseURL) (string, error)` to `internal/fetch/`
  - Handle relative URLs via `ResolveURL`
  - Respect Content-Type (skip non-text responses)
  - Add timeout (5s per stylesheet)
  - Handle `@import` rules (1 level deep max)
- [ ] **M1.2** Move stylesheet discovery into `BuildLayoutTree` or a new `ResolveStyles` phase
  - Find all `<link rel="stylesheet" href="...">` in `<head>`
  - Find all `<style>` tags
  - Fetch external sheets, parse all, merge into single `[]Rule`
  - Respect `media` attribute on `<link>` tags
- [ ] **M1.3** Handle CSS `@import` in parser
  - `css.Parse()` should recognize `@import url("...")` at the top of a stylesheet
  - Return import URLs for the caller to fetch and prepend
- [ ] **M1.4** Test: render `example.com` — should show styled text
- [ ] **M1.5** Test: render `news.ycombinator.com` — should show orange header, ranked list

### M2: CSS Inheritance Fix
**Goal:** Inherited properties (`color`, `font-family`, `font-size`, `font-weight`, `font-style`, `line-height`, `text-align`, `letter-spacing`, `word-spacing`, `text-transform`, `white-space`, `visibility`, `direction`, `list-style-*`) propagate from parent to child.

#### Tasks
- [ ] **M2.1** Define `inheritedProperties` set in `css/style.go`
- [ ] **M2.2** Modify `buildBox()` in `box.go` to pass parent computed style
  - For each inherited property: if child has no author/inline rule, copy from parent
  - Already partially done (parentStyle passed for TextBox) — extend to all elements
- [ ] **M2.3** Test: nested `<div style="color:red"><p>text</p></div>` → paragraph text is red

### M3: Script Suppression + Display Fixes
**Goal:** Clean up rendering artifacts from unsuppressed content and missing display modes.

#### Tasks
- [ ] **M3.1** Skip `<script>` tag content in layout tree building (don't create TextBox for script content)
- [ ] **M3.2** Skip `<style>` tag content in layout tree building (already used for CSS, shouldn't render as text)
- [ ] **M3.3** Implement `display: inline-block`
  - Layout: compute intrinsic width from children, participate in inline flow
  - Render: draw as block within line
- [ ] **M3.4** Implement `display: table` / `display: table-row` / `display: table-cell`
  - Basic auto-layout: equal column distribution
  - Content-based column sizing (measure text width)
- [ ] **M3.5** Test: nav bar with `display: inline-block` items renders horizontally

### M4: Image Fetching + Rendering
**Goal:** Actually fetch and display `<img>` elements.

#### Tasks
- [ ] **M4.1** Add `FetchImage(url, baseURL) (image.Image, error)` to `internal/fetch/`
  - Support JPEG, PNG, GIF, WebP
  - Decode via Go's `image` package
  - Cache decoded images in memory (map[url]image.Image)
  - Size limit: skip images > 10MB
- [ ] **M4.2** Implement image scaling in `canvas.go`
  - Scale to CSS `width`/`height` if specified
  - Maintain aspect ratio with `object-fit` defaults
  - Fall back to intrinsic image dimensions
- [ ] **M4.3** Draw decoded image pixels onto canvas at layout position
- [ ] **M4.4** Handle `<img>` without explicit dimensions
  - Use intrinsic size from decoded image
  - Trigger relayout if image dimensions differ from placeholder
- [ ] **M4.5** Test: page with inline images renders them at correct size

### M5: CSS Shorthand + Value Robustness
**Goal:** Handle the real-world CSS that websites actually use.

#### Tasks
- [ ] **M5.1** `border` shorthand → expand to `border-top-width`, `border-top-style`, `border-top-color` (× 4 sides)
- [ ] **M5.2** `margin` shorthand → already works for 1-4 values, verify edge cases
- [ ] **M5.3** `padding` shorthand → same
- [ ] **M5.4** `font` shorthand → `font-style font-variant font-weight font-size/line-height font-family`
- [ ] **M5.5** `background` shorthand → expand to individual `background-*` properties
- [ ] **M5.6** `list-style` shorthand → `list-style-type list-style-position list-style-image`
- [ ] **M5.7** `flex` shorthand → `flex-grow flex-shrink flex-basis`
- [ ] **M5.8** Handle `!important` in cascade (currently framework-ready but not enforced)
- [ ] **M5.9** Handle `inherit`, `initial`, `unset` keyword values

### M6: Table Layout (Real)
**Goal:** Tables render with proper column sizing and cell alignment.

#### Tasks
- [ ] **M6.1** Column width calculation
  - Measure content width of each cell
  - Distribute remaining width proportionally
  - Respect `width` attributes on `<td>`, `<th>`, `<col>`
- [ ] **M6.2** `colspan` layout — cell spans multiple columns
- [ ] **M6.3** `rowspan` layout — cell spans multiple rows
- [ ] **M6.4** Table borders: `border-collapse: collapse` vs `separate`
- [ ] **M6.5** Cell padding and alignment (`text-align`, `vertical-align`)
- [ ] **M6.6** Test: render a Wikipedia infobox table

### M7: Conformance Testing with html5lib + WPT
**Goal:** Validate correctness against standardized test suites.

#### Tasks
- [ ] **M7.1** Download html5lib-tests tokenizer tests (JSON format)
  - Write Go test harness that runs each test case through tokenizer
  - Compare token output against expected
  - Track pass/fail counts
- [ ] **M7.2** Download html5lib-tests tree-construction tests
  - Write harness that runs HTML through parser
  - Compare DOM tree structure against expected
  - Focus on: adoption agency, foster parenting, template, foreign content
- [ ] **M7.3** Download WPT CSS test suite (css/CSS2 subset)
  - Focus on: box model, visual formatting, colors, backgrounds, fonts
  - Render each test case to PNG
  - Compare against reference image (pixel diff or structural check)
- [ ] **M7.4** Add `go test ./...` CI step
  - Run all existing harness tests
  - Run html5lib tokenizer/parser tests
  - Fail on regressions
- [ ] **M7.5** Create test dashboard: % of html5lib tests passing, tracked over time

### M8: Rendering Quality Polish
**Goal:** Make correctly-structured pages actually look right.

#### Tasks
- [ ] **M8.1** Fix margin collapse between adjacent blocks
  - Currently partially implemented — verify edge cases (negative margins, clearance)
- [ ] **M8.2** Fix inline element wrapping at container boundary
  - Words should break at whitespace, not mid-word (unless `overflow-wrap: break-word`)
- [ ] **M8.3** Fix text baseline alignment within line boxes
  - Multiple inline elements on same line should share a baseline
- [ ] **M8.4** `auto` margins for centering (`margin: 0 auto`)
  - Calculate available space, split equally
- [ ] **M8.5** `max-width` / `min-width` constraints in layout
- [ ] **M8.6** Percentage width/height resolution against containing block
- [ ] **M8.7** Fix `line-height` interaction with font-size (normal = 1.2× font-size)

---

## Testing Strategy

### Unit Tests (per component)
- **Tokenizer:** html5lib tokenizer test corpus (~2700 tests)
- **Parser:** html5lib tree-construction tests (~1100 tests)
- **CSS Parser:** Property parsing, shorthand expansion, specificity
- **CSS Cascade:** Rule matching, inheritance, !important
- **Layout:** Box model math, margin collapse, inline wrapping, flex distribution
- **Render:** Pixel output for known inputs (golden-file PNG comparison)

### Integration Tests
- **Reference pages:** Render 5-10 known pages, diff against baselines
  - `example.com` — simplest possible real page
  - `news.ycombinator.com` — text-heavy, simple layout
  - `lite.cnn.com` — news content, minimal CSS
  - `en.wikipedia.org/wiki/HTML` — tables, infoboxes, structured content
  - Local `sample_pages/*.html` — controlled test cases

### Regression Tests
- Every bug fix gets a test case that reproduces the bug
- Golden PNG snapshots checked into repo for visual comparison
- `go test -run TestRender` suite for automated visual regression

### Fuzz Testing
- `go test -fuzz` on tokenizer and CSS parser
- Goal: no panics on any input

---

## Context File Changes

### Files to Update

1. **PLAN.md** — Outdated. Says "Flexbox/Grid is out of scope" but flexbox is implemented. Says phases 2-5 are IN PROGRESS/TODO but most are done. Should be replaced by this ROADMAP.md or significantly revised to match reality.

2. **README.md** — Says "Planning phase" but the project has 20K lines and renders pages. Update status section. Add sample render output. Update architecture diagram.

3. **backlog.md** — 70KB and growing. Many completed items still listed with strikethrough. Consider:
   - Archive completed items to `CHANGELOG.md` or `backlog-completed.md`
   - Keep only open items in `backlog.md`
   - Cross-reference items to this roadmap's milestones

### Files to Add

4. **CLAUDE.md** — Essential context file for MiniMax/any LLM working on this project. Should contain:
   - Architecture overview (the pipeline: fetch→tokenize→parse→cascade→layout→render)
   - Key files and what they do
   - How to build and test (`go build ./cmd/browser`, `go test ./...`)
   - Current priorities (this roadmap's M1-M3)
   - Known bugs and gotchas
   - Code conventions (package structure, error handling patterns)

5. **Makefile** — `make build`, `make test`, `make render URL=...`, `make clean`

6. **test/references/** — Golden PNG files for visual regression testing

### Files to Remove/Trim

7. **Pre-built binaries** (`browser`, `vibrowing`, `vibrowser-test`) — 25MB+ of binaries in the repo. These should be in `.gitignore`, not committed. They go stale immediately.
8. **`*.png` render outputs** — Should be in `.gitignore` unless they're golden reference images.

---

## Recommended Execution Order

```
M1 (External CSS)     ← This is the #1 unlock. Everything else is polish without it.
  ↓
M2 (Inheritance)      ← Quick win, fixes cascading color/font issues
  ↓
M3 (Script/Display)   ← Removes visual garbage, adds inline-block
  ↓
M7 (Conformance)      ← Set up test infrastructure before more features
  ↓
M5 (Shorthands)       ← Real CSS uses shorthands everywhere
  ↓
M4 (Images)           ← Visual impact but not structural
  ↓
M8 (Polish)           ← Fine-tuning
  ↓
M6 (Tables)           ← Complex, lower priority than the above
```

## Security Considerations

- **Fetch:** Already has redirect limit (5) and timeout. Good.
- **DOM depth:** Max 10,000 levels. Good.
- **Text size:** Max 1MB per text node. Good.
- **Missing:** No SSRF protection (fetching `file:///etc/passwd` or internal IPs). Add URL scheme allowlist (`http`, `https`, `file` only for local mode).
- **Missing:** No Content-Security-Policy for fetched resources (not critical for a toy browser but worth noting).
- **Images:** Add size limit before decode (prevent zip bombs / decompression attacks).

## Maintainability

- **Code organization is good.** Clean package boundaries: `fetch`, `html`, `css`, `layout`, `render`.
- **Test coverage exists but is harness-only.** No integration/visual regression tests yet.
- **The backlog is bloated.** 70KB file with 150+ items. Needs triaging — many items are aspirational features that distract from core rendering bugs.
- **Commit messages are automated sprint dumps.** Consider requiring descriptive commit messages that explain *what* changed and *why*.
