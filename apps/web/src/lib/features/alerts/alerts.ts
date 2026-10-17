import { type LucideIcon, Sun, Wind } from '@lucide/svelte';

import {
  AlertListSchema,
  AlertNameSchema,
  ThresholdSchema,
} from '#lib/api/generated/schemas.gen.ts';
import type { Source, Subsystem } from '#lib/api/generated/types.gen.ts';
import type { FieldErrorCode } from '#lib/forms/errors.ts';

/** The fields of the alert form, in the order it shows them. */
export const alertFields = ['name', 'source', 'subsystem', 'thresholdMwh'] as const;
export type AlertField = (typeof alertFields)[number];

/** Read from the contract, so the form and the API never disagree. */
export const limits = {
  nameMaxLength: AlertNameSchema.maxLength,
  thresholdMin: ThresholdSchema.minimum,
  thresholdMax: ThresholdSchema.maximum,
  perAccount: AlertListSchema.properties.items.maxItems,
} as const;

export const sources = [
  { value: 'wind', label: 'Wind', icon: Wind },
  { value: 'solar', label: 'Solar', icon: Sun },
] as const satisfies readonly { value: Source; label: string; icon: LucideIcon }[];

/** The ONS's names for its subsystems, in its usual order, north to south. */
export const subsystems = [
  { value: 'N', label: 'North' },
  { value: 'NE', label: 'Northeast' },
  { value: 'SE', label: 'Southeast and Centre-West' },
  { value: 'S', label: 'South' },
] as const satisfies readonly { value: Subsystem; label: string }[];

export function sourceOf(value: Source) {
  return sources.find((s) => s.value === value) ?? sources[0];
}

export function subsystemLabel(value: Subsystem): string {
  return subsystems.find((s) => s.value === value)?.label ?? value;
}

const whole = new Intl.NumberFormat('en-US');

/** "1,500 MWh": a whole number, grouped in thousands, with its unit. */
export function formatMWh(value: number): string {
  return `${whole.format(value)} MWh`;
}

// Brasília time, where the ONS's days begin and end; fixed, so the server and
// the browser write the same date and hydration finds nothing to change.
const day = new Intl.DateTimeFormat('en-US', {
  month: 'short',
  day: '2-digit',
  year: 'numeric',
  timeZone: 'America/Sao_Paulo',
});

/** "Oct 17, 2026": the short date every page writes. */
export function formatDay(iso: string): string {
  return day.format(new Date(iso));
}

const range = `${whole.format(limits.thresholdMin)} to ${whole.format(limits.thresholdMax)}`;

const messages: Record<AlertField, Partial<Record<FieldErrorCode, string>>> = {
  name: {
    required: 'Enter a name for the alert',
    too_long: `Use ${limits.nameMaxLength} characters or fewer`,
    invalid: 'Use letters, numbers and punctuation only',
    taken: 'You already have an alert with this name',
  },
  source: {
    required: 'Choose wind or solar',
    invalid: 'Choose wind or solar',
  },
  subsystem: {
    required: 'Choose a subsystem',
    invalid: 'Choose a subsystem',
  },
  thresholdMwh: {
    required: 'Enter a daily threshold in MWh',
    invalid: 'Enter a whole number of MWh, such as 500',
    out_of_range: `Enter a number from ${range}`,
  },
};

/** The sentence for a broken rule, or undefined when the field is fine. */
export function alertError(
  field: AlertField,
  code: FieldErrorCode | undefined,
): string | undefined {
  if (code === undefined) return undefined;
  return messages[field][code] ?? 'Check this field';
}
