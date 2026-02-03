/** @type {import('prettier').Config} */
const prettierConfig = {
  semi: true,
  singleQuote: false,
  trailingComma: "none",
  printWidth: 100,
  tabWidth: 2,
  plugins: ["prettier-plugin-tailwindcss"]
};

const config = {
  ...prettierConfig,
};

export default config;
