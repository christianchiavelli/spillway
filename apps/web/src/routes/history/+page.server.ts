import { api } from '#lib/server/api.ts';
import { published } from '#lib/server/published.ts';

import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async (event) => ({
  history: published(await api(event).listMonths()),
});
