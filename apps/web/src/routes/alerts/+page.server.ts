import { error, fail, redirect } from '@sveltejs/kit';
import * as z from 'zod';

import { resolve } from '$app/paths';
import { readPage } from '#lib/features/curtailment/filters.ts';
import { localized } from '#lib/i18n.ts';
import { m } from '#lib/paraglide/messages.js';
import { api } from '#lib/server/api.ts';
import { setFlash } from '#lib/server/flash.ts';
import { text } from '#lib/server/forms.ts';
import { requireSession, sessionLost } from '#lib/server/session.ts';

import type { Actions, PageServerLoad } from './$types';

/** Firings a page. */
const FIRINGS = 10;

// +page.server.ts, not +page.ts: the list needs the session token, which must
// never reach the browser, and the API is not reachable from it anyway.
export const load: PageServerLoad = async (event) => {
  requireSession(event);
  const page = readPage(event.url);
  if (page === null) error(404, m.alerts_no_page());
  const client = api(event);
  const [alerts, firings] = await Promise.all([
    client.listAlerts(),
    client.listFirings({ limit: FIRINGS, offset: (page - 1) * FIRINGS }),
  ]);
  if (!alerts.ok || !firings.ok) sessionLost(event);
  return { alerts: alerts.data.items, firings: firings.data, page, perPage: FIRINGS };
};

export const actions = {
  delete: async (event) => {
    requireSession(event);
    const form = await event.request.formData();
    const id = z.uuid().safeParse(text(form, 'id'));
    if (!id.success) return fail(400, { delete: { failed: true } });

    const result = await api(event).deleteAlert(id.data);
    // A 404 is an alert already gone, deleted from another tab: the outcome
    // asked for, so it reads as success.
    if (!result.ok && result.status === 401) sessionLost(event);
    if (!result.ok && result.status !== 404)
      return fail(result.status, { delete: { failed: true } });

    setFlash(event.cookies, event.url, {
      kind: 'alert-deleted',
      name: text(form, 'name').slice(0, 80),
    });
    redirect(303, localized(resolve('/alerts')));
  },
} satisfies Actions;
