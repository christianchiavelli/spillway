import SolarPower from '~icons/material-symbols/solar-power-outline-rounded';
import WindPower from '~icons/material-symbols/wind-power-outline-rounded';

import type { Reason, Source, Subsystem } from '#lib/api/generated/types.gen.ts';
import type { Icon } from '#lib/components/ui/icon.ts';

// A turbine and a panel, not the weather: what curtailment holds back is the
// plants' output. The classes are written out whole, so Tailwind finds them.
export const sources = [
  {
    value: 'wind',
    label: 'Wind',
    icon: WindPower,
    fill: 'fill-chart-wind',
    stroke: 'stroke-chart-wind',
    swatch: 'bg-chart-wind',
  },
  {
    value: 'solar',
    label: 'Solar',
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
  { value: 'N', label: 'North' },
  { value: 'NE', label: 'Northeast' },
  { value: 'SE', label: 'Southeast and Centre-West' },
  { value: 'S', label: 'South' },
] as const satisfies readonly { value: Subsystem; label: string }[];

export function subsystemLabel(value: Subsystem): string {
  return subsystems.find((s) => s.value === value)?.label ?? value;
}

/** The reasons the ONS gives, in plain words, and what each one means. */
export const reasons = {
  ENE: {
    label: 'Supply above demand',
    meaning: 'The plants could make more than the country was using',
  },
  CNF: { label: 'Grid reliability', meaning: 'Held back to keep the grid stable' },
  REL: { label: 'Transmission limits', meaning: 'The lines could not carry it' },
  PAR: { label: 'Connection terms', meaning: 'A limit agreed when the plant was connected' },
} as const satisfies Record<Reason, { label: string; meaning: string }>;
