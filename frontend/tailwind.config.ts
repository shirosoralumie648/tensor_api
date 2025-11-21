import type { Config } from 'tailwindcss';
import { tailwindConfig } from './src/theme/tokens';

const config: Config = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  darkMode: 'class',
  theme: {
    extend: tailwindConfig.extend,
  },
  plugins: [],
};
export default config;

