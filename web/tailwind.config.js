const defaultTheme = require('tailwindcss/defaultTheme');

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./app/**/*.{js,ts,jsx,tsx}', './pages/**/*.{js,ts,jsx,tsx}', './components/**/*.{js,ts,jsx,tsx}'],
  plugins: [require('@tailwindcss/forms'), require('@tailwindcss/typography')],
  theme: {
    fontFamily: {
      heading: ['var(--font-lato)', ...defaultTheme.fontFamily.sans],
      sans: ['var(--font-work-sans)', ...defaultTheme.fontFamily.sans]
    },
    colors: {
      inherit: 'inherit',
      current: 'currentColor',
      transparent: 'transparent',
      black: '#222222',
      white: '#ffffff',
      blue: {
        50: '#D9ECFC',
        100: '#B3D8F9',
        200: '#7BBCF5',
        300: '#50A6F1',
        400: '#2690EE',
        500: '#117CDA',
        600: '#0D5EA6',
        700: '#094172',
        800: '#073055',
        900: '#052642'
      },
      green: {
        50: '#E5F6EF',
        100: '#CEEFE1',
        200: '#AAE1CB',
        300: '#8BD6B8',
        400: '#6DCCA6',
        500: '#4FC193',
        600: '#38A076',
        700: '#2A7657',
        800: '#1B4D39',
        900: '#133527'
      },
      red: {
        50: '#FDF1F2',
        100: '#FCE4E6',
        200: '#F9CDD0',
        300: '#F4A4AA',
        400: '#EF818A',
        500: '#EB5D68',
        600: '#E63946',
        700: '#CD1A27',
        800: '#9B141E',
        900: '#690D14'
      },
      orange: {
        50: '#FEF1EC',
        100: '#FDE7DD',
        200: '#FCDACA',
        300: '#FBC5AC',
        400: '#F9A985',
        500: '#F78E5E',
        600: '#F57337',
        700: '#C2440A',
        800: '#7D2C06',
        900: '#481904'
      },
      yellow: {
        50: '#FEF7EC',
        100: '#FDF1DD',
        200: '#FCEBCE',
        300: '#FADBA7',
        400: '#F8CC80',
        500: '#F8BF3D',
        600: '#F4AC33',
        700: '#E3930C',
        800: '#AD7009',
        900: '#784E06'
      },
      indigo: {
        50: '#EEF2FF',
        100: '#E0E7FF',
        200: '#C7D2FE',
        300: '#A5B4FC',
        400: '#818CF8',
        500: '#6366F1',
        600: '#4F46E5',
        700: '#4338CA',
        800: '#3730A3',
        900: '#312E81'
      },
      gray: {
        50: '#f9fafb',
        100: '#f3f4f6',
        200: '#e5e7eb',
        300: '#d1d5db',
        400: '#9ca3af',
        500: '#6b7280',
        600: '#4b5563',
        700: '#374151',
        800: '#1f2937',
        900: '#111827'
      }
    }
  }
};
