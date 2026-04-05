# ViBrowsing Backlog

## 🔴 Critical (Parser/Rendering)

- [x] ~~Fix HTML parser double-html/body bug~~ (2026-04-03 sprint) — tokenizer now skips html/head/body StartTag/EndTag tokens; parser bootstraps them once cleanly; no more duplication
- [x] ~~Fix foster parenting~~ (2026-04-03 sprint) — parser now tracks table context, text inside tables is fostered to parent; implicit <p> close before block elements; table end tags properly close the table context
- [x] ~~Fix unclosed tag handling~~ (2026-04-03 sprint partial) — generic end tag now pops stack until matching tag found; unknown end tags are skipped without crashing; block tags implicitly close open <p> tags
- [x] ~~Implement entity decoding~~ (2026-04-03 sprint) — added decodeEntities() with named entities (amp, lt, gt, quot, apos, nbsp, ndash, mdash, lsquo, rsquo, ldquo, rdquo, hellip, copy, reg, trade, deg, plusmn, times, divide, frac12, frac14, frac34) and numeric entities (&#65;, &#x41;)
- [x] ~~Implement foreign content handling~~ (2026-04-04 sprint) — svg/math tracked via foreignContent counter; HTML-specific rules (p-closing, foster parenting) disabled inside foreign content; end tags only pop matching element in foreign context; bootstrap stack corrected to include head+body so elements append to body not html
- [x] ~~Fix dataPreview undefined variable in block.go~~ (2026-04-05 sprint 2) — removed fmt.Printf debug statements that referenced out-of-scope variable
- [x] ~~Fix MatchMedia width/height offset bugs in supports.go~~ (2026-04-05 sprint 2) — corrected string slice offsets for max-width/min-width/max-height/min-height and fixed orientation portrait/landscape logic (vh > vw not >=); fixed AND condition parsing to properly evaluate both conditions

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
- [x] ~~Implement pseudo-elements~~ (2026-04-05 sprint 2) — `::first-line`, `::first-letter`, `::before`, `::after` with content property
- [x] ~~Implement `:focus-within`~~ (2026-04-05 sprint 2) — matches when element or descendant has focus
- [x] ~~Implement `:checked`, `:disabled`, `:enabled`~~ (2026-04-05 sprint 2) — form state matching pseudo-classes
- [x] ~~Implement `:lang()`~~ (2026-04-05 sprint 2) — language-based selector with parameter support
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
- [x] ~~Implement HTTP cookies~~ (2026-04-05 sprint 2) — CookieJar stores/returns cookies; Set-Cookie header parsing; Cookie header on requests
- [x] ~~Implement browser history~~ (2026-04-05 sprint 2) — BrowserState with history list, back/forward navigation
- [x] ~~Implement link target resolution~~ (2026-04-05 sprint 2) — target="_blank"/"_self"/"_parent"/"_top" handling

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
- [ ] **Implement CSS `unicode-bidi`** — bidirectional text embedding levels (partially implemented in supports.go)
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

- [x] ~~Implement HTTP/1.1 keep-alive~~ (2026-04-10 sprint) — ConnPool with MaxIdleConns/MaxConnsPerHost; http.Transport with DisableKeepAlives=false for connection reuse
- [ ] **Implement HTTP/2 support** — upgrade to HTTP/2 for multiplexed requests
- [ ] **Implement TLS certificate verification** — proper HTTPS with certificate validation
- [ ] **Implement DNS resolution caching** — cache resolved IPs to avoid repeated DNS lookups
- [ ] **Implement connection timeout** — max time to establish TCP connection
- [ ] **Implement read/write timeouts** — prevent hanging on slow connections
- [ ] **Implement retry on connection reset** — automatically retry on transient failures
- [x] ~~Implement conditional GET (If-Modified-Since)~~ (2026-04-10 sprint) — ConditionalCache stores Last-Modified; sends If-Modified-Since header; handles 304 Not Modified
- [x] ~~Implement Content-Encoding~~ (2026-04-10 sprint) — Accept-Encoding: gzip, deflate; decompressBody() handles gzip/deflate; Response.Decompressed field
- [x] ~~Implement streaming fetch~~ (2026-04-10 sprint) — FetchStreaming with progress callback and cancel channel; FetchProgress struct; OnProgress/CancelFunc in StreamingOptions
- [x] ~~Implement `navigator.connection`~~ (2026-04-10 sprint) — NetworkInformation API with EffectiveType, Downlink, RTT, SaveData, Type; Navigator.Connection() method

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

- [x] ~~Implement `<colgroup>` and `<col>`~~ (2026-04-07 sprint) — ColumnBox for colgroup/col; span attribute on col; collectColumnWidths
- [ ] **Implement `<thead>`, `<tbody>`, `<tfoot>` table sections** — proper table section rendering order
- [x] ~~Implement `<td colspan>` and `<td rowspan>`~~ (2026-04-07 sprint) — tableGrid tracks cell positions; layoutTableRowWithGrid
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
- [x] ~~Implement `<details>` and `<summary>` toggle~~ (2026-04-07 sprint) — DetailsBox/SummaryBox types; details toggles open/closed state; summary is clickable label
- [x] ~~Implement CSS `pointer-events`~~ (2026-04-04 sprint 2) — pointer-events stored in style defaults
- [x] ~~Implement `<picture>` element~~ (2026-04-07 sprint) — responsive images with source srcset fallback chain; resolvePictureSource selects best image based on viewport/dpr
- [x] ~~Implement CSS `will-change` hint~~ (2026-04-04 sprint 2) — will-change stored in style defaults
- [x] ~~Implement CSS `image-rendering`~~ (2026-04-04 sprint 2) — image-rendering stored in style defaults
- [x] ~~Implement `window.scrollTo()` and `window.scrollBy()`~~ (2026-04-07 sprint) — Window.ScrollTo/ScrollBy with {top,left,behavior} options; scrollTo for absolute, scrollBy for relative positioning
- [x] ~~Implement `element.scrollIntoView()`~~ (2026-04-07 sprint) — Box.ScrollIntoView method with {block,inline,behavior} options; FindBoxByNode traverses layout tree; Canvas.ScrollIntoView high-level API
- [x] ~~Implement CSS `scroll-behavior`~~ (2026-04-04 sprint 2) — scroll-behavior stored in style defaults

### 🟢 Medium Priority

- [x] ~~Implement `window.innerWidth`, `window.innerHeight`~~ (2026-04-07 sprint) — returns viewport width/height in CSS pixels from Window object
- [x] ~~Implement `navigator.userAgent`~~ (2026-04-07 sprint) — Navigator API with UserAgent, AppCodeName, AppName, AppVersion, Platform; NewNavigator/NewNavigatorWithAgent constructors
- [x] ~~Implement CSS `overscroll-behavior`~~ (2026-04-04 sprint 2) — overscroll-behavior stored in style defaults
- [x] ~~Implement `<del>` and `<ins>` elements~~ (2026-04-04 sprint 2) — del renders strikethrough, ins renders underline
- [x] ~~Implement CSS `text-decoration-line/color/style`~~ (2026-04-04 sprint 2) — individual properties parsed; text-decoration shorthand parses all three; rendering in DrawText
- [x] ~~Implement `<output>` element~~ (2026-04-04 sprint 2) — output element handled as inline box
- [x] ~~Implement CSS `font-stretch`~~ (2026-04-07 sprint) — ultra-condensed to ultra-expanded keywords; font shorthand updated to parse stretch value
- [x] ~~Implement CSS `place-items` and `place-self`~~ (2026-04-07 sprint) — alignment shorthands; place-items: <align> <justify>?; place-self: <align-self> <justify-self>?; justify-items/justify-self individual properties added
- [x] ~~Implement CSS `mix-blend-mode`~~ (2026-04-04 sprint 2) — mix-blend-mode stored in style defaults
- [x] ~~Implement CSS `hanging-punctuation`~~ (2026-04-04 sprint 2) — hanging-punctuation stored in style defaults
- [x] ~~Implement CSS `contain` property~~ (2026-04-04 sprint 2) — contain stored in style defaults
- [x] ~~Implement `CSS.supports()` API~~ (2026-04-07 sprint) — CSS.Supports(property, value) and CSS.Supports(condition) for @supports evaluation from JS
- [x] ~~Implement `window.matchMedia()`~~ (2026-04-07 sprint) — MatchMedia returns MediaQueryList with Matches and MediaQuery properties; supports media types and feature queries
- [x] ~~Implement CSS `transform-box`~~ (2026-04-07 sprint) — transform-box: view-box/content-box/fill-box; determines reference box for CSS transforms

### 🟠 Low Priority

- [x] ~~Implement `<map>` and `<area>`~~ (2026-04-07 sprint) — MapBox/AreaBox types with MapName, AreaShape, AreaCoords; area parses rect/circle/poly/default shapes with coords
- [x] ~~Implement `IntersectionObserver`~~ (2026-04-07 sprint) — IntersectionObserver with Observe/Unobserve/Disconnect; IntersectionObservation tracks elements; CheckAndNotify calculates intersection ratios and fires callback
- [x] ~~Implement `ResizeObserver`~~ (2026-04-07 sprint) — ResizeObserver with Observe/Unobserve/Disconnect; CheckAndNotify detects size changes and fires callback with ResizeObserverEntry
- **Implement `ime-mode` CSS property** — input method editor hints for East Asian input
- [x] ~~Implement `<data>` element~~ (2026-04-07 sprint) — DataBox type with DataValue property; renders as inline box; value attribute stored for programmatic access


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
- [x] ~~Implement `hyphens: auto/manual`~~ (2026-04-10 sprint) — hyphens property stored in defaults; applies soft hyphen during text rendering
- [x] ~~Implement `text-justify: inter-word/inter-character`~~ (2026-04-10 sprint) — text-justify property stored in defaults
- [x] ~~Implement `font-synthesis`~~ (2026-04-10 sprint) — font-synthesis property stored in defaults; controls bold/italic synthesis
- [x] ~~Implement `appearance` property~~ (2026-04-10 sprint) — appearance property stored in defaults; standardize form controls

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

- [x] ~~Implement CSS `clip-path` drawing~~ (2026-04-06 sprint) — inset/circle/ellipse/polygon clipping with PushClip and ellipse mask drawing
- [x] ~~Implement CSS `filter` drawing~~ (2026-04-06 sprint) — blur, brightness, contrast, grayscale, sepia applied via pixel manipulation
- [x] ~~Implement CSS `backdrop-filter` drawing~~ (2026-04-06 sprint) — blur behind fixed/absolute positioned elements by extracting and filtering background pixels
- [x] ~~Implement `<iframe>` rendering~~ (2026-04-06 sprint) — iframe placeholder with src/title display and border
- [ ] **Implement `<canvas>` 2D context** — Render canvas drawing commands to output
- [x] ~~Implement emoji rendering~~ (2026-04-06 sprint) — emoji detection via Unicode ranges and colored circle placeholder fallback
- [ ] **Fix inline box baseline calculation** — inline text boxes should share a common baseline; vertical-align: middle/bottom should position relative to that baseline
- [ ] **Fix table cell collapsing borders** — adjacent table cells should share borders (border-collapse behavior), currently each cell renders its own border
- [ ] **Fix float clearing** — blocks that clear:left/right/both should properly position below the float, not overlap it
- [x] ~~Fix `:nth-child()` selector~~ (2026-04-05 sprint 2) — ParseNthChild now handles even, odd, 2n+1, 3n-1 formulas

### 🟡 High (CSS Selectors & Cascade)

- [x] ~~Implement `:not()` pseudo-class~~ (2026-04-04 sprint) — negation selector with complex selectors inside
- [x] ~~Implement `:checked`, `:disabled`, `:enabled` pseudo-classes~~ (2026-04-05 sprint) — form state matching pseudo-classes
- [x] ~~Implement `::before` and `::after` pseudo-elements~~ (2026-04-05 sprint 2) — with `content` property
- [x] ~~Implement `:first-line` and `::first-letter`~~ (2026-04-05 sprint 2) — text segment pseudo-elements
- [x] ~~Implement CSS combinators fully~~ (2026-04-05 sprint) — descendant/child/adjacent sibling/general sibling with traversal
- [x] ~~Implement `:focus-visible` pseudo-class~~ (2026-04-06 sprint) — matches focusable elements with tabindex or form elements
- [x] ~~Implement `:is()` and `:where()` pseudo-classes~~ (2026-04-06 sprint) — forgiving selector list pseudo-classes with 0 specificity for :where
- [x] ~~Implement `:has()` relative selector~~ (2026-04-06 sprint) — parent/previous-sibling selector (div:has(p), img:has(+ p))
- [x] ~~Implement `:indeterminate` and `:default` pseudo-classes~~ (2026-04-06 sprint) — checkbox/radio state pseudo-classes

### 🟡 High (URL & Navigation)

- [x] ~~Implement HTTP cookies~~ (2026-04-05 sprint 2) — CookieJar stores/returns cookies; Set-Cookie header parsing; Cookie header on requests
- [x] ~~Implement browser history~~ (2026-04-05 sprint 2) — BrowserState with history list, back/forward navigation
- [x] ~~Implement link target resolution~~ (2026-04-05 sprint 2) — target="_blank"/"_self"/"_parent"/"_top" handling
- [x] ~~Handle malformed URLs gracefully~~ (2026-04-06 sprint) — IsValidURL/SanitizeURL validation; FetchError type with descriptive messages
- [x] ~~Handle binary/non-text content~~ (2026-04-06 sprint) — IsTextContent checks Content-Type; DetectBinaryContent sniffs magic numbers (JPEG/PNG/GIF/PDF/ZIP); returns FetchError for non-HTML
- [x] ~~Handle very large pages~~ (2026-04-06 sprint) — MaxDocumentSize (10MB) limit; streaming fetch; large line handling (1MB cap)

### 🟡 High (Platform)

- [ ] **Implement scroll support** — Mouse wheel / scrollbar navigation through page content
- [ ] **Implement click interaction** — Clicking links should navigate to those URLs
- [ ] **Implement text selection** — Highlight text with mouse
- [ ] **Implement window title** — Render document `<title>` in window title bar
- [ ] **Implement favicon** — Fetch and display favicon.ico in window
- [x] ~~Implement `<template>` element~~ (2026-04-06 sprint) — parsed but not rendered until JS activates it; TemplateContent field stores inert document fragment

### 🟢 Medium (Performance)

- [ ] **Lazy image decoding** — Don't decode images until visible in viewport
- [x] ~~CSS selector indexing~~ (2026-04-06 sprint) — BuildSelectorIndex maps byClass/byID/byTag for O(1) lookups
- [ ] **Text measurement caching** — Cache Ebitengine text measurement results per font/size/text combo
- [x] ~~Streaming HTML parse~~ (2026-04-06 sprint) — NewTokenizerFromReader, ReadFrom method, FetchStreaming for incremental parsing
- [ ] **Memory pool for nodes** — Reuse allocated Node/Token objects instead of GC-heavy allocation per parse
- [x] ~~Implement calc() evaluation in layout~~ (2026-04-06 sprint) — resolveLength helpers evaluate calc() expressions during layout

### 🟢 Medium (Content & Rendering)

- [ ] **Implement `<abbr>` with title tooltip** — Abbreviation with full text on hover
- [ ] **Implement `<ruby>` layout** — Ruby text above/below base text for East Asian typography
- [x] ~~Implement calc() evaluation in layout~~ (2026-04-06 sprint) — resolveLength helpers evaluate calc() during layout
- [x] ~~Implement CSS @media query matching~~ (2026-04-06 sprint) — MatchMediaQuery + FilterRulesByMedia; viewport dims passed to BuildLayoutTree
- [ ] **Implement CSS `cursor`** — Show appropriate cursor on interactive elements

### 🟠 Low (Testing & QA)

- [ ] **html5lib test corpus** — Download and run 500+ HTML parsing edge case tests
- [ ] **Visual screenshot tests** — Collect baseline screenshots, diff on changes
- [ ] **Fuzz testing** — Use go-fuzz to generate random HTML/CSS and verify no panics
- [ ] **Performance benchmarks** — Measure parse time, layout time, paint time for pages of varying size

### 🟠 Low (Canvas/Drawing)

- [x] ~~Implement `background-repeat: space/round`~~ (2026-04-04 sprint 2) — space distributes tiles evenly with spacing; round scales tiles to fill
- [x] ~~Implement `background-position` (percentages)~~ (2026-04-04 sprint 2) — percentage values resolve against (container - image) for offset
- [x] ~~Implement `background-size: cover/contain`~~ (2026-04-04 sprint 2) — already partially done; cover fills, contain fits
- [x] ~~Implement `text-shadow` multiple shadows~~ (2026-04-04 sprint 2) — comma-separated shadows via splitShadowParts
- [x] ~~Implement `box-shadow` multiple shadows~~ (2026-04-04 sprint 2) — already supported via comma split
- [ ] **Implement `opacity` per draw call** — Apply alpha blending per element not just whole box

---

## 🆕 New Items (2026-04-04 Sprint 2)

### 🟡 High (Layout/Rendering)

- [ ] **Implement `<abbr>` with title tooltip** — Abbreviation with full text on hover (store title, show on mouseover)
- [ ] **Implement CSS `writing-mode: vertical-rl/vertical-lr`** — Vertical text layout for CJK and other writing systems
- [ ] **Implement `text-justify: inter-word/inter-character`** — Justification algorithm improvements
- [ ] **Implement `hyphens: auto/manual`** — Automatic hyphenation of words at line breaks
- [ ] **Implement CSS `@container` queries** — Container-style responsive design

### 🟡 High (CSS Selectors & Cascade)

- [ ] **Implement `::first-line` pseudo-element** — Style first line of text differently
- [ ] **Implement `::first-letter` pseudo-element** — Style first letter of text differently (drop cap)
- [ ] **Implement `::before` and `::after` with content property** — Generate content before/after elements
- [ ] **Implement `:focus-within` pseudo-class** — Matches when element or descendant has focus

### 🟢 Medium (Features)

- [ ] **Implement window.matchMedia()** — JavaScript API for media query testing
- [ ] **Implement CSS.supports() API** — JavaScript API for CSS feature detection
- [ ] **Implement CSS counter() rendering in content** — Draw counters for list numbering
- [ ] **Implement window.scrollTo() and window.scrollBy()** — Scroll the viewport programmatically
- [ ] **Implement element.scrollIntoView()** — Scroll element into viewport

### 🟢 Medium (Performance)

- [ ] **Implement text measurement caching** — Cache font.MeasureString results per font/size/text
- [ ] **Implement CSS selector indexing** — Build index of elements by class/id/tag for fast matching
- [ ] **Implement image lazy decoding** — Decode images only when they enter viewport

### 🟠 Low (Content & Rendering)

- [ ] **Implement `<ruby>` layout fully** — Ruby annotations above/below base text for ruby elements
- [ ] **Implement `image-rendering: pixelated`** — Crisp pixel art scaling
- [ ] **Implement `shape-outside` for floats** — Non-rectangular float boundaries using polygon
- [ ] **Implement `clip-path: path()` SVG paths** — Support for SVG path data in clip-path

### 🟠 Low (Testing & QA)

- [ ] **Download html5lib test corpus** — Run 500+ HTML parsing edge case tests
- [ ] **Visual screenshot regression tests** — Collect baseline screenshots, diff on changes
- [ ] **go-fuzz fuzz testing** — Generate random HTML/CSS, verify no panics

---

## 🆕 New Items (2026-04-05 Sprint 2)

### 🔴 Critical (Parser/Rendering)

- [ ] **Fix table cell border-collapse rendering** — Adjacent table cells should share borders (border-collapse behavior), currently each cell renders its own border separately
- [ ] **Fix float clearing logic** — Blocks that clear:left/right/both should properly position below the float, not overlap it; ensure ClearStyle is respected
- [ ] **Fix inline box baseline calculation** — Inline text boxes should share a common baseline; vertical-align: middle/bottom should position relative to that baseline

### 🟡 High (Layout/Rendering)

- [ ] **Implement CSS `filter` drawing** — Apply blur, brightness, contrast, grayscale, sepia pixel effects to rendered boxes using image filtering
- [ ] **Implement CSS `backdrop-filter` drawing** — Apply blur to elements behind fixed/absolute positioned elements (for dialog/modals with blur backdrop)
- [ ] **Implement `<iframe>` rendering** — Show placeholder for iframes; optionally recursively render same-origin iframe content
- [ ] **Implement `<canvas>` 2D context** — Render canvas 2D drawing API (rect, arc, path, text, image) to output buffer
- [ ] **Implement emoji rendering** — Proper color emoji display using fontconfig or embedded emoji font

### 🟡 High (HTML Elements)

- [ ] **Implement `<template>`** — Parse template content into inert DOM but don't render until JavaScript activates it
- [ ] **Implement `<slot>` and shadow DOM** — Basic slot projection for web component support; fallback to display:contents
- [ ] **Implement `<dialog>` modal** — Modal dialog with backdrop; showModal() and close() methods
- [ ] **Implement `<colgroup>` and `<col>`** — Column grouping for table column widths and styles
- [ ] **Implement `<td colspan>` and `<td rowspan>`** — Cell spanning for complex tables with merged cells

### 🟡 High (CSS Properties)

- [ ] **Implement CSS `transform` drawing** — Apply rotate(), scale(), translate(), skew() 2D transforms to elements during rendering
- [ ] **Implement CSS `writing-mode: vertical-rl/vertical-lr`** — Vertical text layout for CJK and other writing systems
- [ ] **Implement CSS `hyphens: auto/manual`** — Automatic hyphenation of words at line breaks using hyphenate character
- [ ] **Implement CSS `text-justify: inter-word/inter-character`** — Justification algorithm for better text alignment
- [ ] **Implement CSS `unicode-bidi: isolate/embed/override`** — Bidirectional text isolation and override for RTL content

### 🟡 High (DOM APIs)

- [ ] **Implement `innerText`** — Get rendered text content (like textContent but CSS-aware, respects display and visibility)
- [ ] **Implement `getComputedStyle()`** — Return the computed style object for an element (all CSS properties as they resolve)
- [ ] **Implement `getBoundingClientRect()`** — Return element position relative to viewport (x, y, width, height, top, right, bottom, left)
- [ ] **Implement `element.scrollIntoView()`** — Scroll element into viewport with options (start/center/end/nearest)
- [ ] **Implement `MutationObserver`** — JavaScript API to observe DOM changes (attributes, childList, subtree)

### 🟢 Medium (Features)

- [ ] **Implement find-in-page** — Ctrl+F to search for text in rendered page and highlight matches
- [ ] **Implement right-click context menu** — Copy link, copy text, open in new tab options on right-click
- [ ] **Implement loading progress indicator** — Spinner/progress bar during page fetch
- [ ] **Implement `window.print()`** — Trigger print dialog with current page content
- [ ] **Implement `@media print` styles** — Apply print-specific stylesheet rules and hide non-essential content

### 🟢 Medium (Performance)

- [ ] **Text measurement caching** — Cache Ebitengine font.MeasureString results per font/size/style combination to avoid repeated measurement
- [ ] **CSS selector indexing** — Build index of elements by class/id/tag for fast selector matching instead of tree traversal
- [ ] **Image lazy decoding** — Don't decode image data until it's about to be rendered in viewport
- [ ] **Memory pool for nodes** — Use sync.Pool to reuse allocated Node/Token/Box objects instead of GC-heavy allocation per parse

### 🟠 Low (Platform)

- [ ] **Headless mode** — Generate screenshots without GUI display for CI/testing using Ebitengine headless mode
- [ ] **PDF output** — Render page to PDF using go's pdf libraries or command-line tools
- [ ] **Wayland support** — Currently X11 only via Ebitengine; add Wayland compositor support
- [ ] **Multi-window/tab support** — Multiple browser windows or tabs with independent navigation

---

## 🆕 New Items (2026-04-06 Sprint)

### 🔴 Critical (Parser/Rendering)

- [ ] **Fix inline box baseline calculation** — inline text boxes should share a common baseline; vertical-align: middle/bottom should position relative to that baseline
- [x] ~~Fix table cell border-collapse rendering~~ (2026-04-11 sprint) — adjacent table cells with border-collapse:collapse now share borders; cells at shared edges only draw half their border
- [x] ~~Fix float clearing logic~~ (2026-04-11 sprint) — blocks with clear:left/right/both now properly position below float; ctx.Y is updated after clearing so subsequent blocks don't overlap

### 🟡 High (Layout/Rendering)

- [ ] **Implement `<canvas>` 2D context** — Render canvas 2D drawing API (rect, arc, path, text, image) to output buffer
- [ ] **Implement CSS `transform` drawing** — Apply rotate(), scale(), translate(), skew() 2D transforms to elements during rendering
- [ ] **Implement CSS `writing-mode: vertical-rl/vertical-lr`** — Vertical text layout for CJK and other writing systems
- [ ] **Implement `hyphens: auto/manual`** — Automatic hyphenation of words at line breaks using hyphenate character

### 🟡 High (HTML Elements)

- [ ] **Implement `<colgroup>` and `<col>`** — Column grouping for table column widths and styles
- [ ] **Implement `<td colspan>` and `<td rowspan>`** — Cell spanning for complex tables with merged cells
- [x] ~~Implement `<figure>` and `<figcaption>`~~ (2026-04-07 sprint) — FigureBox and FigcaptionBox types in BoxType enum; buildBox handles figure/figcaption tags; caption positioned above/below figure content

### 🟡 High (CSS Properties)

- [ ] **Implement CSS `unicode-bidi: isolate/embed/override`** — Bidirectional text isolation and override for RTL content
- [ ] **Implement CSS `text-justify: inter-word/inter-character`** — Justification algorithm for better text alignment
- [ ] **Implement CSS `@container` queries** — Container-style responsive design

### 🟡 High (DOM APIs)

- [ ] **Implement `innerText`** — Get rendered text content (like textContent but CSS-aware, respects display and visibility)
- [ ] **Implement `getComputedStyle()`** — Return the computed style object for an element (all CSS properties as they resolve)
- [ ] **Implement `getBoundingClientRect()`** — Return element position relative to viewport (x, y, width, height, top, right, bottom, left)
- [ ] **Implement `MutationObserver`** — JavaScript API to observe DOM changes (attributes, childList, subtree)

### 🟢 Medium (Features)

- [ ] **Implement find-in-page** — Ctrl+F to search for text in rendered page and highlight matches
- [ ] **Implement right-click context menu** — Copy link, copy text, open in new tab options on right-click
- [ ] **Implement loading progress indicator** — Spinner/progress bar during page fetch
- [ ] **Implement `window.print()`** — Trigger print dialog with current page content
- [ ] **Implement `@media print` styles** — Apply print-specific stylesheet rules and hide non-essential content

### 🟢 Medium (Advanced DOM)

- [ ] **Implement `element.scrollIntoView()`** — Scroll element into viewport with options (start/center/end/nearest)
- [ ] **Implement `window.scrollTo()` and `window.scrollBy()`** — Scroll the viewport programmatically
- [ ] **Implement `window.matchMedia()`** — MediaQueryList interface for JavaScript media query checks
- [ ] **Implement `CSS.supports()` API** — JavaScript API for CSS feature detection: `CSS.supports('display', 'grid')`

### 🟠 Low (Canvas/Drawing)

- [ ] **Implement `opacity` per draw call** — Apply alpha blending per element not just whole box
- [ ] **Implement `background-repeat: space/round` drawing** — Tile backgrounds with spacing or scaling
- [ ] **Implement `background-position` (percentages) drawing** — Offset background image by percentage

### 🟠 Low (Testing & QA)

- [ ] **html5lib test corpus** — Download and run 500+ HTML parsing edge case tests from html5lib project
- [ ] **Visual screenshot regression tests** — Collect baseline screenshots of known pages, diff on changes
- [ ] **go-fuzz fuzz testing** — Generate random HTML/CSS, verify parser/renderer doesn't panic on malformed input
- [ ] **Performance regression CI** — Fail build if parse+layout time increases >10% vs baseline

## 🆕 New Items (2026-04-07 Sprint)

### 🔴 Critical (Parser/Rendering)

- [ ] **Fix inline box baseline calculation** — inline text boxes should share a common baseline; vertical-align: middle/bottom should position relative to that baseline
- [ ] **Fix table cell border-collapse rendering** — adjacent table cells should share borders (border-collapse behavior), currently each cell renders its own border
- [ ] **Fix float clearing logic** — blocks that clear:left/right/both should properly position below the float, not overlap it
- [ ] **Fix `<colgroup>` width application** — column widths from colgroup/col elements are collected but not yet applied to table cell widths during layout

### 🟡 High (Layout/Rendering)

- [x] ~~Implement `<canvas>` 2D context~~ (2026-04-08 sprint) — Canvas 2D Context2D with state management, paths, transforms, gradients, text, images
- [ ] **Implement CSS `transform` drawing** — Apply rotate(), scale(), translate(), skew(), matrix() 2D transforms to elements during rendering using transform-box reference
- [ ] **Implement CSS `writing-mode: vertical-rl/vertical-lr`** — Vertical text layout for CJK and other writing systems; text flows top-to-bottom or bottom-to-top
- [ ] **Implement `<thead>`, `<tbody>`, `<tfoot>` table sections** — proper table section rendering order; these elements should not affect visual layout but organize table structure
- [ ] **Implement `hyphens: auto/manual`** — Automatic hyphenation of words at line breaks using hyphenate character (U+00AD)

### 🟡 High (HTML Elements)

- [ ] **Implement `<dialog>` modal fully** — modal dialog with backdrop; showModal() and close() methods; backdrop should apply backdrop-filter blur
- [x] ~~Implement `<details>` and `<summary>` toggle~~ (2026-04-08 sprint 2) — DrawDetails with disclosure triangle; DrawSummary with clickable label
- [x] ~~Implement `<slot>` element~~ (2026-04-08 sprint 2) — DrawSlot renders as display:contents (no-op); slots project light DOM content
- [x] ~~Implement `<template>` rendering~~ (2026-04-08 sprint 2) — DrawTemplate is a no-op; template content is inert and not rendered

### 🟡 High (CSS Properties)

- [x] ~~Implement CSS `unicode-bidi: isolate/embed/override`~~ (2026-04-08 sprint 2) — supports isolate and isolate-override values in CSS.supports()
- [ ] **Implement CSS `text-justify: inter-word/inter-character`** — Justification algorithm for better text alignment; inter-character adjusts spacing between characters
- [x] ~~Implement CSS `@container` queries~~ (2026-04-08 sprint 2) — ParseContainerQuery, MatchContainerQuery, IsContainerQuery in css/media.go
- [ ] **Implement CSS `break-inside: avoid/avoid-page/avoid-column`** — Prevent breaks inside elements; used for keeping tables, figures, cards together
- [x] ~~Implement CSS `mix-blend-mode` drawing~~ (2026-04-08 sprint 2) — blendColors function with all 16 blend modes; rgbToHSL/hslToRGB helpers

### 🟡 High (DOM APIs)

- [x] ~~Implement `getComputedStyle()`~~ (2026-04-08 sprint) — GetComputedStyle function returns cascaded/inherited style map for element
- [x] ~~Implement `getBoundingClientRect()`~~ (2026-04-07 sprint) — GetBoundingClientRect returns DOMRect with viewport-relative position and dimensions
- [x] ~~Implement `innerText`~~ (2026-04-07 sprint) — InnerText returns CSS-aware rendered text (respects display/visibility)
- [x] ~~Implement `closest()`~~ (2026-04-08 sprint) — closest() traverses ancestors (including self) to find matching CSS selector
- [x] ~~Implement `matches()`~~ (2026-04-08 sprint) — matches() returns true if element would be selected by given CSS selector

### 🟢 Medium (Features)

- [ ] **Implement find-in-page** — Ctrl+F to search for text in rendered page and highlight matches; show count of matches
- [ ] **Implement right-click context menu** — Copy link, copy text, open in new tab options on right-click
- [ ] **Implement loading progress indicator** — Spinner/progress bar during page fetch; update from 0-100% as content loads
- [ ] **Implement `window.print()`** — Trigger print dialog with current page content; should apply @media print styles
- [ ] **Implement `@media print` styles** — Apply print-specific stylesheet rules; hide navigation, expand hidden sections, optimize for paper

### 🟢 Medium (Performance)

- [x] ~~Text measurement caching~~ (2026-04-08 sprint 2) — TextMeasurementCache with sync.Map in layout/text_cache.go
- [x] ~~Memory pool for nodes~~ (2026-04-08 sprint 2) — BoxPool and CSSStylePool with sync.Pool in layout/text_cache.go
- [ ] **CSS selector caching** — Cache selector match results for unchanged DOM/subtree; invalidate on DOM mutations

### 🟠 Low (Canvas/Drawing)

- [ ] **Implement `opacity` per draw call** — Apply alpha blending per element not just whole box; each DrawBox should respect element opacity
- [x] ~~Implement `outline` drawing~~ (2026-04-08 sprint 2) — outline already draws in DrawBox at lines 1396-1415
- [ ] **Implement emoji rendering via fontconfig** — Use fontconfig to find and load color emoji fonts for proper emoji display
- [x] ~~Implement `<img>` actual image loading~~ (2026-04-08 sprint 2) — loadImage now fetches and decodes PNG/JPEG/GIF from URL

### 🟡 High (Content & Rendering)

- [x] ~~Implement `<img>` actual image loading~~ (2026-04-08 sprint 2) — loadImage fetches from URL and decodes PNG/JPEG/GIF

### 🟠 Low (Testing & QA)

- [ ] **Property-based testing with go-fuzz** — Generate random HTML/CSS combinations and verify no panics or hangs
- [ ] **Large document stress test** — Parse and render 10MB+ HTML file; verify memory usage stays under 500MB and completes within 30s

---

## 🆕 New Items (2026-04-08 Sprint)

### 🔴 Critical (Parser/Rendering)

- [ ] **Fix table cell border-collapse rendering** — adjacent table cells should share borders (border-collapse behavior), currently each cell renders its own border separately; cells at shared edges should only draw outer half of border
- [x] ~~Fix CSS `font-family` cascading~~ (2026-04-11 sprint) — font-family properly inherited through CSS cascade via GetComputedStyle fix; child without explicit font-family now inherits parent's font-family
- [x] ~~Fix CSS `font-size` em unit resolution~~ (2026-04-11 sprint) — font-size: 1.2em inside parent with font-size: 20px now resolves to 24px using parent's computed font-size for em resolution

### 🟡 High (Layout/Rendering)

- [ ] **Implement CSS `line-height` proper rendering** — line-height should affect the height of inline boxes and the spacing between lines; currently may not be properly applied during inline layout
- [ ] **Implement CSS `letter-spacing` in DrawText** — letter-spacing property should add extra space between characters; currently the property is stored but may not be applied during text rendering
- [ ] **Implement CSS `word-spacing` in DrawText** — word-spacing property should add extra space after spaces/words; currently stored but not applied during text rendering
- [ ] **Implement `<thead>`, `<tbody>`, `<tfoot>` table section layout** — table sections should organize rows properly; currently TableSectionBox type exists but sections may not affect visual rendering order

### 🟡 High (CSS Selectors & Cascade)

- [x] ~~Fix `:first-child` selector~~ (2026-04-11 sprint) — :first-child now correctly matches when element is first child of parent
- [x] ~~Fix `:last-child` selector~~ (2026-04-11 sprint) — :last-child now correctly matches when element is last child of parent
- [x] ~~Implement `:only-child` selector~~ (2026-04-11 sprint) — :only-child matches when element is the only child of its parent
- [x] ~~Implement `:nth-of-type()` pseudo-class~~ (2026-04-11 sprint) — :nth-of-type(an+b) counts siblings of same type; :nth-last-of-type also implemented

### 🟡 High (HTML Elements)

- [x] ~~Implement `<output>` element fully~~ (2026-04-07 sprint) — the output element should show the result of a calculation, similar to a read-only text field with special semantic meaning
- [x] ~~Implement `<datalist>` element~~ (2026-04-09 sprint) — datalist provides autocomplete suggestions for input elements; DataListBox type added
- [ ] **Implement `<meter>` visual rendering** — meter should show a gauge value with color zones (green/yellow/red based on low/high/optimum attributes)
- [ ] **Implement `<progress>` visual rendering** — progress bar should show completion percentage with animated fill when indeterminate

### 🟡 High (DOM APIs)

- [x] ~~Implement `closest()` method~~ (2026-04-10 sprint) — returns the closest ancestor of the current element (including itself) that matches a given CSS selector; returns null if no match found
- [x] ~~Implement `matches()` method~~ (2026-04-10 sprint) — returns true if the element would be selected by the given CSS selector; throws SyntaxError if the selector is invalid
- [x] ~~Implement `scrollBy()` method on Element~~ (2026-04-09 sprint) — scrolls the element by a specified amount; Box.ScrollBy with (dx, dy) and {left, top} options

### 🟢 Medium (Features)

- [ ] **Implement browser zoom** — Ctrl+/Ctrl- should zoom the page by scaling the viewport or applying CSS transform to the root box; should track zoom level and display in window title
- [ ] **Implement focus indicator** — when navigating with Tab key, interactive elements should show visible focus ring (outline or box-shadow) as defined by `:focus-visible` pseudo-class
- [ ] **Implement `<input type="checkbox">` visual** — checkboxes should render as small square boxes with check mark when checked, using unicode ballot box character or custom drawn shape

### 🟢 Medium (Performance)

- [ ] **Parallel layout for independent subtrees** — if a container has multiple independent block children (no float/positioned dependencies), layout them concurrently using goroutines
- [ ] **Cache computed styles** — when the same element is queried for computed style multiple times without DOM changes, reuse the cached result from first computation

### 🟠 Low (Canvas/Drawing)

- [ ] **Implement `background-blend-mode`** — blend background-color and background-image together using specified blend mode (multiply, screen, overlay, etc.)
- [ ] **Implement `mix-blend-mode` on boxes** — apply blend mode when drawing overlapping positioned elements; currently mix-blend-mode is stored in style but not applied during rendering

### 🟠 Low (Networking & Protocol)

- [ ] **Implement HTTP keep-alive** — reuse TCP connections for multiple requests to the same origin server instead of opening a new connection for each request
- [ ] **Implement conditional GET with If-Modified-Since** — when fetching a URL that was previously retrieved, send Last-Modified header; if server responds with 304 Not Modified, use cached content

### 🟠 Low (Internationalization)

- [x] ~~Implement `<bdo dir="rtl">>` (2026-04-10 sprint)~~ — bidirectional text override element that forces text direction; bdo with dir="rtl" or dir="ltr" sets direction CSS property
- [ ] **Implement `lang` attribute on html element** — `<html lang="en">` should set the document's language, affecting speech synthesis, hyphenation, and :lang() selector matching

### 🟠 Low (Accessibility)

- [ ] **Implement `role` attribute rendering** — certain ARIA roles should affect how elements are announced by screen readers; roles like "button", "navigation", "main" have semantic meaning
- [ ] **Implement `<summary>` and `<details>`** — the summary element is the visible label for a details element; clicking summary toggles visibility of details content

---

## 🆕 New Items (2026-04-08 Sprint)

### 🔴 Critical (Parser/Rendering)

- [ ] **Fix table cell border-collapse rendering** — adjacent table cells should share borders (border-collapse behavior), currently each cell renders its own border separately; cells at shared edges should only draw outer half of border
- [ ] **Fix inline box baseline calculation** — inline text boxes should share a common baseline; vertical-align: middle/bottom should position relative to that baseline; currently each text box uses its own baseline
- [ ] **Fix CSS `font-size` em unit resolution** — `font-size: 1.2em` inside a div with `font-size: 20px` should resolve to 24px but may incorrectly use body default 16px instead of parent 20px
- [ ] **Fix `<colgroup>` width application** — column widths from colgroup/col elements are collected but not yet applied to table cell widths during layout

### 🟡 High (Layout/Rendering)

- [ ] **Implement CSS `writing-mode: vertical-rl/vertical-lr`** — Vertical text layout for CJK; text flows top-to-bottom (rl) or bottom-to-top (lr); inline boxes stack vertically instead of horizontally
- [ ] **Implement CSS `unicode-bidi: isolate/embed/override`** — isolate/override values for RTL text; isolate-unicode-level resets embedding level; override forces LTR or RTL direction
- [ ] **Implement CSS `text-justify: inter-word/inter-character`** — inter-word adjusts space between words; inter-character adjusts space between characters for CJK text
- [ ] **Implement CSS `@container` queries** — Container-style responsive design; @container rule with size queries on named containers; container-query library for parsing
- [ ] **Implement CSS `break-inside: avoid/avoid-page/avoid-column`** — Prevent breaks inside elements; when laying out multi-column, don't break inside a box with break-inside: avoid
- [ ] **Implement CSS `transform` drawing for all elements** — Currently transform parsing exists but transforms are not applied during rendering; applyTransform in DrawBox for rotated/scaled elements

### 🟡 High (HTML Elements)

- [ ] **Implement `<dialog>` modal fully** — modal dialog with backdrop; showModal() and close() methods; backdrop applies backdrop-filter blur; ESC key closes modal
- [ ] **Implement `<thead>`, `<tbody>`, `<tfoot>` table section rendering** — these elements organize table rows; tbody rows should render in document order; thead at top, tfoot at bottom regardless of source order
- [ ] **Implement `<slot>` and shadow DOM fully** — slot projection for web components; fallback to display:contents when no slot assignment; named slots support

### 🟡 High (CSS Selectors & Cascade)

- [x] ~~Implement `:nth-last-of-type()` pseudo-class~~ (2026-04-09 sprint) — counts from end of sibling list; `tr:nth-last-of-type(2)` selects second-to-last tr
- [x] ~~Implement `:only-of-type()` pseudo-class~~ (2026-04-09 sprint) — matches elements that are the only child of its type among siblings
- [ ] **Implement `:placeholder-shown` on select** — matches when a select's placeholder option is selected; for styled dropdowns with placeholder text

### 🟡 High (DOM APIs)

- [x] ~~Implement `scrollBy()` on Element~~ (2026-04-09 sprint) — scrolls the element by a specified amount; Box.ScrollBy with (dx, dy) and {left, top} options
- [x] ~~Implement `requestAnimationFrame()`~~ (2026-04-09 sprint) — AnimationManager.RequestAnimationFrame with handle return; CancelAnimationFrame; TickRAF for render loop
- [x] ~~Implement `cancelAnimationFrame()`~~ (2026-04-09 sprint) — AnimationManager.CancelAnimationFrame marks handle for cancellation before next frame

### 🟢 Medium (Features)

- [ ] **Implement find-in-page** — Ctrl+F to search for text in rendered page; highlight all matches; show count; navigate between matches with Enter/Shift+Enter
- [ ] **Implement right-click context menu** — Copy link, copy text, open in new tab on right-click; menu appears at cursor position
- [ ] **Implement loading progress indicator** — Spinner/progress bar during page fetch; updates from 0-100% as content loads; shows in window title
- [ ] **Implement `window.print()`** — Trigger print dialog; applies @media print styles; renders to PDF or system print dialog
- [ ] **Implement `@media print` styles** — Hide navigation, ads; expand hidden sections; optimize for paper (A4/Letter); page break avoidance

### 🟢 Medium (Performance)

- [ ] **Text measurement caching** — Cache Ebitengine font.MeasureString results per font/size/style; use sync.Map keyed by "font:size:text"; invalidate on font change
- [ ] **Image lazy decoding** — Don't decode images until scrolled into viewport; use offscreen buffer; decode on background goroutine
- [ ] **Memory pool for Box allocation** — Use sync.Pool for Box, Node, Token objects; reduce GC pressure during parse/layout of large documents

### 🟠 Low (Canvas/Drawing)

- [ ] **Implement `background-blend-mode`** — Blend background-color and background-image using multiply, screen, overlay, etc.; apply per-element when drawing
- [ ] **Implement `mix-blend-mode` on positioned elements** — Apply blend mode when overlapping elements overlap; use Porter-Duff compositing
- [ ] **Implement emoji rendering via fontconfig** — Use fontconfig to find and load color emoji fonts; detect emoji ranges and use emoji font instead of serif fallback

### 🟠 Low (Networking)

- [ ] **Implement HTTP keep-alive** — Reuse TCP connections for multiple requests to same origin; connection pool with max connections per host
- [ ] **Implement conditional GET with If-Modified-Since** — Send Last-Modified header on repeat requests; handle 304 Not Modified; use cached content on 304
- [ ] **Implement `Accept-Encoding: gzip, deflate, br`** — Accept compressed responses; automatically decompress based on Content-Encoding header

### 🟠 Low (Accessibility)

- [ ] **Implement `role` attribute rendering** — ARIA roles affect semantic meaning; role="button" makes div clickable; role="navigation" marks landmark
- [ ] **Implement `<summary>` and `<details>` toggle** — Summary is visible label; clicking toggles details content visibility; use CSS display toggle or explicit state
- [ ] **Implement `<bdo dir="rtl">`** — Bidirectional text override; bdo with dir="rtl" reverses the direction of text content

### 🟠 Low (Testing & QA)

- [ ] **html5lib test corpus integration** — Download html5lib test files; write test harness; run tokenizer/parser tests; fix mismatches
- [ ] **Visual screenshot regression tests** — Use chromium to capture ground-truth screenshots; diff against ViBrowsing renders; report pixel-level regressions

### 🟢 Medium (Text & Typography)

- [ ] **Implement `letter-spacing` drawing** — Add letter-spacing to character glyph advancement; apply between every character pair during text layout
- [ ] **Implement `word-spacing` drawing** — Add word-spacing between word boundaries; CSS default is normal (0) but explicit values modify word gap
- [ ] **Implement `text-indent` drawing** — First-line indentation; apply extra horizontal offset before first character of first line only
- [ ] **Implement `tab-size` CSS property** — Specify tab character width in spaces; default is 8; affects pre-formatted text rendering

### 🟡 High (CSS Properties)

- [ ] **Implement CSS `transform` drawing** — Apply rotate(), scale(), translate(), skew(), matrix() transforms using transform-box and transform-origin; compose into affine matrix
- [ ] **Implement CSS `transform-origin` parsing** — Parse transform-origin as x/y/z keywords, lengths, or percentages; resolve relative to element's bounding box
- [ ] **Implement CSS `transform-box` drawing** — Use view-box or fill-box as transform reference box; affects how transforms are centered/applied

### 🟡 High (Events & Input)

- [ ] **Implement `wheel` event** — Track mouse wheel scrolling; dispatch wheel event on scrollable elements; default action scrolls content
- [ ] **Implement `compositionstart/compositionupdate/compositionend` events** — IME composition events for CJK input; prevent double-insertion of composed characters
- [ ] **Implement `beforeinput` event** — Fire before input is inserted; supports `getTargetRanges()` for range-based input; can call preventDefault()

### 🟢 Medium (Media & Embeds)

- [ ] **Implement `<video>` element** — Video element with src attribute; use ffmpeg to decode frames; render current frame to canvas
- [ ] **Implement `<audio>` element** — Audio element with play/pause controls; use audio library to decode and play audio stream
- [ ] **Implement `<embed>` element** — Generic embedded content; detect MIME type; for unsupported types show fallback content or icon

### 🟡 High (Layout)

- [ ] **Implement `position: sticky`** — Sticky positioning acts like relative until scrolled past a threshold; then fixes like absolute within containing block
- [ ] **Implement `clip: auto/rect(...)`** — CSS clip property for positioned elements; clip restricts visible area; rect(top, right, bottom, left) syntax
- [ ] **Implement `contain: layout/style/paint`** — CSS contain property hints browser about independent rendering; layout = size changes don't affect children

### 🟠 Low (HTML Elements)

- [ ] **Implement `<abbr>` element** — Abbreviation element; render with dotted underline; optional title attribute for full expansion on hover
- [ ] **Implement `<address>` element** — Contact information; typically styled in italics; represents author/owner contact details for nearest article
- [ ] **Implement `<time>` element** — Time element with datetime attribute; can use machine-readable format; display formatting independent of semantic value

### 🟡 High (JavaScript APIs)

- [ ] **Implement `element.scrollIntoView()`** — Scroll element into viewport; align to top/bottom/start/end; supports behavior: smooth/auto options
- [ ] **Implement `window.scrollBy()`** — Scroll window by delta pixels; scrolls viewport; does not fire scroll event during scroll (only after)
- [ ] **Implement `document.createTreeWalker()`** — TreeWalker for traversing DOM with filters; supports whatToShow flags and NodeFilter callback

### 🟠 Low (Rendering)

- [ ] **Implement `caret-color` CSS property** — Color of text insertion caret; render as vertical bar in caret-color at cursor position during text input
- [ ] **Implement `cursor: grab/grabbing`** — Grab cursor on hover over draggable elements; grabbing cursor when mousedown on draggable
- [ ] **Implement `text-rendering: optimizeLegibility`** — Hint to font renderer; optimizeLegibility enables ligatures and kerning; affects text measurement

### 🟢 Medium (Networking)

- [ ] **Implement HTTP/2 server push** — Accept server push promises; cache pushed resources; serve from cache without extra round-trip
- [ ] **Implement `fetch()` with streaming responses** — Stream response body as it arrives; update progress; allow cancellation mid-download
- [ ] **Implement `navigator.connection`** — NetworkInformation API; expose effectiveType, downlink, rtt, saveData from underlying connection info

### 🟠 Low (Debugging)

- [ ] **Add DevTools protocol for CDP** — Chrome DevTools Protocol support; enable remote debugging; expose DOM, CSS, Network, Page domains
- [ ] **Console panel in HUD** — Show console.log/warn/error messages in overlay HUD; click to expand; copy message text

---

## 🆕 New Items (2026-04-09 Sprint)

### 🔴 Critical (Parser/Rendering)

- [ ] **Fix CSS `font-family` inheritance** — font-family should properly cascade from parent to child elements; ensure inheritedProperties includes font-family
- [ ] **Fix CSS `font-size` em unit resolution** — font-size: 1.2em inside a parent with font-size: 20px should resolve to 24px; currently may use body default 16px
- [ ] **Fix inline box baseline calculation** — inline text boxes should share a common baseline for vertical-align: middle/bottom; each text box currently uses its own baseline

### 🟡 High (CSS Properties)

- [ ] **Implement CSS `writing-mode: vertical-rl/vertical-lr`** — Vertical text layout for CJK; inline boxes stack vertically; text flows top-to-bottom or bottom-to-top
- [x] ~~Implement CSS `color-scheme`~~ (2026-04-10 sprint) — ColorSchemeType with Light/Dark/Only parsing; defaults stored in style
- [x] ~~Implement CSS `accent-color`~~ (2026-04-10 sprint) — AccentColorType with IsAuto and Color parsing; stored in style defaults
- [x] ~~Implement CSS `scrollbar-width` and `scrollbar-color`~~ (2026-04-10 sprint) — ScrollbarWidthType (auto/thin/none) and ScrollbarColor (thumb/track colors) parsed and stored in style
- [x] ~~Implement CSS `@layer`~~ (2026-04-11 sprint) — Cascade layers for organizing CSS rules; @layer name { } and anonymous @layer { }; earlier layers have lower priority; rules outside layers have highest priority
- [x] ~~Implement CSS `@property`~~ (2026-04-11 sprint) — Custom properties with type checking; @property --name { syntax: '<type>'; inherits: true/false; initial-value: '<value>' }

### 🟡 High (Events & Input)

- [x] ~~Implement `scroll` event~~ (2026-04-11 sprint) — Fire scroll event on scrollable elements when content is scrolled; Box.ScrollBy dispatches scroll event; EventTarget system in internal/js/
- [x] ~~Implement `wheel` event~~ (2026-04-11 sprint) — Track mouse wheel input; dispatch wheel event with deltaX/deltaY/deltaZ; default action scrolls content; internal/window/event.go
- [x] ~~Implement `input` event~~ (2026-04-11 sprint) — Fire input event when input/textarea value changes; Box.SetValue dispatches input event
- [ ] **Implement `change` event** — Fire change event on form elements when value changes and focus is lost
- [ ] **Implement `beforeinput` event** — Fire before input is inserted; can call preventDefault() to cancel; supports getTargetRanges()
- [ ] **Implement `drag` and `drop` events** — dragstart, drag, dragenter, dragover, dragleave, drop, dragend for native drag and drop API

### 🟡 High (JavaScript APIs)

- [x] ~~Implement `DOMParser` API~~ (2026-04-10 sprint) — Parse HTML/XML strings into DOM documents; DOMParser.parseFromString(string, mimeType)
- [x] ~~Implement `URL` and `URLSearchParams`~~ (2026-04-10 sprint) — URL class with pathname, search, hash, etc.; URLSearchParams for query string manipulation
- [x] ~~Implement `AbortController` and `AbortSignal`~~ (2026-04-10 sprint) — Cancel fetch requests and other operations; AbortController.signal
- [x] ~~Implement `History` API~~ (2026-04-10 sprint) — window.history.pushState(state, title, url); replaceState; popstate event on back/forward
- [x] ~~Implement `navigator.clipboard`~~ (2026-04-10 sprint) — Clipboard API for reading/writing text; navigator.clipboard.readText() and writeText()
- [x] ~~Implement `navigator.storage`~~ (2026-04-10 sprint) — StorageManager API; estimate() for storage quota; persist() for persistent storage permission

### 🟡 High (Media & Embeds)

- [ ] **Implement `<video>` with playback controls** — Video element with play/pause/seek/volume; use ffmpeg or external library to decode frames
- [ ] **Implement `<audio>` with controls** — Audio element with play/pause/seek/volume controls; waveform visualization

### 🟢 Medium (Performance)

- [ ] **CSS selector caching** — Cache selector match results; invalidate on DOM mutations; avoid re-matching unchanged subtrees
- [ ] **Incremental layout** — Layout visible viewport first; defer off-screen content; update layout on scroll for long documents

### 🟢 Medium (Text & Typography)

- [ ] **Implement `hyphens: auto`** — Automatic hyphenation using U+00AD soft hyphen; detect word boundaries; apply hyphens CSS property
- [ ] **Implement `text-wrap: balance/pretty`** — text-wrap: balance for balanced text wrapping in headings; pretty for optimized word breaks
- [ ] **Implement `font-synthesis`** — Control whether browsers synthesize missing font variants (bold, italic); none/synthesis/style/weight

### 🟠 Low (i18n & i10n)

- [ ] **Implement `Accept-Language` header** — Send preferred languages to servers based on navigator.language
- [ ] **Implement `<bdi>` and `<bdo>` properly** — Bidirectional text isolation and override for mixed LTR/RTL content
- [ ] **Implement `lang` attribute inheritance** — Document language from `<html lang="en">` should cascade and affect :lang() selector

### 🟠 Low (Canvas/Drawing)

- [x] ~~Implement `background-blend-mode` parsing~~ (2026-04-10 sprint) — BlendModeType constants; ParseBlendMode and IsValidBlendMode functions; supported in CSS.supports()
- [x] ~~Implement `mix-blend-mode` parsing~~ (2026-04-10 sprint) — mix-blend-mode already in style defaults; blend mode functions ready for canvas drawing implementation
- [ ] **Implement CSS `scroll-behavior`** — Controls smooth scrolling (auto/smooth); stored in style defaults (auto is default)

### 🟠 Low (Networking)

- [ ] **Implement `fetch()` API with streaming** — fetch(url) with ReadableStream response body; support request/response headers
- [ ] **Implement `Content-Encoding: gzip/deflate/br`** — Automatically decompress compressed responses from servers based on Content-Encoding header

### 🟠 Low (CSS Functions)

- [ ] **Implement `color-mix()` function** — Mix two colors together using specified color space (e.g., color-mix(in srgb, red, blue))
- [ ] **Implement `light-dark()` function** — light-dark(color1, color2) for automatic light/dark mode color selection

---

## 🆕 New Items (2026-04-10 Sprint)

### 🟡 High (Layout/Rendering)

- [ ] **Implement CSS `transform` drawing** — Apply rotate(), scale(), translate(), skew(), matrix() transforms to elements during rendering; use affine matrix transformation when drawing boxes
- [ ] **Implement CSS `transform-origin` drawing** — Parse and apply transform-origin as center/percentage/length for rotate/scale reference point
- [ ] **Implement CSS `writing-mode: vertical-rl/vertical-lr`** — Vertical text layout for CJK; inline boxes stack vertically instead of horizontally
- [ ] **Implement `<thead>`, `<tbody>`, `<tfoot>` table sections** — Proper table section rendering order; these elements should organize table rows visually

### 🟡 High (CSS Properties)

- [ ] **Implement CSS `@layer`** — Cascade layers for organizing CSS rules; @layer directive with named layers for specificity control
- [ ] **Implement CSS `@property`** — Custom properties with type checking; @property --name { syntax: <type>; inherits: true/false; initial-value: <value> }
- [ ] **Implement CSS `text-wrap: balance/pretty`** — text-wrap: balance for balanced text wrapping in headings; pretty for optimized word breaks
- [ ] **Implement CSS `contain: layout/style/paint`** — CSS contain property hints browser about independent rendering; layout = size changes don't affect children

### 🟡 High (Events & Input)

- [ ] **Implement `scroll` event** — Fire scroll event on scrollable elements when content is scrolled; dispatch to registered scroll event listeners
- [ ] **Implement `wheel` event** — Track mouse wheel input; dispatch wheel event with deltaX/deltaY/deltaZ; default action scrolls content
- [ ] **Implement `input` event** — Fire input event when input/textarea value changes; fires on every keystroke, paste, cut
- [ ] **Implement `change` event** — Fire change event on form elements when value changes and focus is lost
- [ ] **Implement `drag` and `drop` events** — dragstart, drag, dragenter, dragover, dragleave, drop, dragend for native drag and drop API

### 🟢 Medium (Media & Embeds)

- [ ] **Implement `<video>` with playback controls** — Video element with play/pause/seek/volume; use ffmpeg or external library to decode frames; render current frame
- [ ] **Implement `<audio>` with controls** — Audio element with play/pause/seek/volume controls; waveform visualization
- [ ] **Implement `<embed>` element** — Generic embedded content; detect MIME type; for unsupported types show fallback content or icon

### 🟢 Medium (Performance)

- [ ] **CSS selector caching** — Cache selector match results; invalidate on DOM mutations; avoid re-matching unchanged subtrees
- [ ] **Incremental layout** — Layout visible viewport first; defer off-screen content; update layout on scroll for long documents
- [ ] **Parallel layout for independent subtrees** — If a container has multiple independent block children, layout them concurrently using goroutines

### 🟠 Low (Canvas/Drawing)

- [ ] **Implement `background-blend-mode` drawing** — Apply blend mode when drawing background-image over background-color using multiply, screen, overlay, etc.
- [ ] **Implement `mix-blend-mode` on positioned elements** — Apply blend mode when overlapping elements overlap; use Porter-Duff compositing
- [ ] **Implement `opacity` per draw call** — Apply alpha blending per element not just whole box; respect element opacity on individual draw calls

### 🟠 Low (i18n & i10n)

- [ ] **Implement `Accept-Language` header** — Send preferred languages to servers based on navigator.language
- [ ] **Implement `lang` attribute inheritance** — Document language from `<html lang="en">` should cascade and affect :lang() selector
- [ ] **Implement `<bdi>` element** — Bidirectional text isolation; creates a separate embedding level for its content

### 🟠 Low (Networking)

- [ ] **Implement HTTP/2 support** — Upgrade to HTTP/2 for multiplexed requests on a single connection
- [ ] **Implement `Content-Encoding: br` (brotli)** — Support brotli decompression in addition to gzip/deflate

### 🟠 Low (Testing & QA)

- [ ] **Property-based testing with go-fuzz** — Generate random HTML/CSS combinations and verify no panics or hangs
- [ ] **Large document stress test** — Parse and render 10MB+ HTML file; verify memory usage stays under 500MB and completes within 30s

### 🟠 Low (CSS Functions)

- [ ] **Implement `color-mix()` function** — Mix two colors together using specified color space (e.g., color-mix(in srgb, red, blue))
- [ ] **Implement `light-dark()` function** — light-dark(color1, color2) for automatic light/dark mode color selection
- [ ] **Implement `OKLCH` color notation** — OKLCH color space support for modern CSS colors (oklch(), oklab())
- [ ] **Implement `hwb()` color notation** — HWB (Hue, Whiteness, Blackness) color notation

---

## 🆕 New Items (2026-04-11 Sprint)

### 🔴 Critical (Parser/Rendering)

- [ ] **Fix inline box baseline calculation** — inline text boxes should share a common baseline; vertical-align: middle/bottom should position relative to that baseline; currently each text box uses its own baseline
- [ ] **Fix table cell border-collapse rendering** — adjacent table cells should share borders (border-collapse behavior), currently each cell renders its own border separately; cells at shared edges should only draw outer half of border
- [ ] **Fix float clearing logic** — blocks that clear:left/right/both should properly position below float; ctx.Y must be updated after clearing so subsequent blocks don't overlap floats

### 🟡 High (Layout/Rendering)

- [ ] **Implement CSS `transform-box` drawing** — Use view-box or fill-box as transform reference box; affects how transforms are centered/applied; currently transform parsing exists but transform-box may not be respected
- [ ] **Implement CSS `transform-origin` drawing** — Parse transform-origin as x/y/z keywords, lengths, or percentages; resolve relative to element's bounding box for rotate/scale reference
- [ ] **Implement CSS `writing-mode: vertical-rl/vertical-lr`** — Vertical text layout for CJK; inline boxes stack vertically; text flows top-to-bottom (rl) or bottom-to-top (lr)
- [ ] **Implement `<thead>`, `<tbody>`, `<tfoot>` table sections** — Proper table section rendering order; thead at top, tbody in middle, tfoot at bottom regardless of source order
- [ ] **Implement CSS `hyphens: auto/manual`** — Automatic hyphenation using U+00AD soft hyphen; detect word boundaries; apply hyphens CSS property
- [ ] **Implement CSS `text-justify: inter-word/inter-character`** — Justification algorithm for better text alignment; inter-character adjusts spacing between characters for CJK text

### 🟡 High (CSS Properties)

- [ ] **Implement CSS `@layer`** — Cascade layers for organizing CSS rules with explicit precedence order; earlier layers have lower priority; rules outside layers have highest priority
- [ ] **Implement CSS `@property`** — Custom properties with type checking; @property --name { syntax: '<type>'; inherits: true/false; initial-value: '<value>' }
- [ ] **Implement CSS `text-wrap: balance/pretty`** — text-wrap: balance for balanced text wrapping in headings; pretty for optimized word breaks
- [ ] **Implement CSS `contain: layout/style/paint`** — CSS contain property hints browser about independent rendering; layout = size changes don't affect children

### 🟡 High (Events & Input)

- [ ] **Implement `scroll` event** — Fire scroll event on scrollable elements when content is scrolled; dispatch to registered scroll event listeners
- [ ] **Implement `wheel` event** — Track mouse wheel input; dispatch wheel event with deltaX/deltaY/deltaZ; default action scrolls content
- [ ] **Implement `input` event** — Fire input event when input/textarea value changes; fires on every keystroke, paste, cut
- [ ] **Implement `change` event** — Fire change event on form elements when value changes and focus is lost
- [ ] **Implement `beforeinput` event** — Fire before input is inserted; can call preventDefault() to cancel; supports getTargetRanges()
- [ ] **Implement `drag` and `drop` events** — dragstart, drag, dragenter, dragover, dragleave, drop, dragend for native drag and drop API

### 🟢 Medium (Media & Embeds)

- [ ] **Implement `<video>` with playback controls** — Video element with play/pause/seek/volume; show video frame or controls UI; no actual video playback required, just visual
- [ ] **Implement `<audio>` with controls** — Audio element with play/pause/seek/volume controls; waveform visualization
- [ ] **Implement `<embed>` element** — Generic embedded content; detect MIME type; for unsupported types show fallback content or icon

### 🟢 Medium (Performance)

- [ ] **CSS selector caching** — Cache selector match results; invalidate on DOM mutations; avoid re-matching unchanged subtrees
- [ ] **Incremental layout** — Layout visible viewport first; defer off-screen content; update layout on scroll for long documents
- [ ] **Parallel layout for independent subtrees** — If a container has multiple independent block children, layout them concurrently using goroutines

### 🟠 Low (Canvas/Drawing)

- [ ] **Implement `background-blend-mode` drawing** — Apply blend mode when drawing background-image over background-color using multiply, screen, overlay, etc.
- [ ] **Implement `mix-blend-mode` on positioned elements** — Apply blend mode when overlapping elements overlap; use Porter-Duff compositing
- [ ] **Implement `opacity` per draw call** — Apply alpha blending per element not just whole box; respect element opacity on individual draw calls

### 🟠 Low (i18n & i10n)

- [ ] **Implement `Accept-Language` header** — Send preferred languages to servers based on navigator.language
- [ ] **Implement `lang` attribute inheritance** — Document language from `<html lang="en">` should cascade and affect :lang() selector
- [ ] **Implement `<bdi>` element** — Bidirectional text isolation; creates a separate embedding level for its content

### 🟠 Low (Networking)

- [ ] **Implement HTTP/2 support** — Upgrade to HTTP/2 for multiplexed requests on a single connection
- [ ] **Implement `Content-Encoding: br` (brotli)** — Support brotli decompression in addition to gzip/deflate

### 🟠 Low (Testing & QA)

- [ ] **Property-based testing with go-fuzz** — Generate random HTML/CSS combinations and verify no panics or hangs
- [ ] **Large document stress test** — Parse and render 10MB+ HTML file; verify memory usage stays under 500MB and completes within 30s

### 🟠 Low (CSS Functions)

- [ ] **Implement `color-mix()` function** — Mix two colors together using specified color space (e.g., color-mix(in srgb, red, blue))
- [ ] **Implement `light-dark()` function** — light-dark(color1, color2) for automatic light/dark mode color selection
- [ ] **Implement `OKLCH` color notation** — OKLCH color space support for modern CSS colors (oklch(), oklab())
- [ ] **Implement `hwb()` color notation** — HWB (Hue, Whiteness, Blackness) color notation

---

## 🆕 New Items (2026-04-07 Sprint) — Already Implemented
