// Minimal ESLint config - zero dependencies
export default [
  {
    files: ["**/*.js"],
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: "module",
      globals: {
        window: "readonly",
        document: "readonly",
        console: "readonly",
        fetch: "readonly",
        setInterval: "readonly",
        clearInterval: "readonly",
        App: "readonly",
        Utils: "readonly",
        ConnectionManager: "readonly",
        StatusManager: "readonly"
      }
    },
    rules: {
      "no-unused-vars": "warn",
      "no-undef": "error",
      "no-console": "off",
      "no-redeclare": "error",
      "no-unreachable": "error"
    }
  }
];
