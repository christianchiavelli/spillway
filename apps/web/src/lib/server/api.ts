import { type RequestEvent, error } from '@sveltejs/kit';
import * as z from 'zod';

import { API_URL } from '$app/env/private';
import { createClient } from '#lib/api/generated/client/index.ts';
import * as sdk from '#lib/api/generated/sdk.gen.ts';
import type {
  Credentials,
  GetCurtailmentData,
  ListFiringsData,
  ListPlantsData,
  NewAccount,
  NewAlert,
  Problem,
} from '#lib/api/generated/types.gen.ts';
import {
  zAlert,
  zAlertList,
  zCurrentSession,
  zCurtailment,
  zDailyCurtailment,
  zFiringPage,
  zMonthlyCurtailment,
  zPlantRanking,
  zProblem,
  zSessionGrant,
} from '#lib/api/generated/zod.gen.ts';

import { log } from './log.ts';

/** Every call to the API gives up after this long. */
const TIMEOUT_MS = 5_000;

/**
 * What a call ends in when the answer is one the contract describes: the
 * data, or the problem the API reported. Anything else (no answer, a 5xx, a
 * body outside the contract) is logged here with the request id and thrown as
 * an HTTP error, which the error page shows with that id as a reference.
 */
export type ApiResult<T> =
  | { ok: true; data: T }
  | { ok: false; status: number; problem: Problem; retryAfterSeconds: number | undefined };

/** The empty success of an operation answered with 204. */
const zNoContent = z.unknown().transform(() => null);

/**
 * The API as one request sees it. Each request gets a client of its own,
 * carrying its session and its id and nothing else: a module-wide client with
 * a token set on it would hand one person's session to the next request.
 */
export function api(event: RequestEvent, token: string | undefined = event.locals.session?.token) {
  const client = createClient({
    baseUrl: API_URL,
    auth: token,
    headers: { 'X-Request-Id': event.locals.requestId },
    fetch: (input: RequestInfo | URL, init?: RequestInit) =>
      fetch(input, { ...init, signal: AbortSignal.timeout(TIMEOUT_MS) }),
  });

  /**
   * Waits for an operation and reads its answer. Success is checked against
   * the operation's own schema: the generated types also allow a problem
   * there, since the contract describes errors as its `default` response.
   */
  async function call<S extends z.ZodType>(
    operation: string,
    success: S,
    pending: Promise<{ data?: unknown; error?: unknown; response?: Response }>,
  ): Promise<ApiResult<z.output<S>>> {
    const { data, error: failure, response } = await pending;
    const context = { request_id: event.locals.requestId, operation };

    if (response === undefined) {
      log.error('the API did not answer', { ...context, error: failure });
      error(503, 'Spillway cannot reach its data right now. Try again in a moment.');
    }
    if (response.ok) {
      const parsed = success.safeParse(data);
      if (!parsed.success) {
        log.error('the API answered outside its contract', {
          ...context,
          status: response.status,
          issues: parsed.error.issues,
        });
        error(502, 'Spillway got an answer it did not expect. Try again in a moment.');
      }
      return { ok: true, data: parsed.data };
    }

    // A 503 that says the data is still loading is an answer, which the page
    // words; any other failure on the API's side is not.
    const problem = zProblem.safeParse(failure);
    const loading = problem.success && problem.data.code === 'data_unavailable';
    if ((response.status >= 500 && !loading) || !problem.success) {
      log.error('the API failed', { ...context, status: response.status, body: failure });
      error(503, 'Spillway cannot reach its data right now. Try again in a moment.');
    }
    const retryAfter = Number(response.headers.get('Retry-After'));
    return {
      ok: false,
      status: response.status,
      problem: problem.data,
      retryAfterSeconds: Number.isInteger(retryAfter) && retryAfter > 0 ? retryAfter : undefined,
    };
  }

  return {
    createAccount: (body: NewAccount) =>
      call('createAccount', zSessionGrant, sdk.createAccount({ client, body })),
    createSession: (body: Credentials) =>
      call('createSession', zSessionGrant, sdk.createSession({ client, body })),
    currentSession: () =>
      call('getCurrentSession', zCurrentSession, sdk.getCurrentSession({ client })),
    endSession: () =>
      call('deleteCurrentSession', zNoContent, sdk.deleteCurrentSession({ client })),
    listAlerts: () => call('listAlerts', zAlertList, sdk.listAlerts({ client })),
    createAlert: (body: NewAlert) => call('createAlert', zAlert, sdk.createAlert({ client, body })),
    deleteAlert: (alertId: string) =>
      call('deleteAlert', zNoContent, sdk.deleteAlert({ client, path: { alertId } })),
    listFirings: (query: ListFiringsData['query']) =>
      call('listFirings', zFiringPage, sdk.listFirings({ client, query })),
    curtailment: (query: GetCurtailmentData['query']) =>
      call('getCurtailment', zCurtailment, sdk.getCurtailment({ client, query })),
    listPlants: (query: ListPlantsData['query']) =>
      call('listPlants', zPlantRanking, sdk.listPlants({ client, query })),
    listMonths: () =>
      call('listMonthlyCurtailment', zMonthlyCurtailment, sdk.listMonthlyCurtailment({ client })),
    listRecentDays: (days: number) =>
      call(
        'listDailyCurtailment',
        zDailyCurtailment,
        sdk.listDailyCurtailment({ client, query: { days } }),
      ),
  };
}
