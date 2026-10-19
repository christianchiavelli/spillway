import type { Period, PeriodKind } from '#lib/api/generated/types.gen.ts';

import { dayAndMonth, dayOfMonth, longDay, monthName, monthYear, weekday } from './format.ts';

export const periodKinds = [
  { value: 'day', label: 'Day' },
  { value: 'week', label: 'Week' },
  { value: 'month', label: 'Month' },
] as const satisfies readonly { value: PeriodKind; label: string }[];

/** The period a page's address asks for: the latest day when it says nothing. */
export type PeriodQuery = { period: PeriodKind; date: string | undefined };

/**
 * Reads ?period and ?date, or returns null for an address no link of the app
 * writes, which the page answers with a 404.
 */
export function readPeriod(url: URL): PeriodQuery | null {
  const period = url.searchParams.get('period') ?? 'day';
  const date = url.searchParams.get('date') ?? undefined;
  if (!periodKinds.some((k) => k.value === period)) return null;
  if (date !== undefined && !/^\d{4}-\d{2}-\d{2}$/.test(date)) return null;
  return { period: period as PeriodKind, date };
}

/** The query of a period, leaving out what is the default. */
export function periodSearch(kind: PeriodKind, date?: string): URLSearchParams {
  const q = new URLSearchParams();
  if (kind !== 'day') q.set('period', kind);
  if (date !== undefined) q.set('date', date);
  return q;
}

/** Whether the period runs past the latest day published. */
export const running = (p: Period) => p.through < p.end;

/** Whether the period holds the latest day published. */
export const latest = (p: Period) => p.later === undefined;

/** "Friday, October 16, 2026", "October 12 to 18, 2026", "October 2026" */
export function periodTitle(p: Period): string {
  switch (p.kind) {
    case 'day':
      return longDay(p.start);
    case 'week':
      return weekTitle(p.start, p.end);
    case 'month':
      return monthYear(p.start);
  }
}

function weekTitle(start: string, end: string): string {
  const year = end.slice(0, 4);
  if (start.slice(0, 4) !== year)
    return `${dayAndMonth(start)}, ${start.slice(0, 4)} to ${dayAndMonth(end)}, ${year}`;
  if (start.slice(5, 7) !== end.slice(5, 7))
    return `${dayAndMonth(start)} to ${dayAndMonth(end)}, ${year}`;
  return `${dayAndMonth(start)} to ${dayOfMonth(end)}, ${year}`;
}

/** The headline of the totals: "Curtailed on Friday", "Curtailed this week so far". */
export function periodHeadline(p: Period): string {
  switch (p.kind) {
    case 'day':
      return `Curtailed on ${weekday(p.start)}`;
    case 'week':
      if (!latest(p)) return 'Curtailed in the week';
      return running(p) ? 'Curtailed this week so far' : 'Curtailed this week';
    case 'month':
      return `Curtailed in ${monthName(p.start)}${running(p) ? ' so far' : ''}`;
  }
}

/** What the period is compared with: "Thursday", "the same days last week". */
export function previousName(p: Period): string | undefined {
  if (p.previousStart === undefined || p.previousEnd === undefined) return undefined;
  switch (p.kind) {
    case 'day':
      return weekday(p.previousStart);
    case 'week':
      return running(p) ? 'the same days the week before' : 'the week before';
    case 'month':
      return running(p)
        ? `${monthName(p.previousStart)} 1 to ${dayOfMonth(p.previousEnd)}`
        : monthName(p.previousStart);
  }
}

/** "in the day", "in the week", "in the month" */
export const inThePeriod = (p: Period) => `in the ${p.kind}`;
