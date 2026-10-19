import { error } from '@sveltejs/kit';

import { readPeriod } from '#lib/features/curtailment/periods.ts';
import { m } from '#lib/paraglide/messages.js';
import { api } from '#lib/server/api.ts';
import { published } from '#lib/server/published.ts';

import type { PageServerLoad } from './$types';

/** The plants the overview ranks; the plants page has them all. */
const TOP_PLANTS = 8;

// Public: the ONS's data is open, and the overview is the front door.
export const load: PageServerLoad = async (event) => {
  const query = readPeriod(event.url);
  if (query === null) error(404, m.error_no_period());
  const client = api(event);
  const [overview, plants] = await Promise.all([
    client.curtailment(query),
    client.listPlants({ ...query, limit: TOP_PLANTS }),
  ]);
  return { query, overview: published(overview), plants: published(plants) };
};
