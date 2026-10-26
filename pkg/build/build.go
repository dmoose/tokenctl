// tokenctl/pkg/build/build.go

// Package build turns token trees into a stylesheet.
package build

import (
	"errors"
	"fmt"
	"io/fs"
	"sort"

	"github.com/dmoose/tokenctl/v2/internal/load"
	"github.com/dmoose/tokenctl/v2/pkg/generators"
	"github.com/dmoose/tokenctl/v2/pkg/tokens"
)

// Format selects the stylesheet flavour.
type Format int

const (
	FormatCSS      Format = iota // plain CSS, as --format=css
	FormatTailwind               // Tailwind 4, as --format=tailwind
)

// Options control a build.
type Options struct {
	Format Format

	// StrictUnknownKeys fails the build on input tokenctl does not consume.
	// Otherwise those keys are only reported in Result.Findings.
	StrictUnknownKeys bool
}

// Result is a build's output. On failure, Conflicts and Findings hold what
// was collected before the error.
type Result struct {
	CSS       string
	Findings  []tokens.Finding // input the build did not consume
	Conflicts []string         // tokens redefined by a later file or tree
}

// ErrUnknownKeys is returned when StrictUnknownKeys is set and the input
// has keys tokenctl does not consume.
var ErrUnknownKeys = errors.New("unconsumed keys in token input")

// BuildCSS merges the token trees left to right, as the CLI merges its
// directory arguments, and generates the stylesheet. The root of each FS is
// a token tree: brand/, semantic/, components/, themes/ and so on.
func BuildCSS(trees []fs.FS, opts Options) (Result, error) {
	srcs := make([]load.Source, len(trees))
	for i, t := range trees {
		srcs[i] = load.FS(fmt.Sprintf("tree %d", i), t)
	}
	return buildCSS(srcs, opts)
}

// BuildCSSDirs is BuildCSS over directories on disk.
func BuildCSSDirs(dirs []string, opts Options) (Result, error) {
	srcs := make([]load.Source, len(dirs))
	for i, d := range dirs {
		srcs[i] = load.Dir(d)
	}
	return buildCSS(srcs, opts)
}

func buildCSS(srcs []load.Source, opts Options) (Result, error) {
	var res Result

	loader := tokens.NewLoader()
	loader.OnConflict = func(msg string) { res.Conflicts = append(res.Conflicts, msg) }

	baseDict, themes, err := load.Merge(loader, srcs)
	if err != nil {
		return res, err
	}
	res.Findings = loader.Findings
	if opts.StrictUnknownKeys && len(res.Findings) > 0 {
		return res, fmt.Errorf("%w (%d)", ErrUnknownKeys, len(res.Findings))
	}

	resolver, err := tokens.NewResolver(baseDict)
	if err != nil {
		return res, fmt.Errorf("failed to initialize resolver: %w", err)
	}
	resolvedBase, err := resolver.ResolveAll()
	if err != nil {
		return res, fmt.Errorf("resolution failed: %w", err)
	}

	emittedBase := resolver.EmitAll(baseDict.Layers, resolvedBase)

	res.CSS, err = generate(opts.Format, baseDict, resolvedBase, emittedBase, themes)
	return res, err
}

func generate(format Format, baseDict *tokens.Dictionary, resolvedBase, emittedBase map[string]any, themes map[string]*tokens.Dictionary) (string, error) {
	inheritedThemes, err := tokens.ResolveThemeInheritance(baseDict, themes)
	if err != nil {
		return "", fmt.Errorf("failed to resolve theme inheritance: %w", err)
	}

	// Build theme contexts (sorted for deterministic error reporting)
	themeContexts := make(map[string]generators.ThemeContext)
	sortedNames := make([]string, 0, len(inheritedThemes))
	for name := range inheritedThemes {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	for _, name := range sortedNames {
		mergedDict := inheritedThemes[name]
		themeResolver, err := tokens.NewResolver(mergedDict)
		if err != nil {
			return "", fmt.Errorf("failed to resolve theme %s: %w", name, err)
		}
		resolvedTheme, err := themeResolver.ResolveAll()
		if err != nil {
			return "", fmt.Errorf("resolution failed for theme %s: %w", name, err)
		}

		emittedTheme := themeResolver.EmitAll(mergedDict.Layers, resolvedTheme)
		themeContexts[name] = generators.ThemeContext{
			Dict:           mergedDict,
			ResolvedTokens: resolvedTheme,
			DiffTokens:     tokens.DiffEmitted(resolvedTheme, resolvedBase, emittedTheme, emittedBase),
			EmitTokens:     emittedTheme,
		}
	}

	components, err := baseDict.ExtractComponents()
	if err != nil {
		return "", fmt.Errorf("failed to extract components: %w", err)
	}

	ctx := &generators.GenerationContext{
		BaseDict:           baseDict,
		ResolvedTokens:     resolvedBase,
		EmitTokens:         emittedBase,
		Components:         components,
		Themes:             themeContexts,
		DefaultTheme:       tokens.DetectDefaultTheme(themes),
		PropertyTokens:     tokens.ExtractPropertyTokens(baseDict, resolvedBase),
		Keyframes:          tokens.ExtractKeyframes(baseDict),
		Breakpoints:        tokens.ExtractBreakpoints(baseDict),
		ResponsiveTokens:   tokens.ExtractResponsiveTokens(baseDict),
		ContainerOverrides: tokens.ExtractContainerOverrides(components),
	}

	if format == FormatCSS {
		return generators.NewCSSGenerator().Generate(ctx)
	}
	return generators.NewTailwindGenerator().Generate(ctx)
}
