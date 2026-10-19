import { error } from '@sveltejs/kit';

import { m } from '#lib/paraglide/messages.js';

import type { ApiResult } from './api.ts';

/**
 * The answer to a read of the ONS's data, or null while the first files are
 * still loading, which a page shows as a wait rather than an error. A day the
 * data does not reach is a 404, like any other address with nothing at it.
 */
export function published<T>(result: ApiResult<T>): T | null {
  if (result.ok) return result.data;
  if (result.problem.code === 'data_unavailable') return null;
  if (result.status === 404) error(404, m.error_nothing_published());
  error(400, m.error_bad_address());
}
