import adapter from '@sveltejs/adapter-node';
import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import icons from 'unplugin-icons/vite';
import { playwright } from '@vitest/browser-playwright';
import { defineConfig } from 'vitest/config';

export default defineConfig(({ command }) => ({
  plugins: [
    tailwindcss(),
    sveltekit({
      adapter: adapter(),
      compilerOptions: {
        // Runes everywhere in the app; libraries keep their own mode.
        runes: ({ filename }) =>
          filename.split(/[/\\]/).includes('node_modules') ? undefined : true,
      },
      // Form posts from any other origin are refused. That is the default,
      // written down so nobody widens it without reading this line.
      csrf: { trustedOrigins: [] },
      // The address the built app is opened at, which the CSRF check compares
      // each form's Origin with. Without it, adapter-node 6 takes every request
      // for https unless a proxy says otherwise, and refuses the plain-HTTP
      // forms of http://localhost. The dev server keeps its own address.
      paths: {
        origin:
          command === 'build' ? (process.env['ORIGIN'] ?? 'http://localhost:8090') : undefined,
      },
      // Nonces for pages rendered per request, hashes for prerendered ones, and
      // no 'unsafe-inline' anywhere: the app writes no inline styles or scripts.
      csp: {
        mode: 'auto',
        directives: {
          'default-src': ['self'],
          'script-src': ['self'],
          // One style attribute is allowed, by its hash: the one SvelteKit's
          // route announcer hides itself with. If an upgrade changes it, the
          // browser reports the new hash and the e2e suite fails on it.
          'style-src': [
            'self',
            'unsafe-hashes',
            'sha256-S8qMpvofolR8Mpjy4kQvEm7m1q8clzU4dfDH0AmvZjo=',
          ],
          'img-src': ['self', 'data:'],
          'font-src': ['self'],
          'connect-src': ['self'],
          'form-action': ['self'],
          'base-uri': ['self'],
          'frame-ancestors': ['none'],
          'object-src': ['none'],
        },
      },
    }),
    // Material Symbols, each one compiled into a Svelte component at build
    // time: inline SVG, no runtime and no style attribute. They sit next to a
    // label wherever they appear, so a screen reader skips them.
    icons({
      compiler: 'svelte',
      scale: 1,
      iconCustomizer: (_collection, _icon, props) => {
        props['aria-hidden'] = 'true';
      },
    }),
  ],
  // Bundled before the browser tests start: found halfway through a run, a new
  // dependency reloads the page under the tests, and every one of them fails.
  optimizeDeps: { include: ['d3-scale', 'd3-shape'] },
  test: {
    expect: { requireAssertions: true },
    projects: [
      {
        extends: './vite.config.ts',
        test: {
          name: 'components',
          browser: {
            enabled: true,
            provider: playwright(),
            instances: [{ browser: 'chromium', headless: true }],
          },
          include: ['src/**/*.svelte.{test,spec}.ts'],
          exclude: ['src/lib/server/**'],
        },
      },
      {
        extends: './vite.config.ts',
        test: {
          name: 'server',
          environment: 'node',
          include: ['src/**/*.{test,spec}.ts'],
          exclude: ['src/**/*.svelte.{test,spec}.ts'],
        },
      },
    ],
  },
}));
