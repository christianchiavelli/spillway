import { describe, expect, it, vi } from 'vitest';
import { page } from 'vitest/browser';
import { render } from 'vitest-browser-svelte';

import type { ResolvedPathname } from '$app/types';

import BarRows from './BarRows.svelte';

const series = [
  { label: 'Wind', fill: 'fill-chart-wind' },
  { label: 'Solar', fill: 'fill-chart-solar' },
];

const rows = [
  {
    key: 'NE',
    label: 'Northeast',
    values: [55_700, 40_800],
    value: '96.5 GWh',
    href: '/plants?subsystem=NE' as ResolvedPathname,
  },
  {
    key: 'N',
    label: 'North',
    values: [1_900, 0],
    value: '1.9 GWh',
    href: '/plants?subsystem=N' as ResolvedPathname,
  },
];

describe('BarRows', () => {
  it('reads as a ranked list, each row a link with its figure', async () => {
    const { container } = await render(BarRows, { title: 'By subsystem', series, rows });

    const list = page.getByRole('list', { name: 'By subsystem' });
    await expect.element(list).toBeVisible();
    // In the order given, the largest first.
    await expect.element(list).toHaveTextContent('Northeast 96.5 GWh North 1.9 GWh');
    await expect
      .element(list.getByRole('link', { name: 'North', exact: true }))
      .toHaveAttribute('href', '/plants?subsystem=N');
    // Both sources in the first bar, wind alone in the second: nothing for zero.
    expect(container.querySelectorAll('.fill-chart-wind')).toHaveLength(2);
    expect(container.querySelectorAll('.fill-chart-solar')).toHaveLength(1);
  });

  it('says what each row means, under its label', async () => {
    await render(BarRows, {
      title: 'By reason',
      series,
      rows: [
        {
          key: 'REL',
          label: 'Transmission limits',
          description: 'The lines could not carry it',
          values: [4_500, 0],
          value: '4.5 GWh',
        },
      ],
    });

    await expect.element(page.getByText('The lines could not carry it')).toBeVisible();
    await expect.element(page.getByRole('link')).not.toBeInTheDocument();
  });

  it('tells which row is pointed at or focused, and highlights the one it is told', async () => {
    const onactive = vi.fn();
    await render(BarRows, { title: 'By subsystem', series, rows, active: 'N', onactive });

    await expect
      .element(page.getByRole('listitem').nth(1))
      .toHaveClass(/bg-surface-container-hover/);
    const northeast = page.getByRole('listitem').first();
    await northeast.hover();
    expect(onactive).toHaveBeenLastCalledWith('NE');
    await northeast.unhover();
    expect(onactive).toHaveBeenLastCalledWith(null);
    // The keyboard too: the row's link, named by its label, takes focus.
    page.getByRole('link', { name: 'North', exact: true }).element().focus();
    expect(onactive).toHaveBeenLastCalledWith('N');
  });
});
