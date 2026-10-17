import { describe, expect, it } from 'vitest';

import { text, wholeNumber } from './forms.ts';

function form(entries: Record<string, string | Blob>) {
  const data = new FormData();
  for (const [key, value] of Object.entries(entries)) data.append(key, value);
  return data;
}

describe('text', () => {
  it('reads a field, and treats a missing one or a file as blank', () => {
    const data = form({ name: ' Wind ', upload: new Blob(['x']) });
    expect(text(data, 'name')).toBe(' Wind ');
    expect(text(data, 'missing')).toBe('');
    expect(text(data, 'upload')).toBe('');
  });
});

describe('wholeNumber', () => {
  it('reads digits and nothing else', () => {
    const data = form({
      a: ' 500 ',
      b: '',
      c: '1,500',
      d: '1.5',
      e: 'abc',
      f: '-3',
      g: '9999999999',
    });
    expect(wholeNumber(data, 'a')).toBe(500);
    expect(wholeNumber(data, 'b')).toBeUndefined();
    expect(wholeNumber(data, 'missing')).toBeUndefined();
    expect(wholeNumber(data, 'c')).toBeNaN();
    expect(wholeNumber(data, 'd')).toBeNaN();
    expect(wholeNumber(data, 'e')).toBeNaN();
    expect(wholeNumber(data, 'f')).toBe(-3);
    expect(wholeNumber(data, 'g')).toBeNaN();
  });
});
