/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: { DEFAULT: '#2563eb', dark: '#1d4ed8' },
        dark: { DEFAULT: '#0f172a', card: '#1e293b', border: '#334155' },
      },
    },
  },
  plugins: [],
};
