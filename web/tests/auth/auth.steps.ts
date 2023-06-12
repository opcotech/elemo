import { Page } from '@playwright/test';
import * as commonSteps from '../helpers/common.steps';

export const SELECTORS = {
  loginFormEmailField: 'input[name=username]',
  loginFormPasswordField: 'input[name=password]',
  loginFormSubmitButton: 'button[type=submit]',
  logoutFormSubmitButton: 'button[type=submit]'
};

export async function login(page: Page) {
  await page.waitForSelector(SELECTORS.loginFormSubmitButton);
  await page.click(SELECTORS.loginFormSubmitButton);
}

export async function logout(page: Page) {
  await commonSteps.clickMenuItem(
    page,
    commonSteps.SELECTORS.navbar.userMenuOpenButton,
    commonSteps.SELECTORS.navbar.userMenu.logoutButton
  );

  await page.waitForSelector(SELECTORS.logoutFormSubmitButton);
  await page.click(SELECTORS.logoutFormSubmitButton);
}

export async function fillLoginForm(page: Page, email: string, password: string) {
  await page.fill(SELECTORS.loginFormEmailField, email);
  await page.fill(SELECTORS.loginFormPasswordField, password);
}

export async function authenticate(page: Page, email: string, password: string) {
  await page.goto(commonSteps.URLS.login);
  await fillLoginForm(page, email, password);
  await login(page);
}
