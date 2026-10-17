import { EmailSchema, NewAccountSchema } from '#lib/api/generated/schemas.gen.ts';
import type { FieldErrorCode } from '#lib/forms/errors.ts';

/** The fields of the sign-in and sign-up forms, in the order they show them. */
export const accountFields = ['email', 'password'] as const;
export type AccountField = (typeof accountFields)[number];

/** Read from the contract, so the form and the API never disagree. */
export const limits = {
  emailMaxLength: EmailSchema.maxLength,
  passwordMinLength: NewAccountSchema.properties.password.minLength,
  passwordMaxLength: NewAccountSchema.properties.password.maxLength,
} as const;

export const passwordHint = `At least ${limits.passwordMinLength} characters. A few unrelated words in a row make a strong one.`;

const messages: Record<AccountField, Partial<Record<FieldErrorCode, string>>> = {
  email: {
    required: 'Enter your email address',
    invalid: 'Enter an email address like name@example.com',
    too_long: `Use an email address of ${limits.emailMaxLength} characters or fewer`,
    taken: 'An account already uses this email address',
  },
  password: {
    required: 'Enter a password',
    too_short: `Use at least ${limits.passwordMinLength} characters`,
    too_long: `Use ${limits.passwordMaxLength} characters or fewer`,
  },
};

/** The sentence for a broken rule, or undefined when the field is fine. */
export function accountError(
  field: AccountField,
  code: FieldErrorCode | undefined,
): string | undefined {
  if (code === undefined) return undefined;
  return messages[field][code] ?? 'Check this field';
}
