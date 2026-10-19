import {
  AlertListSchema,
  AlertNameSchema,
  ThresholdSchema,
} from '#lib/api/generated/schemas.gen.ts';
import { int } from '#lib/features/curtailment/format.ts';
import type { FieldErrorCode } from '#lib/forms/errors.ts';
import { m } from '#lib/paraglide/messages.js';

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

/** "1,500 MWh" or "1.500 MWh": a whole number, grouped in thousands, with its unit. */
export function formatMWh(value: number): string {
  return `${int(value)} MWh`;
}

// Functions, not strings, so each request reads them in its own edition.
const messages: Record<AlertField, Partial<Record<FieldErrorCode, () => string>>> = {
  name: {
    required: m.alert_name_required,
    too_long: () => m.alert_name_too_long({ max: limits.nameMaxLength }),
    invalid: m.alert_name_invalid,
    taken: m.alert_name_taken,
  },
  source: {
    required: m.alert_source_required,
    invalid: m.alert_source_required,
  },
  subsystem: {
    required: m.alert_subsystem_required,
    invalid: m.alert_subsystem_required,
  },
  thresholdMwh: {
    required: m.alert_threshold_required,
    invalid: m.alert_threshold_invalid,
    out_of_range: () =>
      m.alert_threshold_range({ min: int(limits.thresholdMin), max: int(limits.thresholdMax) }),
  },
};

/** The sentence for a broken rule, or undefined when the field is fine. */
export function alertError(
  field: AlertField,
  code: FieldErrorCode | undefined,
): string | undefined {
  if (code === undefined) return undefined;
  return (messages[field][code] ?? m.form_check_field)();
}
