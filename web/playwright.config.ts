import { defineConfig, devices } from '@playwright/test';

// Type declaration for Node.js process (needed for Playwright config files)
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
  /* Set workers in CI to 1 for better test reliability */
  workers: IS_CI_ENV ? 1 : '75%',
  /* Fail fast in CI to save resources */
  ...(IS_CI_ENV && { maxFailures: 10 }),
  /* Global timeout for each test */
  timeout: 45 * 1000, // 30 seconds
  /* Output directory for test artifacts */
  outputDir: './test-results',
  /* Reporter configuration - use list in CI for better logs, html for local */
  reporter: IS_CI_ENV
    ? [['list'], ['html', { outputFolder: 'playwright-report' }]]
    : [['html', { outputFolder: 'playwright-report' }]],
  /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
  use: {
    /* Base URL to use in actions like `await page.goto('/')`. */
    baseURL: 'http://localhost:3000',
    /* Run in headless mode in CI, headed locally for debugging */
    headless: IS_CI_ENV,
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
    /* Timeout for each expectation */
    timeout: 5 * 1000, // 5 seconds
  },

  /* Run your local dev server before starting the tests */
  webServer: {
    command: process.env.NO_COMMAND === 'true' ? '' : 'pnpm start',
    url: 'http://localhost:3000',
    reuseExistingServer: !IS_CI_ENV,
    timeout: 120 * 1000,
    stdout: 'pipe',
    stderr: 'pipe',
  },
});
