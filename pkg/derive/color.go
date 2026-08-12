// tokenctl/pkg/derive/color.go
//
// Hex ↔ OKLCH conversion, ported from the extension's color-convert.ts
// (which delegated to colorjs.io). This is the entry point for "my brand
// colour is #3b6de0": the hex is reduced to a hue and chroma, and the
// derivation proceeds from those exactly as a preset does.
//
// The pipeline below reproduces colorjs.io's, matrix for matrix:
//
//	sRGB → linear sRGB → XYZ (D65) → LMS → cbrt → OKLab → OKLCh
//
// go-colorful also offers OkLab, but it reaches it by a different route
// and lands 1e-4 away in lightness and chroma and up to 0.04° away in
// hue. That is small until it crosses a rounding boundary: it moved
// #10b981's hue from 162 to 163 and #123456's chroma from 0.072 to
// 0.073, which is a visible difference in the emitted token and would
// have forced a tolerance into the golden comparison. Carrying the
// matrices here keeps the agreement exact.
package derive

import (
	"fmt"
	"math"
	"strings"
)

// Recalculated for a consistent reference white; these are colorjs.io's
// values, themselves from the CSS Color 4 discussion.
var (
	linearSRGBToXYZ = [3][3]float64{
		{0.41239079926595934, 0.357584339383878, 0.1804807884018343},
		{0.21263900587151027, 0.715168678767756, 0.07219231536073371},
		{0.01933081871559182, 0.11919477979462598, 0.9505321522496607},
	}
	xyzToLinearSRGB = [3][3]float64{
		{3.2409699419045226, -1.537383177570094, -0.4986107602930034},
		{-0.9692436362808796, 1.8759675015077202, 0.04155505740717559},
		{0.05563007969699366, -0.20397695888897652, 1.0569715142428786},
	}
	xyzToLMS = [3][3]float64{
		{0.8190224379967030, 0.3619062600528904, -0.1288737815209879},
		{0.0329836539323885, 0.9292868615863434, 0.0361446663506424},
		{0.0481771893596242, 0.2642395317527308, 0.6335478284694309},
	}
	lmsToXYZ = [3][3]float64{
		{1.2268798758459243, -0.5578149944602171, 0.2813910456659647},
		{-0.0405757452148008, 1.1122868032803170, -0.0717110580655164},
		{-0.0763729366746601, -0.4214933324022432, 1.5869240198367816},
	}
	lmsToLab = [3][3]float64{
		{0.2104542683093140, 0.7936177747023054, -0.0040720430116193},
		{1.9779985324311684, -2.4285922420485799, 0.4505937096174110},
		{0.0259040424655478, 0.7827717124575296, -0.8086757549230774},
	}
	labToLMS = [3][3]float64{
		{1.0000000000000000, 0.3963377773761749, 0.2158037573099136},
		{1.0000000000000000, -0.1055613458156586, -0.0638541728258133},
		{1.0000000000000000, -0.0894841775298119, -1.2914855480194092},
	}
)

// achromaticEpsilon matches colorjs.io: below it the a/b components are
// noise and the hue is undefined. The engine coerced that undefined hue
// to 0, so a grey brand colour derives a zero hue rather than whatever
// angle the rounding noise happened to point at.
const achromaticEpsilon = 0.0002

// HexToOklchParts converts a hex colour to OKLCH components, returning
// lightness as a percentage (0–100), chroma, and hue in degrees.
func HexToOklchParts(hex string) (l, c, h float64, err error) {
	r, g, b, err := parseHex(hex)
	if err != nil {
		return 0, 0, 0, err
	}
	lab := srgbToOklab(r, g, b)
	ll, cc, hh := oklabToOklch(lab)
	return ll * 100, cc, hh, nil
}

// OklchToHex converts OKLCH components back to a six-digit hex string,
// clamping out-of-gamut channels into sRGB the way the engine did.
func OklchToHex(l, c, h float64) string {
	r, g, b := oklchToSRGB(l/100, c, h)
	return fmt.Sprintf("#%02x%02x%02x",
		channelToByte(r), channelToByte(g), channelToByte(b))
}

// ParamsFromHex builds derivation params from a brand colour, taking the
// remaining controls from defaults. Chroma is floored at 0.08 so a
// near-grey brand colour still yields a usable primary — the engine's
// Math.max(primaryC, 0.08).
func ParamsFromHex(hex string) (Params, error) {
	_, c, h, err := HexToOklchParts(hex)
	if err != nil {
		return Params{}, err
	}
	p := DefaultParams
	p.Hue = h
	p.Chroma = math.Max(c, 0.08)
	return p, nil
}

func srgbToOklab(r, g, b float64) [3]float64 {
	lin := [3]float64{srgbToLinear(r), srgbToLinear(g), srgbToLinear(b)}
	xyz := mul3(linearSRGBToXYZ, lin)
	lms := mul3(xyzToLMS, xyz)
	for i := range lms {
		lms[i] = math.Cbrt(lms[i])
	}
	return mul3(lmsToLab, lms)
}

func oklabToOklch(lab [3]float64) (l, c, h float64) {
	l, a, b := lab[0], lab[1], lab[2]
	c = math.Sqrt(a*a + b*b)
	if math.Abs(a) < achromaticEpsilon && math.Abs(b) < achromaticEpsilon {
		return l, c, 0 // hue undefined; the engine's `|| 0`
	}
	h = math.Atan2(b, a) * 180 / math.Pi
	return l, c, constrainAngle(h)
}

func oklchToSRGB(l, c, h float64) (r, g, b float64) {
	rad := h * math.Pi / 180
	lab := [3]float64{l, c * math.Cos(rad), c * math.Sin(rad)}
	lms := mul3(labToLMS, lab)
	for i := range lms {
		lms[i] = lms[i] * lms[i] * lms[i]
	}
	xyz := mul3(lmsToXYZ, lms)
	lin := mul3(xyzToLinearSRGB, xyz)
	return linearToSRGB(lin[0]), linearToSRGB(lin[1]), linearToSRGB(lin[2])
}

func srgbToLinear(v float64) float64 {
	abs := math.Abs(v)
	if abs <= 0.04045 {
		return v / 12.92
	}
	return math.Copysign(math.Pow((abs+0.055)/1.055, 2.4), v)
}

func linearToSRGB(v float64) float64 {
	abs := math.Abs(v)
	if abs > 0.0031308 {
		return math.Copysign(1.055*math.Pow(abs, 1/2.4)-0.055, v)
	}
	return 12.92 * v
}

func mul3(m [3][3]float64, v [3]float64) [3]float64 {
	return [3]float64{
		m[0][0]*v[0] + m[0][1]*v[1] + m[0][2]*v[2],
		m[1][0]*v[0] + m[1][1]*v[1] + m[1][2]*v[2],
		m[2][0]*v[0] + m[2][1]*v[1] + m[2][2]*v[2],
	}
}

func constrainAngle(deg float64) float64 {
	return math.Mod(math.Mod(deg, 360)+360, 360)
}

func channelToByte(v float64) int {
	return int(math.Round(math.Max(0, math.Min(1, v)) * 255))
}

func parseHex(hex string) (r, g, b float64, err error) {
	s := strings.TrimSpace(hex)
	s = strings.TrimPrefix(s, "#")
	// Expand the three-digit shorthand.
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex colour %q: want #rgb or #rrggbb", hex)
	}
	var ri, gi, bi int
	if _, err := fmt.Sscanf(strings.ToLower(s), "%02x%02x%02x", &ri, &gi, &bi); err != nil {
		return 0, 0, 0, fmt.Errorf("invalid hex colour %q: %w", hex, err)
	}
	return float64(ri) / 255, float64(gi) / 255, float64(bi) / 255, nil
}
