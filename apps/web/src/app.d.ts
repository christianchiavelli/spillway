// The app's own types for SvelteKit: what a request carries and what an error
// says. See https://svelte.dev/docs/kit/types#app.d.ts
import type { Account } from '#lib/api/generated/types.gen.ts';

declare global {
  namespace App {
    interface Error {
      status: number;
      message: string;
      /** Shown on the error page, and logged next to the cause. */
      reference?: string;
    }
    interface Locals {
      /** Made here, never taken from the browser, and sent on to the API. */
      requestId: string;
      /** The signed-in session, or null. The token never reaches the browser. */
      session: { token: string; account: Account; expiresAt: string } | null;
    }
  }
}

export {};
