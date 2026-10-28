import { readdirSync, readFileSync } from 'node:fs';
import { join, relative, sep } from 'node:path';
import { fileURLToPath } from 'node:url';

import { describe, expect, it } from 'vitest';

const SRC = fileURLToPath(new URL('.', import.meta.url));
const theme =
  readFileSync(join(SRC, 'app.css'), 'utf8').match(/@theme \{([\s\S]*?)\n\}/)?.[1] ?? '';

/** Every colour token in app.css's @theme, and what it is set to. */
const tokens = [...theme.matchAll(/--color-([a-z0-9-]+):\s*([^;]+);/g)].map(
  ([, name = '', value = '']) => ({ name, value: value.trim() }),
);

/** A role says what a colour is for; every other colour token is a primitive the roles use. */
const ROLE = /^(?:text|surface|stroke|icon|visual|chart)-/;
const roles = tokens.filter(({ name }) => ROLE.test(name));
const primitives = tokens.filter(({ name }) => !ROLE.test(name)).map(({ name }) => name);

/** Roles with one colour in both themes, on purpose: what sits on the header and the rail, dark in both, or on a primary fill. */
const SAME_IN_BOTH = [
  'text-on-dark-primary',
  'text-on-dark-secondary',
  'stroke-neutral-on-dark',
  'icon-on-dark',
];

// The theme menu's scraps show each theme in its own colours, whatever the
// page wears.
const OWN_COLOURS = ['lib/components/layout/ThemeScrap.svelte'];

const names = primitives.join('|');

/** Ways of painting a primitive, which would not change with the theme. */
const PAINTS = [
  // A utility, with any variant before it and any opacity or ! with it.
  new RegExp(
    String.raw`(?:^|[\s'"\x60:!(])(?:bg|text|fill|stroke|border(?:-[trblxyse])?|ring(?:-offset)?|inset-ring|outline|decoration|divide|accent|caret|from|via|to|shadow|inset-shadow|drop-shadow|text-shadow|placeholder)-(?:(?:${names})(?:\/(?:\d+|\[[^\]]+\]))?|\(--color-(?:${names})\))!?(?=[\s'"\x60=!}]|$)`,
    'm',
  ),
  // The custom property itself, as in an arbitrary value.
  new RegExp(String.raw`var\(--color-(?:${names})\)`),
  // A colour written out: in an arbitrary value, or an SVG paint attribute.
  /\[[^\]\s]*(?:#[0-9a-f]{3,8}|\b(?:rgba?|hsla?|oklch|oklab|lab|lch)\()|\b(?:fill|stroke|color|stop-color|flood-color)=["'](?:#|rgba?\(|hsla?\()/i,
];

const paints = (text: string) => PAINTS.some((pattern) => pattern.test(text));

function sources(dir: string): string[] {
  return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const path = join(dir, entry.name);
    if (entry.isDirectory()) {
      return ['paraglide', 'generated'].includes(entry.name) ? [] : sources(path);
    }
    return /\.(svelte|ts)$/.test(entry.name) && !/\.spec\.ts$/.test(entry.name) ? [path] : [];
  });
}

describe('the colour tokens', () => {
  it('give every role a colour for each theme, unless it is meant to keep one', () => {
    const one = roles.filter(({ value }) => !value.includes('light-dark('));
    expect(one.map(({ name }) => name).sort()).toEqual([...SAME_IN_BOTH].sort());
  });

  it('are what components paint with: a primitive would not change with the theme', () => {
    const painted = sources(SRC)
      .map((file) => relative(SRC, file).split(sep).join('/'))
      .filter((file) => !OWN_COLOURS.includes(file))
      .filter((file) => paints(readFileSync(join(SRC, file), 'utf8')));
    expect(painted).toEqual([]);
  });

  it.each([
    'class="bg-white"',
    'class="hover:bg-white/60"',
    'class="bg-white/[0.6]"',
    "'!bg-white'",
    "'bg-white!'",
    'class:bg-white={open}',
    'class="bg-(--color-white)"',
    'class="border-s-grey-100"',
    'class="ring-offset-white"',
    'class="inset-shadow-white"',
    'class="drop-shadow-cool-grey-490"',
    'class="shadow-[0_0_0_1px_var(--color-blue-300)]"',
    'class="bg-[#fff]"',
    'class="text-[rgb(0_0_0)]"',
    '<rect fill="#ffffff" />',
  ])('find a primitive or a colour of its own in %s', (text) => {
    expect(paints(text)).toBe(true);
  });

  it.each([
    'class="bg-surface-container-primary"',
    'class="hover:bg-surface-container-primary/60"',
    'class="whitespace-nowrap text-text-on-dark-primary"',
    'class="fill-chart-wind stroke-stroke-neutral-muted"',
    'class="autofill:shadow-[inset_0_0_0_2rem_var(--color-surface-container-primary)]"',
    '<path fill="none" stroke="currentColor" />',
  ])('find nothing in %s', (text) => {
    expect(paints(text)).toBe(false);
  });

  it('read the primitives and the roles out of app.css', () => {
    expect(primitives).toEqual(
      expect.arrayContaining(['white', 'grey-25', 'grey-150', 'cool-grey-490', 'blue-tint']),
    );
    expect(roles.map(({ name }) => name)).toEqual(
      expect.arrayContaining(['surface-canvas', 'chart-wind', 'stroke-neutral-muted']),
    );
  });
});
