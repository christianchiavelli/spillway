import js from '@eslint/js';
import prettier from 'eslint-config-prettier';
import { defineConfig, globalIgnores } from 'eslint/config';
import ts from 'typescript-eslint';

// The type-checked rules are here for one fault above all: an assertion left
// without its await passes whatever the page shows, and the rules call that
// a floating promise.
export default defineConfig(
  globalIgnores(['playwright-report/', 'test-results/', 'blob-report/']),
  js.configs.recommended,
  ts.configs.strictTypeChecked,
  prettier,
  {
    languageOptions: {
      parserOptions: { projectService: true, tsconfigRootDir: import.meta.dirname },
    },
  },
  {
    files: ['**/*.js'],
    extends: [ts.configs.disableTypeChecked],
  },
);
