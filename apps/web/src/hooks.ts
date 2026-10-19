import type { Reroute } from '@sveltejs/kit/hooks';

import { deLocalizeUrl } from '#lib/paraglide/runtime.js';

// Both editions share one set of routes: /pt-br/plants is /plants, read in
// Portuguese. Here, not in hooks.server.ts, so the browser routes the same way.
export const reroute: Reroute = ({ url }) => deLocalizeUrl(url).pathname;
