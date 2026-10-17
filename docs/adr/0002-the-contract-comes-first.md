# 2. The contract comes first: OpenAPI 3.1, ogen and Hey API

- Status: accepted
- Date: 2026-10-17

## Context

Two languages share every request and every answer. Types written twice drift apart, and the drift shows up at run time, as a 500 or as a form that says nothing.

## Decision

- **`contracts/openapi.yaml` is the only place a request, an answer or a limit is written.** Both sides are generated from it, and the generated code is committed.
- **Go: ogen 1.24** writes the server: routing, decoding, validation of every field against the contract, encoding, and an OpenTelemetry span and metrics per operation. Handlers implement a generated interface, so an operation added to the contract does not compile until it is handled.
- **TypeScript: Hey API 0.99** writes the types, a fetch client and Zod schemas. The web's forms parse with the generated request schemas, and `src/lib/server/api.ts` parses every answer with its success schema before the app reads it.
- **Limits live in the contract** (`maxLength`, `minimum`, `maxItems`). The web reads its hints from the generated schemas, and a Go test checks that each limit equals the domain's constant.
- **The error vocabulary lives there too.** `ProblemCode` is a closed enum, so the web can switch on it and the compiler says when a case is missing.
- **CI fails** when either side's generated code differs from what the contract generates today, and when the contract breaks a client of the previous commit's (oasdiff).

## Consequences

- A field renamed in the contract fails to compile on both sides.
- The generated code is large. It is read in review only when the contract changed.
- ogen reads YAML 1.1, where a bare `N` is false: the subsystem `N` is quoted, with a comment saying why.
- ogen covers most of OpenAPI 3.1, not all of it. The contract keeps to what it covers, and the generator warns when it does not.

## Alternatives considered

- **Code first, with the contract generated from Go types (huma, swaggo):** the contract becomes an output, reviewed after the fact, and the TypeScript side still has to be generated from it.
- **oapi-codegen:** mature, and it leaves validation to a separate middleware. ogen validates in the decoder it writes, without reflection.
- **gRPC or Connect:** typed both ways, and the BFF would still turn it into JSON. HTTP with OpenAPI keeps the API one `curl` away.
- **openapi-typescript with openapi-fetch:** types and a client, and no runtime schemas to parse forms and answers with.
