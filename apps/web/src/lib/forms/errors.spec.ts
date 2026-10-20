import { describe, expect, it } from 'vitest';
import * as z from 'zod';

import { zNewAccount, zNewAlert } from '#lib/api/generated/zod.gen.ts';

import { firstInvalid, fromIssues, fromProblem, fromValidity } from './errors.ts';

const alertFields = ['name', 'source', 'subsystem', 'thresholdMwh'] as const;

function check<Field extends string>(
  schema: z.ZodType,
  fields: readonly Field[],
  input: Partial<Record<Field, unknown>>,
) {
  const result = schema.safeParse(input);
  if (result.success) throw new Error('expected the input to fail');
  return fromIssues(result.error.issues, fields, input);
}

describe('fromIssues', () => {
  it('names the rule each field broke, in the API’s words', () => {
    const errors = check(zNewAlert, alertFields, {
      name: '',
      source: 'coal',
      subsystem: undefined,
      thresholdMwh: 0,
    });
    expect(errors).toEqual({
      name: 'required',
      source: 'invalid',
      subsystem: 'required',
      thresholdMwh: 'out_of_range',
    });
  });

  it('tells a long name and a huge threshold apart', () => {
    const errors = check(zNewAlert, alertFields, {
      name: 'a'.repeat(81),
      source: 'wind',
      subsystem: 'NE',
      thresholdMwh: 1_000_001,
    });
    expect(errors).toEqual({ name: 'too_long', thresholdMwh: 'out_of_range' });
  });

  it('reads a threshold that is not a whole number as invalid', () => {
    const errors = check(zNewAlert, alertFields, {
      name: 'x',
      source: 'wind',
      subsystem: 'NE',
      thresholdMwh: Number.NaN,
    });
    expect(errors).toEqual({ thresholdMwh: 'invalid' });
  });

  it('tells a short password from a missing one', () => {
    const fields = ['email', 'password'] as const;
    expect(check(zNewAccount, fields, { email: 'ana@example.com', password: 'short' })).toEqual({
      password: 'too_short',
    });
    expect(check(zNewAccount, fields, { email: '', password: '' })).toEqual({
      email: 'required',
      password: 'required',
    });
    expect(check(zNewAccount, fields, { email: 'ana', password: 'x'.repeat(15) })).toEqual({
      email: 'invalid',
    });
  });
});

describe('fromProblem', () => {
  it('follows each pointer to its field', () => {
    expect(
      fromProblem(
        {
          errors: [
            { pointer: '/name', code: 'taken' },
            { pointer: '/thresholdMwh', code: 'out_of_range' },
            { pointer: '/name', code: 'too_long' },
          ],
        },
        alertFields,
      ),
    ).toEqual({ name: 'taken', thresholdMwh: 'out_of_range' });
  });

  it('shows a code it does not know as invalid, and skips fields the form lacks', () => {
    expect(
      fromProblem(
        {
          errors: [
            { pointer: '/source', code: 'brand_new_rule' },
            { pointer: '/owner', code: 'invalid' },
          ],
        },
        alertFields,
      ),
    ).toEqual({ source: 'invalid' });
  });
});

describe('firstInvalid', () => {
  it('finds the first field with an error in form order', () => {
    expect(firstInvalid({ thresholdMwh: 'invalid', source: 'invalid' }, alertFields)).toBe(
      'source',
    );
    expect(firstInvalid({}, alertFields)).toBeUndefined();
    expect(firstInvalid(undefined, alertFields)).toBeUndefined();
  });
});

describe('fromValidity', () => {
  // Only what the check reads: a browser's ValidityState has every flag false
  // but the broken one, and valid true when none is set.
  const validity = (broken?: keyof ValidityState) =>
    ({ valid: broken === undefined, ...(broken && { [broken]: true }) }) as ValidityState;

  it('names the browser’s verdict with the code the server would send', () => {
    expect(fromValidity(validity('valueMissing'))).toBe('required');
    expect(fromValidity(validity('typeMismatch'))).toBe('invalid');
    expect(fromValidity(validity('tooShort'))).toBe('too_short');
    expect(fromValidity(validity('tooLong'))).toBe('too_long');
    expect(fromValidity(validity('rangeOverflow'))).toBe('out_of_range');
  });

  it('finds nothing wrong with a valid field', () => {
    expect(fromValidity(validity())).toBeUndefined();
  });
});
