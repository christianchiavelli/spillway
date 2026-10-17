# syntax=docker/dockerfile:1
# The web image: SvelteKit's Node server. adapter-node bundles every package
# the app imports into its output, so the image holds that output and Node.
# Build from the repository root: docker build -f apps/web/Dockerfile .

ARG NODE_IMAGE=node:24.22.0-alpine3.24@sha256:44fce696f38ba8c987c4db89954051000bbd05ffd8d8659ba15ed4a8602e937f

FROM ${NODE_IMAGE} AS pnpm
ENV PNPM_HOME=/pnpm \
    COREPACK_ENABLE_DOWNLOAD_PROMPT=0
WORKDIR /repo
COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
# pnpm at the exact version and checksum `packageManager` names.
RUN corepack enable pnpm && corepack install

FROM pnpm AS build
# Downloads depend on the lockfile alone, so this layer outlives source edits.
RUN --mount=type=cache,id=pnpm-store,target=/pnpm/store \
    pnpm fetch --store-dir /pnpm/store
COPY apps/web apps/web
RUN --mount=type=cache,id=pnpm-store,target=/pnpm/store \
    pnpm install --offline --frozen-lockfile --store-dir /pnpm/store --filter @spillway/web
# The address the app is opened at, which the CSRF check compares each form's
# Origin with. SvelteKit reads it at build time (ADR 0015).
ARG ORIGIN=http://localhost:8090
RUN ORIGIN=${ORIGIN} pnpm --filter @spillway/web build

FROM ${NODE_IMAGE} AS runtime
ENV NODE_ENV=production \
    HOST=0.0.0.0 \
    PORT=3000
WORKDIR /app
# Owned by root and only read by the server, which runs as `node`.
COPY --from=build /repo/apps/web/package.json ./
COPY --from=build /repo/apps/web/build ./build
USER node
EXPOSE 3000
# No curl in the image, so Node asks. The sign-in page renders without the API,
# so a slow API never gets this container restarted.
HEALTHCHECK --interval=10s --timeout=5s --start-period=20s --start-interval=2s --retries=3 \
    CMD ["node", "-e", "fetch(`http://127.0.0.1:${process.env.PORT}/sign-in`).then((r) => process.exit(r.ok ? 0 : 1), () => process.exit(1))"]
CMD ["node", "build"]
