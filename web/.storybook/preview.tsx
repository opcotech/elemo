import React from 'react';
import type { Preview, ReactRenderer } from '@storybook/react';
import { withThemeByClassName } from '@storybook/addon-themes';
import { Title, Subtitle, Description, Primary, Controls, Stories } from '@storybook/blocks';

import { ThemeProvider } from '../components/providers/theme-provider';
import { Toaster } from '../components/ui/toaster';

import '../app/styles/globals.css';

const preview: Preview = {
  decorators: [
    withThemeByClassName<ReactRenderer>({
      themes: {
        light: 'ight',
        dark: 'dark',
        system: 'system',
      },
      defaultTheme: 'system',
    }),
    (Story, context) => {
      return (
        <ThemeProvider attribute='class' forcedTheme={context.globals.theme}>
          <Story />
          <Toaster />
        </ThemeProvider>
      );
    },
  ],
  parameters: {
    nextjs: {
      appDirectory: true,
    },
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
    backgrounds: {
      disable: true,
    },
    docs: {
      toc: true,
      page: () => (
        <>
          <Title />
          <Subtitle />
          <Description />
          <Primary />
          <Controls />
          <Stories />
        </>
      ),
    },
  },
};

export default preview;
