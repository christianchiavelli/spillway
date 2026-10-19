import {
  AlertListSchema,
  AlertNameSchema,
  ThresholdSchema,
} from '#lib/api/generated/schemas.gen.ts';
import type { FieldErrorCode } from '#lib/forms/errors.ts';

/** The fields of the alert form, in the order it shows them. */
export const alertFields = ['source', 'name', 'subsystem', 'thresholdMwh'] as const;
export type AlertField = (typeof alertFields)[number];

/** Read from the contract, so the form and the API never disagree. */
export const limits = {
  nameMaxLength: AlertNameSchema.maxLength,
  thresholdMin: ThresholdSchema.minimum,
  thresholdMax: ThresholdSchema.maximum,
  perAccount: AlertListSchema.properties.items.maxItems,
} as const;

const whole = new Intl.NumberFormat('en-US');

/** "1,500 MWh": a whole number, grouped in thousands, with its unit. */
export function formatMWh(value: number): string {
  return `${whole.format(value)} MWh`;
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
