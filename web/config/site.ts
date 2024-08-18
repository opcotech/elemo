export type SiteConfig = typeof siteConfig;

export const links = {};

export const emailAddresses = {
  support: 'support@elemo.app',
};

export const siteConfig = {
  name: 'Elemo',
  description: 'The next-generation project management tool',
  frontendUrl: process.env.NEXT_PUBLIC_SITE_URL ?? 'http://127.0.0.1:3000',
};
