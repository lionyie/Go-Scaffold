import js from '@eslint/js'
import prettier from 'eslint-config-prettier'
import pluginVue from 'eslint-plugin-vue'
import tseslint from 'typescript-eslint'

export default [
  {
    ignores: [
      '**/dist/**',
      '**/node_modules/**',
      'services/internal/webui/dist/**',
      'runtime-plugins/**/dist/**',
      '**/*.d.ts',
      'packages/api-client/src/generated/**',
      'packages/shared/src/permissions.gen.ts',
      '**/features.gen.ts',
      '**/.turbo/**',
      'services/api/**',
    ],
  },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  ...pluginVue.configs['flat/recommended'],
  {
    files: ['**/*.vue'],
    languageOptions: {
      parserOptions: {
        parser: tseslint.parser,
      },
    },
  },
  {
    files: ['apps/**/*.{ts,vue}', 'packages/**/*.{ts,vue}', 'extensions/templates/**/*.{ts,vue}', 'runtime-plugins/**/*.{ts,vue}'],
    languageOptions: {
      globals: {
        window: 'readonly',
        fetch: 'readonly',
        localStorage: 'readonly',
        console: 'readonly',
        URL: 'readonly',
        Response: 'readonly',
        setTimeout: 'readonly',
      },
    },
  },
  {
    files: ['scripts/**/*.mjs', '*.config.js', 'eslint.config.js'],
    languageOptions: {
      globals: {
        process: 'readonly',
        console: 'readonly',
        URL: 'readonly',
      },
    },
  },
  {
    files: ['apps/admin/src/**/*.ts'],
    rules: {
      'no-restricted-syntax': [
        'error',
        {
          selector: "Literal[value=/^[a-z]+:[a-z]+$/]",
          message: 'Use the generated Permission constants instead of permission string literals.',
        },
      ],
    },
  },
  {
    rules: {
      '@typescript-eslint/no-explicit-any': 'off',
      '@typescript-eslint/no-unused-vars': ['warn', { argsIgnorePattern: '^_' }],
      'no-empty': ['error', { allowEmptyCatch: true }],
      'vue/multi-word-component-names': 'off',
      'vue/require-default-prop': 'off',
      'vue/attributes-order': 'off',
      'vue/valid-v-slot': ['error', { allowModifiers: true }],
    },
  },
  prettier,
]
