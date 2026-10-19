import { describe, expect, it } from 'vitest';

import { inPortuguese } from '#lib/testing/portuguese.ts';

import {
  compared,
  energy,
  gwh,
  longDay,
  mwh,
  onWeekday,
  pct,
  publishedThrough,
  shareOf,
  time,
  twh,
} from './format.ts';

describe('energy', () => {
  it('writes MWh under a gigawatt hour and GWh above, none when it rounds to nothing', () => {
    expect(energy(0.3)).toBe('None');
    expect(energy(850.4)).toBe('850 MWh');
    expect(energy(14_012)).toBe('14.0 GWh');
    expect(energy(150_400)).toBe('150 GWh');
  });

  it('keeps one decimal while a figure is small', () => {
    expect(gwh(7_389.7)).toBe('7.4');
    expect(gwh(123_456)).toBe('123');
    expect(twh(1_493_731.9)).toBe('1.5');
    expect(twh(9_651.6)).toBe('Under 0.1');
    expect(mwh(0.2)).toBe('Under 1 MWh');
    expect(mwh(169_327.4)).toBe('169,327 MWh');
  });
});

describe('shares', () => {
  it('writes one decimal under 10% and none above', () => {
    expect(pct(0.035)).toBe('3.5%');
    expect(pct(0.257)).toBe('26%');
  });

  it('measures the loss against what could have been generated', () => {
    expect(shareOf({ curtailedMwh: 25, generatedMwh: 75 })).toBe(0.25);
    expect(shareOf({ curtailedMwh: 0, generatedMwh: 0 })).toBe(0);
  });
});

describe('compared', () => {
  it('says how a period compares with the one before, in words', () => {
    expect(compared(14_000, 4_375, 'Thursday')).toBe('3.2 times as much as Thursday');
    expect(compared(110, 100, 'the week before')).toBe('10% more than the week before');
    expect(compared(90, 100, 'September')).toBe('10% less than September');
    expect(compared(500, 0, 'Thursday')).toBe('None on Thursday');
    expect(compared(0, 0, 'Thursday')).toBe('None, as on Thursday');
  });
});

describe('publishedThrough', () => {
  it('writes the last half hour on Brasília’s clock, in any zone', () => {
    // 02:30 UTC on the 17th is 23:30 on the 16th in Brasília.
    expect(publishedThrough('2026-10-17T02:30:00Z')).toBe('ONS data to October 16, 23:30');
  });
});

describe('in Portuguese', () => {
  inPortuguese();

  it('groups thousands with a dot and writes decimals with a comma', () => {
    expect(energy(0.3)).toBe('Nada');
    expect(energy(14_012)).toBe('14,0 GWh');
    expect(mwh(169_327.4)).toBe('169.327 MWh');
    expect(twh(9_651.6)).toBe('Menos de 0,1');
    expect(pct(0.035)).toBe('3,5%');
  });

  it('writes days as Brazil does, and the clock as both editions do', () => {
    expect(longDay('2026-10-16')).toBe('sexta-feira, 16 de outubro de 2026');
    expect(time('2026-10-17T02:30:00Z')).toBe('23:30');
    expect(publishedThrough('2026-10-17T02:30:00Z')).toBe('Dados do ONS até 16 de outubro, 23:30');
  });

  it('puts the preposition each weekday takes before it', () => {
    expect(onWeekday('2026-10-16')).toBe('na sexta-feira');
    expect(onWeekday('2026-10-17')).toBe('no sábado');
    expect(onWeekday('2026-10-18')).toBe('no domingo');
  });

  it('compares in Portuguese, the name carrying its own preposition', () => {
    expect(compared(14_000, 4_375, 'na quinta-feira')).toBe(
      '3,2 vezes o que foi cortado na quinta-feira',
    );
    expect(compared(110, 100, 'na semana anterior')).toBe('10% a mais que na semana anterior');
    expect(compared(0, 0, 'na quinta-feira')).toBe('Nada, assim como na quinta-feira');
  });
});
