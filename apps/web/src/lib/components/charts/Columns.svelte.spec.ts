import { describe, expect, it } from 'vitest';
import { page } from 'vitest/browser';
import { render } from 'vitest-browser-svelte';

import Columns from './Columns.svelte';

const series = [
  { label: 'Wind', fill: 'fill-chart-wind' },
  { label: 'Solar', fill: 'fill-chart-solar' },
];

describe('Columns', () => {
  it('names the chart for a screen reader and stacks a column for each slot published', async () => {
    const { container } = await render(Columns, {
      title: 'Energy curtailed each day',
      series,
      values: [[1, 2], null, [0, 0], [3, 0]],
      tick: String,
      xLabel: (i: number) => `Day ${i + 1}`,
      class: '',
    });

    await expect
      .element(page.getByRole('img', { name: 'Energy curtailed each day' }))
      .toBeVisible();
    // Two wind segments, one under solar and one alone; nothing for zero.
    expect(container.querySelectorAll('.fill-chart-wind')).toHaveLength(2);
    expect(container.querySelectorAll('.fill-chart-solar')).toHaveLength(1);
    // The slot not published yet keeps its label, greyed.
    expect(container.querySelector('.fill-text-primary-disabled')?.textContent).toBe('Day 2');
  });

  it('paints a whole column one way when told, as the preview does under the threshold', async () => {
    const { container } = await render(Columns, {
      title: 'Against the threshold',
      series: [series[0] ?? { label: 'Wind', fill: 'fill-chart-wind' }],
      values: [[50], [150]],
      tick: String,
      xLabel: () => null,
      rule: { value: 100, label: '100 MWh', caption: 'threshold' },
      fill: (i: number) => (i === 1 ? 'fill-chart-wind' : 'fill-chart-muted'),
      class: '',
    });

    expect(container.querySelectorAll('.fill-chart-muted')).toHaveLength(1);
    expect(container.querySelectorAll('.fill-chart-wind')).toHaveLength(1);
    expect(container.querySelector('[stroke-dasharray]')).not.toBeNull();
  });

  it('leaves out the labels that would overlap on a narrow chart', async () => {
    await page.viewport(320, 640);
    const { container } = await render(Columns, {
      title: 'Thirty days',
      series,
      values: Array.from({ length: 30 }, () => [1, 1]),
      tick: String,
      xLabel: () => 'Oct 17',
      class: '',
    });

    await expect
      .poll(
        () =>
          [...container.querySelectorAll('text')].filter((t) => t.textContent === 'Oct 17').length,
      )
      .toBeLessThan(10);
  });
});
