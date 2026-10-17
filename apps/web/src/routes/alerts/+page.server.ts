import { fail, redirect } from '@sveltejs/kit';
import * as z from 'zod';

import { zNewAlert } from '#lib/api/generated/zod.gen.ts';
import { type AlertField, alertFields } from '#lib/features/alerts/alerts.ts';
import { type FieldErrors, fromIssues, fromProblem } from '#lib/forms/errors.ts';
import { api } from '#lib/server/api.ts';
import { setFlash } from '#lib/server/flash.ts';
import { text, wholeNumber } from '#lib/server/forms.ts';
import { requireSession, sessionLost } from '#lib/server/session.ts';

import type { Actions, PageServerLoad } from './$types';

// +page.server.ts, not +page.ts: the list needs the session token, which must
// never reach the browser, and the API is not reachable from it anyway.
export const load: PageServerLoad = async (event) => {
  requireSession(event);
  const result = await api(event).listAlerts();
  if (!result.ok) sessionLost(event);
  return { alerts: result.data.items };
};

export const actions = {
  create: async (event) => {
    requireSession(event);
    const form = await event.request.formData();
    const input = {
      name: text(form, 'name'),
      source: text(form, 'source'),
      subsystem: text(form, 'subsystem'),
      thresholdMwh: wholeNumber(form, 'thresholdMwh'),
    };
    // Sent back as typed, so a failed submit loses nothing, "1,500" included.
    const values = { ...input, thresholdMwh: text(form, 'thresholdMwh') };
    const refuse = (status: number, errors: FieldErrors<AlertField>, full = false) =>
      fail(status, { create: { values, errors, full } });

    const parsed = zNewAlert.safeParse(input);
    if (!parsed.success) return refuse(422, fromIssues(parsed.error.issues, alertFields, input));

    const result = await api(event).createAlert(parsed.data);
    if (!result.ok) {
      if (result.status === 401) sessionLost(event);
      if (result.problem.code === 'alert_limit_reached') return refuse(409, {}, true);
      return refuse(result.status, fromProblem(result.problem, alertFields));
    }

    setFlash(event.cookies, event.url, { kind: 'alert-created', name: result.data.name });
    redirect(303, '/alerts');
  },

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
    redirect(303, '/alerts');
  },
} satisfies Actions;
