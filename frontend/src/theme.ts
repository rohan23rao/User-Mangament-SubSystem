import { createTheme } from '@mantine/core';

export const theme = createTheme({
  primaryColor: 'blue',
  fontFamily: 'Inter, -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif',
  headings: {
    fontFamily: 'Greycliff CF, Inter, -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif',
    fontWeight: '700',
    textWrap: 'wrap',
    sizes: {
      h1: { fontSize: '2rem', fontWeight: '700' },
      h2: { fontSize: '1.5rem', fontWeight: '700' },
      h3: { fontSize: '1.25rem', fontWeight: '600' },
      h4: { fontSize: '1.125rem', fontWeight: '600' },
      h5: { fontSize: '1rem', fontWeight: '600' },
      h6: { fontSize: '0.875rem', fontWeight: '600' },
    },
  },
  // Remove the old v6 components structure - v7 uses CSS variables instead
});
