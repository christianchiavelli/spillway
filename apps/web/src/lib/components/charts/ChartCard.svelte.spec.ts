import { createRawSnippet } from 'svelte';
import { describe, expect, it } from 'vitest';
import { page } from 'vitest/browser';
import { render } from 'vitest-browser-svelte';

import ChartCard from './ChartCard.svelte';

const html = (markup: string) => createRawSnippet(() => ({ render: () => markup }));

describe('ChartCard', () => {
  it('opens on the chart, with the table one choice away', async () => {
    await render(ChartCard, {
      title: 'Day by Day',
      children: html('<svg role="img" aria-label="The chart"></svg>'),
      table: html('<table aria-label="The figures"></table>'),
    });

    const region = page.getByRole('region', { name: 'Day by Day' });
    const choice = region.getByRole('group', { name: 'Show as' });
    await expect.element(choice.getByRole('radio', { name: 'Chart' })).toBeChecked();
    await expect.element(choice.getByRole('radio', { name: 'Table' })).not.toBeChecked();
  });

  it('gives each card its own switch, so one choice never moves another', async () => {
    const one = await render(ChartCard, {
      title: 'One',
      children: html('<p>One chart</p>'),
      table: html('<p>One table</p>'),
    });
    const two = await render(ChartCard, {
      title: 'Two',
      children: html('<p>Two chart</p>'),
      table: html('<p>Two table</p>'),
    });

    const name = (c: HTMLElement) => c.querySelector('input[type=radio]')?.getAttribute('name');
    expect(name(one.container)).not.toBe(name(two.container));
  });
});
