import js from "@eslint/js";
import pluginVue from "eslint-plugin-vue";
import tsEslint from "typescript-eslint";
import prettier from "@vue/eslint-config-prettier";

export default [
  { ignores: ["**/dist/**", "**/node_modules/**", "**/templates/**", "**/env.d.ts", "**/*.d.ts"] },
  js.configs.recommended,
  ...pluginVue.configs["flat/recommended"],
  prettier,
  // TypeScript 规则仅应用于 .ts 文件，避免与 Vue 模板解析冲突
  ...tsEslint.configs.recommended.map((config) => ({
    ...config,
    files: ["**/*.ts"],
  })),
  {
    files: ["**/*.ts", "**/*.vue"],
    rules: {
      "no-console": "warn",
      "@typescript-eslint/no-explicit-any": "warn",
      "@typescript-eslint/no-unused-vars": ["warn", { argsIgnorePattern: "^_" }],
      "@typescript-eslint/no-empty-object-type": "off",
      "vue/multi-word-component-names": "off",
    },
  },
];
