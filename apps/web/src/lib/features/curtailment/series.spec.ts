import { describe, expect, it } from 'vitest';

import type { Curtailment } from '#lib/api/generated/types.gen.ts';

import { dailySlots, days, halfHours } from './series.ts';

const base: Curtailment = {
  period: { kind: 'day', start: '2026-10-16', end: '2026-10-16', through: '2026-10-16' },
  publishedThrough: '2026-10-16T16:30:00Z', // 13:30 in Brasília
  wind: { curtailedMwh: 0, generatedMwh: 0 },
  solar: { curtailedMwh: 0, generatedMwh: 0 },
  plantsLimited: 0,
  plantsReporting: 0,
  series: [{ start: '2026-10-16T16:00:00Z', windMwh: 12, solarMwh: 30 }],
  bySubsystem: [],
  byReason: [],
};

describe('halfHours', () => {
  it('lays a day out on Brasília’s clock, zero until the latest half hour and empty after', () => {
    const slots = halfHours(base);

    expect(slots).toHaveLength(48);
    expect(slots[26]).toEqual({ time: '13:00', end: '13:30', values: { wind: 12, solar: 30 } });
    expect(slots[0]?.values).toEqual({ wind: 0, solar: 0 });
    expect(slots[27]?.values).toEqual({ wind: 0, solar: 0 });
    expect(slots[28]?.values).toBeNull();
    expect(slots[47]?.end).toBe('24:00');
  });
});

describe('dailySlots', () => {
  it('fills the days published and leaves the rest of the period empty', () => {
    const slots = dailySlots({
      ...base,
      period: { kind: 'week', start: '2026-10-12', end: '2026-10-18', through: '2026-10-16' },
      series: [{ start: '2026-10-14T03:00:00Z', windMwh: 1500, solarMwh: 0 }],
    });

    expect(slots.map((s) => s.day)).toEqual(days('2026-10-12', '2026-10-18'));
    expect(slots[2]?.values).toEqual({ wind: 1500, solar: 0 });
    expect(slots[4]?.values).toEqual({ wind: 0, solar: 0 });
    expect(slots[5]?.values).toBeNull();
  });
});
