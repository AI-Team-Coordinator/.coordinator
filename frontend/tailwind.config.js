/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ['class'],
  content: [
    './index.html',
    './src/**/*.{ts,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          50: '#eef2ff',
          100: '#e0e7ff',
          200: '#c7d2fe',
          300: '#a5b4fc',
          400: '#818cf8',
          500: '#6366f1',
          600: '#4f46e5',
          700: '#4338ca',
          800: '#3730a3',
          900: '#312e81',
        },
      },
      animation: {
        'pulse-slow': 'pulse-slow 2.5s infinite ease-in-out',
        'conflict-title-shimmer': 'conflict-title-shimmer 2.8s ease-in-out infinite',
        'conflict-border-pulse': 'conflict-border-pulse 2.5s ease-in-out infinite',
      },
      keyframes: {
        'pulse-slow': {
          '0%, 100%': { opacity: '1', transform: 'scale(1)' },
          '50%': { opacity: '0.6', transform: 'scale(0.96)' },
        },
        'conflict-title-shimmer': {
          '0%, 100%': {
            backgroundPosition: '0% 50%',
          },
          '50%': {
            backgroundPosition: '100% 50%',
          },
        },
        'conflict-border-pulse': {
          '0%, 100%': {
            borderColor: 'rgb(254 202 202 / 0.55)',
            boxShadow: '0 0 0 0 rgb(244 63 94 / 0)',
          },
          '50%': {
            borderColor: 'rgb(244 63 94 / 0.75)',
            boxShadow: '0 0 12px -2px rgb(244 63 94 / 0.35)',
          },
        },
      },
    },
  },
  plugins: [],
}
