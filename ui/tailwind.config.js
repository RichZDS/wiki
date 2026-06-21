/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        ink: '#172033',
        paper: '#f7f8fc',
        mint: '#59d6b3',
        coral: '#ff8f70',
        violet: '#8e7cff',
      },
      boxShadow: {
        soft: '0 18px 60px rgba(28, 37, 64, 0.10)',
      },
    },
  },
  plugins: [],
}
