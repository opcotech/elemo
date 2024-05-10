import type { PlaywrightTestConfig } from '@playwright/test';

const commonProjectConfig = {
  outputDir: 'tests/results',
  snaphsotDir: 'tests/results/snapshots'
};

const serverHost = process.env.PLAYWRIGHT_SERVER_HOST || '127.0.0.1';
const serverPort = parseInt(process.env.PLAYWRIGHT_SERVER_PORT || '3000');
const serverCommand = `pnpm build && pnpm start --hostname ${serverHost} --port ${serverPort}`;

const config: PlaywrightTestConfig = {
  webServer: {
    command: serverCommand,
    port: serverPort,
    reuseExistingServer: !process.env.CI
  },
  testDir: 'tests',
  use: {
    baseURL: `http://${serverHost}:${serverPort}`,
    headless: process.env.PLAYWRIGHT_HEADLESS !== 'false',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    trace: 'retain-on-failure',
    ignoreHTTPSErrors: true
  },
  projects: [
    {
      name: 'Desktop',
      ...commonProjectConfig,
      use: {
        viewport: { width: 1366, height: 768 },
        deviceScaleFactor: 1
      }
    },
    {
      name: 'Desktop FHD',
      ...commonProjectConfig,
      use: {
        viewport: { width: 1920, height: 1080 },
        deviceScaleFactor: 1
      }
    }
    /*{
      name: 'Mobile',
      ...commonProjectConfig,
      use: {
        viewport: { width: 414, height: 896 },
        deviceScaleFactor: 2
      }
    }*/
  ]
};

export default config;
