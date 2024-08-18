import { addons } from '@storybook/manager-api';
import { create } from '@storybook/theming';

addons.setConfig({
  theme: create({
    base: 'light',

    brandTitle: 'Elemo design system',
    brandUrl: 'https://github.com/opcotech/elemo',
    brandImage: '',
    brandTarget: '_self',

    colorPrimary: '#0f73e6',
    colorSecondary: '#0f73e6',
    textColor: '#171c27',
    barTextColor: '#61616b',
  }),
});
