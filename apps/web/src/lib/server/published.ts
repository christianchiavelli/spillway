import { error } from '@sveltejs/kit';

import type { ApiResult } from './api.ts';

/**
 * The answer to a read of the ONS's data, or null while the first files are
 * still loading, which a page shows as a wait rather than an error. A day the
 * data does not reach is a 404, like any other address with nothing at it.
 */
export function published<T>(result: ApiResult<T>): T | null {
  if (result.ok) return result.data;
  if (result.problem.code === 'data_unavailable') return null;
  if (result.status === 404) error(404, 'The ONS has published nothing for that day.');
  error(400, 'Spillway cannot show what this address asks for.');
}
