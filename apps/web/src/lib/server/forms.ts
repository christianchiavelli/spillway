/** A text field as submitted, or '' when it is missing or a file. */
export function text(form: FormData, name: string): string {
  const value = form.get(name);
  return typeof value === 'string' ? value : '';
}

/**
 * A whole number as submitted: undefined when blank and NaN when not digits,
 * so the contract's check names the rule, required or invalid. Thousands
 * separators are refused, not guessed: "1,500" and "1.500" mean different
 * numbers in English and in Portuguese.
 */
export function wholeNumber(form: FormData, name: string): number | undefined {
  const value = text(form, name).trim();
  if (value === '') return undefined;
  return /^-?\d{1,9}$/.test(value) ? Number(value) : Number.NaN;
}
