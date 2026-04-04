# ViBrowsing — Project Plan

A from-scratch web browser written in Go, with no external browser engine libraries.

## What are we building?

A toy browser project that renders web pages using its own HTML/CSS engine, built entirely from scratch in Go. JavaScript is out of scope. Flexbox/Grid is out of scope. Production use is also out of scope.

## Language & Stack

- **Go 1.21+** — fast compilation, simple concurrency, good stdlib
- **No external browser engine libraries** — everything from scratch
- **Ebitengine** (planned) — for window chrome and input

## Module Structure

```
github.com/nickcoury/ViBrowsing/
├── cmd/browser/main.go      # Entry point
├── internal/
│   ├── fetch/fetch.go      # HTTP client, URL resolution, redirects
│   ├── html/
│   │   ├── tokenizer.go    # HTML5 tokenizer (state machine)
│   │   ├── node.go        # DOM node types
│   │   └── parser.go      # Token → DOM tree builder
│   ├── css/
│   │   ├── values.go      # Length units, colors
│   │   ├── parser.go      # CSS rule parser
│   │   └── style.go       # Cascaded style computation
│   ├── layout/
│   │   ├── box.go         # LayoutBox tree from DOM
│   │   └── block.go      # Block-level layout algorithm
│   └── render/
│       └── canvas.go      # RGBA pixel buffer → PNG
└── sample_pages/          # Test pages served on GitHub Pages
```

## Architecture

```
URL Input
    ↓
net/http fetch
    ↓
HTML5 Tokenizer → DOM Tree
    ↓
CSS Parser → Cascaded Styles
    ↓
Box Layout Engine → Render Tree
    ↓
Pixel Buffer → PNG (CLI)
```

## Task Breakdown

### Phase 0: Scaffold (DONE)
- [x] Create repo structure
- [x] go mod init
- [x] Choose project name (ViBrowsing)

### Phase 1: Fetch + HTML Tokenizer (DONE)
- [x] `fetch.go`: GET request, follow redirects (max 5), extract Content-Type
- [x] `tokenizer.go`: HTML5 tokenizer — state machine
- [x] `node.go`: Document, Element, Text, Comment node types

### Phase 2: DOM Tree + CSS Parser (IN PROGRESS)
- [x] `parser.go`: Tokens → DOM tree
- [x] `css/values.go`: CSS length units, colors
- [x] `css/parser.go`: Parse stylesheets
- [ ] `css/style.go`: Full cascade (origin, specificity, !important)
- [ ] Default stylesheet (block elements default to block display)

### Phase 3: Layout Engine (IN PROGRESS)
- [x] `layout/box.go`: DOM → LayoutBox tree
- [x] `layout/block.go`: Block layout algorithm
- [ ] `layout/inline.go`: Inline text layout with wrapping

### Phase 4: Renderer (IN PROGRESS)
- [x] `render/canvas.go`: RGBA pixel buffer
- [x] Basic box painting (background, border, padding)
- [ ] Proper text rendering (load font, measure glyphs)
- [ ] Paint borders correctly (all 4 sides)

### Phase 5: Integration (TODO)
- [ ] Connect: fetch → parse → layout → render → PNG output
- [ ] CLI: `go run ./cmd/browser [url]`
- [ ] Ebitengine window (planned)

### Phase 6: Polish & Iteration (TODO)
- [ ] Support `<img>` tags
- [ ] Support `<a>` tags with click → navigate
- [ ] Support `<style>` tags and inline styles
- [ ] Back/forward history
- [ ] Tab support (multiple pages)
- [ ] Scroll bars
- [ ] CSS `display: inline`, `display: none`
- [ ] Margin collapse behavior
- [ ] Font loading (woff/ttf)
- [ ] Ebitengine window chrome

## What's Hard

1. **Text wrapping** — Word-breaking, measuring text width against container
2. **Inline layout** — Mixing text nodes and inline elements on the same line
3. **Cascading** — CSS specificity, `!important`, user-agent stylesheet
4. **Margin collapse** — Adjacent block margins merging
5. **Border rendering** — All 4 sides drawn correctly

## Milestones

1. **M1** — Window opens, fetches URL, prints raw HTML to console
2. **M2** — DOM tree printed as indentation
3. **M3** — Rendered page (text only) to PNG, styled with defaults
4. **M4** — CSS-applied styles visible (colors, font sizes)
5. **M5** — Navigation works (type URL, hit Enter, page renders)
6. **M6** — Back button, history
7. **M7** — Window chrome with URL bar, tabs
