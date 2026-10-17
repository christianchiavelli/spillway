import { defineEnvVars } from '@sveltejs/kit/env';
import * as z from 'zod';

import { building } from '$app/env';

const apiUrl = z.url({ protocol: /^https?$/ });

// Checked when the server starts: a missing or malformed value stops it with a
// message, instead of failing on the first request. The build reads this file
// too, with no API to point at, so there the value may be absent.
export const variables = defineEnvVars({
  API_URL: {
    description: 'Where the Go API listens, as this server reaches it. Never sent to the browser.',
    schema: (value) => (building ? (value ?? '') : apiUrl.parse(value)),
  },
});
