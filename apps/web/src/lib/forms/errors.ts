import type { Problem } from '#lib/api/generated/types.gen.ts';

/**
 * The rules a field can break, in the words the API uses. The form shows each
 * one as a sentence of its own, so the code decides the copy, not the layer
 * that caught it: the server's Zod check and the Go API name a rule alike.
 */
export type FieldErrorCode =
  'required' | 'too_short' | 'too_long' | 'out_of_range' | 'invalid' | 'taken';

export type FieldErrors<Field extends string> = Partial<Record<Field, FieldErrorCode>>;

const codes = new Set<string>([
  'required',
  'too_short',
  'too_long',
  'out_of_range',
  'invalid',
  'taken',
]);

/** The part of a Zod issue this module reads; every Zod 4 issue has it. */
interface Issue {
  code: string;
  path: readonly PropertyKey[];
  origin?: string;
}

/**
 * Names the rule behind each Zod issue on one of fields, keeping the first
 * issue per field and ignoring the rest. Zod leaves the input out of its
 * issues, so whether a field was left blank is read from values.
 */
export function fromIssues<Field extends string>(
  issues: readonly Issue[],
  fields: readonly Field[],
  values: Partial<Record<Field, unknown>>,
): FieldErrors<Field> {
  const errors: FieldErrors<Field> = {};
  for (const issue of issues) {
    const field = fields.find((f) => f === issue.path[0]);
    if (field === undefined || errors[field] !== undefined) continue;
    const value = values[field];
    errors[field] = codeOf(issue, value === undefined || value === '');
  }
  return errors;
}

function codeOf(issue: Issue, blank: boolean): FieldErrorCode {
  if (blank) return 'required';
  const numeric = issue.origin === 'number' || issue.origin === 'int';
  switch (issue.code) {
    case 'too_small':
      return numeric ? 'out_of_range' : 'too_short';
    case 'too_big':
      return numeric ? 'out_of_range' : 'too_long';
    default:
      return 'invalid';
  }
}

/**
 * Reads the field errors of an API problem. A pointer is a JSON Pointer into
 * the request body, `/name` for the field `name`. A code the form does not
 * know becomes `invalid`, which every field can show.
 */
export function fromProblem<Field extends string>(
  problem: Pick<Problem, 'errors'>,
  fields: readonly Field[],
): FieldErrors<Field> {
  const errors: FieldErrors<Field> = {};
  for (const { pointer, code } of problem.errors ?? []) {
    const field = fields.find((f) => `/${f}` === pointer);
    if (field === undefined || errors[field] !== undefined) continue;
    errors[field] = codes.has(code) ? (code as FieldErrorCode) : 'invalid';
  }
  return errors;
}

/**
 * Names the rule a field breaks by the browser's own check of its attributes,
 * `required`, `type`, `minlength` and the rest, which the form sets from the
 * contract: the same rules the server applies, read before anything is sent.
 */
export function fromValidity(validity: ValidityState): FieldErrorCode | undefined {
  if (validity.valueMissing) return 'required';
  if (validity.tooShort) return 'too_short';
  if (validity.tooLong) return 'too_long';
  if (validity.rangeUnderflow || validity.rangeOverflow) return 'out_of_range';
  return validity.valid ? undefined : 'invalid';
}

/** The first field with an error, in the order the form shows them. */
export function firstInvalid<Field extends string>(
  errors: FieldErrors<Field> | undefined,
  order: readonly Field[],
): Field | undefined {
  return order.find((f) => errors?.[f] !== undefined);
}
