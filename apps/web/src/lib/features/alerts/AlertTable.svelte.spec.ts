import { describe, expect, it } from 'vitest';
import { page } from 'vitest/browser';
import { render } from 'vitest-browser-svelte';

import type { WatchedAlert } from '#lib/api/generated/types.gen.ts';

import AlertTable from './AlertTable.svelte';

const alerts: WatchedAlert[] = [
  {
    id: '019a3b6e-0000-7000-8000-000000000002',
    name: 'Solar in the Northeast',
    source: 'solar',
    subsystem: 'NE',
    thresholdMwh: 12_500,
    createdAt: '2026-10-17T02:30:00Z',
    lastFiredOn: '2026-10-11',
    daysFired: 7,
  },
  {
    id: '019a3b6e-0000-7000-8000-000000000001',
    name: 'Wind in the South',
    source: 'wind',
    subsystem: 'S',
    thresholdMwh: 800,
    createdAt: '2026-10-15T12:00:00Z',
    daysFired: 0,
  },
];

describe('AlertTable', () => {
  it('names each row after its alert, so every cell is read with it', async () => {
    await render(AlertTable, { alerts, ondelete: () => undefined, deleting: null });

    const rows = page.getByRole('row');
    await expect
      .element(rows.nth(1).getByRole('rowheader'))
      .toHaveTextContent('Solar in the Northeast');
    await expect.element(rows.nth(2).getByRole('rowheader')).toHaveTextContent('Wind in the South');
  });

  it('writes numbers, units, places and days the way every page does', async () => {
    await render(AlertTable, { alerts, ondelete: () => undefined, deleting: null });

    // Vitest 5's toHaveTextContent wants the whole text; toMatchTextContent a part.
    const first = page.getByRole('row').nth(1);
    await expect.element(first).toMatchTextContent('12,500 MWh');
    await expect.element(first).toMatchTextContent('Northeast');
    // A day of the ONS's calendar, written as the date it is, in any zone.
    await expect.element(first).toMatchTextContent('Oct 11, 2026');
    await expect.element(first).toMatchTextContent('7 days');

    const quiet = page.getByRole('row').nth(2);
    await expect.element(quiet).toMatchTextContent('Never');
    await expect.element(quiet).toMatchTextContent('None');
  });

  it('gives every delete button the name of what it deletes', async () => {
    await render(AlertTable, {
      alerts,
      ondelete: () => undefined,
      deleting: alerts[1]?.id ?? null,
    });

    await expect
      .element(page.getByRole('button', { name: 'Delete “Solar in the Northeast”' }))
      .toBeInTheDocument();
    const busy = page.getByRole('button', { name: 'Deleting… “Wind in the South”' });
    await expect.element(busy).toHaveAttribute('aria-disabled', 'true');
  });
});
