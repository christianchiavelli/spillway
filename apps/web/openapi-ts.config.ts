import { defineConfig } from '@hey-api/openapi-ts';

// The web app's half of the contract: types, a fetch client, Zod schemas for
// the request bodies and the constraints as constants, all generated from
// contracts/openapi.yaml, which also generates the Go server.
export default defineConfig({
  input: '../../contracts/openapi.yaml',
  output: { path: 'src/lib/api/generated' },
  plugins: [
    // No base URL: the module's default client cannot reach the API, so a
    // call that forgets the per-request client fails instead of borrowing
    // someone else's session.
    { name: '@hey-api/client-fetch', baseUrl: false },
    '@hey-api/typescript',
    '@hey-api/schemas',
    { name: 'zod', requests: true },
    // Plain functions, not the class: the class keeps every instance in a
    // static registry, and one per request would leave sessions in memory.
    // No validator here: the contract's `default` problem response makes the
    // generated one accept a problem as success, so #lib/server/api.ts checks
    // each answer against its exact success schema instead.
    '@hey-api/sdk',
  ],
});
