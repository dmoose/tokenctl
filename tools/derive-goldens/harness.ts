/**
 * Golden generator for `tokenctl derive`.
 *
 * Runs the retired extension's theme-engine.ts — the real TypeScript
 * implementation, not a transcription of it — over the preset list and a
 * parameter grid, and writes one JSON file per case. The Go port is
 * checked against these files.
 *
 * The engine is imported from its original location by absolute path so
 * that this harness never copies or forks the math. Its only dependency
 * is colorjs.io, which resolves out of the extension's node_modules.
 *
 * Run via ../../tools/derive-goldens/regenerate.sh — never from `go test`.
 */
import {
  PRESETS,
  TYPOGRAPHY_SYSTEMS,
  DEFAULT_PARAMS,
  generateThemeFromParams,
  hexToOklchParts,
  oklchToHex,
  type ThemeParams,
} from '@engine';

import { writeFileSync, mkdirSync } from 'node:fs';
import { join } from 'node:path';

const outDir = process.argv[2];
if (!outDir) {
  console.error('usage: harness <output-dir>');
  process.exit(1);
}
mkdirSync(outDir, { recursive: true });

interface Case {
  name: string;
  params: ThemeParams;
}

const cases: Case[] = [];

// ── 1. Every built-in preset, both modes, defaults elsewhere ──────
for (const preset of PRESETS) {
  for (const isDark of [false, true]) {
    cases.push({
      name: `preset-${preset.name.toLowerCase()}-${isDark ? 'dark' : 'light'}`,
      params: {
        ...DEFAULT_PARAMS,
        hue: preset.hue,
        chroma: preset.chroma ?? DEFAULT_PARAMS.chroma,
        isDark,
      },
    });
  }
}

// ── 2. One axis at a time: edges and midpoint of every range ──────
// Ranges are the engine's documented domains: tint 0–100,
// saturation 0–150, density 75–130. Hue wraps at 360; chroma is
// clamped to 0.4 inside the engine.
const axes: { axis: keyof ThemeParams; values: (number | string)[] }[] = [
  { axis: 'hue', values: [0, 180, 359] },
  { axis: 'chroma', values: [0, 0.2, 0.4] },
  { axis: 'tint', values: [0, 50, 100] },
  { axis: 'saturation', values: [0, 75, 150] },
  { axis: 'density', values: [75, 102.5, 130] },
  { axis: 'fontPairing', values: TYPOGRAPHY_SYSTEMS.map((s) => s.key) },
];

for (const { axis, values } of axes) {
  for (const value of values) {
    for (const isDark of [false, true]) {
      cases.push({
        name: `axis-${axis}-${String(value).replace('.', '_')}-${isDark ? 'dark' : 'light'}`,
        params: { ...DEFAULT_PARAMS, [axis]: value, isDark } as ThemeParams,
      });
    }
  }
}

// ── 3. Corners: every axis at a bound simultaneously ──────────────
const corners: { name: string; params: Partial<ThemeParams> }[] = [
  { name: 'corner-all-min', params: { hue: 0, chroma: 0, tint: 0, saturation: 0, density: 75 } },
  { name: 'corner-all-max', params: { hue: 359, chroma: 0.4, tint: 100, saturation: 150, density: 130 } },
  { name: 'corner-flat-neutral', params: { hue: 250, chroma: 0, tint: 0, saturation: 0, density: 100 } },
  { name: 'corner-max-tint-min-sat', params: { hue: 90, chroma: 0.3, tint: 100, saturation: 0, density: 100 } },
  { name: 'corner-min-tint-max-sat', params: { hue: 300, chroma: 0.1, tint: 0, saturation: 150, density: 130 } },
];
for (const corner of corners) {
  for (const isDark of [false, true]) {
    cases.push({
      name: `${corner.name}-${isDark ? 'dark' : 'light'}`,
      params: { ...DEFAULT_PARAMS, ...corner.params, isDark } as ThemeParams,
    });
  }
}

// ── Emit one file per case ────────────────────────────────────────
const index: string[] = [];
for (const c of cases) {
  const tokens = generateThemeFromParams(c.params);
  const payload = { name: c.name, params: c.params, tokens };
  writeFileSync(join(outDir, `${c.name}.json`), JSON.stringify(payload, null, 2) + '\n');
  index.push(c.name);
}

// ── Colour-conversion goldens (color-convert.ts math) ─────────────
// hexToOklchParts / oklchToHex are the hex entry point into the same
// engine; the Go port has to agree on both directions.
const hexes = [
  '#000000', '#ffffff', '#3b6de0', '#8b5cf6', '#10b981', '#d97706',
  '#dc2626', '#0d9488', '#ec4899', '#64748b', '#ff0000', '#00ff00',
  '#0000ff', '#808080', '#123456', '#fedcba',
];
const hexCases = hexes.map((hex) => {
  const [l, c, h] = hexToOklchParts(hex);
  return { hex, l, c, h, roundTrip: oklchToHex(l, c, h) };
});
writeFileSync(
  join(outDir, '_colorconvert.json'),
  JSON.stringify({ cases: hexCases }, null, 2) + '\n',
);

writeFileSync(
  join(outDir, '_index.json'),
  JSON.stringify({ count: index.length, cases: index.sort() }, null, 2) + '\n',
);

console.log(`wrote ${index.length} theme goldens + ${hexCases.length} colour-convert goldens to ${outDir}`);
