# ViBrowsing Backlog

## 🔴 Critical (Parser/Rendering)

- [x] ~~Fix HTML parser double-html/body bug~~ (2026-04-03 sprint) — tokenizer now skips html/head/body StartTag/EndTag tokens; parser bootstraps them once cleanly; no more duplication
- [x] ~~Fix foster parenting~~ (2026-04-03 sprint) — parser now tracks table context, text inside tables is fostered to parent; implicit <p> close before block elements; table end tags properly close the table context
- [x] ~~Fix unclosed tag handling~~ (2026-04-03 sprint partial) — generic end tag now pops stack until matching tag found; unknown end tags are skipped without crashing; block tags implicitly close open <p> tags
- [x] ~~Implement entity decoding~~ (2026-04-03 sprint) — added decodeEntities() with named entities (amp, lt, gt, quot, apos, nbsp, ndash, mdash, lsquo, rsquo, ldquo, rdquo, hellip, copy, reg, trade, deg, plusmn, times, divide, frac12, frac14, frac34) and numeric entities (&#65;, &#x41;)
- [ ] **Implement foreign content handling** — `<svg>` and `<math>` have special nested tokenization rules

## 🟡 High (Layout/Rendering)

- [ ] **Implement CSS box model properly** — margin, border, padding, content areas need pixel-accurate sizing. Currently may be conflating these
- [ ] **Implement flexbox layout** — HN stories and most modern UIs use flexbox. Need to support `display: flex`, `flex-direction`, `justify-content`, `align-items`, `gap`
- [ ] **Implement inline layout** — text wraps within block containers, `<span>` flows with text, `white-space: pre` preserves whitespace
- [ ] **Implement float** — `float: left/right` removes element from flow, surrounding content wraps around it
- [ ] **Implement positioned layout** — `position: absolute/relative/fixed` with `top/left/right/bottom` offsets
- [ ] **Implement z-index stacking** — elements stack in layers, `z-index` controls stacking order
- [ ] **Implement `visibility: hidden` and `display: none`** — hidden elements occupy no space; none elements are removed from layout entirely
- [ ] **Implement overflow handling** — `overflow: hidden/scroll/auto` clips or adds scrollbars to content

## 🟡 High (HTML/CSS Coverage)

### HTML Elements
- [ ] **Implement all HTML void elements** — `<img>` (with alt text rendering), `<br>`, `<hr>`, `<input>`, `<meta>`, `<link>`, `<source>`, `<track>`, `<wbr>`, `<area>`, `<base>`, `<col>`, `<embed>`, `<param>`
- [ ] **Implement table layout** — `<table>`, `<thead>`, `<tbody>`, `<tfoot>`, `<tr>`, `<td>`, `<th>`, `colspan`, `rowspan`, `border` attribute. Tables are complex in HTML/CSS
- [ ] **Implement list layout** — `<ul>`, `<ol>`, `<li>` with bullet/number markers. Need to handle `list-style-type`, `list-style-image`, `list-style-position`
- [ ] **Implement form elements** — `<input>`, `<button>`, `<select>`, `<textarea>`, `<label>` (visual only, no interactivity)
- [ ] **Implement media elements** — `<img>` (display), `<video>`, `<audio>` (show controls UI)
- [ ] **Implement semantic elements** — `<header>`, `<footer>`, `<nav>`, `<article>`, `<section>`, `<aside>`, `<main>` (these should render as blocks)
- [ ] **Implement `<script>` and `<style>`** — style content parsed as CSS; script content may be JS (don't execute, just skip)
- [ ] **Implement `<noscript>`** — render content when JS is disabled (show noscript content)
- [ ] **Implement `<template>`** — parse but don't render template content

### CSS Properties
- [ ] **Implement CSS `color` property** — foreground text color for all elements. Need RGB, RGBA, hex, HSL, HSLA, named colors
- [ ] **Implement CSS `background-color`** — verify hex (`#fff`, `#ffffff`) and rgb() work. Also rgba(), hsl(), hsla(), named colors
- [ ] **Implement CSS `background-image`, `background-repeat`, `background-position`, `background-size`** — for gradients and images
- [ ] **Implement CSS `background` shorthand** — `background: #fff url(img.png) no-repeat center top`
- [ ] **Implement CSS `border-radius`** — rounded corners on boxes, including per-corner (`border-radius: 10px 5px 10px 5px`)
- [ ] **Implement CSS `box-shadow`** — drop shadows: `box-shadow: 2px 2px 4px rgba(0,0,0,0.5)`
- [ ] **Implement CSS `text-align`** — left/center/right/justify
- [ ] **Implement CSS `font-weight`, `font-style`, `text-decoration`** — bold, italic, underline, strikethrough, overline
- [ ] **Implement CSS `line-height`** — proper text spacing: unitless, px, em, %
- [ ] **Implement CSS `vertical-align`** — for inline and table cells: top/middle/bottom/baseline/sub/super
- [ ] **Implement CSS `opacity`** — transparency on elements (0-1, 0%-100%)
- [ ] **Implement CSS `transform`** — rotate, scale, translate (2D transforms)
- [ ] **Implement CSS `@media` queries** — responsive design breakpoints
- [ ] **Implement CSS `cursor`** — show appropriate cursor on interactive elements (pointer, text, wait, etc)
- [ ] **Implement CSS `outline`** — focus ring around elements (like border but doesn't affect layout)
- [ ] **Implement CSS `overflow`** — `overflow: hidden/scroll/auto/visible`, `overflow-x`, `overflow-y`
- [ ] **Implement CSS `white-space`** — `normal`, `pre`, `nowrap`, `pre-wrap`, `pre-line`
- [ ] **Implement CSS `word-wrap` / `overflow-wrap`** — long word breaking
- [ ] **Implement CSS `text-overflow`** — `text-overflow: ellipsis` for clipped text
- [ ] **Implement CSS `content`** — for ::before and ::after pseudo-elements
- [ ] **Implement CSS `@keyframes` and `animation`** — CSS animations (for visual completeness)
- [ ] **Implement CSS `transition`** — smooth property transitions on hover/focus

### CSS Selectors
- [ ] **Implement attribute selectors** — `[attr]`, `[attr=value]`, `[attr~=value]`, `[attr|=value]`
- [ ] **Implement pseudo-classes** — `:hover`, `:focus`, `:active`, `:visited`, `:link`, `:first-child`, `:last-child`, `:nth-child()`, `:nth-of-type()`, `:not()`
- [ ] **Implement pseudo-elements** — `::before`, `::after`, `::first-line`, `::first-letter`
- [ ] **Implement combinators** — descendant (space), child (>), adjacent sibling (+), general sibling (~)

### CSS Layout
- [ ] **Implement flexbox fully** — `display: flex`, `flex-direction`, `flex-wrap`, `flex-flow`, `justify-content`, `align-items`, `align-content`, `gap`, `flex-grow`, `flex-shrink`, `flex-basis`, `order`
- [ ] **Implement CSS grid** — `display: grid`, `grid-template-columns`, `grid-template-rows`, `grid-column`, `grid-row`, `gap`, `span`
- [ ] **Implement float** — `float: left/right` with wrap-around content
- [ ] **Implement `display` values** — `display: block/inline/inline-block/none/grid/flex`
- [ ] **Implement positioned layout** — `position: absolute/relative/fixed/sticky` with `top/left/right/bottom` offsets. Stacking context with z-index
- [ ] **Implement `visibility: hidden` and `display: none`** — hidden elements (visibility: hidden) occupy space; display:none removed from layout entirely
- [ ] **Implement `position: fixed`** — viewport-locked positioning (header bars, modals)

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

- [ ] **Implement `<img>` actual rendering** — fetch image URL, decode JPEG/PNG/WebP/GIF, display at correct size within content box
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
