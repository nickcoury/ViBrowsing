package layout

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nickcoury/ViBrowsing/internal/css"
	"github.com/nickcoury/ViBrowsing/internal/html"
)

// BoxType represents the type of layout box.
type BoxType int

const (
	BlockBox BoxType = iota
	InlineBox
	InlineBlock
	TextBox
	FlexBox
	PositionedBox
	ImageBox
	ListItemBox
	HorizontalRuleBox
	TableBox
	TableRowBox
	TableCellBox
	TableCaptionBox
	TableSectionBox
	ButtonBox
	SelectBox
	TextAreaBox
	LabelBox
	MediaBox
	SlotBox
	DialogBox
	MeterBox
	ProgressBox
	RubyBox
	RubyTextBox
	RubyBaseBox
	BdiBox
	BdoBox
	AbbrBox
	MarkBox
	ColumnBox // Multi-column layout column
	DetailsBox
	SummaryBox
	PictureBox
	FieldsetBox
	LegendBox
	MapBox
	DataBox
)

// Box represents a CSS box in the layout tree.
type Box struct {
	Type       BoxType
	Node       *html.Node // source DOM node
	Parent     *Box       // parent box in layout tree

	// Content area
	ContentX, ContentY float64
	ContentW, ContentH float64

	// Aspect ratio (width/height), e.g., 16/9 = 1.78
	AspectRatio float64

	// Box dimensions (CSS box model)
	MarginTop, MarginRight, MarginBottom, MarginLeft css.Length
	BorderTop, BorderRight, BorderBottom, BorderLeft css.Length
	PaddingTop, PaddingRight, PaddingBottom, PaddingLeft css.Length

	// Computed style
	Style map[string]string

	// Children
	Children []*Box

	// Table-specific properties
	ColSpan    int // number of columns this cell spans (1 = normal)
	RowSpan    int // number of rows this cell spans (1 = normal)
	ColumnIndex int // column index for table cells

	// Map-specific properties (for <map> and <area> elements)
	MapName string // name attribute of the map element
	AreaShape string // shape attribute of area: rect, circle, poly, default
	AreaCoords []int // coordinates for area element shape

	// Data element property
	DataValue string // value attribute of <data> element
}

// GetWidth returns the computed width of the content area.
func (b *Box) GetWidth() float64 {
	if w, ok := b.Style["width"]; ok {
		l := css.ParseLength(w)
		if !l.IsAuto {
			return l.Value
		}
	}
	return b.ContentW
}

// GetHeight returns the computed height of the content area.
func (b *Box) GetHeight() float64 {
	if h, ok := b.Style["height"]; ok {
		l := css.ParseLength(h)
		return l.Value
	}
	return b.ContentH
}

// GetMargin returns the margin lengths.
func (b *Box) GetMargin() (top, right, bottom, left css.Length) {
	return b.MarginTop, b.MarginRight, b.MarginBottom, b.MarginLeft
}

// GetPadding returns the padding lengths.
func (b *Box) GetPadding() (top, right, bottom, left css.Length) {
	return b.PaddingTop, b.PaddingRight, b.PaddingBottom, b.PaddingLeft
}

// GetBorder returns the border widths.
func (b *Box) GetBorder() (top, right, bottom, left css.Length) {
	return b.BorderTop, b.BorderRight, b.BorderBottom, b.BorderLeft
}

// TotalWidth returns the full width including margin, border, padding.
func (b *Box) TotalWidth() float64 {
	m := b.MarginLeft.Value + b.MarginRight.Value
	w := b.ContentW
	t := b.BorderLeft.Value + b.BorderRight.Value
	p := b.PaddingLeft.Value + b.PaddingRight.Value
	return m + w + t + p
}

// TotalHeight returns the full height including margin, border, padding.
func (b *Box) TotalHeight() float64 {
	m := b.MarginTop.Value + b.MarginBottom.Value
	h := b.ContentH
	t := b.BorderTop.Value + b.BorderBottom.Value
	p := b.PaddingTop.Value + b.PaddingBottom.Value
	return m + h + t + p
}

// OuterWidth returns the width from outer edge of margin to outer edge of margin.
func (b *Box) OuterWidth() float64 {
	return b.MarginLeft.Value + b.ContentW + b.BorderLeft.Value +
		b.PaddingLeft.Value + b.PaddingRight.Value + b.BorderRight.Value +
		b.MarginRight.Value
}

// OuterHeight returns the height from outer edge of margin to outer edge of margin.
func (b *Box) OuterHeight() float64 {
	return b.MarginTop.Value + b.ContentH + b.BorderTop.Value +
		b.PaddingTop.Value + b.PaddingBottom.Value + b.BorderBottom.Value +
		b.MarginBottom.Value
}

// BuildLayoutTree converts a DOM tree to a layout tree.
func BuildLayoutTree(doc *html.Node, rules []css.Rule) *Box {
	// Find body
	body := findBody(doc)
	if body == nil {
		return nil
	}

	// Build layout tree recursively
	box := buildBox(body, rules, 0, nil)
	return box
}

func findBody(n *html.Node) *html.Node {
	if n.Type == html.NodeElement && n.TagName == "body" {
		return n
	}
	for _, child := range n.Children {
		if b := findBody(child); b != nil {
			return b
		}
	}
	return nil
}

func buildBox(node *html.Node, rules []css.Rule, depth int, parentStyle map[string]string) *Box {
	if node == nil {
		return nil
	}

	// Text nodes inherit parent style
	if node.Type == html.NodeText {
		// Skip text inside template (template content is not rendered)
		if node.InsideTemplate() {
			return nil
		}
		style := map[string]string{}
		if parentStyle != nil {
			for k, v := range parentStyle {
				style[k] = v
			}
		}
		return &Box{
			Type:  TextBox,
			Node:  node,
			Style: style,
			Children: []*Box{},
		}
	}

	// Skip non-element nodes
	if node.Type != html.NodeElement {
		return nil
	}

	// Skip template element itself and its children (template content is not rendered)
	// Template element is parsed into DOM but not rendered
	if node.TagName == "template" || node.InsideTemplate() {
		return nil
	}

	// Compute style
	class := node.GetAttribute("class")
	id := node.GetAttribute("id")
	tagName := node.TagName
	inlineDecls := css.ParseInline(node.GetAttribute("style"))
	style := css.ComputeStyle(tagName, class, id, inlineDecls, rules)

	// Also apply node-aware matching for attribute selectors
	nodeStyle := css.ComputeStyleForNode(node, rules)
	// Merge nodeStyle into style (nodeStyle has attribute selectors applied)
	for k, v := range nodeStyle {
		style[k] = v
	}

	display := style["display"]

	// Create box
	box := &Box{
		Type:  BlockBox,
		Node:  node,
		Style: style,
	}

	// Override display type
	switch display {
	case "block":
		box.Type = BlockBox
	case "inline":
		box.Type = InlineBox
	case "none":
		// Hidden — don't create a box
		return nil
	}

	// Default block elements
	switch tagName {
	case "div", "p", "h1", "h2", "h3", "h4", "h5", "h6",
		"ul", "ol", "form", "pre",
		"blockquote", "address", "article", "aside",
		"footer", "header", "main", "nav", "section",
		"figure", "figcaption", "noscript":
		box.Type = BlockBox
	case "table":
		box.Type = TableBox
	case "tr":
		box.Type = TableRowBox
	case "td", "th":
		box.Type = TableCellBox
	case "tbody", "thead", "tfoot":
		box.Type = TableSectionBox
	case "caption":
		box.Type = TableCaptionBox
	case "button":
		box.Type = ButtonBox
	case "select":
		box.Type = SelectBox
	case "textarea":
		box.Type = TextAreaBox
	case "label":
		box.Type = LabelBox
	case "video", "audio":
		box.Type = MediaBox
	case "span", "a", "strong", "em", "b", "i", "code", "small", "br", "wbr", "cite":
		box.Type = InlineBox
	case "img":
		box.Type = ImageBox
	case "li":
		box.Type = ListItemBox
	case "input":
		box.Type = InlineBox
	case "hr":
		box.Type = HorizontalRuleBox
	case "slot":
		box.Type = SlotBox
		box.Style["display"] = "contents"
	case "dialog":
		box.Type = DialogBox
		box.Style["display"] = "block"
	case "meter":
		box.Type = MeterBox
		box.Style["display"] = "inline-block"
	case "progress":
		box.Type = ProgressBox
		box.Style["display"] = "inline-block"
	case "ruby":
		box.Type = RubyBox
		// Ruby is inline by default
	case "rt", "rp":
		box.Type = RubyTextBox
	case "bdi":
		box.Type = BdiBox
	case "bdo":
		box.Type = BdoBox
	case "abbr":
		box.Type = AbbrBox
		// Default abbr styling: dotted underline
		if box.Style["text-decoration-line"] == "" {
			box.Style["text-decoration-line"] = "underline"
			box.Style["text-decoration-style"] = "dotted"
		}
	case "details":
		box.Type = DetailsBox
	case "summary":
		box.Type = SummaryBox
	case "picture":
		box.Type = PictureBox
	case "fieldset":
		box.Type = FieldsetBox
	case "legend":
		box.Type = LegendBox
		// Legend is block-level by default
	case "mark":
		box.Type = MarkBox
		// Default mark styling
		box.Style["background-color"] = "yellow"
		box.Style["color"] = "black"
	case "output":
		box.Type = InlineBox
		// output is an inline output element
	case "del", "ins":
		box.Type = InlineBox
		// del/ins are inline by default (like u and s)
	case "data":
		box.Type = DataBox
		// data element - machine-readable data wrapper with value attribute
	case "map":
		box.Type = MapBox
		// map element - client-side image map (container for area elements)
		// Maps are display: inline but define regions for use by img usemap
	case "area":
		box.Type = InlineBox
		// area element - clickable region in image map
		// areas are void elements that define shapes within a map
	}

	// Flex container
	if display == "flex" {
		box.Type = FlexBox
	}

	// Table display types
	if display == "table" {
		box.Type = TableBox
	} else if display == "table-row" {
		box.Type = TableRowBox
	} else if display == "table-cell" {
		box.Type = TableCellBox
	} else if display == "table-caption" {
		box.Type = TableCaptionBox
	} else if display == "table-row-group" || display == "table-header-group" || display == "table-footer-group" {
		box.Type = TableSectionBox
	}

	// Positioned elements
	if position := style["position"]; position == "absolute" || position == "relative" || position == "fixed" {
		box.Type = PositionedBox
	}

	// Parse box model properties
	box.MarginTop = css.ParseLength(style["margin-top"])
	box.MarginRight = css.ParseLength(style["margin-right"])
	box.MarginBottom = css.ParseLength(style["margin-bottom"])
	box.MarginLeft = css.ParseLength(style["margin-left"])
	box.PaddingTop = css.ParseLength(style["padding-top"])
	box.PaddingRight = css.ParseLength(style["padding-right"])
	box.PaddingBottom = css.ParseLength(style["padding-bottom"])
	box.PaddingLeft = css.ParseLength(style["padding-left"])

	borderWidth := css.ParseLength(style["border-width"])
	box.BorderTop = borderWidth
	box.BorderRight = borderWidth
	box.BorderBottom = borderWidth
	box.BorderLeft = borderWidth

	// Image element default size: if no width/height specified, use 150x150
	if tagName == "img" {
		if _, ok := style["width"]; !ok {
			box.Style["width"] = "150"
		}
		if _, ok := style["height"]; !ok {
			box.Style["height"] = "150"
		}
	}

	// Apply aspect-ratio when width is set but height is auto
	if arStr, ok := style["aspect-ratio"]; ok && arStr != "auto" {
		ratio := css.ParseAspectRatio(arStr)
		if ratio > 0 {
			box.AspectRatio = ratio
			// Check if width is set and height is auto
			widthLen := css.ParseLength(style["width"])
			heightLen := css.ParseLength(style["height"])
			if !widthLen.IsAuto && heightLen.IsAuto {
				// Width is set, height is auto - compute height from aspect ratio
				// Use a default width if ContentW not yet set (will be computed later)
				if box.ContentW > 0 {
					box.ContentH = box.ContentW / ratio
				} else {
					// Width as length value
					box.ContentW = widthLen.Value
					box.ContentH = widthLen.Value / ratio
				}
			}
		}
	}

	// Table cell: parse colspan and rowspan attributes
	if tagName == "td" || tagName == "th" {
		if colspan := node.GetAttribute("colspan"); colspan != "" {
			if v, err := strconv.Atoi(colspan); err == nil && v > 1 {
				box.ColSpan = v
			} else {
				box.ColSpan = 1
			}
		} else {
			box.ColSpan = 1
		}
		if rowspan := node.GetAttribute("rowspan"); rowspan != "" {
			if v, err := strconv.Atoi(rowspan); err == nil && v > 1 {
				box.RowSpan = v
			} else {
				box.RowSpan = 1
			}
		} else {
			box.RowSpan = 1
		}
	}

	// Recurse for children (pass our computed style as parent style)
	for _, child := range node.Children {
		childBox := buildBox(child, rules, depth+1, style)
		if childBox != nil {
			childBox.Parent = box
			box.Children = append(box.Children, childBox)
		}
	}

	// Handle picture element: select the best source based on viewport
	if tagName == "picture" {
		box = resolvePictureSource(box, 800) // default viewport width
	}

	// Handle details element: if open attribute is missing or "open", show children
	// The summary is the first summary child; content is other children
	if tagName == "details" {
		isOpen := node.GetAttribute("open")
		// If not open, we still include the summary but mark details as collapsed
		// For now, just pass through - rendering can handle visibility
		if isOpen == "" {
			// Not open - mark that content should be hidden
			box.Style["_details_open"] = "false"
		} else {
			box.Style["_details_open"] = "true"
		}
	}

	// Handle abbr element: store title attribute for tooltip rendering
	if tagName == "abbr" {
		title := node.GetAttribute("title")
		if title != "" {
			box.Style["_abbr_title"] = title
		}
	}

	// Handle map element: store name attribute for usemap reference
	if tagName == "map" {
		box.MapName = node.GetAttribute("name")
	}

	// Handle area element: parse shape and coords attributes for image map regions
	if tagName == "area" {
		box.AreaShape = strings.ToLower(node.GetAttribute("shape"))
		if box.AreaShape == "" {
			box.AreaShape = "rect" // default shape
		}
		// Parse coords attribute (comma or space separated integers)
		coordsStr := node.GetAttribute("coords")
		if coordsStr != "" {
			coordsStr = strings.ReplaceAll(coordsStr, ",", " ")
			parts := strings.Fields(coordsStr)
			for _, part := range parts {
				if coord, err := strconv.Atoi(part); err == nil {
					box.AreaCoords = append(box.AreaCoords, coord)
				}
			}
		}
	}

	// Handle data element: store value attribute for machine-readable data
	if tagName == "data" {
		box.DataValue = node.GetAttribute("value")
	}

	return box
}

// String returns a formatted string representation of the layout tree.
func (b *Box) String() string {
	return b.stringWithIndent(0)
}

func (b *Box) stringWithIndent(indent int) string {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	var tagName string
	if b.Node != nil {
		tagName = b.Node.TagName
	} else {
		tagName = "(root)"
	}

	var boxType string
	switch b.Type {
	case BlockBox:
		boxType = "block"
	case InlineBox:
		boxType = "inline"
	case InlineBlock:
		boxType = "inline-block"
	case TextBox:
		boxType = "text"
	}

	s := prefix + fmt.Sprintf("[%s] %s (%.0fx%.0f at (%.0f,%.0f))\n",
		boxType, tagName, b.ContentW, b.ContentH, b.ContentX, b.ContentY)

	for _, child := range b.Children {
		s += child.stringWithIndent(indent + 1)
	}

	return s
}

// resolvePictureSource handles the <picture> element by selecting the best <source>
// based on viewport width and media queries. Returns a modified box with a single
// img child (or the original img if no picture sources apply).
func resolvePictureSource(pictureBox *Box, viewportWidth float64) *Box {
	if pictureBox == nil {
		return pictureBox
	}

	// Find all source elements and the img element
	var sources []*Box
	var imgBox *Box

	for _, child := range pictureBox.Children {
		if child.Node != nil && child.Node.TagName == "source" {
			sources = append(sources, child)
		} else if child.Node != nil && child.Node.TagName == "img" {
			imgBox = child
		}
	}

	// If no img, return picture box as-is
	if imgBox == nil {
		return pictureBox
	}

	// Evaluate sources in order (first matching source wins)
	selectedSrc := ""

	for _, source := range sources {
		srcset := source.Node.GetAttribute("srcset")
		media := source.Node.GetAttribute("media")

		// Check if media query matches viewport
		if media != "" && !matchesMediaQuery(media, viewportWidth) {
			continue
		}

		// This source matches - use its srcset
		if srcset != "" {
			selectedSrc = parseSrcset(srcset, viewportWidth)
			break
		}
	}

	// If no source matched, use img's src as fallback
	if selectedSrc == "" {
		selectedSrc = imgBox.Node.GetAttribute("src")
	}

	// Update the img element's src attribute in the DOM
	if selectedSrc != "" {
		imgBox.Node.SetAttribute("src", selectedSrc)
	}

	// Replace picture's children with just the img
	pictureBox.Children = []*Box{imgBox}
	imgBox.Parent = pictureBox

	return pictureBox
}

// matchesMediaQuery checks if a media query string matches the given viewport width.
func matchesMediaQuery(media string, viewportWidth float64) bool {
	// Simple media query parser for common patterns
	media = strings.ToLower(strings.TrimSpace(media))

	// Handle min-width queries
	if strings.Contains(media, "min-width") {
		// Extract pixel value from e.g. "(min-width: 768px)"
		start := strings.Index(media, "min-width")
		if start >= 0 {
			rest := media[start+9:] // skip "min-width"
			rest = strings.TrimLeft(rest, ": ")
			// Extract number
			numStr := ""
			for _, c := range rest {
				if c >= '0' && c <= '9' {
					numStr += string(c)
				} else if c == 'p' { // end of "px"
					break
				}
			}
			if numStr != "" {
				minW := 0
				fmt.Sscanf(numStr, "%d", &minW)
				return viewportWidth >= float64(minW)
			}
		}
	}

	// Handle max-width queries
	if strings.Contains(media, "max-width") {
		start := strings.Index(media, "max-width")
		if start >= 0 {
			rest := media[start+9:]
			rest = strings.TrimLeft(rest, ": ")
			numStr := ""
			for _, c := range rest {
				if c >= '0' && c <= '9' {
					numStr += string(c)
				} else if c == 'p' {
					break
				}
			}
			if numStr != "" {
				maxW := 0
				fmt.Sscanf(numStr, "%d", &maxW)
				return viewportWidth <= float64(maxW)
			}
		}
	}

	// Default: matches
	return true
}

// parseSrcset parses a srcset attribute value and returns the URL that best
// matches the given viewport width. Format: "url1 100w, url2 200w, url3 300w"
func parseSrcset(srcset string, viewportWidth float64) string {
	if srcset == "" {
		return ""
	}

	// Split by comma
	parts := strings.Split(srcset, ",")
	var candidates []struct {
		url   string
		width int
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split to get url and descriptor
		fields := strings.Fields(part)
		if len(fields) < 1 {
			continue
		}

		url := fields[0]
		width := 0

		if len(fields) >= 2 {
			desc := fields[len(fields)-1]
			if strings.HasSuffix(desc, "w") {
				// Width descriptor
				wStr := strings.TrimSuffix(desc, "w")
				if w, err := strconv.Atoi(wStr); err == nil {
					width = w
				}
			} else if strings.HasSuffix(desc, "x") {
				// Density descriptor - convert to width using a base of 1
				dStr := strings.TrimSuffix(desc, "x")
				if d, err := strconv.ParseFloat(dStr, 64); err == nil && d > 0 {
					// Assume base width of 100vw for density
					width = int(100 * d)
				}
			}
		}

		// Default width if not specified
		if width == 0 {
			width = int(viewportWidth)
		}

		candidates = append(candidates, struct {
			url   string
			width int
		}{url, width})
	}

	if len(candidates) == 0 {
		return ""
	}

	// Find the best match: select the smallest image that is >= viewportWidth
	// If none are >= viewportWidth, select the largest
	bestIdx := 0
	bestWidth := candidates[0].width

	for i, cand := range candidates {
		if cand.width >= int(viewportWidth) {
			if bestWidth < int(viewportWidth) || cand.width < bestWidth {
				bestIdx = i
				bestWidth = cand.width
			}
		} else if cand.width > bestWidth {
			bestIdx = i
			bestWidth = cand.width
		}
	}

	return candidates[bestIdx].url
}
