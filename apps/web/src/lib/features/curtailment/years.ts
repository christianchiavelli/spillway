import type { MonthCurtailment, SourceEnergy } from '#lib/api/generated/types.gen.ts';

export type Year = {
  year: string;
  /** The first and the last month in it, for a year the data covers in part. */
  first: string;
  last: string;
  wind: SourceEnergy | undefined;
  solar: SourceEnergy | undefined;
};

function add(a: SourceEnergy | undefined, b: SourceEnergy | undefined): SourceEnergy | undefined {
  if (b === undefined) return a;
  return {
    curtailedMwh: (a?.curtailedMwh ?? 0) + b.curtailedMwh,
    generatedMwh: (a?.generatedMwh ?? 0) + b.generatedMwh,
  };
}

/** The months summed by year, the latest year first. */
export function byYear(months: readonly MonthCurtailment[]): Year[] {
  const years = new Map<string, Year>();
  for (const m of months) {
    const year = m.month.slice(0, 4);
    const y = years.get(year) ?? {
      year,
      first: m.month,
      last: m.month,
      wind: undefined,
      solar: undefined,
    };
    y.last = m.month;
    y.wind = add(y.wind, m.wind);
    y.solar = add(y.solar, m.solar);
    years.set(year, y);
  }
  return [...years.values()].reverse();
}
