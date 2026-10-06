import js from "@eslint/js";
import tseslint from "typescript-eslint";

export default tseslint.config(
  { ignores: ["node_modules/"] },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ["**/*.tsx"],
    rules: {
      // `h` is the JSX factory; TypeScript-ESLint cannot see JSX use it.
      "@typescript-eslint/no-unused-vars": ["error", { varsIgnorePattern: "^h$" }],
    },
  },
);
