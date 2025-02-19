module.exports = {
  root: true,
  extends: ['@blueking/eslint-config-bk/tsvue3'],
  plugins: [
    'simple-import-sort',
  ],
  rules: {
    'simple-import-sort/imports': ['error', {
      groups: [
        ['^[a-zA-Z]'],
        ['^@\\w'],
        ['^\\.\\.'],
        ['^\\.'],
      ],
    }],
    '@typescript-eslint/consistent-type-imports': 'error',
    '@typescript-eslint/consistent-type-exports': 'error',
    'vue/multi-word-component-names': 'off'
  },
  parserOptions: {
    parser: '@typescript-eslint/parser',
    project: [
      './tsconfig.json',
    ],
    extraFileExtensions: ['.vue', '.ts']
  },
};
