import type { Config } from 'tailwindcss'

export default {
  content: [
    './components/**/*.{js,vue,ts}',
    './layouts/**/*.vue',
    './pages/**/*.vue',
    './plugins/**/*.{js,ts}',
    './app.vue',
  ],
  theme: {
    extend: {
      colors: {
        // Primary accent colors
        lime: '#B8F36B',
        sky: '#7CC4FF',
        pink: '#F49AC1',
        teal: '#6EE7C8',
        lavender: '#C9B6FF',
        orange: '#FFB86B',
        amber: '#FFD66B',
        coral: '#FF9B8A',
        magenta: '#C2185B',

        // Neutral/Background
        cream: '#faf9f5',
        darkSlate: '#121B17',
        darkGreen: '#1A2520',
        darkGray: '#2A3832',
        lightGray: '#E8EFEA',
        medGray: '#9DB0A6',

        // UI Colors
        primary: {
          50: '#f0fdf4',
          100: '#dcfce7',
          200: '#bbf7d0',
          300: '#86efac',
          400: '#4ade80',
          500: '#B8F36B',
          600: '#22c55e',
          700: '#16a34a',
          800: '#15803d',
          900: '#166534',
        },
        secondary: {
          50: '#f0f9ff',
          100: '#e0f2fe',
          200: '#bae6fd',
          300: '#7dd3fc',
          400: '#38bdf8',
          500: '#7CC4FF',
          600: '#0284c7',
          700: '#0369a1',
          800: '#075985',
          900: '#0c4a6e',
        },
        accent: {
          lime: '#B8F36B',
          sky: '#7CC4FF',
          pink: '#F49AC1',
          teal: '#6EE7C8',
          lavender: '#C9B6FF',
          orange: '#FFB86B',
          amber: '#FFD66B',
          coral: '#FF9B8A',
        },
        dark: {
          bg: '#121B17',
          surface: '#1A2520',
          border: '#2A3832',
          text: '#E8EFEA',
          textSecondary: '#9DB0A6',
        }
      },
      fontFamily: {
        sans: ['DM Sans', 'system-ui', 'sans-serif'],
        display: ['Space Grotesk', 'system-ui', 'sans-serif'],
        mono: ['ui-monospace', 'monospace'],
      },
      borderRadius: {
        xs: '4px',
        sm: '8px',
        md: '12px',
        lg: '16px',
        xl: '20px',
        '2xl': '24px',
      },
      spacing: {
        xs: '4px',
        sm: '8px',
        md: '12px',
        lg: '16px',
        xl: '24px',
        '2xl': '32px',
      },
      boxShadow: {
        xs: '0 1px 2px rgba(0, 0, 0, 0.05)',
        sm: '0 1px 4px rgba(0, 0, 0, 0.12)',
        md: '0 4px 8px rgba(0, 0, 0, 0.15)',
        lg: '0 8px 16px rgba(0, 0, 0, 0.2)',
        xl: '0 12px 24px rgba(0, 0, 0, 0.25)',
      },
    },
  },
  darkMode: 'class',
  plugins: [],
} satisfies Config
