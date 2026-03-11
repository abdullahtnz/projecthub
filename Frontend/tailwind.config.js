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
          50: '#e4e4f0',   // Lightest lavender
          100: '#bdbdd9',
          200: '#9595c2',
          300: '#6e6eab',
          400: '#474794',
          500: '#2a2a5c',   // Deep purple-blue (main)
          600: '#222248',
          700: '#1a1a36',
          800: '#111124',
          900: '#090912',   // Almost black
        },
        secondary: {
          50: '#f2f2f2',
          100: '#e6e6e6',
          200: '#cccccc',
          300: '#b3b3b3',
          400: '#999999',
          500: '#808080',   // Medium gray
          600: '#666666',
          700: '#4d4d4d',
          800: '#333333',
          900: '#1a1a1a',   // Dark gray
        },
        accent: {
          500: '#ff6b6b',   // Coral red for highlights
          600: '#e05555',
        }
      },
    },
  },
  plugins: [],
}