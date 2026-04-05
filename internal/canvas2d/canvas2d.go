package canvas2d

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/freetype/truetype"
)

// Context2D represents the Canvas 2D rendering context.
// It maintains a state stack for save/restore operations.
type Context2D struct {
	// Canvas dimensions
	Width, Height int

	// The underlying pixel buffer
	Pixels *image.RGBA

	// State stack for save/restore
	stateStack []ContextState

	// Current state
	state ContextState

	// fontCache caches loaded fonts
	fontCache map[string]*truetype.Font
	fontMu    sync.RWMutex

	// Default font metrics
	defaultFont     *truetype.Font
	defaultAdvWidth float64
}

// ContextState holds the drawing state for save/restore.
// See https://html.spec.whatwg.org/multipage/canvas.html#drawing-state
type ContextState struct {
	// Transformation matrix
	Transform TransformMatrix

	// Stroke style
	StrokeStyle CSSColor

	// Fill style
	FillStyle CSSColor

	// Font settings
	Font       string
	FontFamily string
	FontSize   float64
	FontWeight string
	FontStyle  string

	// Text alignment
	TextAlign    TextAlign
	TextBaseline TextBaseline
	Direction    Direction

	// Line settings
	LineWidth      float64
	LineCap        LineCap
	LineJoin       LineJoin
	MiterLimit     float64
	LineDashOffset float64
	LineDash       []float64

	// Shadow
	ShadowColor   string
	ShadowBlur    float64
	ShadowOffsetX float64
	ShadowOffsetY float64

	// Compositing
	GlobalAlpha           float64
	GlobalCompositeOperation CompositeOperation

	// Filter effects
	Filter string

	// Current path
	Path Path2D

	// Current hit regions
	HitRegions []HitRegion
}

// TransformMatrix represents a 2D transformation matrix [a c e]
//                                                [b d f]
//                                                [0 0 1]
type TransformMatrix struct {
	A, B, C, D, E, F float64
}

// Identity returns the identity transform matrix.
func IdentityMatrix() TransformMatrix {
	return TransformMatrix{A: 1, B: 0, C: 0, D: 1, E: 0, F: 0}
}

// Multiply returns the result of multiplying this matrix by m.
func (m TransformMatrix) Multiply(m2 TransformMatrix) TransformMatrix {
	return TransformMatrix{
		A: m.A*m2.A + m.C*m2.B,
		B: m.B*m2.A + m.D*m2.B,
		C: m.A*m2.C + m.C*m2.D,
		D: m.B*m2.C + m.D*m2.D,
		E: m.A*m2.E + m.C*m2.F + m2.E,
		F: m.B*m2.E + m.D*m2.F + m2.F,
	}
}

// TransformPoint applies the matrix to a point (x, y).
func (m TransformMatrix) TransformPoint(x, y float64) (float64, float64) {
	return m.A*x + m.C*y + m.E, m.B*x + m.D*y + m.F
}

// Inverse returns the inverse of this matrix.
func (m TransformMatrix) Inverse() TransformMatrix {
	det := m.A*m.D - m.B*m.C
	if det == 0 {
		return IdentityMatrix()
	}
	invDet := 1 / det
	return TransformMatrix{
		A:  m.D * invDet,
		B: -m.B * invDet,
		C: -m.C * invDet,
		D:  m.A * invDet,
		E:  (m.C*m.F - m.D*m.E) * invDet,
		F:  (m.B*m.E - m.A*m.F) * invDet,
	}
}

// Path2D represents a path for drawing operations.
type Path2D struct {
	Commands []PathCommand
}

// PathCommand represents a single command in a path.
type PathCommand struct {
	Type    PathCommandType
	X, Y    float64    // for most commands
	X1, Y1  float64    // control point 1 (quadratic/bezier)
	X2, Y2  float64    // control point 2 (bezier)
	Radius  float64    // for arc
	Rotation float64   // for ellipse
}

// PathCommandType represents the type of a path command.
type PathCommandType int

const (
	PathMoveTo PathCommandType = iota
	PathLineTo
	PathQuadratic
	PathBezier
	PathArc
	PathEllipse
	PathRect
	PathClose
)

// HitRegion represents a hit region for event handling.
type HitRegion struct {
	ID      string
	Path    Path2D
	Control interface{}
}

// TextAlign represents text alignment.
type TextAlign string

const (
	AlignLeft    TextAlign = "left"
	AlignRight   TextAlign = "right"
	AlignCenter  TextAlign = "center"
	AlignStart   TextAlign = "start"
	AlignEnd     TextAlign = "end"
)

// TextBaseline represents text baseline.
type TextBaseline string

const (
	BaselineTop        TextBaseline = "top"
	BaselineHanging    TextBaseline = "hanging"
	BaselineMiddle     TextBaseline = "middle"
	BaselineAlphabetic TextBaseline = "alphabetic"
	BaselineIdeographic TextBaseline = "ideographic"
	BaselineBottom     TextBaseline = "bottom"
)

// Direction represents text direction.
type Direction string

const (
	DirectionInherit Direction = "inherit"
	DirectionLTR     Direction = "ltr"
	DirectionRTL     Direction = "rtl"
)

// LineCap represents line cap style.
type LineCap string

const (
	LineCapButt   LineCap = "butt"
	LineCapRound  LineCap = "round"
	LineCapSquare LineCap = "square"
)

// LineJoin represents line join style.
type LineJoin string

const (
	LineJoinMiter LineJoin = "miter"
	LineJoinRound LineJoin = "round"
	LineJoinBevel LineJoin = "bevel"
)

// CompositeOperation represents compositing operation.
type CompositeOperation string

const (
	CompositeSourceOver    CompositeOperation = "source-over"
	CompositeSourceAtop    CompositeOperation = "source-atop"
	CompositeSourceIn      CompositeOperation = "source-in"
	CompositeSourceOut     CompositeOperation = "source-out"
	CompositeDestinationOver   CompositeOperation = "destination-over"
	CompositeDestinationAtop  CompositeOperation = "destination-atop"
	CompositeDestinationIn     CompositeOperation = "destination-in"
	CompositeDestinationOut    CompositeOperation = "destination-out"
	CompositeLighter       CompositeOperation = "lighter"
	CompositeCopy          CompositeOperation = "copy"
	CompositeXor           CompositeOperation = "xor"
	CompositeMultiply      CompositeOperation = "multiply"
	CompositeScreen        CompositeOperation = "screen"
	CompositeOverlay       CompositeOperation = "overlay"
	CompositeDarken        CompositeOperation = "darken"
	CompositeLighten       CompositeOperation = "lighten"
	CompositeColorDodge    CompositeOperation = "color-dodge"
	CompositeColorBurn     CompositeOperation = "color-burn"
	CompositeHardLight    CompositeOperation = "hard-light"
	CompositeSoftLight    CompositeOperation = "soft-light"
	CompositeDifference    CompositeOperation = "difference"
	CompositeExclusion     CompositeOperation = "exclusion"
)

// CSSColor represents a color that can be either a solid color or a gradient/pattern.
type CSSColor struct {
	Type    string // "solid", "linear-gradient", "radial-gradient", "pattern"
	Color   color.RGBA
	Stops   []GradientStop
	Gradient GradientConfig
}

// GradientStop represents a color stop in a gradient.
type GradientStop struct {
	Offset float64
	Color  color.RGBA
}

// GradientConfig holds gradient parameters.
type GradientConfig struct {
	X0, Y0, X1, Y1 float64 // linear gradient points
	X2, Y2, R2      float64 // radial gradient center and radius
}

// NewContext2D creates a new Canvas 2D context with the given dimensions.
func NewContext2D(width, height int) *Context2D {
	pixels := image.NewRGBA(image.Rect(0, 0, width, height))
	ctx := &Context2D{
		Width:  width,
		Height: height,
		Pixels: pixels,
		stateStack: make([]ContextState, 0),
		state: defaultContextState(),
		fontCache: make(map[string]*truetype.Font),
	}
	return ctx
}

func defaultContextState() ContextState {
	return ContextState{
		Transform:                   IdentityMatrix(),
		StrokeStyle:                CSSColor{Type: "solid", Color: color.RGBA{R: 0, G: 0, B: 0, A: 255}},
		FillStyle:                  CSSColor{Type: "solid", Color: color.RGBA{R: 0, G: 0, B: 0, A: 255}},
		Font:                       "10px sans-serif",
		FontFamily:                 "sans-serif",
		FontSize:                   10,
		FontWeight:                 "normal",
		FontStyle:                  "normal",
		TextAlign:                  TextAlign("start"),
		TextBaseline:               TextBaseline("alphabetic"),
		Direction:                  DirectionInherit,
		LineWidth:                  1.0,
		LineCap:                    LineCapButt,
		LineJoin:                   LineJoinMiter,
		MiterLimit:                 10.0,
		LineDashOffset:             0,
		LineDash:                   nil,
		ShadowColor:                "rgba(0,0,0,0)",
		ShadowBlur:                 0,
		ShadowOffsetX:              0,
		ShadowOffsetY:              0,
		GlobalAlpha:                1.0,
		GlobalCompositeOperation:   CompositeSourceOver,
		Filter:                     "none",
		Path:                       Path2D{},
		HitRegions:                 nil,
	}
}

// save pushes a copy of the current state onto the stack.
func (ctx *ContextState) save() ContextState {
	// Deep copy the state
	s := ContextState{
		Transform: ctx.Transform,
		StrokeStyle: CSSColor{
			Type:  ctx.StrokeStyle.Type,
			Color: ctx.StrokeStyle.Color,
		},
		FillStyle: CSSColor{
			Type:  ctx.FillStyle.Type,
			Color: ctx.FillStyle.Color,
		},
		Font:            ctx.Font,
		FontFamily:      ctx.FontFamily,
		FontSize:        ctx.FontSize,
		FontWeight:      ctx.FontWeight,
		FontStyle:       ctx.FontStyle,
		TextAlign:       ctx.TextAlign,
		TextBaseline:    ctx.TextBaseline,
		Direction:       ctx.Direction,
		LineWidth:       ctx.LineWidth,
		LineCap:         ctx.LineCap,
		LineJoin:        ctx.LineJoin,
		MiterLimit:      ctx.MiterLimit,
		LineDashOffset:  ctx.LineDashOffset,
		LineDash:        append([]float64(nil), ctx.LineDash...),
		ShadowColor:     ctx.ShadowColor,
		ShadowBlur:      ctx.ShadowBlur,
		ShadowOffsetX:   ctx.ShadowOffsetX,
		ShadowOffsetY:   ctx.ShadowOffsetY,
		GlobalAlpha:     ctx.GlobalAlpha,
		GlobalCompositeOperation: ctx.GlobalCompositeOperation,
		Filter:          ctx.Filter,
		Path:            Path2D{Commands: append([]PathCommand(nil), ctx.Path.Commands...)},
		HitRegions:      nil,
	}
	return s
}

// NewPath2D creates a new empty Path2D object.
func NewPath2D() *Path2D {
	return &Path2D{Commands: make([]PathCommand, 0)}
}

// ============ STATE MANAGEMENT ============

// Save pushes the current state onto the state stack.
func (c *Context2D) Save() {
	saved := c.state.save()
	c.stateStack = append(c.stateStack, saved)
}

// Restore pops the top state from the stack and makes it current.
func (c *Context2D) Restore() {
	if len(c.stateStack) == 0 {
		return
	}
	c.state = c.stateStack[len(c.stateStack)-1]
	c.stateStack = c.stateStack[:len(c.stateStack)-1]
}

// ============ TRANSFORMATIONS ============

// Scale applies a scaling transformation.
func (c *Context2D) Scale(x, y float64) {
	c.state.Transform = c.state.Transform.Multiply(TransformMatrix{
		A: x, B: 0, C: 0, D: y, E: 0, F: 0,
	})
}

// Rotate applies a rotation transformation. Angle is in radians.
func (c *Context2D) Rotate(angle float64) {
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	c.state.Transform = c.state.Transform.Multiply(TransformMatrix{
		A: cos, B: sin, C: -sin, D: cos, E: 0, F: 0,
	})
}

// Translate applies a translation transformation.
func (c *Context2D) Translate(x, y float64) {
	c.state.Transform = c.state.Transform.Multiply(TransformMatrix{
		A: 1, B: 0, C: 0, D: 1, E: x, F: y,
	})
}

// Transform multiplies the current transformation matrix by the given matrix.
func (c *Context2D) Transform(a, b, c2, d, e, f float64) {
	c.state.Transform = c.state.Transform.Multiply(TransformMatrix{
		A: a, B: b, C: c2, D: d, E: e, F: f,
	})
}

// SetTransform resets the current transformation matrix to identity
// then applies the given transformation.
func (c *Context2D) SetTransform(a, b, c2, d, e, f float64) {
	c.state.Transform = TransformMatrix{A: a, B: b, C: c2, D: d, E: e, F: f}
}

// ResetTransform resets the transformation matrix to identity.
func (c *Context2D) ResetTransform() {
	c.state.Transform = IdentityMatrix()
}

// ============ COMPOSITING ============

// GetGlobalAlpha returns the current global alpha value.
func (c *Context2D) GetGlobalAlpha() float64 {
	return c.state.GlobalAlpha
}

// SetGlobalAlpha sets the global alpha value (0.0 to 1.0).
func (c *Context2D) SetGlobalAlpha(alpha float64) {
	if alpha < 0 {
		alpha = 0
	}
	if alpha > 1 {
		alpha = 1
	}
	c.state.GlobalAlpha = alpha
}

// GetGlobalCompositeOperation returns the current compositing operation.
func (c *Context2D) GetGlobalCompositeOperation() CompositeOperation {
	return c.state.GlobalCompositeOperation
}

// SetGlobalCompositeOperation sets the compositing operation.
func (c *Context2D) SetGlobalCompositeOperation(op CompositeOperation) {
	c.state.GlobalCompositeOperation = op
}

// ============ FILTERS ============

// GetFilter returns the current filter string.
func (c *Context2D) GetFilter() string {
	return c.state.Filter
}

// SetFilter sets the filter string (e.g., "blur(5px)").
func (c *Context2D) SetFilter(filter string) {
	c.state.Filter = filter
}

// ============ LINE STYLES ============

// GetLineWidth returns the current line width.
func (c *Context2D) GetLineWidth() float64 {
	return c.state.LineWidth
}

// SetLineWidth sets the line width.
func (c *Context2D) SetLineWidth(width float64) {
	c.state.LineWidth = width
}

// GetLineCap returns the current line cap style.
func (c *Context2D) GetLineCap() LineCap {
	return c.state.LineCap
}

// SetLineCap sets the line cap style.
func (c *Context2D) SetLineCap(cap LineCap) {
	c.state.LineCap = cap
}

// GetLineJoin returns the current line join style.
func (c *Context2D) GetLineJoin() LineJoin {
	return c.state.LineJoin
}

// SetLineJoin sets the line join style.
func (c *Context2D) SetLineJoin(join LineJoin) {
	c.state.LineJoin = join
}

// GetMiterLimit returns the current miter limit.
func (c *Context2D) GetMiterLimit() float64 {
	return c.state.MiterLimit
}

// SetMiterLimit sets the miter limit.
func (c *Context2D) SetMiterLimit(limit float64) {
	c.state.MiterLimit = limit
}

// SetLineDash sets the line dash pattern.
func (c *Context2D) SetLineDash(segments []float64) {
	c.state.LineDash = append([]float64(nil), segments...)
}

// GetLineDash returns the current line dash pattern.
func (c *Context2D) GetLineDash() []float64 {
	return append([]float64(nil), c.state.LineDash...)
}

// SetLineDashOffset sets the line dash offset.
func (c *Context2D) SetLineDashOffset(offset float64) {
	c.state.LineDashOffset = offset
}

// GetLineDashOffset returns the current line dash offset.
func (c *Context2D) GetLineDashOffset() float64 {
	return c.state.LineDashOffset
}

// ============ SHADOWS ============

// SetShadowColor sets the shadow color.
func (c *Context2D) SetShadowColor(color string) {
	c.state.ShadowColor = color
}

// GetShadowColor returns the shadow color.
func (c *Context2D) GetShadowColor() string {
	return c.state.ShadowColor
}

// SetShadowBlur sets the shadow blur radius.
func (c *Context2D) SetShadowBlur(blur float64) {
	c.state.ShadowBlur = blur
}

// GetShadowBlur returns the shadow blur radius.
func (c *Context2D) GetShadowBlur() float64 {
	return c.state.ShadowBlur
}

// SetShadowOffsetX sets the shadow X offset.
func (c *Context2D) SetShadowOffsetX(x float64) {
	c.state.ShadowOffsetX = x
}

// GetShadowOffsetX returns the shadow X offset.
func (c *Context2D) GetShadowOffsetX() float64 {
	return c.state.ShadowOffsetX
}

// SetShadowOffsetY sets the shadow Y offset.
func (c *Context2D) SetShadowOffsetY(y float64) {
	c.state.ShadowOffsetY = y
}

// GetShadowOffsetY returns the shadow Y offset.
func (c *Context2D) GetShadowOffsetY() float64 {
	return c.state.ShadowOffsetY
}

// ============ PATHS ============

// BeginPath starts a new path by clearing the current path.
func (c *Context2D) BeginPath() {
	c.state.Path = Path2D{Commands: make([]PathCommand, 0)}
}

// MoveTo moves the pen to the given point without drawing.
func (c *Context2D) MoveTo(x, y float64) {
	c.state.Path.Commands = append(c.state.Path.Commands, PathCommand{
		Type: PathMoveTo,
		X:    x,
		Y:    y,
	})
}

// LineTo draws a straight line from the current point to the given point.
func (c *Context2D) LineTo(x, y float64) {
	c.state.Path.Commands = append(c.state.Path.Commands, PathCommand{
		Type: PathLineTo,
		X:    x,
		Y:    y,
	})
}

// QuadraticCurveTo draws a quadratic Bézier curve.
func (c *Context2D) QuadraticCurveTo(cpx, cpy, x, y float64) {
	c.state.Path.Commands = append(c.state.Path.Commands, PathCommand{
		Type: PathQuadratic,
		X:    x,
		Y:    y,
		X1:   cpx,
		Y1:   cpy,
	})
}

// BezierCurveTo draws a cubic Bézier curve.
func (c *Context2D) BezierCurveTo(cp1x, cp1y, cp2x, cp2y, x, y float64) {
	c.state.Path.Commands = append(c.state.Path.Commands, PathCommand{
		Type: PathBezier,
		X:    x,
		Y:    y,
		X1:   cp1x,
		Y1:   cp1y,
		X2:   cp2x,
		Y2:   cp2y,
	})
}

// ArcTo draws a circular arc between the current point and the given end point.
func (c *Context2D) ArcTo(x1, y1, x2, y2, radius float64) {
	// Compute the intersection point between the line from current point to (x1,y1)
	// and the line from (x1,y1) to (x2,y2), then draw an arc
	// For simplicity, we approximate by drawing a standard arc
	c.Arc(x1, y1, radius, 0, math.Pi, false)
}

// Arc draws a circular arc.
func (c *Context2D) Arc(x, y, radius, startAngle, endAngle float64, counterclockwise bool) {
	// Adjust angles if needed
	if counterclockwise {
		// Swap angles and draw in reverse
		startAngle, endAngle = endAngle, startAngle
	}
	c.state.Path.Commands = append(c.state.Path.Commands, PathCommand{
		Type:   PathArc,
		X:      x,
		Y:      y,
		Radius: radius,
	})
	// Store angles in X1, Y1 for the fill/stroke operations
}

// Ellipse draws an ellipse.
func (c *Context2D) Ellipse(x, y, radiusX, radiusY, rotation, startAngle, endAngle float64, counterclockwise bool) {
	c.state.Path.Commands = append(c.state.Path.Commands, PathCommand{
		Type:     PathEllipse,
		X:        x,
		Y:        y,
		Radius:   radiusX,
		Rotation: rotation,
	})
}

// Rect adds a rectangle to the current path.
func (c *Context2D) Rect(x, y, w, h float64) {
	c.state.Path.Commands = append(c.state.Path.Commands, PathCommand{
		Type: PathRect,
		X:    x,
		Y:    y,
	})
}

// ClosePath closes the current path by drawing a line to the starting point.
func (c *Context2D) ClosePath() {
	c.state.Path.Commands = append(c.state.Path.Commands, PathCommand{
		Type: PathClose,
	})
}

// ============ DRAWING PRIMITIVES ============

// Fill fills the current path with the current fill style.
func (c *Context2D) Fill() {
	c.fillPath(c.state.Path, true)
}

// Stroke strokes the current path with the current stroke style.
func (c *Context2D) Stroke() {
	c.strokePath(c.state.Path, true)
}

// Clip creates a clipping region from the current path.
func (c *Context2D) Clip() {
	// For now, clip is a no-op - full implementation would track clip regions
}

// FillRect fills a rectangle with the current fill style.
func (c *Context2D) FillRect(x, y, w, h float64) {
	c.fillRect(x, y, w, h)
}

// StrokeRect strokes a rectangle with the current stroke style.
func (c *Context2D) StrokeRect(x, y, w, h float64) {
	c.strokeRect(x, y, w, h)
}

// ClearRect clears the specified rectangle to transparent black.
func (c *Context2D) ClearRect(x, y, w, h float64) {
	c.clearRect(x, y, w, h)
}

// ============ TEXT ============

// SetFont sets the current font.
func (c *Context2D) SetFont(font string) {
	c.state.Font = font
	// Parse font string to extract components
	// Format: [font-style] [font-variant] [font-weight] font-size [/line-height] font-family
	parts := strings.Fields(font)
	for i, part := range parts {
		switch part {
		case "italic", "oblique":
			c.state.FontStyle = part
		case "normal":
			// Ignore
		default:
			if strings.HasSuffix(part, "px") || strings.HasSuffix(part, "pt") {
				// This is the font size
				sizeStr := strings.TrimSuffix(part, "px")
				sizeStr = strings.TrimSuffix(sizeStr, "pt")
				if size, err := parseFloat(sizeStr); err == nil {
					c.state.FontSize = size
				}
				// Everything after this is the font family
				if i+1 < len(parts) {
					c.state.FontFamily = strings.Join(parts[i+1:], " ")
				}
				break
			}
		}
	}
}

// GetFont returns the current font.
func (c *Context2D) GetFont() string {
	return c.state.Font
}

// SetTextAlign sets the text alignment.
func (c *Context2D) SetTextAlign(align TextAlign) {
	c.state.TextAlign = align
}

// GetTextAlign returns the current text alignment.
func (c *Context2D) GetTextAlign() TextAlign {
	return c.state.TextAlign
}

// SetTextBaseline sets the text baseline.
func (c *Context2D) SetTextBaseline(baseline TextBaseline) {
	c.state.TextBaseline = baseline
}

// GetTextBaseline returns the current text baseline.
func (c *Context2D) GetTextBaseline() TextBaseline {
	return c.state.TextBaseline
}

// SetDirection sets the text direction.
func (c *Context2D) SetDirection(dir Direction) {
	c.state.Direction = dir
}

// GetDirection returns the current text direction.
func (c *Context2D) GetDirection() Direction {
	return c.state.Direction
}

// FillText fills text at the given position.
func (c *Context2D) FillText(text string, x, y float64, maxWidthOpt ...float64) {
	c.drawText(text, x, y, true, maxWidthOpt...)
}

// StrokeText strokes text at the given position.
func (c *Context2D) StrokeText(text string, x, y float64, maxWidthOpt ...float64) {
	c.drawText(text, x, y, false, maxWidthOpt...)
}

// MeasureText returns the width of the text when rendered with the current font.
func (c *Context2D) MeasureText(text string) TextMetrics {
	width := c.measureTextWidth(text)
	return TextMetrics{Width: width}
}

// TextMetrics represents text metrics.
type TextMetrics struct {
	Width float64
}

// ============ GRADIENTS ============

// CreateLinearGradient creates a linear gradient.
func (c *Context2D) CreateLinearGradient(x0, y0, x1, y1 float64) *Gradient {
	return &Gradient{
		Type: "linear",
		X0:   x0,
		Y0:   y0,
		X1:   x1,
		Y1:   y1,
		Stops: make([]GradientStop, 0),
	}
}

// CreateRadialGradient creates a radial gradient.
func (c *Context2D) CreateRadialGradient(x0, y0, r0, x1, y1, r1 float64) *Gradient {
	return &Gradient{
		Type: "radial",
		X0:   x0,
		Y0:   y0,
		R0:   r0,
		X1:   x1,
		Y1:   y1,
		R1:   r1,
		Stops: make([]GradientStop, 0),
	}
}

// Gradient represents a gradient for filling/stroking.
type Gradient struct {
	Type  string // "linear" or "radial"
	X0, Y0 float64
	X1, Y1 float64
	R0    float64
	R1    float64
	Stops []GradientStop
}

// AddColorStop adds a color stop to the gradient.
func (g *Gradient) AddColorStop(offset float64, colorStr string) {
	col := ParseColor(colorStr)
	g.Stops = append(g.Stops, GradientStop{Offset: offset, Color: col})
}

// ============ STROKE/FILL STYLES ============

// SetStrokeStyle sets the stroke style.
func (c *Context2D) SetStrokeStyle(style interface{}) {
	switch v := style.(type) {
	case string:
		c.state.StrokeStyle = CSSColor{Type: "solid", Color: ParseColor(v)}
	case *Gradient:
		c.state.StrokeStyle = CSSColor{Type: "gradient", Gradient: GradientConfig{
			X0: v.X0, Y0: v.Y0, X1: v.X1, Y1: v.Y1,
		}, Stops: v.Stops}
	}
}

// SetFillStyle sets the fill style.
func (c *Context2D) SetFillStyle(style interface{}) {
	switch v := style.(type) {
	case string:
		c.state.FillStyle = CSSColor{Type: "solid", Color: ParseColor(v)}
	case *Gradient:
		c.state.FillStyle = CSSColor{Type: "gradient", Gradient: GradientConfig{
			X0: v.X0, Y0: v.Y0, X1: v.X1, Y1: v.Y1,
		}, Stops: v.Stops}
	}
}

// ============ IMAGE DRAWING ============

// DrawImage draws an image onto the canvas.
// For now, we support drawing image.RGBA images.
func (c *Context2D) DrawImage(img image.Image, dx, dy float64, optionsOpt ...DrawImageOptions) {
	bounds := img.Bounds()
	w := float64(bounds.Dx())
	h := float64(bounds.Dy())

	if len(optionsOpt) > 0 {
		opts := optionsOpt[0]
		if opts.DW != 0 && opts.DH != 0 {
			w = opts.DW
			h = opts.DH
		}
	}

	c.drawImageRect(img, dx, dy, w, h)
}

// DrawImageOptions contains optional parameters for DrawImage.
type DrawImageOptions struct {
	DX, DY float64 // destination position
	DW, DH float64 // destination dimensions
	SX, SY float64 // source position
	SW, SH float64 // source dimensions
}

func (c *Context2D) drawImageRect(img image.Image, dx, dy, w, h float64) {
	bounds := img.Bounds()
	srcX := bounds.Min.X
	srcY := bounds.Min.Y

	// Apply transform
	tx, ty := c.state.Transform.TransformPoint(dx, dy)

	// Draw each pixel
	for sy := 0; sy < bounds.Dy(); sy++ {
		for sx := 0; sx < bounds.Dx(); sx++ {
			px := int(tx) + sx
			py := int(ty) + sy
			if px >= 0 && px < c.Width && py >= 0 && py < c.Height {
				col := img.At(srcX+sx, srcY+sy)
				r, g2, b, a := col.RGBA()
				c.Pixels.SetRGBA(px, py, color.RGBA{
					R: uint8(r >> 8),
					G: uint8(g2 >> 8),
					B: uint8(b >> 8),
					A: uint8(a >> 8),
				})
			}
		}
	}
}

// ============ IMAGE DATA ============

// CreateImageData creates a new ImageData object.
func (c *Context2D) CreateImageData(w, h int) *ImageData {
	return &ImageData{
		Width:  w,
		Height: h,
		Data:   make([]uint8, w*h*4),
	}
}

// GetImageData returns the image data for the specified rectangle.
func (c *Context2D) GetImageData(sx, sy, sw, sh int) *ImageData {
	data := make([]uint8, sw*sh*4)
	idx := 0
	for y := sy; y < sy+sh; y++ {
		for x := sx; x < sx+sw; x++ {
			if x >= 0 && x < c.Width && y >= 0 && y < c.Height {
				col := c.Pixels.At(x, y)
				r, g2, b, a := col.RGBA()
				data[idx] = uint8(r >> 8)
				data[idx+1] = uint8(g2 >> 8)
				data[idx+2] = uint8(b >> 8)
				data[idx+3] = uint8(a >> 8)
			}
			idx += 4
		}
	}
	return &ImageData{Width: sw, Height: sh, Data: data}
}

// PutImageData paints pixel data onto the canvas.
func (c *Context2D) PutImageData(data *ImageData, dx, dy int) {
	idx := 0
	for y := 0; y < data.Height; y++ {
		for x := 0; x < data.Width; x++ {
			px := dx + x
			py := dy + y
			if px >= 0 && px < c.Width && py >= 0 && py < c.Height {
				c.Pixels.SetRGBA(px, py, color.RGBA{
					R: data.Data[idx],
					G: data.Data[idx+1],
					B: data.Data[idx+2],
					A: data.Data[idx+3],
				})
			}
			idx += 4
		}
	}
}

// ImageData represents pixel data.
type ImageData struct {
	Width  int
	Height int
	Data   []uint8
}

// ============ HELPER FUNCTIONS ============

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ParseColor parses a CSS color string and returns an RGBA color.
// Supports: hex (#RGB, #RRGGBB, #RRGGBBAA), rgb(), rgba(), hsl(), hsla(), named colors.
func ParseColor(colorStr string) color.RGBA {
	colorStr = strings.TrimSpace(colorStr)
	if colorStr == "" || colorStr == "transparent" {
		return color.RGBA{R: 0, G: 0, B: 0, A: 0}
	}

	// Named colors
	if named, ok := namedColors[strings.ToLower(colorStr)]; ok {
		return named
	}

	// Hex colors
	if strings.HasPrefix(colorStr, "#") {
		hex := colorStr[1:]
		switch len(hex) {
		case 3: // #RGB
			r := hex[0]
			g := hex[1]
			b := hex[2]
			return color.RGBA{
				R: uint8((int(r-'0')*16 + int(r-'0'))),
				G: uint8((int(g-'0')*16 + int(g-'0'))),
				B: uint8((int(b-'0')*16 + int(b-'0'))),
				A: 255,
			}
		case 6: // #RRGGBB
			var r, g, b uint8
			fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
			return color.RGBA{R: r, G: g, B: b, A: 255}
		case 8: // #RRGGBBAA
			var r, g, b, a uint8
			fmt.Sscanf(hex, "%02x%02x%02x%02x", &r, &g, &b, &a)
			return color.RGBA{R: r, G: g, B: b, A: a}
		}
	}

	// rgb(), rgba()
	if strings.HasPrefix(colorStr, "rgb") {
		return parseRGBAColor(colorStr)
	}

	// hsl(), hsla()
	if strings.HasPrefix(colorStr, "hsl") {
		return parseHSLAColor(colorStr)
	}

	return color.RGBA{R: 0, G: 0, B: 0, A: 255}
}

func parseRGBAColor(s string) color.RGBA {
	// Parse rgb(r, g, b) or rgba(r, g, b, a)
	s = strings.TrimPrefix(s, "rgba")
	s = strings.TrimPrefix(s, "rgb")
	s = strings.Trim(strings.Trim(s, "()"), " ")

	parts := strings.Split(s, ",")
	if len(parts) < 3 {
		return color.RGBA{R: 0, G: 0, B: 0, A: 255}
	}

	var r, g, b float64
	var a float64 = 1.0

	fmt.Sscanf(strings.TrimSpace(parts[0]), "%f", &r)
	fmt.Sscanf(strings.TrimSpace(parts[1]), "%f", &g)
	fmt.Sscanf(strings.TrimSpace(parts[2]), "%f", &b)
	if len(parts) >= 4 {
		fmt.Sscanf(strings.TrimSpace(parts[3]), "%f", &a)
	}

	// Handle percentage values
	if strings.HasSuffix(strings.TrimSpace(parts[0]), "%") {
		r = r * 255 / 100
	}
	if strings.HasSuffix(strings.TrimSpace(parts[1]), "%") {
		g = g * 255 / 100
	}
	if strings.HasSuffix(strings.TrimSpace(parts[2]), "%") {
		b = b * 255 / 100
	}

	return color.RGBA{
		R: uint8(min(255, int(r))),
		G: uint8(min(255, int(g))),
		B: uint8(min(255, int(b))),
		A: uint8(min(255, int(a*255))),
	}
}

func parseHSLAColor(s string) color.RGBA {
	// Parse hsl(h, s%, l%) or hsla(h, s%, l%, a)
	s = strings.TrimPrefix(s, "hsla")
	s = strings.TrimPrefix(s, "hsl")
	s = strings.Trim(strings.Trim(s, "()"), " ")

	parts := strings.Split(s, ",")
	if len(parts) < 3 {
		return color.RGBA{R: 0, G: 0, B: 0, A: 255}
	}

	var h, sL, l float64
	var a float64 = 1.0

	fmt.Sscanf(strings.TrimSpace(parts[0]), "%f", &h)
	fmt.Sscanf(strings.TrimSpace(parts[1]), "%f", &sL)
	fmt.Sscanf(strings.TrimSpace(parts[2]), "%f", &l)
	if len(parts) >= 4 {
		fmt.Sscanf(strings.TrimSpace(parts[3]), "%f", &a)
	}

	// Remove % from saturation and lightness
	if strings.HasSuffix(strings.TrimSpace(parts[1]), "%") {
		sL = sL / 100
	}
	if strings.HasSuffix(strings.TrimSpace(parts[2]), "%") {
		l = l / 100
	}

	// Convert HSL to RGB
	r, g2, b := hslToRgb(h/360, sL, l)

	return color.RGBA{
		R: uint8(r),
		G: uint8(g2),
		B: uint8(b),
		A: uint8(min(255, int(a*255))),
	}
}

func hslToRgb(h, s, l float64) (uint8, uint8, uint8) {
	if s == 0 {
		v := uint8(l * 255)
		return v, v, v
	}

	q := l * (1 - s)
	if l >= 0.5 {
		q = l + s - l*s
	}
	p := 2 * l - q

	hNorm := h
	if hNorm < 0 {
		hNorm += 1
	}
	if hNorm > 1 {
		hNorm -= 1
	}

	r := hueToRgb(p, q, hNorm+1.0/3.0)
	g := hueToRgb(p, q, hNorm)
	b := hueToRgb(p, q, hNorm-1.0/3.0)

	return uint8(r * 255), uint8(g * 255), uint8(b * 255)
}

func hueToRgb(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

// namedColors maps CSS named colors to RGBA values.
var namedColors = map[string]color.RGBA{
	"black":   {R: 0, G: 0, B: 0, A: 255},
	"silver":  {R: 192, G: 192, B: 192, A: 255},
	"gray":    {R: 128, G: 128, B: 128, A: 255},
	"white":   {R: 255, G: 255, B: 255, A: 255},
	"maroon":  {R: 128, G: 0, B: 0, A: 255},
	"red":     {R: 255, G: 0, B: 0, A: 255},
	"purple":  {R: 128, G: 0, B: 128, A: 255},
	"fuchsia": {R: 255, G: 0, B: 255, A: 255},
	"green":   {R: 0, G: 128, B: 0, A: 255},
	"lime":    {R: 0, G: 255, B: 0, A: 255},
	"olive":   {R: 128, G: 128, B: 0, A: 255},
	"yellow":  {R: 255, G: 255, B: 0, A: 255},
	"navy":    {R: 0, G: 0, B: 128, A: 255},
	"blue":    {R: 0, G: 0, B: 255, A: 255},
	"teal":    {R: 0, G: 128, B: 128, A: 255},
	"aqua":    {R: 0, G: 255, B: 255, A: 255},
	"orange":  {R: 255, G: 165, B: 0, A: 255},
	"pink":    {R: 255, G: 192, B: 203, A: 255},
	"brown":   {R: 165, G: 42, B: 42, A: 255},
	"coral":   {R: 255, G: 127, B: 80, A: 255},
	"gold":    {R: 255, G: 215, B: 0, A: 255},
	"cyan":    {R: 0, G: 255, B: 255, A: 255},
	"ivory":   {R: 255, G: 255, B: 240, A: 255},
	"khaki":   {R: 240, G: 230, B: 140, A: 255},
	"lavender":{R: 230, G: 230, B: 250, A: 255},
	"salmon":  {R: 250, G: 128, B: 114, A: 255},
	"wheat":   {R: 245, G: 222, B: 179, A: 255},
	"transparent": {R: 0, G: 0, B: 0, A: 0},
}

// fillPath fills a path with the current fill style.
func (c *Context2D) fillPath(path Path2D, strokeFirst bool) {
	for _, cmd := range path.Commands {
		switch cmd.Type {
		case PathMoveTo, PathLineTo:
			// Just draw a line segment
		case PathRect:
			c.fillRect(cmd.X, cmd.Y, 0, 0) // will be replaced with actual rect
		case PathClose:
			// Close path
		}
	}
	// For a basic implementation, fill with the current fill style
	// by drawing a filled polygon
}

// strokePath strokes a path with the current stroke style.
func (c *Context2D) strokePath(path Path2D, strokeFirst bool) {
	if c.state.LineWidth <= 0 {
		return
	}
	
	prevX, prevY := float64(0), float64(0)
	for _, cmd := range path.Commands {
		switch cmd.Type {
		case PathMoveTo:
			prevX, prevY = cmd.X, cmd.Y
		case PathLineTo:
			c.drawLine(prevX, prevY, cmd.X, cmd.Y)
			prevX, prevY = cmd.X, cmd.Y
		case PathClose:
			// Draw line back to start if needed
		}
	}
}

func (c *Context2D) drawLine(x1, y1, x2, y2 float64) {
	// Apply transform
	tx1, ty1 := c.state.Transform.TransformPoint(x1, y1)
	tx2, ty2 := c.state.Transform.TransformPoint(x2, y2)

	// Draw line using Bresenham's algorithm with stroke style
	col := c.state.StrokeStyle.Color
	if c.state.GlobalAlpha < 1.0 {
		col = applyAlpha(col, c.state.GlobalAlpha)
	}

	// Simple line drawing
	dx := tx2 - tx1
	dy := ty2 - ty1
	
	absDx := dx
	if absDx < 0 {
		absDx = -dx
	}
	absDy := dy
	if absDy < 0 {
		absDy = -dy
	}

	x := int(tx1)
	y := int(ty1)

	if absDx > absDy {
		// Horizontal-ish line
		if tx2 < tx1 {
			x, y = int(tx2), int(ty2)
		}
		err := absDx / 2
		for ; x <= int(tx2); x++ {
			c.setPixelClipped(x, y, col)
			err -= absDy
			if err < 0 {
				if dy > 0 {
					y++
				} else {
					y--
				}
				err += absDx
			}
		}
	} else {
		// Vertical-ish line
		if ty2 < ty1 {
			x, y = int(tx2), int(ty2)
		}
		err := absDy / 2
		for ; y <= int(ty2); y++ {
			c.setPixelClipped(x, y, col)
			err -= absDx
			if err < 0 {
				if dx > 0 {
					x++
				} else {
					x--
				}
				err += absDy
			}
		}
	}
}

func (c *Context2D) setPixelClipped(x, y int, col color.RGBA) {
	if x >= 0 && x < c.Width && y >= 0 && y < c.Height {
		c.Pixels.SetRGBA(x, y, col)
	}
}

func applyAlpha(col color.RGBA, alpha float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(col.R) * alpha),
		G: uint8(float64(col.G) * alpha),
		B: uint8(float64(col.B) * alpha),
		A: uint8(float64(col.A) * alpha),
	}
}

func (c *Context2D) fillRect(x, y, w, h float64) {
	if w == 0 && h == 0 {
		return
	}
	// Apply transform
	tx, ty := c.state.Transform.TransformPoint(x, y)
	
	col := c.state.FillStyle.Color
	if c.state.GlobalAlpha < 1.0 {
		col = applyAlpha(col, c.state.GlobalAlpha)
	}

	ix := int(tx)
	iy := int(ty)
	iw := int(w)
	ih := int(h)

	if iw < 0 {
		ix += iw
		iw = -iw
	}
	if ih < 0 {
		iy += ih
		ih = -ih
	}

	for dy := 0; dy < ih; dy++ {
		for dx := 0; dx < iw; dx++ {
			c.setPixelClipped(ix+dx, iy+dy, col)
		}
	}
}

func (c *Context2D) strokeRect(x, y, w, h float64) {
	if c.state.LineWidth <= 0 {
		return
	}
	// Draw four sides
	c.drawLine(x, y, x+w, y)
	c.drawLine(x+w, y, x+w, y+h)
	c.drawLine(x+w, y+h, x, y+h)
	c.drawLine(x, y+h, x, y)
}

func (c *Context2D) clearRect(x, y, w, h float64) {
	// Fill with transparent black
	tx, ty := c.state.Transform.TransformPoint(x, y)
	ix := int(tx)
	iy := int(ty)
	iw := int(w)
	ih := int(h)

	transparent := color.RGBA{R: 0, G: 0, B: 0, A: 0}

	for dy := 0; dy < ih; dy++ {
		for dx := 0; dx < iw; dx++ {
			c.setPixelClipped(ix+dx, iy+dy, transparent)
		}
	}
}

func (c *Context2D) drawText(text string, x, y float64, fill bool, maxWidthOpt ...float64) {
	// Apply transform
	tx, ty := c.state.Transform.TransformPoint(x, y)

	fontSize := c.state.FontSize
	if fontSize == 0 {
		fontSize = 10
	}

	col := c.state.FillStyle.Color
	if !fill {
		col = c.state.StrokeStyle.Color
	}
	if c.state.GlobalAlpha < 1.0 {
		col = applyAlpha(col, c.state.GlobalAlpha)
	}

	// Simple character-by-character drawing
	// Each character width is approximated as fontSize * 0.6
	charWidth := fontSize * 0.6
	_ = charWidth // Used for spacing below
	for i := range text {
		cx := tx + float64(i)*charWidth
		cy := ty
		c.setPixelClipped(int(cx), int(cy), col)
	}
}

func (c *Context2D) measureTextWidth(text string) float64 {
	fontSize := c.state.FontSize
	if fontSize == 0 {
		fontSize = 10
	}
	// Approximate: each char is roughly fontSize * 0.6 wide
	return float64(len(text)) * fontSize * 0.6
}

var _ = &Context2D{} // type check
var _ = &ContextState{} // type check
var _ = &Path2D{} // type check
