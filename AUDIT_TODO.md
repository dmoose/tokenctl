# tokenctl Code Audit TODO

Ranked by importance. Items marked [FIXED] have been addressed.

## Tier 1 -- Correctness / Safety Bugs

- [x] **1. `defer f.Close()` inside a loop leaks file descriptors**
  `cmd/tokenctl/init.go` -- Extracted `writeTokenFile` helper; each file is now closed within its own scope.

- [x] **2. `os.Exit(1)` inside a `RunE` function bypasses cobra error handling**
  `cmd/tokenctl/validate.go` -- Replaced with `return fmt.Errorf("validation failed")`.

- [x] **3. Unchecked type assertion panics on bad input**
  `pkg/tokens/components.go` -- Changed to comma-ok assertion; non-string `$class` is now silently skipped.

- [x] **4. `strictMode` flag is registered but never read**
  `cmd/tokenctl/validate.go` -- Removed dead `--strict` flag and `strictMode` variable entirely.

- [x] **5. `validateColorFormat` skips `contrast/darken/lighten` but not `shade`**
  `pkg/tokens/validator.go` -- Added `shade(` to the expression skip list.

## Tier 2 -- Significant Code Duplication

- [x] **6. Massive duplication between `CSSGenerator` and `TailwindGenerator`**
  Extracted `generatePropertyDeclarations`, `buildStateSelector`, `writeProperties`, and
  `resolveTokenReferences` into `pkg/generators/shared.go` as free functions. Hoisted regex
  to package-level `var`. `generateThemeVariations` and `generateComponents` kept as methods
  (they have real structural differences beyond formatting).

- [x] **7. `roundTo` vs `roundFloat` -- identical functions with different names**
  Removed `roundFloat` from `expressions.go`; all call sites now use `roundTo` from `dimension.go`.

- [~] **8. Repeated tree-walking pattern across 8+ functions** -- WON'T FIX
  The walkers have different return types, recursion conditions, and side effects.
  A generic walker would require complex callbacks that obscure rather than simplify.

- [x] **9. "Load base + load themes + resolve" sequence duplicated across 3 commands**
  Extracted `loadTokens` and `resolveTokens` helpers into `cmd/tokenctl/helpers.go`;
  `validate.go`, `build.go`, and `search.go` now use the shared helpers.

## Tier 3 -- Go Idiom Violations

- [x] **10. `filepath.Walk` instead of `filepath.WalkDir`**
  Both call sites in `loader.go` updated to `filepath.WalkDir` with `fs.DirEntry`.

- [x] **11. Regex compiled on every call in hot path**
  Fixed as part of #6 -- `resolveTokenReferences` moved to `shared.go` with regex hoisted to package-level `var`.

- [x] **12. `fmt.Sprintf` instead of `filepath.Join` for paths**
  `cmd/tokenctl/build.go` -- All 5 path constructions now use `filepath.Join`.

- [x] **13. Empty structs as method namespaces**
  Converted `Validator` and `ThemeGenerator` to standalone functions (unique names).
  `CSSGenerator`/`TailwindGenerator` kept as structs (method names like `Generate`,
  `generateThemeVariations`, `generateComponents` collide without the receiver).

- [x] **14. `errors` variable shadows the `errors` package**
  Renamed `errors` to `errs` in all four validator functions.

- [x] **15. Custom `contains`/`containsString` reimplementing `strings.Contains`**
  Replaced both with `strings.Contains` and deleted the custom functions.

## Tier 4 -- Swallowed Errors

- [x] **16. `strconv.ParseFloat` errors silently ignored (4 sites)**
  Added error checking with descriptive messages for all 4 `strconv.ParseFloat` calls in `expressions.go`.

- [x] **17. `fmt.Sscanf` return value ignored**
  Replaced with `strconv.Atoi` in both `responsive.go` and `keyframes.go`; default-to-zero behavior documented.

- [x] **18. `log.Printf` in library code**
  Replaced with `fmt.Fprintf(os.Stderr, ...)` in `loader.go`; removed `log` import from package.

## Tier 5 -- Missing Test Coverage

- [x] **19. Five files have zero test coverage:**
  Added test files: `layers_test.go` (37 tests), `responsive_test.go` (25 tests),
  `components_test.go` (22 tests), `metadata_test.go` (20 tests), `themes_test.go` (9 tests).

## Tier 6 -- Dead Code

- [x] **20. `Token` struct defined but never used**
  Removed from `types.go`.

- [x] **21. Unused exported functions:**
  - `colors.FromRGB` (`colors.go:340`)
  - `colors.CreateFromColorful` (`content.go:224`)
  - `colors.ContentColorPreserveHue` (`content.go:159`)
  - `colors.SufficientContrast` (`contrast.go:79`)

- [x] **22. Unused `value` parameter in `matchesSearch`**
  Removed from function signature and call site in `search.go`.

- [x] **23. Unused `_` parameter in `MergeWithPath`**
  Removed third parameter from signature; updated all callers in `loader.go` and `loader_test.go`.

## Tier 7 -- Design / Maintainability

- [x] **24. `runBuild` is a 209-line god function**
  Split into `buildCSSOutput`, `buildCatalogOutput`, and `writeOutput` helpers.

- [~] **25. Over-reliance on `map[string]any` with no typed intermediate representation** -- WON'T FIX
  Massive rewrite for marginal benefit; the W3C spec's dynamic nesting maps naturally to `map[string]any`.

- [x] **26. `MeetsWCAAA` / `MeetsWCAAAAA` -- misleading function names**
  Renamed to `MeetsWCAG_AA` and `MeetsWCAG_AAA`.

- [x] **27. `serializeValueForCSS` lives in `tailwind.go` but is used by `css.go`**
  Moved to `pkg/generators/css_utils.go` alongside the other serialization helpers.

- [x] **28. Non-deterministic map iteration order in output**
  Added sorted theme-name iteration in `validate.go` and both theme loops in `build.go`.

## Tier 8 -- Minor / Cosmetic

- [x] **29. Typo: `spaceSeperatedProps`**
  Renamed to `spaceSeparatedProps`.

- [x] **30. Hardcoded `"light"` theme name in 3 places**
  Extracted `DefaultThemeName` constant in `themes.go`; all 3 sites now reference it.

- [~] **31. Magic numbers without named constants** -- WON'T FIX
  Most numeric literals are CSS spec values (e.g., WCAG thresholds, permission bits);
  naming them adds indirection without clarity.

- [x] **32. Path traversal via unsanitized `category` in `--format=manifest:CATEGORY`**
  `parseFormat` now rejects categories containing `/`, `\`, or `..`.

- [x] **33. No `t.Parallel()` on independent tests**
  Added `t.Parallel()` to all top-level tests and table-driven subtests across all 4 packages.
  Excluded 3 loader tests that mutate `os.Stderr` (inherently non-parallelizable).

- [~] **34. No golden file update mechanism (`-update` flag pattern)** -- WON'T FIX
  Tests use substring checks rather than exact golden file comparisons;
  an `-update` flag would add complexity without addressing a real maintenance burden.
