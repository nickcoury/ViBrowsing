# ViBrowsing Backlog

## 🔴 Critical (Parser/Rendering)

- [x] ~~Fix HTML parser double-html/body bug~~ (2026-04-03 sprint) — tokenizer now skips html/head/body StartTag/EndTag tokens; parser bootstraps them once cleanly; no more duplication
- [x] ~~Fix foster parenting~~ (2026-04-03 sprint) — parser now tracks table context, text inside tables is fostered to parent; implicit <p> close before block elements; table end tags properly close the table context
- [x] ~~Fix unclosed tag handling~~ (2026-04-03 sprint partial) — generic end tag now pops stack until matching tag found; unknown end tags are skipped without crashing; block tags implicitly close open <p> tags
- [x] ~~Implement entity decoding~~ (2026-04-03 sprint) — added decodeEntities() with named entities (amp, lt, gt, quot, apos, nbsp, ndash, mdash, lsquo, rsquo, ldquo, rdquo, hellip, copy, reg, trade, deg, plusmn, times, divide, frac12, frac14, frac34) and numeric entities (&#65;, &#x41;)
- [x] ~~Implement foreign content handling~~ (2026-04-04 sprint) — svg/math tracked via foreignContent counter; HTML-specific rules (p-closing, foster parenting) disabled inside foreign content; end tags only pop matching element in foreign context; bootstrap stack corrected to include head+body so elements append to body not html

## 🟡 High (Layout/Rendering)

- [x] ~~Implement CSS box model properly~~ (2026-04-03 sprint) — TotalWidth/TotalHeight now use Box fields; canvas draws margin/background/padding/border in correct order with proper geometry
- [x] ~~Implement flexbox layout~~ (2026-04-03 sprint) — added FlexBox type, flex-direction (row/column/reverse), justify-content, align-items, align-self, flex-grow, flex-basis, gap CSS properties
- [x] ~~Implement inline layout~~ (2026-04-03 sprint) — text wraps at container width; white-space:normal collapses whitespace, pre/pre-wrap preserves it; text boxes inherit parent style including white-space; explicit newline handling in pre mode
- [x] ~~Implement float~~ (2026-04-04 sprint) — float:left/right with LayoutContext.FloatLeftEdge/FloatRightEdge/FloatBottom tracking; blocks below float edge clear floats and reflow below
- [x] ~~Implement positioned layout~~ (2026-04-03 sprint) — position:absolute/relative/fixed with top/left offsets; positioned elements use PositionedBox type; fixed uses viewport as containing block
- [x] ~~Implement z-index stacking~~ (2026-04-03 sprint) — children sorted by z-index before drawing; positioned elements drawn after normal flow; z-index parsed as integer
- [x] ~~Implement `visibility: hidden` and `display: none`~~ (2026-04-03 sprint) — visibility:hidden now paints background/border/padding but hides content and children; display:none skips box entirely
- [x] ~~Implement overflow handling~~ (2026-04-03 sprint) — overflow:hidden/scroll/auto clips children to content box via clip stack; visible is no-op (default)
- [x] ~~Implement `position: sticky`~~ (2026-04-05 sprint) — treated as relative during layout with sticky offset values stored for render-time adjustment
- [x] ~~Implement `visibility: collapse`~~ (2026-04-05 sprint) — table rows/cells with collapse take no layout space
- [x] ~~Implement `clip-path` CSS rendering~~ (2026-04-05 sprint) — inset/circle/ellipse/polygon clip shapes applied via PushClip in DrawBox
- [x] ~~Implement `overflow: visible` handling~~ (2026-04-05 sprint) — confirmed visible (default) applies no clipping

## 🟡 High (HTML/CSS Coverage)

### HTML Elements
- [x] ~~Implement `<br>` void element~~ (2026-04-04 sprint) — br forces line break in inline layout; advances Y, resets X, flushes line box
- [x] ~~Implement `<hr>` void element~~ (2026-04-04 sprint) — HorizontalRuleBox type; renders as 1px border with margin
- [x] ~~Implement `<wbr>` void element~~ (2026-04-04 sprint) — zero-width break opportunity in inline layout
- [x] ~~Implement `<meta>` and `<link>` void elements~~ (2026-04-04 sprint) — handled via voidElements list in parser (no crash)
- [x] ~~Implement `<input>` form element (visual)~~ (2026-04-04 sprint) — renders as inline-block with border and padding
- [x] ~~Implement remaining void elements~~ (2026-04-04 sprint 2) — `<area>`, `<base>`, `<col>`, `<embed>`, `<param>`, `<source>`, `<track>` handled in parser
- [ ] **Implement table layout** — `<table>`, `<thead>`, `<tbody>`, `<tfoot>`, `<tr>`, `<td>`, `<th>`, `colspan`, `rowspan`, `border` attribute. Tables are complex in HTML/CSS
- [x] ~~Implement table `caption-side`~~ (2026-04-05 sprint) — captions positioned top/bottom via layoutCaption
- [x] ~~Implement `empty-cells: show/hide`~~ (2026-04-05 sprint) — empty cells detected and marked; rendering uses empty-cells style
- [x] ~~Implement list layout~~ (2026-04-04 sprint) — ListItemBox type, layoutListItemChild with bullet/number markers, DrawListMarker renders disc/square/circle/number markers, list-style-type/position/image properties
- [ ] **Implement form elements** — `<input>`, `<button>`, `<select>`, `<textarea>`, `<label>` (visual only, no interactivity)
- [ ] **Implement media elements** — `<img>` (display), `<video>`, `<audio>` (show controls UI)
- [x] ~~Implement semantic block elements~~ (2026-04-04 sprint) — header, footer, nav, article, section, aside, main, figure, figcaption, details, summary all render as block
- [x] ~~Implement `<noscript>`~~ (2026-04-04 sprint) — treated as block element, content rendered
- [ ] **Implement `<script>` and `<style>`** — style content parsed as CSS; script content may be JS (don't execute, just skip)
- [ ] **Implement `<template>`** — parse but don't render template content

### CSS Properties
- [x] ~~Implement CSS `color` property~~ (2026-04-04 sprint) — ParseColor now supports RGB, RGBA, HSL, HSLA, hex (#RGB, #RRGGBB), and named colors; rgba() alpha accepts 0-1 and percentage; fixed RGBA() to properly expand 8-bit to 16-bit
- [x] ~~Implement CSS `background-color`~~ (2026-04-04 sprint) — same ParseColor parser used for all color values including hsl()/hsla(); ParseFloat255 fixed to cap at 255 not 1
- [x] ~~Implement CSS `background` shorthand~~ (2026-04-05 sprint) — improved: color parsing, linear/radial/conic gradients, background-color properly set from shorthand
- [x] ~~Implement CSS `caret-color`~~ (2026-04-05 sprint) — caret rendered in input/textarea at insertion point; parseCaretColor handles auto/explicit colors
- [x] ~~Implement CSS `cursor` property storage~~ (2026-04-05 sprint) — cursor CSS property stored in style (actual cursor shape rendering needs windowing system)
- [x] ~~Implement CSS `text-shadow`~~ (2026-04-04 sprint) — parse offset-x offset-y blur color; draw shadow when rendering text
- [x] ~~Implement CSS `background-image` (parsing)~~ (2026-04-04 sprint) — parses url() values, stores in style; placeholder drawing for url() images
- [ ] **Implement CSS `background-image` (drawing)** — actually draw background image from URL (currently placeholder only)
- [ ] **Implement CSS `background-repeat`, `background-position`, `background-size`** — drawing with repeat patterns and positioned/sized backgrounds
- [ ] **Implement CSS `background` gradients** — `linear-gradient()`, `radial-gradient()` as background-image values
- [x] ~~Implement CSS `border-radius`~~ (2026-04-04 sprint partial) — ParseBorderRadius parses 1-4 values; DrawBorder now uses rounded corner arcs with filled quarter circles; DrawRoundedRect helper added for background with rounded corners
- [x] ~~Implement CSS `box-shadow`~~ (2026-04-04 sprint) — parse box-shadow value; draw shadow rectangle offset from content box
- [x] ~~Implement CSS `text-align`~~ (2026-04-04 sprint) — stored in style props; DrawText respects alignment offset
- [x] ~~Implement CSS `font-weight`, `font-style`, `text-decoration`~~ (2026-04-04 sprint) — stored in style props; DrawText uses font-weight for char width and font-style for italic slant; text-decoration rendering (underline/overline/line-through) implemented in sprint 2
- [x] ~~Implement CSS `line-height`~~ (2026-04-04 sprint) — already worked; ParseLength handles unitless values
- [x] ~~Implement CSS `vertical-align`~~ (2026-04-04 sprint) — top/middle/bottom/baseline/sub/super and length values; LayoutContext tracks LineBoxBaseline/MaxAscent/MaxDescent for deferred vertical-align application
- [x] ~~Implement CSS `opacity`~~ (2026-04-04 sprint) — opacity value stored in style; DrawBox applies applyOpacity() to background and border colors; opacity 0-1 range clamped
- [ ] **Implement CSS `transform`** — rotate, scale, translate (2D transforms)
- [ ] **Implement CSS `@media` queries** — responsive design breakpoints
- [ ] **Implement CSS `cursor`** — show appropriate cursor on interactive elements (pointer, text, wait, etc)
- [x] ~~Implement CSS `outline`~~ (2026-04-04 sprint) — parse outline-width/style/color and outline shorthand; draw outside border box
- [x] ~~Implement CSS `overflow`~~ (2026-04-05 sprint) — hidden/scroll/auto clip via PushClip; visible (default) no clipping
- [x] ~~Implement CSS `white-space`~~ (2026-04-03 sprint) — normal/pre/pre-wrap values; collapses spaces in normal mode; preserves newlines and spaces in pre mode
- [x] ~~Implement CSS `word-wrap` / `overflow-wrap`~~ (2026-04-04 sprint 2) — already in defaults, values stored and used
- [x] ~~Implement CSS `text-overflow`~~ (2026-04-04 sprint 2) — already in defaults; ellipsis drawing implemented in canvas
- [ ] **Implement CSS `content`** — for ::before and ::after pseudo-elements
- [x] ~~Implement CSS `@keyframes` and `animation`~~ (2026-04-05 sprint) — AnimationManager with StartAnimation/StopAnimation/UpdateAnimation; RegisterKeyframes; timing function and iteration count parsing
- [x] ~~Implement CSS `transition`~~ (2026-04-05 sprint) — parse transition-property/duration/timing-function/delay; parseTransitionShorthand handles combined values

### CSS Selectors
- [x] ~~Implement attribute selectors~~ (2026-04-04 sprint 2) — `[attr]`, `[attr=value]`, `[attr~=value]`, `[attr|=value]`, `^=`, `$=`, `*=` implemented with MatchNodeSelector
- [x] ~~Implement pseudo-classes~~ (2026-04-05 sprint) — `:not()` with complex selector support, `:nth-child()` with formula (2n+1, odd, even, 3n), `:valid/:invalid` with format checking, `:placeholder-shown`
- [ ] **Implement pseudo-elements** — `::before`, `::after`, `::first-line`, `::first-letter`
- [x] ~~Implement combinators~~ (2026-04-05 sprint) — descendant (space), child (>), adjacent sibling (+), general sibling (~) all work via splitSelectorParts + matchSelectorChain

### CSS Layout
- [x] ~~Implement flexbox fully~~ (2026-04-05 sprint) — `flex-wrap`, `align-content` added; rest was done in prior sprints
- [ ] **Implement CSS grid** — `display: grid`, `grid-template-columns`, `grid-template-rows`, `grid-column`, `grid-row`, `gap`, `span`
- [x] ~~Implement float~~ (2026-04-04 sprint) — float:left/right with wrap-around content
- [x] ~~Implement `display` values~~ (2026-04-03 sprint partial) — display:block/inline/none/flex handled; inline-block/grid not yet implemented
- [x] ~~Implement positioned layout~~ (2026-04-05 sprint) — `position: absolute/relative/fixed/sticky` with `top/left/right/bottom` offsets; sticky treated as relative
- [x] ~~Implement `visibility: hidden` and `display: none`~~ (2026-04-03 sprint) — hidden elements occupy space; display:none removed from layout
- [x] ~~Implement `position: fixed`~~ (2026-04-03 sprint) — viewport-locked positioning (header bars, modals)

## 🟢 Medium (Features)

- [ ] **Scroll support** — mouse wheel / scrollbar navigation through page content
- [ ] **Click interaction** — clicking links should navigate to those URLs
- [ ] **Text selection** — highlight text with mouse
- [ ] **Input/textarea typing** — keyboard input in form fields
- [ ] **DevTools / Inspector** — show DOM tree, computed styles, box model dimensions for any element
- [ ] **Console panel** — show JavaScript console messages from the page
- [ ] **Performance profiling** — show layout/paint timing
- [ ] **Download progress indicator** — show fetch progress for large pages

## 🟢 Medium (Platform)

- [ ] **Wayland support** — currently X11 only via Ebitengine. Wayland compositor support
- [ ] **Headless mode** — generate screenshots without GUI display (for CI)
- [ ] **PNG output** — save rendered output to file (partially working)
- [ ] **PDF output** — render page to PDF document
- [ ] **Window management** — resize, fullscreen, multiple windows, tabs

## 🟡 High (HTML/CSS Coverage)

### Missing CSS properties
- [ ] **Implement CSS multi-column layout fully** — `column-span: all`, `break-before/after: column`, `column-fill: auto/balance`
- [ ] **Implement CSS `border-image`** — draw border using image slices; stretch/repeat/round modes; fallback to border-color when no image
- [ ] **Implement CSS `filter`** — blur(), brightness(), contrast(), grayscale(), sepia(), drop-shadow() filter effects on elements
- [ ] **Implement CSS `backdrop-filter`** — apply filter effects to area behind an element (for dialog/modals with blur backdrop)
- [ ] **Implement CSS `transform`** — rotate(), scale(), translate(), skew(), matrix() 2D transforms on elements
- [ ] **Implement CSS `@media` queries** — responsive breakpoints with media types (screen, print) and feature queries (width, height, orientation)
- [ ] **Implement CSS `aspect-ratio`** — enforce intrinsic aspect ratio on boxes (used heavily for iframe/video elements)
- [ ] **Implement CSS `object-fit`** — cover/contain/fill/scale-down for replaced element sizing (img, video, iframe)
- [ ] **Implement CSS `resize`** — make divs resizable via resize: both/horizontal/vertical
- [ ] **Implement CSS `user-select`** — none/text/all control text selection behavior
- [ ] **Implement `<img>` drawing** — actually fetch and draw image from URL; handle loading/error states with placeholder
- [ ] **Implement `<video>` and `<audio>`** — show video frame or audio controls UI; play/pause/volume (no actual playback, just visual)
- [ ] **Implement `<canvas>` element** — 2D canvas drawing API for web canvas compatibility
- [ ] **Implement `<iframe>`** — render iframe content if same-origin, show placeholder if cross-origin
- [ ] **Implement `<svg>` inline SVG** — parse and render basic SVG elements (rect, circle, path, line, polyline, polygon, text)
- [ ] **Implement `<select>` dropdown** — render as styled box with dropdown arrow; option elements listed in popup
- [ ] **Implement `<textarea>` fully** — multi-line text input with scroll support; resize handle
- [ ] **Implement `<button>` element** — styled button with hover/active states; submit/reset types
- [ ] **Implement `<label>` element** — associate label with form control; clicking label focuses associated input
- [ ] **Implement `<form>` element** — form submission via GET/POST; action/tethod/enctype attributes
- [ ] **Implement form validation UI** — :valid/:invalid styling applied based on input state; required attribute handling
- [ ] **Implement `<noscript>` properly** — show content when JS is disabled; hide when enabled
- [ ] **Implement `<template>`** — parse template content into DocumentFragment but don't render until JS activates it
- [ ] **Implement `<slot>` and Shadow DOM** — basic slot projection for web component support
- [ ] **Implement CSS `::before` and `::after`** — draw pseudo-elements with content property before/after element content
- [ ] **Implement CSS `::first-line` and `::first-letter`** — apply special formatting to first line/letter of text
- [ ] **Implement CSS `:focus-visible`** — keyboard focus ring separate from mouse focus styling
- [ ] **Implement CSS `:is()` and `:where()`** — forgiving matching for selector lists
- [ ] **Implement CSS `@supports`** — conditional CSS based on browser feature support
- [ ] **Implement CSS `counter()` and `counters()`** — automatic numbering with counters for ordered lists/sections
- [ ] **Implement CSS `unicode-range`** — specify character ranges for web font loading
- [ ] **Implement `pointer-events`** — `pointer-events: none` prevents element from receiving pointer events (clicks, hover)
- [x] ~~Implement `letter-spacing`, `word-spacing`, `text-indent`, `text-transform`~~ (2026-04-04 sprint) — DrawText: text-transform applies uppercase/lowercase/capitalize; letter-spacing adds per-char extra; word-spacing adds after spaces; text-indent offsets first line; font-weight affects char width (bold=0.65em, light=0.55em); font-style italic makes chars 10% wider

### Missing CSS Selectors
- [ ] **Implement attribute selectors** — `[attr]`, `[attr=value]`, `[attr~=value]`, `[attr|=value]`
- [ ] **Implement pseudo-classes** — `:hover`, `:focus`, `:active`, `:first-child`, `:last-child`, `:nth-child()`
- [ ] **Implement pseudo-elements** — `::before`, `::after` (with `content` property)

### URL Handling
- [x] ~~Implement `base` href~~ (2026-04-04 sprint 2) — ResolveURL in internal/fetch/url.go handles relative URLs
- [x] ~~Implement absolute URL resolution~~ (2026-04-04 sprint 2) — ResolveURL handles all relative URL forms

### Browser Features
- [ ] **Link click navigation** — clicking `<a href>` elements navigates to those URLs
- [ ] **Page scroll** — mouse wheel / scrollbar navigation through page content
- [x] ~~404 / error page handling~~ (2026-04-04 sprint 2) — displays styled error page on HTTP 4xx/5xx responses

## 🟠 Low (Testing)

- [ ] **html5lib test corpus** — download and run 500+ HTML parsing edge case tests from html5lib project. Compare tokenizer + parser output against reference. Fix failures. This validates correctness against real-world HTML from all browsers
- [ ] **Visual screenshot tests** — collect baseline screenshots of known pages (HN, Reddit, etc). Run rendering pipeline, diff PNGs. Track visual regressions over time
- [ ] **Fuzz testing** — use go-fuzz or custom mutator to generate random HTML/CSS and verify parser/renderer doesn't panic on malformed input
- [ ] **Performance benchmarks** — measure parse time, layout time, paint time for pages of varying complexity. Set budgets and alert on regressions

## 🟠 Low (Maintenance)

- [ ] **Go module tidy and dependency pinning** — ensure reproducible builds
- [ ] **Add Makefile** — `make build`, `make test`, `make clean`, `make install` targets
- [ ] **CI/CD on GitHub Actions** — run tests, build, and deploy on every push
- [ ] **Add CHANGELOG.md** — track version history
- [ ] **Add CONTRIBUTING.md** — guide for new contributors
- [ ] **Document architecture** — write up the rendering pipeline: URL → Fetch → Tokenize → Parse → Layout → Render → Display

## Visual QA Notes (2026-01-03)

### news.ycombinator.com ✅ Usable
- Layout: ranked numbered list (1-30), orange HN banner (Y icon), nav links top-right
- Content: tech news stories — title, domain (parens), points, submitter username, relative time, comment count
- Design: intentionally minimalist, text-focused, high contrast, easy to scan
- Structure: thin metadata row under each title; footer with More button + search
- **Rendering challenge:** simple vertical list — should be easy to render correctly. Text-heavy, no images, no complex layout

### www.reddit.com 🚫 Blocked
- Network-level security block page (corporate/ISP filter)
- Cannot access Reddit without proxy or different network
- **Rendering challenge:** would be very complex — nested comments tree, vote arrows, collapsible threads, heavy JS

### x.com (Twitter) 🔒 Login-gated
- Shows login/signup page for unauthenticated users
- Cannot access the feed without being logged in
- **Rendering challenge:** social media feed — would be complex but achievable. Tweet cards, threaded replies, compose box

### www.yahoo.com ⚠️ Partial (screenshot too large for vision)
- Title: "Yahoo | Mail, Weather, Search, Politics, News, Finance, Sports & Videos"
- Layout (from accessibility tree): skip links, Yahoo logo + search bar, nav tabs (News, Finance, Sports, More, Mail, Sign in)
- Features bar: horoscope dropdown, events near me, What to Watch, Today in History, Game of the Day, NCAAW score
- Trending section: Tiger Woods, Trump executive order, Bondi replacement, South Carolina vs UConn, US fighter jet
- Major Markets: S&P 500 widget with live price
- **Rendering challenge:** extremely complex — portal with 200+ elements, multiple columns, embedded widgets, ads, dynamic content. This is the hardest test case

## Realistic Testing Approach

### html5lib Test Corpus
The html5lib Python project has comprehensive HTML parsing tests:
- URL: https://github.com/html5lib/html5lib-tests
- Tests cover: tokenizer edge cases, DOM construction, tree building
- Each test is a JSON file with `input`, `errors`, and `output` (expected DOM)
- Strategy: fork/copy test files, write Go test harness that runs same inputs through tokenizer+parser, compare against expected output
- This is how Go's own `net/html` package validates

### Visual Screenshot Testing
1. Collect ~10 known websites of varying complexity
2. Take ground-truth screenshots using a real browser
3. Run same URLs through ViBrowsing's rendering pipeline
4. Diff the two images (pixel-level or perceptual hash)
5. Report regressions

### Realistic Site Coverage
- **news.ycombinator.com** — simple, text-heavy, ranked list
- **reddit.com** — complex layout, nested comments, vote buttons
- **x.com** — dynamic content, social media feed
- **yahoo.com** — portal site, news, ads, heavy media
- **wikipedia.org** — structured content, tables, infoboxes, references
- **stackoverflow.com** — code blocks, syntax highlighting, Q&A layout
- **github.com** — repo UI, markdown, file trees
- **amazon.com** — e-commerce, product listings, grids, filters

---

## 🟡 Medium (Performance)

- [ ] **Benchmark parsing speed** — measure tokens/second for pages of varying size (1KB, 10KB, 100KB, 1MB). Set baseline and alert on regressions
- [ ] **Benchmark layout speed** — measure layout pass time for complex DOM trees. Identify bottlenecks
- [ ] **Benchmark render speed** — measure pixel output rate (pixels/second) for full-page renders
- [ ] **Optimize tokenizer** — avoid repeated string comparisons in hot path. Use bytes.HasPrefix instead of string matching where possible
- [ ] **Optimize layout tree walks** — reduce repeated parent/child traversal during box tree construction
- [ ] **Cache computed styles** — avoid re-computing inherited properties on every element. Build cascade once, reuse
- [ ] **Parallelize independent subtrees** — if DOM has multiple independent branches, layout/render them concurrently (goroutines)
- [ ] **Lazy load images** — don't decode image data until it's about to be rendered to screen
- [ ] **Incremental rendering** — for long documents, render the visible viewport first, then background sections
- [ ] **Memory pool for nodes** — reuse allocated Node/Token objects instead of GC-heavy allocation per parse

## 🟠 Low (Developer Experience)

- [x] ~~Add verbose/debug logging flag~~ (2026-04-03 sprint) — `browser --debug` enables verbose output during fetch/parse/render
- [x] ~~Add `--profile` flag~~ — output timing profile (CPU/memory) for parse + layout + render phases
- [x] ~~Add `--dump-dom` flag~~ (2026-04-03 sprint) — added --dump-dom flag to browser CLI
- [x] ~~Add `--dump-layout` flag~~ (2026-04-03 sprint) — added --dump-layout flag with Box.String() method on layout tree
- [ ] **Add `--benchmark` flag** — run parse+layout+render N times and print timing stats
- [x] ~~Add `--viewport` flag~~ (2026-04-03 sprint) — added --viewport WxH flag (e.g. 375x667)
- [x] ~~Add `--user-agent` flag~~ (2026-04-03 sprint) — added --user-agent flag to set HTTP User-Agent header
- [ ] **Colorize terminal output** — use ANSI colors for DOM/tree dumps in debug mode
- [ ] **TUI devtools panel** — ncurses-based panel alongside browser showing DOM tree, style computed values, network requests
- [x] ~~Show file:// URL support~~ (2026-04-03 sprint) — browser already handles local file paths, auto-prefixes with file://

## 🟡 Medium (Error Handling & Robustness)

- [ ] **Handle malformed URLs gracefully** — show error page instead of panic on bad URL
- [x] ~~Handle fetch timeouts~~ (2026-04-03 sprint) — fetch.Fetch() now uses configurable timeout via HTTP client; default 30s, wired to --user-agent flag
- [x] ~~Handle HTTP errors~~ (2026-04-03 sprint) — HTTP error codes now print error and exit cleanly
- [ ] **Handle binary/non-text content** — if server returns image/binary for HTML content-type, don't try to parse as HTML
- [ ] **Handle very large pages** — pages > 10MB should be truncated or streaming-parsed, not loaded entirely into memory
- [ ] **Handle deeply nested DOM** — pages with >10,000 levels of nesting shouldn't stack overflow in recursive layout
- [ ] **Handle extremely long lines in HTML** — a single line with 10MB of text should not cause memory issues
- [ ] **Handle missing/invalid CSS** — malformed CSS declarations should be skipped, not crash the cascade
- [ ] **Handle circular CSS references** — `width: 50%` of parent where parent width depends on child should not infinite loop

## 🟠 Low (Accessibility)

- [ ] **Implement ARIA roles** — `role="button"`, `role="navigation"`, etc. affect rendering semantics
- [ ] **Implement `<summary>` and `<details>`** — collapsible disclosure widget (toggle visibility of summary content)
- [ ] **Implement `<dialog>` and `<form>`** — modal dialog element
- [ ] **Implement `<fieldset>` and `<legend>`** — form grouping with border and label
- [ ] **Implement `<meter>` and `<progress>`** — gauge and progress bar elements
- [ ] **Implement `<time>`** — machine-readable date/time element
- [ ] **Implement `<abbr>`** — abbreviation with tooltip for full text
- [ ] **Implement `<mark>`** — highlighted/marked text styling
- [ ] **Implement `<ruby>`** — ruby annotation for East Asian typography (ruby text above/below base text)
- [ ] **Implement `<bdi>` and `<bdo>`** — bidirectional text isolation and override

## 🟡 Medium (Content & Rendering Quality)

- [x] ~~Implement `<img>` actual rendering~~ (2026-04-04 sprint) — ImageBox type added; DrawImage shows alt text / broken image icon; img elements get 150x150 default size; image scaling infrastructure ready (loadImage is stub)
- [ ] **Implement CSS `background-image`** — background images on elements (URL-based)
- [ ] **Implement CSS gradients** — `linear-gradient()`, `radial-gradient()` as background-image values
- [ ] **Implement CSS `clip-path`** — masking shapes on elements
- [ ] **Implement `<video>` and `<audio>`** — show video player frame or audio player with controls UI
- [ ] **Implement `<canvas>`** — render canvas 2D context content to output
- [ ] **Implement `<iframe>`** — for embedded content, show placeholder or recursively render same-origin iframes
- [ ] **Implement emoji rendering** — proper emoji character display (these are complex Unicode, may need a library)
- [ ] **Implement symbol rendering** — `&copy;`, `&reg;`, `&trade;`, `&mdash;`, `&ndash;`, `&hellip;`, `&nbsp;` named entities
- [ ] **Implement `calc()` in CSS** — `width: calc(100% - 20px)` support in CSS value parsing

## 🟡 High (CSS Text & Typography)

- [ ] **Implement CSS `font-size`** — absolute sizes (px, pt, em, rem), relative sizes (larger, smaller), keywords (small, medium, large, xx-large)
- [ ] **Implement CSS `font-family`** — serif, sans-serif, monospace, cursive, fantasy, and generic fallback chain
- [ ] **Implement CSS `letter-spacing`** — tracking between characters
- [ ] **Implement CSS `word-spacing`** — spacing between words
- [ ] **Implement CSS `text-indent`** — first-line indentation
- [ ] **Implement CSS `text-transform`** — uppercase, lowercase, capitalize
- [ ] **Implement CSS `text-shadow`** — text shadow effects
- [ ] **Implement CSS `font-variant`** — small-caps, ligatures
- [ ] **Implement CSS `quotes`** — custom quote characters for `<q>` elements
- [ ] **Implement CSS `counter-increment` and `counter-reset`** — automatic numbering for lists/headings
- [ ] **Implement CSS `direction`** — ltr vs rtl (for Arabic, Hebrew pages)
- [ ] **Implement CSS `unicode-bidi`** — bidirectional text embedding levels
- [ ] **Implement CSS `writing-mode`** — horizontal-tb, vertical-rl, vertical-lr
- [ ] **Implement CSS `tab-size`** — tab character rendering width

## 🟡 High (URL & Navigation)

- [ ] **Implement `<base href>` support** — resolve relative URLs against base tag in document head
- [ ] **Implement proper URL resolution** — absolute vs relative URL handling (scheme, host, path, query, fragment)
- [ ] **Implement HTTP redirects** — follow 301/302/303/307/308 redirects with proper URL updating
- [ ] **Implement HTTP cookies** — send cookies on subsequent requests to same origin
- [ ] **Implement HTTP Referer header** — send Referer on navigation
- [ ] **Implement browser history** — back/forward navigation between visited URLs
- [ ] **Implement link target resolution** — `<a target="_blank">` opens in new tab (or same tab if not supported)

## 🟡 Medium (Window & UI)

- [ ] **Implement window title** — render document `<title>` in window title bar
- [ ] **Implement favicon** — fetch and display favicon.ico in window
- [ ] **Implement right-click context menu** — copy link, copy text, open in new tab options
- [ ] **Implement address bar** — show current URL in a text field at top
- [ ] **Implement reload/stop buttons** — toolbar with reload, stop, back, forward buttons
- [ ] **Implement loading indicator** — spinner/progress bar during page fetch
- [ ] **Implement find-in-page** — Ctrl+F to search for text in rendered page
- [ ] **Implement zoom** — Ctrl+/Ctrl- for page zoom (CSS transforms or viewport scaling)
- [ ] **Implement focus ring** — visible focus indicator on interactive elements for keyboard navigation

---

## 🟠 Low (Networking & Protocol)

- [ ] **Implement HTTP/1.1 keep-alive** — reuse TCP connections for multiple requests to same origin
- [ ] **Implement HTTP/2 support** — upgrade to HTTP/2 for multiplexed requests
- [ ] **Implement TLS certificate verification** — proper HTTPS with certificate validation
- [ ] **Implement DNS resolution caching** — cache resolved IPs to avoid repeated DNS lookups
- [ ] **Implement connection timeout** — max time to establish TCP connection
- [ ] **Implement read/write timeouts** — prevent hanging on slow connections
- [ ] **Implement retry on connection reset** — automatically retry on transient failures
- [ ] **Implement conditional GET (If-Modified-Since)** — send Last-Modified header, handle 304 Not Modified
- [ ] **Implement Content-Encoding** — handle gzip/deflate/br content encoding from servers
- [ ] **Implement streaming fetch** — for large pages, stream HTML as it's received rather than buffering all

## 🟠 Low (Internationalization & i18n)

- [x] ~~Implement UTF-8 charset detection~~ (2026-04-03 sprint) — DetectCharset() checks UTF-8 BOM, <meta charset="">, and <meta http-equiv="Content-Type" content="...charset=...">; defaults to utf-8
- [x] ~~Implement `<meta http-equiv="Content-Type">` charset~~ (2026-04-03 sprint) — covered by DetectCharset()
- [x] ~~Implement `<meta charset="UTF-8">`~~ (2026-04-03 sprint) — covered by DetectCharset()
- [ ] **Implement CSS `lang` attribute selector** — `:lang(en)` pseudo-class
- [ ] **Implement HTML `lang` attribute** — `<html lang="en">` for accessibility
- [ ] **Implement `<bdo dir="rtl">`** — right-to-left text override
- [ ] **Implement emoji rendering** — proper emoji display (color emoji fonts)
- [ ] **Implement `Accept-Language` header** — send preferred languages to servers
- [ ] **Implement number formatting per locale** — for Arabic/Hindic numerials

## 🟠 Low (Print & Export)

- [ ] **Implement `@media print`** — apply print-specific stylesheet rules
- [ ] **Implement print styles** — hide navigation, expand hidden sections, optimize for paper
- [ ] **Implement PDF export** — render page to PDF using go's pdf libraries or command-line tools
- [ ] **Implement SVG export** — save rendered output as SVG vector format
- [ ] **Implement screenshot of specific element** — `dom.toImage()` or screenshot a div
- [ ] **Implement `window.print()`** — trigger print dialog with current page

## 🟡 Medium (Additional CSS Features)

- [ ] **Implement CSS `clip`** — legacy clipping (replaced by clip-path)
- [ ] **Implement CSS `clip-path: polygon()`** — complex polygon clipping shapes
- [ ] **Implement CSS `mask-image`** — image masking
- [ ] **Implement CSS `backdrop-filter`** — blur behind fixed-position elements
- [ ] **Implement CSS `filter`** — blur, brightness, contrast, grayscale, sepia on elements
- [ ] **Implement CSS `object-fit`** — how img/video fill their container
- [ ] **Implement CSS `object-position`** — positioning of replaced content
- [ ] **Implement CSS `aspect-ratio`** — forced aspect ratio on boxes
- [ ] **Implement CSS `column-width` and `column-count`** — multi-column layout
- [ ] **Implement CSS `column-gap`, `column-rule`** — column spacing and dividers
- [ ] **Implement CSS `break-inside`, `break-before`, `break-after`** — pagination control
- [ ] **Implement CSS `page-break-*`** — print pagination

## 🟡 Medium (Advanced DOM APIs)

- [ ] **Implement `querySelector()`** — CSS selector-based element lookup
- [ ] **Implement `querySelectorAll()`** — return all matching elements
- [ ] **Implement `getElementById()`** — fast ID-based lookup with index
- [ ] **Implement `getElementsByClassName()`** — class-based element collection
- [ ] **Implement `getElementsByTagName()`** — tag-based element collection
- [ ] **Implement `innerHTML`** — get/set inner HTML of elements
- [ ] **Implement `outerHTML`** — get/set outer HTML of elements
- [ ] **Implement `textContent`** — get/set text content of elements
- [ ] **Implement `innerText`** — get/set rendered text (like textContent but CSS-aware)
- [ ] **Implement `createElement()`** — DOM API for creating elements
- [ ] **Implement `createTextNode()`** — DOM API for creating text nodes
- [ ] **Implement `appendChild()`** — DOM API (may already exist)
- [ ] **Implement `removeChild()`** — DOM API to remove nodes
- [ ] **Implement `insertBefore()`** — DOM API to insert before reference node
- [ ] **Implement `classList` API** — add/remove/toggle/contains CSS classes
- [ ] **Implement `getAttribute()` / `setAttribute()`** — attribute access
- [ ] **Implement `style` property** — inline style get/set
- [ ] **Implement `dataset` property** — `data-*` attribute access

## 🟠 Low (Testing & QA)

- [ ] **Property-based fuzzing** — use go-fuzz to generate random HTML/CSS combinations
- [ ] **Regression test suite** — save known-good outputs for each sample page, diff on change
- [ ] **Parse error recovery tests** — malformed HTML should not crash, should produce best-effort DOM
- [ ] **Unicode boundary tests** — emoji, combining characters, RTL, surrogate pairs
- [ ] **Very large document test** — 10MB+ HTML file should parse without OOM or timeout
- [ ] **Deeply nested document test** — 10,000 levels of nesting should not stack overflow
- [ ] **Memory leak tests** — run parse 1000 times, ensure memory doesn't grow unbounded
- [ ] **Performance regression CI** — fail build if parse+layout time increases >10% vs baseline

## 🟡 Medium (Code Quality)

- [ ] **Extract CSS parser into own package** — `internal/css/parser.go` from layout
- [ ] **Extract layout engine into own package** — `internal/layout/box.go` from render
- [ ] **Add package-level documentation** — godoc for each internal package
- [ ] **Add inline comments for complex algorithms** — foster parenting, float algorithm, etc
- [ ] **Add benchmarking to `html.Parse()`** — measure and log parse time
- [ ] **Add benchmarking to layout** — measure box tree construction time
- [ ] **Profile with pprof** — identify CPU and memory bottlenecks
- [ ] **Reduce string allocations in tokenizer** — use []byte/[]rune pooling
- [ ] **Use sync.Pool for Node allocation** — reduce GC pressure in hot path
- [ ] **Add error type hierarchy** — `ParseError`, `FetchError`, `LayoutError` with stack traces

## 🟠 Low (Canvas/Drawing Improvements)

- [ ] **Implement `DrawRoundedRect` for background with border-radius** — currently border-radius draws square corners; bg must clip to rounded shape
- [ ] **Implement actual `background-image` fetching and decoding** — download image URL, decode JPEG/PNG/GIF/WebP, draw to canvas
- [ ] **Implement `background-repeat` drawing** — tile background image across element (repeat, repeat-x, repeat-y, space, round)
- [ ] **Implement `background-position` drawing** — offset background image from element edge
- [ ] **Implement `background-size` drawing** — scale background image to specified dimensions (cover, contain, explicit W H)
- [ ] **Implement `linear-gradient()` parsing and drawing** — draw gradient from color-stop list
- [ ] **Implement `radial-gradient()` parsing and drawing** — draw radial gradient with center/radius parameters
- [ ] **Implement `box-shadow` multiple shadows** — `box-shadow` can have multiple comma-separated shadows
- [ ] **Implement `text-shadow` multiple shadows** — multiple shadows on text via comma separation
- [ ] **Implement `outline` drawing** — currently outline is stored but never drawn; draw outside border box
- [ ] **Implement `opacity` on individual draw calls** — apply alpha blending per element not just whole box

## 🟠 Low (Form Elements)

- [ ] **Implement `<button>` visual rendering** — styled button with border, padding, text
- [ ] **Implement `<select>` dropdown (visual)** — show select as a bordered box with current option text
- [ ] **Implement `<textarea>` visual rendering** — bordered multiline text area
- [ ] **Implement `<label>` association** — label visually linked to associated form element
- [ ] **Implement form focus styling** — `:focus` pseudo-class on inputs/buttons shows outline

## 🟠 Low (Advanced CSS)

- [ ] **Implement CSS `calc()` function** — `width: calc(100% - 20px)` in CSS value parsing and layout
- [ ] **Implement CSS `clamp()` function** — `width: clamp(100px, 50%, 300px)` clamping values
- [ ] **Implement CSS `min()` and `max()` functions** — `width: min(100px, 50%)`
- [ ] **Implement CSS `counter()` and `counters()`** — automatic numbering for lists and headings
- [ ] **Implement CSS `attr()` function** — `content: attr(data-label)` reading attribute values

## 🟠 Low (HTML Elements)

- [ ] **Implement `<colgroup>` and `<col>`** — column grouping for table column widths
- [ ] **Implement `<thead>`, `<tbody>`, `<tfoot>` table sections** — proper table section rendering order
- [ ] **Implement `<td colspan>` and `<td rowspan>`** — cell spanning for complex tables
- [ ] **Implement `<figure>` and `<figcaption>`** — figure with caption rendered below/above
- [ ] **Implement `<dialog>` modal** — dialog element with backdrop
- [ ] **Implement `<slot>` and shadow DOM** — web component slot projection

## 🟠 Low (Performance)

- [ ] **Lazy image decoding** — don't decode image data until visible in viewport
- [ ] **Streaming HTML parse** — for large pages, parse HTML incrementally without buffering all
- [ ] **Incremental layout** — layout visible viewport first, then off-screen content
- [ ] **CSS selector indexing** — build index of elements by class/id/tag for fast selector matching
- [ ] **Text measurement caching** — cache Ebitengine text measurement results per font/size/text combo

---

## 🆕 New Items (2026-04-05 Sprint)

### 🟡 High Priority

- [x] ~~Implement CSS `resize` property~~ (2026-04-04 sprint 2) — resize property stored in style defaults
- **Implement `<details>` and `<summary>` toggle** — disclosure widget, click summary to toggle content visibility
- [x] ~~Implement CSS `pointer-events`~~ (2026-04-04 sprint 2) — pointer-events stored in style defaults
- **Implement `<picture>` element** — responsive images with `<source srcset="...">` fallback chain
- [x] ~~Implement CSS `will-change` hint~~ (2026-04-04 sprint 2) — will-change stored in style defaults
- [x] ~~Implement CSS `image-rendering`~~ (2026-04-04 sprint 2) — image-rendering stored in style defaults
- **Implement `window.scrollTo()` and `window.scrollBy()`** — programmatic scroll APIs
- **Implement `element.scrollIntoView()`** — scroll element into viewport (start/center/end/nearest)
- [x] ~~Implement CSS `scroll-behavior`~~ (2026-04-04 sprint 2) — scroll-behavior stored in style defaults

### 🟢 Medium Priority

- **Implement `window.innerWidth`, `window.innerHeight`** — viewport dimensions accessible via JavaScript
- **Implement `navigator.userAgent`** — expose user agent string via navigator object
- [x] ~~Implement CSS `overscroll-behavior`~~ (2026-04-04 sprint 2) — overscroll-behavior stored in style defaults
- [x] ~~Implement `<del>` and `<ins>` elements~~ (2026-04-04 sprint 2) — del renders strikethrough, ins renders underline
- [x] ~~Implement CSS `text-decoration-line/color/style`~~ (2026-04-04 sprint 2) — individual properties parsed; text-decoration shorthand parses all three; rendering in DrawText
- [x] ~~Implement `<output>` element~~ (2026-04-04 sprint 2) — output element handled as inline box
- **Implement CSS `font-stretch`** — narrow/wider font variants (ultra-condensed to ultra-expanded)
- **Implement CSS `place-items` and `place-self`** — alignment shorthands for grid/flex
- [x] ~~Implement CSS `mix-blend-mode`~~ (2026-04-04 sprint 2) — mix-blend-mode stored in style defaults
- [x] ~~Implement CSS `hanging-punctuation`~~ (2026-04-04 sprint 2) — hanging-punctuation stored in style defaults
- [x] ~~Implement CSS `contain` property~~ (2026-04-04 sprint 2) — contain stored in style defaults
- **Implement `CSS.supports()` API** — programmatic @supports check: `CSS.supports('display', 'grid')`
- **Implement `window.matchMedia()`** — MediaQueryList interface for JavaScript media query checks
- **Implement CSS `transform-box`** — `transform-box: view-box/content-box/fill-box` for transform reference box

### 🟠 Low Priority

- **Implement `<map>` and `<area>`** — client-side image maps with clickable regions (legacy but still used)
- **Implement `IntersectionObserver`** — viewport visibility detection for lazy loading and scroll tracking
- **Implement `ResizeObserver`** — element resize detection for responsive components
- **Implement `ime-mode` CSS property** — input method editor hints for East Asian input
- **Implement `<data>` element** — machine-readable data wrapper with `value=""` attribute


## 🆕 New Items (2026-04-04 Sprint)

### ✅ Completed (2026-04-04)

**CSS Functions & Values:**
- [x] ~~Implement CSS `calc()` function~~ — Parse `calc(100% - 20px)` with +, -, *, / operators
- [x] ~~Implement CSS `clamp()` function~~ — Parse `clamp(min, preferred, max)`
- [x] ~~Implement CSS `min()` and `max()` functions~~ — Parse `min(val1, val2)` and `max(val1, val2)`
- [x] ~~Implement CSS `counter()` and `counters()`~~ — Automatic numbering for lists
- [x] ~~Implement CSS `attr()` function~~ — `attr(data-xxx)` for reading attribute values
- [x] ~~Implement CSS `aspect-ratio`~~ — Forced aspect ratio on boxes
- [x] ~~Implement CSS `object-fit` and `object-position`~~ — Image sizing within containers
- [x] ~~Implement CSS `filter` parsing~~ — blur, brightness, contrast, grayscale, sepia, hue-rotate, drop-shadow
- [x] ~~Implement CSS `backdrop-filter`~~ — Same syntax as filter
- [x] ~~Implement CSS `clip-path` parsing~~ — inset, circle, ellipse, polygon shapes
- [x] ~~Implement CSS `column-*` properties~~ — column-width, column-count, column-gap, column-rule, break-*
- [x] ~~Implement CSS `@keyframes` and `animation`~~ — Keyframe rule parsing and animation shorthand
- [x] ~~Implement CSS `transition`~~ — transition-property, duration, timing-function, delay

**DOM APIs:**
- [x] ~~Implement `GetElementById()`~~ — Fast ID-based element lookup
- [x] ~~Implement `GetElementsByClassName()`~~ — Class-based element collection
- [x] ~~Implement `GetElementsByTagName()`~~ — Tag-based element collection
- [x] ~~Implement `QuerySelector()` and `QuerySelectorAll()`~~ — CSS selector-based element lookup
- [x] ~~Implement `classList` API~~ — add(), remove(), toggle(), contains(), replace(), item(), length
- [x] ~~Implement `textContent` and `innerHTML`/`outerHTML`~~ — Content getter/setter APIs
- [x] ~~Implement `createElement()` and `createTextNode()`~~ — DOM node creation
- [x] ~~Implement `appendChild()`, `removeChild()`, `insertBefore()`, `replaceChild()`~~ — DOM manipulation
- [x] ~~Implement `cloneNode()` (shallow and deep)~~ — Node cloning
- [x] ~~Implement `getAttribute()`, `setAttribute()`, `removeAttribute()`, `hasAttribute()`~~ — Attribute API
- [x] ~~Implement `dataset` property~~ — data-* attribute access
- [x] ~~Implement `style` property (get/set inline styles)~~ — Via SetAttribute("style", ...)

**HTML Elements:**
- [x] ~~Implement `<slot>` element~~ — Web component slot (display: contents)
- [x] ~~Implement `<dialog>` element~~ — Modal dialog with backdrop
- [x] ~~Implement `<meter>` element~~ — Gauge with value coloring
- [x] ~~Implement `<progress>` element~~ — Progress bar
- [x] ~~Implement `<mark>` element~~ — Highlighted text (yellow background)
- [x] ~~Implement `<ruby>`, `<rt>`, `<rp>` elements~~ — Ruby annotation
- [x] ~~Implement `<bdi>` and `<bdo>` elements~~ — Bidirectional text

**CSS Properties:**
- [x] ~~Implement `clip` (legacy)~~ — rect() clipping
- [x] ~~Implement `flex-flow`, `order`, `align-content`~~ — Additional flexbox properties
- [x] ~~Implement `font-variant`, `unicode-bidi`, `direction`, `writing-mode`~~ — Typography & i18n
- [x] ~~Implement `tab-size`, `quotes`~~ — Text formatting properties

---

### 🔴 Critical (Parser/Rendering)
- [ ] **Fix inline box baseline calculation** — inline text boxes should share a common baseline; vertical-align: middle/bottom should position relative to that baseline
- [ ] **Fix table cell collapsing borders** — adjacent table cells should share borders (border-collapse behavior), currently each cell renders its own border
- [ ] **Fix float clearing** — blocks that clear:left/right/both should properly position below the float, not overlap it
- [ ] **Implement CSS transform on inline boxes** — transforms on inline elements should create a transform box and not affect line layout
- [ ] **Fix `:nth-child()` selector** — Complex `an+b` formulas (even, odd, 2n+1, 3n-1) don't work

### 🟡 High (Layout/Rendering)
- [ ] **Implement CSS `filter` drawing** — Actually apply blur, brightness, contrast, grayscale, sepia effects to rendered pixels
- [ ] **Implement CSS `backdrop-filter` drawing** — Apply blur to elements behind fixed/absolute positioned elements
- [ ] **Implement CSS `clip-path` drawing** — Clip element rendering to inset/circle/polygon shapes
- [ ] **Implement `<iframe>` rendering** — Show placeholder for iframes (recursive rendering is complex)
- [ ] **Implement `<canvas>` 2D context** — Render canvas drawing commands to output
- [ ] **Implement emoji rendering** — Proper color emoji display (may need fontconfig)
- [ ] **Implement CSS `clip-path: polygon()` drawing** — Fill polygons for clip-path

### 🟡 High (CSS Selectors & Cascade)
- [ ] **Implement `:not()` pseudo-class** — Negation selector with complex selectors inside
- [ ] **Implement `:checked`, `:disabled`, `:enabled` pseudo-classes** — For form state matching
- [ ] **Implement `::before` and `::after` pseudo-elements** — With `content` property
- [ ] **Implement `:first-line` and `::first-letter`** — Text segment pseudo-elements
- [ ] **Implement CSS combinators fully** — Descendant (space), child (>), adjacent sibling (+), general sibling (~)
- [ ] **Implement `:lang()` pseudo-class** — Language-based selector

### 🟡 High (URL & Navigation)
- [ ] **Implement `<base>` href support** — Resolve relative URLs against base tag
- [ ] **Implement HTTP cookies** — Send cookies on subsequent requests to same origin
- [ ] **Implement browser history** — Back/forward navigation between visited URLs
- [ ] **Implement link target resolution** — `<a target="_blank">` opens in new tab

### 🟡 High (Platform)
- [ ] **Implement scroll support** — Mouse wheel / scrollbar navigation through page content
- [ ] **Implement click interaction** — Clicking links should navigate to those URLs
- [ ] **Implement text selection** — Highlight text with mouse
- [ ] **Implement window title** — Render document `<title>` in window title bar
- [ ] **Implement favicon** — Fetch and display favicon.ico in window

### 🟢 Medium (Performance)
- [ ] **Lazy image decoding** — Don't decode images until visible in viewport
- [ ] **CSS selector indexing** — Build index of elements by class/id/tag for fast selector matching
- [ ] **Text measurement caching** — Cache Ebitengine text measurement results per font/size/text combo
- [ ] **Streaming HTML parse** — For large pages, parse HTML incrementally without buffering all
- [ ] **Memory pool for nodes** — Reuse allocated Node/Token objects instead of GC-heavy allocation per parse

### 🟢 Medium (Content & Rendering)
- [ ] **Implement `<abbr>` with title tooltip** — Abbreviation with full text on hover
- [ ] **Implement `<ruby>` layout** — Ruby text above/below base text for East Asian typography
- [ ] **Implement `calc()` evaluation in layout** — Actually compute `width: calc(100% - 20px)` during layout
- [ ] **Implement CSS `@media` query matching** — Apply rules only when viewport matches
- [ ] **Implement CSS `cursor`** — Show appropriate cursor on interactive elements

### 🟠 Low (Testing & QA)
- [ ] **html5lib test corpus** — Download and run 500+ HTML parsing edge case tests
- [ ] **Visual screenshot tests** — Collect baseline screenshots, diff on changes
- [ ] **Fuzz testing** — Use go-fuzz to generate random HTML/CSS and verify no panics
- [ ] **Performance benchmarks** — Measure parse time, layout time, paint time for pages of varying size

### 🟠 Low (Canvas/Drawing)
- [ ] **Implement `background-repeat: space/round`** — Tile backgrounds with spacing or scaling
- [ ] **Implement `background-position` (percentages)** — Offset background image by percentage
- [ ] **Implement `background-size: cover/contain`** — Scale background to fill or fit
- [ ] **Implement `text-shadow` multiple shadows** — Multiple comma-separated shadows on text
- [ ] **Implement `box-shadow` multiple shadows** — Multiple comma-separated drop shadows
- [ ] **Implement `opacity` per draw call** — Apply alpha blending per element not just whole box

---

## 🆕 New Items (2026-04-05 Sprint 2)

### 🟡 High (Layout/Rendering)

- [ ] **Implement CSS `clip-path` drawing** — Actually clip element rendering to inset/circle/polygon shapes
- [ ] **Implement CSS `filter` drawing** — Actually apply blur, brightness, contrast, grayscale, sepia effects to rendered pixels
- [ ] **Implement CSS `backdrop-filter` drawing** — Apply blur to elements behind fixed/absolute positioned elements
- [ ] **Implement `<iframe>` rendering** — Show placeholder for iframes (recursive rendering is complex)
- [ ] **Implement `<canvas>` 2D context** — Render canvas drawing commands to output
- [ ] **Implement emoji rendering** — Proper color emoji display (may need fontconfig)
- [ ] **Fix inline box baseline calculation** — inline text boxes should share a common baseline; vertical-align: middle/bottom should position relative to that baseline
- [ ] **Fix table cell collapsing borders** — adjacent table cells should share borders (border-collapse behavior), currently each cell renders its own border
- [ ] **Fix float clearing** — blocks that clear:left/right/both should properly position below the float, not overlap it
- [ ] **Fix `:nth-child()` selector** — Complex `an+b` formulas (even, odd, 2n+1, 3n-1) don't work

### 🟡 High (CSS Selectors & Cascade)

- [ ] **Implement `:not()` pseudo-class** — Negation selector with complex selectors inside
- [ ] **Implement `:checked`, `:disabled`, `:enabled` pseudo-classes** — For form state matching
- [ ] **Implement `::before` and `::after` pseudo-elements** — With `content` property
- [ ] **Implement `:first-line` and `::first-letter`** — Text segment pseudo-elements
- [ ] **Implement CSS combinators fully** — Descendant (space), child (>), adjacent sibling (+), general sibling (~) with ancestor/sibling traversal

### 🟡 High (URL & Navigation)

- [ ] **Implement HTTP cookies** — Send cookies on subsequent requests to same origin
- [ ] **Implement browser history** — Back/forward navigation between visited URLs
- [ ] **Implement link target resolution** — `<a target="_blank">` opens in new tab

### 🟡 High (Platform)

- [ ] **Implement scroll support** — Mouse wheel / scrollbar navigation through page content
- [ ] **Implement click interaction** — Clicking links should navigate to those URLs
- [ ] **Implement text selection** — Highlight text with mouse
- [ ] **Implement window title** — Render document `<title>` in window title bar
- [ ] **Implement favicon** — Fetch and display favicon.ico in window

### 🟢 Medium (Performance)

- [ ] **Lazy image decoding** — Don't decode images until visible in viewport
- [ ] **CSS selector indexing** — Build index of elements by class/id/tag for fast selector matching
- [ ] **Text measurement caching** — Cache Ebitengine text measurement results per font/size/text combo
- [ ] **Streaming HTML parse** — For large pages, parse HTML incrementally without buffering all
- [ ] **Memory pool for nodes** — Reuse allocated Node/Token objects instead of GC-heavy allocation per parse

### 🟢 Medium (Content & Rendering)

- [ ] **Implement `<abbr>` with title tooltip** — Abbreviation with full text on hover
- [ ] **Implement `<ruby>` layout** — Ruby text above/below base text for East Asian typography
- [ ] **Implement `calc()` evaluation in layout** — Actually compute `width: calc(100% - 20px)` during layout
- [ ] **Implement CSS `@media` query matching** — Apply rules only when viewport matches
- [ ] **Implement CSS `cursor`** — Show appropriate cursor on interactive elements

### 🟠 Low (Testing & QA)

- [ ] **html5lib test corpus** — Download and run 500+ HTML parsing edge case tests
- [ ] **Visual screenshot tests** — Collect baseline screenshots, diff on changes
- [ ] **Fuzz testing** — Use go-fuzz to generate random HTML/CSS and verify no panics
- [ ] **Performance benchmarks** — Measure parse time, layout time, paint time for pages of varying size

### 🟠 Low (Canvas/Drawing)

- [ ] **Implement `background-repeat: space/round`** — Tile backgrounds with spacing or scaling
- [ ] **Implement `background-position` (percentages)** — Offset background image by percentage
- [ ] **Implement `background-size: cover/contain`** — Scale background to fill or fit
- [ ] **Implement `text-shadow` multiple shadows** — Multiple comma-separated shadows on text
- [ ] **Implement `box-shadow` multiple shadows** — Multiple comma-separated drop shadows
- [ ] **Implement `opacity` per draw call** — Apply alpha blending per element not just whole box
