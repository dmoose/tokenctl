// tokctl/pkg/tokens/expressions.go

package tokens

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/dmoose/tokctl/pkg/colors"
)

// ExpressionEvaluator evaluates expressions in token values
// Supported expressions:
//   - calc({token} * 0.5) - arithmetic with dimensions
//   - contrast({color.primary}) - generate content color
//   - darken({color.primary}, 10%) - darken a color
//   - lighten({color.primary}, 10%) - lighten a color
type ExpressionEvaluator struct {
	resolver *Resolver
}

// NewExpressionEvaluator creates a new expression evaluator
func NewExpressionEvaluator(r *Resolver) *ExpressionEvaluator {
	return &ExpressionEvaluator{resolver: r}
}

// Expression patterns
var (
	// calcRegex matches calc(...) expressions
	calcRegex = regexp.MustCompile(`^calc\((.+)\)$`)

	// contrastRegex matches contrast({token}) expressions
	contrastRegex = regexp.MustCompile(`^contrast\(\s*\{([^}]+)\}\s*\)$`)

	// darkenRegex matches darken({token}, amount) expressions
	darkenRegex = regexp.MustCompile(`^darken\(\s*\{([^}]+)\}\s*,\s*([0-9.]+)%?\s*\)$`)

	// lightenRegex matches lighten({token}, amount) expressions
	lightenRegex = regexp.MustCompile(`^lighten\(\s*\{([^}]+)\}\s*,\s*([0-9.]+)%?\s*\)$`)

	// scaleRegex matches scale({token}, factor) expressions
	scaleRegex = regexp.MustCompile(`^scale\(\s*\{([^}]+)\}\s*,\s*([0-9.]+)\s*\)$`)

	// tokenRefRegex matches {token.path} in expressions
	tokenRefRegex = regexp.MustCompile(`\{([^}]+)\}`)
)

// IsExpression checks if a value contains an expression that needs evaluation
func IsExpression(value string) bool {
	value = strings.TrimSpace(value)
	return strings.HasPrefix(value, "calc(") ||
		strings.HasPrefix(value, "contrast(") ||
		strings.HasPrefix(value, "darken(") ||
		strings.HasPrefix(value, "lighten(") ||
		strings.HasPrefix(value, "scale(")
}

// Evaluate processes an expression string and returns the computed value
func (e *ExpressionEvaluator) Evaluate(expr string) (interface{}, error) {
	expr = strings.TrimSpace(expr)

	// Try each expression type
	if matches := calcRegex.FindStringSubmatch(expr); matches != nil {
		return e.evaluateCalc(matches[1])
	}

	if matches := contrastRegex.FindStringSubmatch(expr); matches != nil {
		return e.evaluateContrast(matches[1])
	}

	if matches := darkenRegex.FindStringSubmatch(expr); matches != nil {
		amount, _ := strconv.ParseFloat(matches[2], 64)
		return e.evaluateDarken(matches[1], amount/100)
	}

	if matches := lightenRegex.FindStringSubmatch(expr); matches != nil {
		amount, _ := strconv.ParseFloat(matches[2], 64)
		return e.evaluateLighten(matches[1], amount/100)
	}

	if matches := scaleRegex.FindStringSubmatch(expr); matches != nil {
		factor, _ := strconv.ParseFloat(matches[2], 64)
		return e.evaluateScale(matches[1], factor)
	}

	return nil, fmt.Errorf("unrecognized expression: %s", expr)
}

// evaluateCalc evaluates a calc() expression
// Supports: +, -, *, / with dimensions and token references
func (e *ExpressionEvaluator) evaluateCalc(inner string) (interface{}, error) {
	// First, resolve all token references in the expression
	resolved, err := e.resolveTokensInExpression(inner)
	if err != nil {
		return nil, fmt.Errorf("calc: %w", err)
	}

	// Parse and evaluate the arithmetic expression
	result, err := e.evaluateArithmetic(resolved)
	if err != nil {
		return nil, fmt.Errorf("calc: %w", err)
	}

	return result, nil
}

// resolveTokensInExpression replaces all {token} references with their resolved values
func (e *ExpressionEvaluator) resolveTokensInExpression(expr string) (string, error) {
	result := expr
	matches := tokenRefRegex.FindAllStringSubmatch(expr, -1)

	for _, match := range matches {
		fullMatch := match[0] // {foo.bar}
		tokenPath := match[1] // foo.bar

		// Resolve the token
		resolved, err := e.resolver.resolveReference(tokenPath)
		if err != nil {
			return "", fmt.Errorf("failed to resolve %s: %w", tokenPath, err)
		}

		// Convert to string
		resolvedStr := fmt.Sprintf("%v", resolved)
		result = strings.Replace(result, fullMatch, resolvedStr, 1)
	}

	return result, nil
}

// evaluateArithmetic evaluates a simple arithmetic expression with dimensions
// This is a simplified parser that handles: dimension op number or dimension op dimension
func (e *ExpressionEvaluator) evaluateArithmetic(expr string) (string, error) {
	expr = strings.TrimSpace(expr)

	// Try to parse as multiplication/division first (higher precedence in our simple case)
	// Pattern: <dimension> * <number> or <dimension> / <number>
	if idx := strings.LastIndex(expr, "*"); idx > 0 {
		left := strings.TrimSpace(expr[:idx])
		right := strings.TrimSpace(expr[idx+1:])
		return e.evaluateMultiply(left, right)
	}

	if idx := strings.LastIndex(expr, "/"); idx > 0 {
		left := strings.TrimSpace(expr[:idx])
		right := strings.TrimSpace(expr[idx+1:])
		return e.evaluateDivide(left, right)
	}

	// Try addition/subtraction
	// Need to be careful not to match negative numbers
	for i := len(expr) - 1; i > 0; i-- {
		if expr[i] == '+' {
			left := strings.TrimSpace(expr[:i])
			right := strings.TrimSpace(expr[i+1:])
			return e.evaluateAdd(left, right)
		}
		if expr[i] == '-' && i > 0 && expr[i-1] != '*' && expr[i-1] != '/' && expr[i-1] != '(' {
			left := strings.TrimSpace(expr[:i])
			right := strings.TrimSpace(expr[i+1:])
			return e.evaluateSubtract(left, right)
		}
	}

	// No operators, return as-is (should be a dimension or number)
	return expr, nil
}

func (e *ExpressionEvaluator) evaluateMultiply(left, right string) (string, error) {
	// Left should be a dimension, right should be a number (or vice versa)
	leftDim, leftErr := ParseDimension(left)
	rightNum, rightErr := strconv.ParseFloat(right, 64)

	if leftErr == nil && rightErr == nil {
		// dimension * number
		result := leftDim.Multiply(rightNum)
		// Round to avoid floating point artifacts
		result.Value = roundFloat(result.Value, 4)
		return result.String(), nil
	}

	// Try the reverse: number * dimension
	leftNum, leftNumErr := strconv.ParseFloat(left, 64)
	rightDim, rightDimErr := ParseDimension(right)

	if leftNumErr == nil && rightDimErr == nil {
		result := rightDim.Multiply(leftNum)
		// Round to avoid floating point artifacts
		result.Value = roundFloat(result.Value, 4)
		return result.String(), nil
	}

	return "", fmt.Errorf("cannot multiply: %s * %s", left, right)
}

func (e *ExpressionEvaluator) evaluateDivide(left, right string) (string, error) {
	leftDim, leftErr := ParseDimension(left)
	rightNum, rightErr := strconv.ParseFloat(right, 64)

	if leftErr == nil && rightErr == nil {
		result, err := leftDim.Divide(rightNum)
		if err != nil {
			return "", err
		}
		// Round to avoid floating point artifacts
		result.Value = roundFloat(result.Value, 4)
		return result.String(), nil
	}

	return "", fmt.Errorf("cannot divide: %s / %s", left, right)
}

func (e *ExpressionEvaluator) evaluateAdd(left, right string) (string, error) {
	leftDim, leftErr := ParseDimension(left)
	rightDim, rightErr := ParseDimension(right)

	if leftErr == nil && rightErr == nil {
		result, err := leftDim.Add(rightDim)
		if err != nil {
			return "", err
		}
		return result.String(), nil
	}

	return "", fmt.Errorf("cannot add: %s + %s", left, right)
}

func (e *ExpressionEvaluator) evaluateSubtract(left, right string) (string, error) {
	leftDim, leftErr := ParseDimension(left)
	rightDim, rightErr := ParseDimension(right)

	if leftErr == nil && rightErr == nil {
		result, err := leftDim.Subtract(rightDim)
		if err != nil {
			return "", err
		}
		return result.String(), nil
	}

	return "", fmt.Errorf("cannot subtract: %s - %s", left, right)
}

// evaluateContrast generates a content color for the given color token
func (e *ExpressionEvaluator) evaluateContrast(tokenPath string) (string, error) {
	// Resolve the color token
	resolved, err := e.resolver.resolveReference(tokenPath)
	if err != nil {
		return "", fmt.Errorf("contrast: failed to resolve %s: %w", tokenPath, err)
	}

	colorStr, ok := resolved.(string)
	if !ok {
		return "", fmt.Errorf("contrast: %s is not a string color value", tokenPath)
	}

	// Parse the color
	bgColor, err := colors.Parse(colorStr)
	if err != nil {
		return "", fmt.Errorf("contrast: invalid color %s: %w", colorStr, err)
	}

	// Generate content color
	contentColor := colors.ContentColor(bgColor)

	// Return in the same format as input, or OKLCH for oklch inputs
	if bgColor.OriginalFormat() == colors.FormatOKLCH {
		return contentColor.ToOKLCH(), nil
	}
	return contentColor.Hex(), nil
}

// evaluateDarken darkens a color by the given amount (0-1)
func (e *ExpressionEvaluator) evaluateDarken(tokenPath string, amount float64) (string, error) {
	resolved, err := e.resolver.resolveReference(tokenPath)
	if err != nil {
		return "", fmt.Errorf("darken: failed to resolve %s: %w", tokenPath, err)
	}

	colorStr, ok := resolved.(string)
	if !ok {
		return "", fmt.Errorf("darken: %s is not a string color value", tokenPath)
	}

	c, err := colors.Parse(colorStr)
	if err != nil {
		return "", fmt.Errorf("darken: invalid color %s: %w", colorStr, err)
	}

	// Darken by reducing lightness in OKLCH space
	l, ch, h := c.OkLch()
	newL := l * (1 - amount)
	if newL < 0 {
		newL = 0
	}

	result := colors.FromOkLch(newL, ch, h).Clamped()

	if c.OriginalFormat() == colors.FormatOKLCH {
		return result.ToOKLCH(), nil
	}
	return result.Hex(), nil
}

// evaluateLighten lightens a color by the given amount (0-1)
func (e *ExpressionEvaluator) evaluateLighten(tokenPath string, amount float64) (string, error) {
	resolved, err := e.resolver.resolveReference(tokenPath)
	if err != nil {
		return "", fmt.Errorf("lighten: failed to resolve %s: %w", tokenPath, err)
	}

	colorStr, ok := resolved.(string)
	if !ok {
		return "", fmt.Errorf("lighten: %s is not a string color value", tokenPath)
	}

	c, err := colors.Parse(colorStr)
	if err != nil {
		return "", fmt.Errorf("lighten: invalid color %s: %w", colorStr, err)
	}

	// Lighten by increasing lightness in OKLCH space
	l, ch, h := c.OkLch()
	newL := l + (1-l)*amount
	if newL > 1 {
		newL = 1
	}

	result := colors.FromOkLch(newL, ch, h).Clamped()

	if c.OriginalFormat() == colors.FormatOKLCH {
		return result.ToOKLCH(), nil
	}
	return result.Hex(), nil
}

// evaluateScale multiplies a dimension by a factor
func (e *ExpressionEvaluator) evaluateScale(tokenPath string, factor float64) (string, error) {
	resolved, err := e.resolver.resolveReference(tokenPath)
	if err != nil {
		return "", fmt.Errorf("scale: failed to resolve %s: %w", tokenPath, err)
	}

	dimStr, ok := resolved.(string)
	if !ok {
		return "", fmt.Errorf("scale: %s is not a string dimension value", tokenPath)
	}

	dim, err := ParseDimension(dimStr)
	if err != nil {
		return "", fmt.Errorf("scale: invalid dimension %s: %w", dimStr, err)
	}

	result := dim.Multiply(factor)
	// Round to avoid floating point artifacts
	result.Value = roundFloat(result.Value, 4)
	return result.String(), nil
}

// roundFloat rounds a float to n decimal places
func roundFloat(val float64, places int) float64 {
	pow := math.Pow(10, float64(places))
	return math.Round(val*pow) / pow
}
