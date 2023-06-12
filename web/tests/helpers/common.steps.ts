import { Page } from '@playwright/test';

export const URLS = {
  home: '/',
  homeRegex: /\/$/,
  login: '/api/auth/signin',
  profile: (userKey?: string) => `/profile/${userKey || ''}`,
  profileRegex: /\/profile(\/[a-zA-Z0-9]+)?$/,
  settings: '/settings',
  settingsRegex: /\/settings$/
};

export const SELECTORS = {
  navbar: {
    userMenuOpenButton: '#navbar #btn-avatar',
    userMenu: {
      profileButton: '#navbar #navbar-user-dropdown #menu-item-profile',
      settingsButton: '#navbar #navbar-user-dropdown #menu-item-settings',
      logoutButton: '#navbar #navbar-user-dropdown #menu-item-logout'
    }
  }
};

export async function visitPage(page: Page, path: string, pathRegex: RegExp, operation?: string, status?: number) {
  const conditions: Promise<any>[] = [page.waitForURL(pathRegex), page.waitForLoadState('domcontentloaded')];

  if (operation && status) {
    conditions.push(
      page.waitForResponse(
        (response) =>
          response.url().endsWith('graphql') &&
          response.status() === status &&
          response.request().method() === 'POST' &&
          (response.request().postData()?.includes(operation) || false)
      )
    );
  } else {
    conditions.push(page.waitForLoadState('networkidle'));
  }

  await page.goto(path);
  await Promise.all(conditions);
}

export async function switchTab(page: Page, selector: string) {
  await Promise.all([
    page.waitForLoadState('domcontentloaded'),
    page.waitForLoadState('networkidle'),
    page.click(selector)
  ]);
}

export async function switchTabAndWaitForResponse(page: Page, selector: string, operation: string, status: number) {
  await Promise.all([
    switchTab(page, selector),
    page.waitForResponse(
      async (response) =>
        response.url().endsWith('graphql') &&
        response.status() === status &&
        response.request().method() === 'POST' &&
        (response.request().postData()?.includes(operation) || false) &&
        !(await response.body()).includes('errors')
    )
  ]);
}

export async function clickMenuItem(page: Page, menu: string, item: string) {
  await Promise.all([
    page.waitForNavigation(),
    page.waitForLoadState('domcontentloaded'),
    page.waitForLoadState('networkidle'),
    page.click(menu),
    page.click(item)
  ]);
}

export async function selectComboOption(page: Page, comboSelector: string, optionSelectors: string[]) {
  await page.click(comboSelector);
  for (const optionSelector of optionSelectors) {
    await page.click(optionSelector);
  }
  await page.click(comboSelector);
}

export async function clickAndWaitForResponse(
  page: Page,
  selector: string,
  operation: string,
  status: number,
  error: boolean
) {
  await Promise.all([
    page.waitForResponse(
      async (response) =>
        response.url().endsWith('graphql') &&
        response.status() === status &&
        response.request().method() === 'POST' &&
        (response.request().postData()?.includes(operation) || false) &&
        (await response.body()).includes('errors') === error
    ),
    page.click(selector)
  ]);
}
