import type { StorybookConfig } from '@storybook/nextjs';
import * as path from 'path';

const config: StorybookConfig = {
  stories: ['../components/**/*.stories.@(js|jsx|ts|tsx)'],
  staticDirs: ['../public'],
  addons: [
    '@storybook/addon-a11y',
    '@storybook/addon-essentials',
    '@storybook/addon-interactions',
    '@storybook/addon-links',
    '@storybook/client-api',
    '@storybook/addon-styling'
  ],
  framework: {
    name: '@storybook/nextjs',
    options: {}
  },
  docs: {
    autodocs: 'tag'
  },
  core: {
    disableTelemetry: true
  },
  webpackFinal: async (config) => {
    if (config.resolve) {
      config.resolve = {
        ...config.resolve,
        alias: {
          ...(config.resolve.alias ?? {}),
          '@/components': path.resolve(__dirname, '../components'),
          '@/lib/auth': path.resolve(__dirname, '../lib/auth.ts'),
          '@/lib/api': path.resolve(__dirname, '../lib/api/index.ts'),
          '@/lib/helpers': path.resolve(__dirname, '../lib/helpers/index.ts'),
          '@/lib/hooks/useTimeout': path.resolve(__dirname, '../lib/hooks/useTimeout.ts'),
          '@/store': path.resolve(__dirname, '../store')
        }
      };
    }

    return config;
  }
};

export default config;
