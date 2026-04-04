# ViBrowsing Backlog

## 🔴 Critical (Parser/Rendering)

- [ ] **Fix HTML parser double-html/body bug** — tokenizer correctly skips synthetic root tags (html/head/body opening/closing) but parser still creates duplicate nodes. Move all synthetic tag skipping to tokenizer's StartTag case; remove synthetic element bootstrapping from parser entirely
- [ ] **Fix foster parenting** — when </table> closes a table, unclosed `<tr>`/`<td>` children should be moved to the table's parent, not silently dropped
- [ ] **Fix unclosed tag handling** — when encountering a closing tag with no matching open tag, don't silently drop the close. Compare behavior against html5lib reference
- [ ] **Implement entity decoding** — `&amp;` `&lt;` `&gt;` `&quot;` `&nbsp;` in text nodes must decode to `&` `<` `>` `"` `` (space). Also numeric: `&#65;` and `&#x41;`
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
