import type { Period, PeriodKind } from '#lib/api/generated/types.gen.ts';
import { m } from '#lib/paraglide/messages.js';

import {
  dayAndMonth,
  dayOfMonth,
  longDay,
  monthName,
  monthYear,
  onWeekday,
  weekday,
} from './format.ts';

// Labels are getters, read in the edition of the page that shows them: a
// constant would hold the first edition the server happened to render.
export const periodKinds = [
  {
    value: 'day',
    get label() {
      return m.period_day();
    },
  },
  {
    value: 'week',
    get label() {
      return m.period_week();
    },
  },
  {
    value: 'month',
    get label() {
      return m.period_month();
    },
  },
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
  const parts = {
    start: dayAndMonth(start),
    end: dayAndMonth(end),
    startDay: dayOfMonth(start),
    endDay: dayOfMonth(end),
    month: monthName(end),
    startYear: start.slice(0, 4),
    year,
  };
  if (parts.startYear !== year) return m.week_two_years(parts);
  if (start.slice(5, 7) !== end.slice(5, 7)) return m.week_two_months(parts);
  return m.week_same_month(parts);
}

/** The headline of the totals: "Curtailed on Friday", "Curtailed this week so far". */
export function periodHeadline(p: Period): string {
  switch (p.kind) {
    case 'day':
      return m.headline_day({ weekday: weekday(p.start), onWeekday: onWeekday(p.start) });
    case 'week':
      if (!latest(p)) return m.headline_week();
      return running(p) ? m.headline_this_week_so_far() : m.headline_this_week();
    case 'month': {
      const month = monthName(p.start);
      return running(p) ? m.headline_month_so_far({ month }) : m.headline_month({ month });
    }
  }
}

/**
 * What the period is compared with: "Thursday", "the same days the week
 * before". Portuguese puts its preposition in: "na quinta-feira", "nos mesmos
 * dias da semana anterior", for compared() to read on.
 */
export function previousName(p: Period): string | undefined {
  if (p.previousStart === undefined || p.previousEnd === undefined) return undefined;
  switch (p.kind) {
    case 'day':
      return onWeekday(p.previousStart);
    case 'week':
      return running(p) ? m.previous_same_days() : m.previous_week();
    case 'month': {
      const month = monthName(p.previousStart);
      return running(p)
        ? m.previous_month_days({ month, endDay: dayOfMonth(p.previousEnd) })
        : m.previous_month({ month });
    }
  }
}

/** "in the day", "in the week", "in the month" */
export const inThePeriod = (p: Period) =>
  ({ day: m.in_the_day, week: m.in_the_week, month: m.in_the_month })[p.kind]();

/** The arrows' names: "Previous week", "Next month". */
export const previousLabel = (kind: PeriodKind) =>
  ({ day: m.previous_period_day, week: m.previous_period_week, month: m.previous_period_month })[
    kind
  ]();
export const nextLabel = (kind: PeriodKind) =>
  ({ day: m.next_period_day, week: m.next_period_week, month: m.next_period_month })[kind]();
