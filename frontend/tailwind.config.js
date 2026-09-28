/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ['class'],
  content: [
    './index.html',
    './src/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      colors: {
        background: '#09090b',
        surface: {
          DEFAULT: '#121215',
          hover: '#18181c',
          border: '#27272a',
          card: '#141418',
        },
        sidebar: {
          DEFAULT: '#0d0d0f',
          hover: '#151518',
        },
        brand: {
          yellow: '#EAB308',
          gold: '#F59E0B',
          emerald: '#10B981',
          cyan: '#06B6D4',
        }
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
      },
      boxShadow: {
        'glow-yellow': '0 0 20px rgba(234, 179, 8, 0.15)',
        'glow-card': '0 8px 32px rgba(0, 0, 0, 0.4)',
      }
    },
  },
  plugins: [],
};
