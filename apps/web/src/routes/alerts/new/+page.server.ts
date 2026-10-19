import { fail, redirect } from '@sveltejs/kit';

import { zNewAlert } from '#lib/api/generated/zod.gen.ts';
import { type AlertField, alertFields } from '#lib/features/alerts/alerts.ts';
import { type FieldErrors, fromIssues, fromProblem } from '#lib/forms/errors.ts';
import { api } from '#lib/server/api.ts';
import { setFlash } from '#lib/server/flash.ts';
import { text, wholeNumber } from '#lib/server/forms.ts';
import { published } from '#lib/server/published.ts';
import { requireSession, sessionLost } from '#lib/server/session.ts';

import type { Actions, PageServerLoad } from './$types';

/** The days the preview shows: the ones an alert's activity counts. */
const PREVIEW_DAYS = 30;

export const load: PageServerLoad = async (event) => {
  requireSession(event);
  const client = api(event);
  const [alerts, recent] = await Promise.all([
    client.listAlerts(),
    client.listRecentDays(PREVIEW_DAYS),
  ]);
  if (!alerts.ok) sessionLost(event);
  return { held: alerts.data.items.length, recent: published(recent) };
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
} satisfies Actions;
