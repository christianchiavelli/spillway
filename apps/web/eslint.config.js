import js from '@eslint/js';
import prettier from 'eslint-config-prettier';
import svelte from 'eslint-plugin-svelte';
import { defineConfig, globalIgnores } from 'eslint/config';
import globals from 'globals';
import ts from 'typescript-eslint';

export default defineConfig(
  globalIgnores(['build/', '.svelte-kit/', 'src/lib/api/generated/']),
  js.configs.recommended,
  ts.configs.strictTypeChecked,
  svelte.configs.recommended,
  prettier,
  svelte.configs.prettier,
  {
    languageOptions: {
      globals: { ...globals.browser, ...globals.node },
      parserOptions: {
        projectService: true,
        extraFileExtensions: ['.svelte'],
        parser: ts.parser,
        tsconfigRootDir: import.meta.dirname,
      },
    },
    rules: {
      // TypeScript already knows every global; this rule only gets them wrong.
      'no-undef': 'off',
      // Template literals with numbers are how copy is written here.
      '@typescript-eslint/restrict-template-expressions': ['error', { allowNumber: true }],
      'svelte/button-has-type': 'error',
      'svelte/require-each-key': 'error',
      'svelte/no-target-blank': 'error',
      // The generated client is reached only through #lib/server/api.ts, which
      // gives every call the request's own client.
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            {
              group: ['**/api/generated/sdk.gen*', '**/api/generated/client*'],
              message: 'Call the API through #lib/server/api.ts.',
            },
          ],
        },
      ],
    },
  },
  {
    files: ['src/lib/server/api.ts'],
    rules: { 'no-restricted-imports': 'off' },
  },
  {
    files: ['**/*.config.{js,ts}'],
    extends: [ts.configs.disableTypeChecked],
  },
);
