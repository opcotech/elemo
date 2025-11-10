import { defineConfig, devices } from '@playwright/test';

// Fix for "Cannot find name 'process'"
declare const process: {
  env: {
    [key: string]: string | undefined;
  };
};

const IS_CI_ENV = process.env.CI === 'true';

/**
 * @see https://playwright.dev/docs/test-configuration
 */
export default defineConfig({
  testDir: './tests/e2e',
  /* Global setup runs once before all tests */
  globalSetup: './tests/e2e/global-setup.ts',
  /* Run tests in files in parallel */
  fullyParallel: true,
  /* Fail the build on CI if test.only left in the source code. */
  forbidOnly: IS_CI_ENV,
  /* Retry more times on CI */
  retries: IS_CI_ENV ? 3 : 1,
  /* Opt out of parallel tests on CI. */
  workers: IS_CI_ENV ? 1 : '75%',
  /* Reporter to use. See https://playwright.dev/docs/test-reporters */
  reporter: 'html',
  /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
  use: {
    /* Base URL to use in actions like `await page.goto('/')`. */
    baseURL: 'http://localhost:3000',

    /* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
    trace: 'on-first-retry',

    /* Take screenshot on failure */
    screenshot: 'only-on-failure',

    /* Record video on failure */
    video: 'retain-on-failure',
  },

  /* Configure projects for major browsers */
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },

    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },

    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },

    /* Test against mobile viewports. */
    /*FIXME: enable "Mobile Chrome" once we support mobile viewports
    {
      name: 'Mobile Chrome',
      use: { ...devices['Pixel 5'] },
    },*/

    /*FIXME: enable "Mobile Safari" once we support mobile viewports
    {
      name: 'Mobile Safari',
      use: { ...devices['iPhone 12'] },
    },*/
  ],

  /* Expectation configuration */
  expect: {
    /* Timeout for each expectation, intentionally low to speed up tests */
    timeout: 3000,
  },

  /* Run your local dev server before starting the tests */
  webServer: {
    command: process.env.NO_COMMAND === 'true' ? '' : 'pnpm build && pnpm start',
    url: 'http://localhost:3000',
    reuseExistingServer: true,
    timeout: 120 * 1000,
    stdout: 'pipe',
    stderr: 'pipe',
  },
});
