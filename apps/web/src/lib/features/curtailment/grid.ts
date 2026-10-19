import SolarPower from '~icons/material-symbols/solar-power-outline-rounded';
import WindPower from '~icons/material-symbols/wind-power-outline-rounded';

import type { Reason, Source, Subsystem } from '#lib/api/generated/types.gen.ts';
import type { Icon } from '#lib/components/ui/icon.ts';
import { m } from '#lib/paraglide/messages.js';

// A turbine and a panel, not the weather: what curtailment holds back is the
// plants' output. The classes are written out whole, so Tailwind finds them.
// Every label here is a getter, read in the edition of the page that shows it:
// a constant would hold the first edition the server happened to render.
export const sources = [
  {
    value: 'wind',
    get label() {
      return m.source_wind();
    },
    icon: WindPower,
    fill: 'fill-chart-wind',
    stroke: 'stroke-chart-wind',
    swatch: 'bg-chart-wind',
  },
  {
    value: 'solar',
    get label() {
      return m.source_solar();
    },
    icon: SolarPower,
    fill: 'fill-chart-solar',
    stroke: 'stroke-chart-solar',
    swatch: 'bg-chart-solar',
  },
] as const satisfies readonly {
  value: Source;
  label: string;
  icon: Icon;
  fill: string;
  stroke: string;
  swatch: string;
}[];

export function sourceOf(value: Source) {
  return sources.find((s) => s.value === value) ?? sources[0];
}

/** The ONS's names for its subsystems, in its usual order, north to south. */
export const subsystems = [
  {
    value: 'N',
    get label() {
      return m.subsystem_n();
    },
  },
  {
    value: 'NE',
    get label() {
      return m.subsystem_ne();
    },
  },
  {
    value: 'SE',
    get label() {
      return m.subsystem_se();
    },
  },
  {
    value: 'S',
    get label() {
      return m.subsystem_s();
    },
  },
] as const satisfies readonly { value: Subsystem; label: string }[];

export function subsystemLabel(value: Subsystem): string {
  return subsystems.find((s) => s.value === value)?.label ?? value;
}

/** The reasons the ONS gives, in plain words, and what each one means. */
const reason = (label: () => string, meaning: () => string) => ({
  get label() {
    return label();
  },
  get meaning() {
    return meaning();
  },
});

export const reasons = {
  ENE: reason(m.reason_ene, m.reason_ene_meaning),
  CNF: reason(m.reason_cnf, m.reason_cnf_meaning),
  REL: reason(m.reason_rel, m.reason_rel_meaning),
  PAR: reason(m.reason_par, m.reason_par_meaning),
} as const satisfies Record<Reason, { label: string; meaning: string }>;
