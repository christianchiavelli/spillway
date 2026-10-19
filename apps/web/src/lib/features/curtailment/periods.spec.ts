import { describe, expect, it } from 'vitest';

import type { Period } from '#lib/api/generated/types.gen.ts';
import { inPortuguese } from '#lib/testing/portuguese.ts';

import {
  nextLabel,
  periodHeadline,
  periodKinds,
  periodSearch,
  periodTitle,
  previousLabel,
  previousName,
  readPeriod,
} from './periods.ts';

const url = (search: string) => new URL(`http://localhost/${search}`);

describe('readPeriod', () => {
  it('reads the period and the day, the latest day when they are left out', () => {
    expect(readPeriod(url(''))).toEqual({ period: 'day', date: undefined });
    expect(readPeriod(url('?period=week&date=2026-10-14'))).toEqual({
      period: 'week',
      date: '2026-10-14',
    });
  });

  it('refuses what no link of the app writes', () => {
    expect(readPeriod(url('?period=year'))).toBeNull();
    expect(readPeriod(url('?date=14/10/2026'))).toBeNull();
  });

  it('writes the default period out of the address', () => {
    expect(periodSearch('day').toString()).toBe('');
    expect(periodSearch('month', '2026-03-31').toString()).toBe('period=month&date=2026-03-31');
  });
});

const week: Period = {
  kind: 'week',
  start: '2026-10-12',
  end: '2026-10-18',
  through: '2026-10-16',
  previousStart: '2026-10-05',
  previousEnd: '2026-10-09',
  earlier: '2026-10-11',
};

describe('the words of a period', () => {
  it('titles a week by its days, across a month or a year', () => {
    expect(periodTitle(week)).toBe('October 12 to 18, 2026');
    expect(periodTitle({ ...week, start: '2026-09-28', end: '2026-10-04' })).toBe(
      'September 28 to October 4, 2026',
    );
    expect(periodTitle({ ...week, start: '2025-12-29', end: '2026-01-04' })).toBe(
      'December 29, 2025 to January 4, 2026',
    );
  });

  it('says so while a period runs, and compares it with as many days before', () => {
    expect(periodHeadline(week)).toBe('Curtailed this week so far');
    expect(previousName(week)).toBe('the same days the week before');

    const past = { ...week, through: '2026-10-18', later: '2026-10-19' };
    expect(periodHeadline(past)).toBe('Curtailed in the week');
    expect(previousName(past)).toBe('the week before');
  });

  it('names the days compared with in a month so far', () => {
    const month: Period = {
      kind: 'month',
      start: '2026-10-01',
      end: '2026-10-31',
      through: '2026-10-16',
      previousStart: '2026-09-01',
      previousEnd: '2026-09-16',
    };
    expect(periodHeadline(month)).toBe('Curtailed in October so far');
    expect(previousName(month)).toBe('September 1 to 16');
    expect(previousName({ ...month, previousStart: undefined, previousEnd: undefined })).toBe(
      undefined,
    );
  });
});

describe('the words of a period, in Portuguese', () => {
  inPortuguese();

  it('titles a week the way Brazil writes a span of days', () => {
    expect(periodTitle(week)).toBe('12 a 18 de outubro de 2026');
    expect(periodTitle({ ...week, start: '2026-09-28', end: '2026-10-04' })).toBe(
      '28 de setembro a 4 de outubro de 2026',
    );
    expect(periodTitle({ ...week, start: '2025-12-29', end: '2026-01-04' })).toBe(
      '29 de dezembro de 2025 a 4 de janeiro de 2026',
    );
  });

  it('says what it compares with as a place in time, ready to follow a comparison', () => {
    expect(periodHeadline(week)).toBe('Corte nesta semana, até agora');
    expect(previousName(week)).toBe('nos mesmos dias da semana anterior');
    expect(previousName({ ...week, through: '2026-10-18', later: '2026-10-19' })).toBe(
      'na semana anterior',
    );
  });

  it('names the kinds and the arrows in their gender', () => {
    expect(periodKinds.map((k) => k.label)).toEqual(['Dia', 'Semana', 'Mês']);
    expect(previousLabel('week')).toBe('Semana anterior');
    expect(nextLabel('day')).toBe('Próximo dia');
    expect(nextLabel('week')).toBe('Próxima semana');
  });
});
