module.exports = {
  '*.go': () => 'golangci-lint run ./...',
  'web/**/*.{ts,tsx}': [
    'prettier --write',
  ],
  'web/**/*.{json,css,md}': 'prettier --write',
};
