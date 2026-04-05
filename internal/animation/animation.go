package animation

import (
	"fmt"
	"strings"
	"time"

	"github.com/nickcoury/ViBrowsing/internal/css"
)

// Animation represents a running CSS animation.
type Animation struct {
	Name         string
	ElementID    string // Identifier for the element undergoing this animation
	Duration     time.Duration
	Delay        time.Duration
	TimingFunc   string // ease, linear, ease-in, ease-out, ease-in-out, etc.
	IterationCount int   // 1 for normal, -1 for infinite
	Direction    string // normal, reverse, alternate, alternate-reverse
	FillMode     string // none, forwards, backwards, both
	
	StartTime    time.Time
	Keyframes    map[float64]map[string]string // percentage -> properties
	ComputedProps map[string]string // Current computed properties during animation
}

// AnimationManager manages running animations.
type AnimationManager struct {
	Animations []*Animation
	// KeyframeRules maps animation name to keyframe definitions
	KeyframeRules map[string]map[float64]map[string]string
	// RAF callbacks scheduled via requestAnimationFrame
	RAFCallbacks map[int]func(float64) // callback handle -> callback
	RAFHandle     int                   // next RAF handle
	RAFCallbacksToCancel map[int]bool  // handles to cancel next frame
}

// NewAnimationManager creates a new animation manager.
func NewAnimationManager() *AnimationManager {
	return &AnimationManager{
		Animations:           []*Animation{},
		KeyframeRules:        make(map[string]map[float64]map[string]string),
		RAFCallbacks:         make(map[int]func(float64)),
		RAFCallbacksToCancel: make(map[int]bool),
	}
}

// ParseAnimationTimingFunction parses CSS timing function values.
// Returns multiplier for progress calculation (simplified).
func ParseAnimationTimingFunction(tf string) func(float64) float64 {
	switch strings.ToLower(tf) {
	case "linear":
		return func(t float64) float64 { return t }
	case "ease":
		// ease is the default, equivalent to cubic-bezier(0.25, 0.1, 0.25, 1.0)
		return func(t float64) float64 {
			if t <= 0 {
				return 0
			}
			if t >= 1 {
				return 1
			}
			// Simplified bezier approximation for ease
			return t * t * (3 - 2*t)
		}
	case "ease-in":
		return func(t float64) float64 {
			if t <= 0 {
				return 0
			}
			if t >= 1 {
				return 1
			}
			return t * t
		}
	case "ease-out":
		return func(t float64) float64 {
			if t <= 0 {
				return 0
			}
			if t >= 1 {
				return 1
			}
			return 1 - (1-t)*(1-t)
		}
	case "ease-in-out":
		return func(t float64) float64 {
			if t <= 0 {
				return 0
			}
			if t >= 1 {
				return 1
			}
			if t < 0.5 {
				return 2 * t * t
			}
			return 1 - (-2*t+2)*(-2*t+2)/2
		}
	case "step-start":
		return func(t float64) float64 {
			if t >= 1 {
				return 1
			}
			return 0
		}
	case "step-end":
		return func(t float64) float64 {
			if t <= 0 {
				return 0
			}
			return 1
		}
	default:
		// Default to ease
		return func(t float64) float64 {
			if t <= 0 {
				return 0
			}
			if t >= 1 {
				return 1
			}
			return t * t * (3 - 2*t)
		}
	}
}

// ParseIterationCount parses animation-iteration-count value.
func ParseIterationCount(s string) int {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "infinite" {
		return -1
	}
	// Try to parse as integer
	var count int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			count = count*10 + int(c-'0')
		} else if c == '.' {
			break
		}
	}
	if count == 0 {
		return 1
	}
	return count
}

// RegisterKeyframes registers keyframes from CSS.Keyframes into the animation manager.
func (am *AnimationManager) RegisterKeyframes(keyframes []css.KeyframeRule) {
	for _, kf := range keyframes {
		if _, exists := am.KeyframeRules[kf.Name]; !exists {
			am.KeyframeRules[kf.Name] = make(map[float64]map[string]string)
		}
		for pct, props := range kf.Keyframes {
			am.KeyframeRules[kf.Name][pct] = props
		}
	}
}

// StartAnimation starts a new animation for an element.
func (am *AnimationManager) StartAnimation(name, elementID string, duration, delay time.Duration, 
	timingFunc string, iterationCount int, direction, fillMode string) {
	
	keyframes, hasKeyframes := am.KeyframeRules[name]
	if !hasKeyframes || len(keyframes) == 0 {
		return // No keyframes found for this animation name
	}
	
	anim := &Animation{
		Name:           name,
		ElementID:      elementID,
		Duration:       duration,
		Delay:          delay,
		TimingFunc:     timingFunc,
		IterationCount: iterationCount,
		Direction:      direction,
		FillMode:       fillMode,
		StartTime:      time.Now().Add(delay),
		Keyframes:      keyframes,
		ComputedProps:  make(map[string]string),
	}
	
	am.Animations = append(am.Animations, anim)
}

// StopAnimation stops all animations for an element.
func (am *AnimationManager) StopAnimation(elementID string) {
	var remaining []*Animation
	for _, anim := range am.Animations {
		if anim.ElementID != elementID {
			remaining = append(remaining, anim)
		}
	}
	am.Animations = remaining
}

// UpdateAnimation updates animation state and returns computed properties.
// Returns nil if animation is not active or has finished.
func (am *AnimationManager) UpdateAnimation(elementID string) map[string]string {
	for i, anim := range am.Animations {
		if anim.ElementID != elementID {
			continue
		}
		
		elapsed := time.Since(anim.StartTime)
		
		// Check if animation is still in delay phase
		if elapsed < 0 {
			return nil
		}
		
		// Calculate total iteration time
		duration := anim.Duration
		if duration <= 0 {
			duration = time.Second // Default to 1s if no duration
		}
		
		// Calculate how many iterations have completed
		totalElapsed := elapsed
		iteration := 0
		if anim.IterationCount != -1 { // Not infinite
			iterationsCompleted := int(totalElapsed / duration)
			iteration = iterationsCompleted
			if iteration >= anim.IterationCount {
				// Animation complete
				am.Animations = append(am.Animations[:i], am.Animations[i+1:]...)
				return nil
			}
		}
		
		// Calculate progress within current iteration (0 to 1)
		progress := float64(totalElapsed % duration) / float64(duration)
		
		// Apply direction
		actualProgress := progress
		if anim.Direction == "alternate" && iteration%2 == 1 {
			actualProgress = 1 - progress
		} else if anim.Direction == "reverse" {
			actualProgress = 1 - progress
		}
		
		// Apply timing function
		timingFn := ParseAnimationTimingFunction(anim.TimingFunc)
		actualProgress = timingFn(actualProgress)
		
		// Interpolate properties between keyframes
		computed := am.interpolateProperties(anim.Keyframes, actualProgress)
		anim.ComputedProps = computed
		
		return computed
	}
	return nil
}

// interpolateProperties interpolates CSS property values at a given progress (0-1).
func (am *AnimationManager) interpolateProperties(keyframes map[float64]map[string]string, progress float64) map[string]string {
	result := make(map[string]string)
	
	if len(keyframes) == 0 {
		return result
	}
	
	// Find the two keyframes to interpolate between
	var fromPct, toPct float64 = 0, 100
	foundFrom := false
	foundTo := false
	
	// Find keyframes surrounding progress
	for pct := range keyframes {
		if pct <= progress*100 && pct > fromPct {
			fromPct = pct
			foundFrom = true
		}
		if pct >= progress*100 && (pct < toPct || !foundTo) {
			toPct = pct
			foundTo = true
		}
	}
	
	// Handle edge cases
	if !foundFrom {
		fromPct = 0
	}
	if !foundTo {
		toPct = 100
	}
	
	if fromPct == toPct {
		// Exact match or single keyframe
		for k, v := range keyframes[fromPct] {
			result[k] = v
		}
		return result
	}
	
	// Interpolate
	t := (progress*100 - fromPct) / (toPct - fromPct)
	
	fromProps := keyframes[fromPct]
	toProps := keyframes[toPct]
	
	// Collect all properties from both keyframes
	propSet := make(map[string]bool)
	for k := range fromProps {
		propSet[k] = true
	}
	for k := range toProps {
		propSet[k] = true
	}
	
	for prop := range propSet {
		fromVal, hasFrom := fromProps[prop]
		toVal, hasTo := toProps[prop]
		
		if !hasFrom {
			result[prop] = toVal
		} else if !hasTo {
			result[prop] = fromVal
		} else {
			// Both have values - interpolate if possible
			interpVal := interpolateCSSValue(fromVal, toVal, t)
			if interpVal != "" {
				result[prop] = interpVal
			} else {
				// Cannot interpolate, use start value
				result[prop] = fromVal
			}
		}
	}
	
	return result
}

// interpolateCSSValue attempts to interpolate between two CSS values.
// Returns empty string if interpolation is not possible.
func interpolateCSSValue(from, to string, t float64) string {
	// Try numeric interpolation first
	fromNum := parseCSSNumber(from)
	toNum := parseCSSNumber(to)
	
	if fromNum != nil && toNum != nil {
		// Interpolate numeric values
		result := (*fromNum) + t*((*toNum)-(*fromNum))
		return formatCSSNumber(result, from)
	}
	
	// Cannot interpolate - return start value
	return from
}

// parseCSSNumber tries to parse a CSS numeric value (length, percentage, number).
func parseCSSNumber(s string) *float64 {
	s = strings.TrimSpace(s)
	
	// Remove color functions for now - would need more complex handling
	if strings.HasPrefix(s, "rgb") || strings.HasPrefix(s, "#") {
		return nil
	}
	
	// Try to parse number with optional unit
	numStr := ""
	unitIdx := 0
	for unitIdx = 0; unitIdx < len(s); unitIdx++ {
		c := s[unitIdx]
		if (c >= '0' && c <= '9') || c == '.' || c == '-' || c == '+' {
			numStr += string(c)
		} else {
			break
		}
	}
	
	if numStr == "" {
		return nil
	}
	
	val, err := parseFloat(numStr)
	if err != nil {
		return nil
	}
	
	return &val
}

// formatCSSNumber formats a number, preserving the unit from the original if present.
func formatCSSNumber(val float64, original string) string {
	// Determine if original had a unit and what it is
	hasUnit := false
	unit := ""
	for i, c := range original {
		if (c >= '0' && c <= '9') || c == '.' || c == '-' || c == '+' {
			continue
		}
		hasUnit = true
		unit = original[i:]
		break
	}
	
	// Format the number
	if hasUnit && unit == "%" {
		return fmt.Sprintf("%.2f%%", val)
	} else if hasUnit {
		return fmt.Sprintf("%.2f%s", val, unit)
	}
	
	// No unit or special case - just return number
	if val == float64(int(val)) {
		return fmt.Sprintf("%.0f", val)
	}
	return fmt.Sprintf("%.2f", val)
}

func parseFloat(s string) (float64, error) {
	var val float64
	var fraction float64 = 1
	var inFraction bool
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '-' || c == '+' {
			continue
		}
		if c == '.' {
			inFraction = true
			continue
		}
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not a number")
		}
		if inFraction {
			fraction /= 10
			val += float64(c-'0') * fraction
		} else {
			val = val*10 + float64(c-'0')
		}
	}
	return val, nil
}

// RequestAnimationFrame schedules a callback to be called before the next repaint.
// Returns a request ID that can be passed to cancelAnimationFrame to cancel the request.
func (am *AnimationManager) RequestAnimationFrame(callback func(float64)) int {
	am.RAFHandle++
	handle := am.RAFHandle
	am.RAFCallbacks[handle] = callback
	return handle
}

// CancelAnimationFrame cancels a scheduled animation frame request.
func (am *AnimationManager) CancelAnimationFrame(handle int) {
	// Mark for cancellation - will be removed before next frame runs
	am.RAFCallbacksToCancel[handle] = true
}

// TickRAF executes all scheduled RAF callbacks with the given timestamp.
// Called by the render loop before each paint.
// Returns how many callbacks were executed.
func (am *AnimationManager) TickRAF(timestamp float64) int {
	// First, cancel any pending cancellations
	for handle := range am.RAFCallbacksToCancel {
		delete(am.RAFCallbacks, handle)
		delete(am.RAFCallbacksToCancel, handle)
	}

	count := 0
	for handle, callback := range am.RAFCallbacks {
		if !am.RAFCallbacksToCancel[handle] {
			callback(timestamp)
			count++
		}
	}
	return count
}

// FlushRAF cancels all pending RAF callbacks without executing them.
func (am *AnimationManager) FlushRAF() {
	am.RAFCallbacks = make(map[int]func(float64))
	am.RAFCallbacksToCancel = make(map[int]bool)
}
