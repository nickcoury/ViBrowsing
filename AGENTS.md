# ViBrowsing — Developer Context

A from-scratch web browser in Go. No Chromium, no WebKit — everything from HTML tokenizer to pixel output is handwritten.

## Architecture

```
URL → Fetch (internal/fetch/) → HTML bytes
  → Tokenize (internal/html/tokenizer.go) → Token stream
  → Parse (internal/html/parser.go) → DOM tree (html.Node)
  → Cascade (internal/css/style.go) → Computed styles per element
  → Layout (internal/layout/) → Box tree with positions/dimensions
  → Render (internal/render/canvas.go) → RGBA pixels → PNG
```

## Key Files

| File | Purpose | Lines |
|------|---------|-------|
| `cmd/browser/main.go` | CLI entry point, wires pipeline together | 315 |
| `internal/fetch/fetch.go` | HTTP client, redirects, cookies | — |
| `internal/html/tokenizer.go` | HTML5 state machine tokenizer | — |
| `internal/html/parser.go` | Token → DOM tree builder | 412 |
| `internal/html/node.go` | DOM node types, querySelector, innerHTML | — |
| `internal/css/parser.go` | CSS rule parser (@media, selectors, declarations) | — |
| `internal/css/style.go` | Cascade: specificity, matching, computed styles | 2035 |
| `internal/css/values.go` | CSS value parsing (colors, lengths, calc()) | — |
| `internal/css/selector.go` | CSS selector matching (combinators, pseudo-classes) | — |
| `internal/layout/box.go` | DOM → Box tree conversion, box types | 776 |
| `internal/layout/block.go` | Layout algorithms (block, inline, flex, float, positioned) | 1743 |
| `internal/render/canvas.go` | Software rasterizer, text rendering via freetype | 3940 |

## Build & Run

```bash
go build -o browser ./cmd/browser
./browser https://example.com                    # renders to output.png
./browser https://example.com -dump-dom          # print DOM tree
./browser https://example.com -dump-layout       # print layout tree
./browser https://example.com -viewport 1024x768 # custom viewport
./browser file:///path/to/test.html              # local file
go test ./...                                    # run all tests
```

## Current Priorities (see ROADMAP.md for full plan)

**M1: External CSS Loading** — THE critical blocker. `<link rel="stylesheet">` tags are parsed but their CSS is never fetched. This is why real websites render blank. Fix this first.

**M2: CSS Inheritance** — Inherited properties (color, font-*, line-height, text-align) don't propagate from parent to child. Children get user-agent defaults instead of inheriting.

**M3: Script/Style Suppression + inline-block** — `<script>` content leaks into rendered output. `display: inline-block` is not implemented.

## What Works

- HTML5 tokenizer (state machine, entity decoding, foreign content)
- DOM tree builder (foster parenting, template, void elements)
- CSS parser (rules, @media, @supports, @keyframes)
- CSS cascade (specificity, 140+ properties, attribute/pseudo selectors)
- Layout: block, inline with wrapping, flexbox, float, positioned (abs/rel/fixed/sticky)
- Rendering: backgrounds, borders (rounded), text (freetype), shadows, opacity, clipping, list markers
- Fetch: HTTP with redirects, cookies, charset detection

## What's Broken / Missing

- **External stylesheets never fetched** (root cause of blank pages)
- **CSS inheritance doesn't work** (children don't inherit parent styles)
- **`<script>` content renders as text**
- **`display: inline-block` not implemented**
- **Images show placeholder, never fetched/decoded**
- **Table layout is structural only** (no column width distribution)
- **`!important` not enforced in cascade**
- **No `inherit`/`initial`/`unset` keyword handling**

## Code Patterns

- All CSS properties stored as `map[string]string` on each box
- `css.ParseLength(value)` returns a `Length` struct (value + unit)
- `css.ParseColor(value)` returns `color.Color`
- Layout uses recursive `layoutChildren()` with a `LayoutContext` for cursor tracking
- Canvas uses a clip stack for `overflow: hidden`
- Safety limits: max DOM depth 10,000, max text node 1MB, max document 10MB

## Testing

- 14 test files, ~4,900 lines of test code
- Harness-based: tokenizer, parser, CSS parser, CSS values, CSS selectors, layout
- Run: `go test ./...`
- No visual regression tests yet (planned in M7)

## Gotchas

- `ComputeStyle()` takes `(tagName, class, id, inlineDecls, rules)` — it doesn't take the DOM node, so attribute selectors are handled separately via `ComputeStyleForNode()`
- `buildBox()` hardcodes element-to-BoxType mapping (large switch statement in box.go:258+)
- The `background` property in the default styles means "background shorthand", but code sometimes reads it as "background-color"
- Pre-built binaries in repo root go stale — always rebuild with `go build`
