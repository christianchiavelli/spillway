import type { Curtailment } from '#lib/api/generated/types.gen.ts';

import { dayOf } from './format.ts';

/** What each source lost in a slot, in MWh; null for a slot not published yet. */
export type SourceValues = { wind: number; solar: number };

const HALF_HOUR = 30 * 60 * 1000;

/** The days from start to end, both 'YYYY-MM-DD', end included. */
export function days(start: string, end: string): string[] {
  const out: string[] = [];
  for (let d = new Date(`${start}T00:00:00Z`); d.toISOString().slice(0, 10) <= end;) {
    out.push(d.toISOString().slice(0, 10));
    d = new Date(d.getTime() + 24 * 60 * 60 * 1000);
  }
  return out;
}

/** 'YYYY-MM-DD' plus n days. */
export function addDays(day: string, n: number): string {
  return new Date(Date.parse(`${day}T00:00:00Z`) + n * 24 * 60 * 60 * 1000)
    .toISOString()
    .slice(0, 10);
}

/**
 * The 48 half hours of a day's overview, on Brasília's clock: zero where the
 * ONS published no loss, null where it has published nothing yet.
 */
export function halfHours(o: Curtailment) {
  // Brasília keeps no summer time since 2019, so its days start at -03:00.
  const midnight = Date.parse(`${o.period.start}T00:00:00-03:00`);
  const latest = Date.parse(o.publishedThrough);
  const by = new Map(o.series.map((s) => [Date.parse(s.start), s]));
  const clock = (i: number) =>
    `${String(Math.floor(i / 2)).padStart(2, '0')}:${i % 2 ? '30' : '00'}`;
  return Array.from({ length: 48 }, (_, i) => {
    const at = midnight + i * HALF_HOUR;
    const slot = by.get(at);
    const values: SourceValues | null = slot
      ? { wind: slot.windMwh, solar: slot.solarMwh }
      : at <= latest
        ? { wind: 0, solar: 0 }
        : null;
    return { time: clock(i), end: i === 47 ? '24:00' : clock(i + 1), values };
  });
}

/** The days of a week's or a month's overview, null after the last published. */
export function dailySlots(o: Curtailment) {
  const by = new Map(o.series.map((s) => [dayOf(s.start), s]));
  return days(o.period.start, o.period.end).map((day) => {
    const slot = by.get(day);
    const values: SourceValues | null = slot
      ? { wind: slot.windMwh, solar: slot.solarMwh }
      : day <= o.period.through
        ? { wind: 0, solar: 0 }
        : null;
    return { day, values };
  });
}
