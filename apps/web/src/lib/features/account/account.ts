import { EmailSchema, NewAccountSchema } from '#lib/api/generated/schemas.gen.ts';
import type { FieldErrorCode } from '#lib/forms/errors.ts';
import { m } from '#lib/paraglide/messages.js';

/** The fields of the sign-in and sign-up forms, in the order they show them. */
export const accountFields = ['email', 'password'] as const;
export type AccountField = (typeof accountFields)[number];

/** Read from the contract, so the form and the API never disagree. */
export const limits = {
  emailMaxLength: EmailSchema.maxLength,
  passwordMinLength: NewAccountSchema.properties.password.minLength,
  passwordMaxLength: NewAccountSchema.properties.password.maxLength,
} as const;

export const passwordHint = () => m.account_password_hint({ min: limits.passwordMinLength });

// Functions, not strings, so each request reads them in its own edition.
const messages: Record<AccountField, Partial<Record<FieldErrorCode, () => string>>> = {
  email: {
    required: m.account_email_required,
    invalid: m.account_email_invalid,
    too_long: () => m.account_email_too_long({ max: limits.emailMaxLength }),
    taken: m.account_email_taken,
  },
  password: {
    required: m.account_password_required,
    too_short: () => m.account_password_too_short({ min: limits.passwordMinLength }),
    too_long: () => m.account_password_too_long({ max: limits.passwordMaxLength }),
  },
};

/** The sentence for a broken rule, or undefined when the field is fine. */
export function accountError(
  field: AccountField,
  code: FieldErrorCode | undefined,
): string | undefined {
  if (code === undefined) return undefined;
  return (messages[field][code] ?? m.form_check_field)();
}

type Input = Event & { currentTarget: EventTarget & HTMLInputElement };

/** What AccountForm hands each of its fields, spread onto the field. */
export type AccountFieldProps = {
  name: AccountField;
  error: string | undefined;
  autofocus: boolean;
  onblur: (event: Input) => void;
  oninput: (event: Input) => void;
};

/** The props of each field of AccountForm, by name. */
export type FieldFor = (name: AccountField) => AccountFieldProps;
