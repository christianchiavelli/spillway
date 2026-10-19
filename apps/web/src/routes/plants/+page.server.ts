import { error } from '@sveltejs/kit';

import { readPage, readSource, readSubsystem } from '#lib/features/curtailment/filters.ts';
import { readPeriod } from '#lib/features/curtailment/periods.ts';
import { m } from '#lib/paraglide/messages.js';
import { api } from '#lib/server/api.ts';
import { published } from '#lib/server/published.ts';

import type { PageServerLoad } from './$types';

const PAGE_SIZE = 20;

export const load: PageServerLoad = async (event) => {
  const query = readPeriod(event.url);
  const source = readSource(event.url);
  const subsystem = readSubsystem(event.url);
  const page = readPage(event.url);
  if (query === null || source === null || subsystem === null || page === null) {
    error(404, m.error_no_plants());
  }
  const result = await api(event).listPlants({
    ...query,
    source,
    subsystem,
    limit: PAGE_SIZE,
    offset: (page - 1) * PAGE_SIZE,
  });
  return { query, source, subsystem, page, pageSize: PAGE_SIZE, ranking: published(result) };
};
