import { describe, expect, it } from 'vitest';
import { page } from 'vitest/browser';
import { render } from 'vitest-browser-svelte';

import type { DailyCurtailment } from '#lib/api/generated/types.gen.ts';
import { inPortuguese } from '#lib/testing/portuguese.ts';

import Preview from './Preview.svelte';

const recent: DailyCurtailment = {
  from: '2026-10-14',
  to: '2026-10-16',
  series: [
    { source: 'wind', subsystem: 'NE', curtailedMwh: [150_000, 99_000, 100_000] },
    { source: 'solar', subsystem: 'NE', curtailedMwh: [0, 0, 0] },
  ],
};

describe('Preview', () => {
  it('counts the days over the threshold as the API does: as much is not over', async () => {
    await render(Preview, { recent, source: 'wind', subsystem: 'NE', threshold: 100_000 });

    await expect
      .element(page.getByText(/It would have fired on/))
      .toHaveTextContent('It would have fired on 1 of the last 3 days, last on October 14.');
  });

  it('says so when nothing would have set it off', async () => {
    await render(Preview, { recent, source: 'solar', subsystem: 'NE', threshold: 10 });

    await expect
      .element(page.getByText('It would not have fired in the last 3 days.'))
      .toBeVisible();
  });

  it('asks for what it still needs', async () => {
    await render(Preview, { recent, source: 'wind', subsystem: undefined, threshold: undefined });
    await expect.element(page.getByRole('heading', { name: 'No subsystem chosen' })).toBeVisible();
    await expect
      .element(page.getByText('Choose one to see what its wind plants lost.'))
      .toBeVisible();
  });
});

describe('Preview, in Portuguese', () => {
  inPortuguese();

  it('counts the days in a sentence of its own, not the English one translated word by word', async () => {
    await render(Preview, { recent, source: 'wind', subsystem: 'NE', threshold: 100_000 });

    await expect
      .element(page.getByText(/Ele teria disparado em/))
      .toHaveTextContent('Ele teria disparado em 1 dos últimos 3 dias, o último em 14 de outubro.');
  });

  it('makes the source agree with the plants it names', async () => {
    await render(Preview, { recent, source: 'solar', subsystem: undefined, threshold: undefined });

    await expect
      .element(page.getByText('Escolha um para ver o que as usinas solares dele perderam.'))
      .toBeVisible();
  });
});
