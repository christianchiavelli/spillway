import type { Source, Subsystem } from '#lib/api/generated/types.gen.ts';

import { sources, subsystems } from './grid.ts';

// Each reader returns undefined when the address leaves the filter out, and
// null for a value no link of the app writes, which the page answers with 404.

export function readSource(url: URL): Source | undefined | null {
  const v = url.searchParams.get('source');
  if (v === null) return undefined;
  return sources.find((s) => s.value === v)?.value ?? null;
}

export function readSubsystem(url: URL): Subsystem | undefined | null {
  const v = url.searchParams.get('subsystem');
  if (v === null || v === '') return undefined;
  return subsystems.find((s) => s.value === v)?.value ?? null;
}

/** ?page, counted from 1. */
export function readPage(url: URL): number | null {
  const v = url.searchParams.get('page') ?? '1';
  if (!/^[1-9]\d{0,4}$/.test(v)) return null;
  return Number(v);
}
