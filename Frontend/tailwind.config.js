/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#e6f0ff',
          100: '#b3d1ff',
          200: '#80b3ff',
          300: '#4d94ff',
          400: '#1a75ff',
          500: '#0052cc',  // Main dark blue
          600: '#0042a3',
          700: '#00317a',
          800: '#002152',
          900: '#001029',
        }
      }
    },
  },
  plugins: [],
}