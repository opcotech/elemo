import { expect, test } from '@playwright/test';
import { User } from 'elemo-client';
import * as auth from '../helpers/auth';
import * as commonSteps from '../helpers/common.steps';
import * as authSteps from './auth.steps';

test.describe('Authentication @auth', () => {
  let user: User;

  test.beforeAll(async () => {
    user = await auth.createDBUser();
  });

  test('unauthenticated user is redirected to login', async ({ page }) => {
    await page.goto(commonSteps.URLS.home);
    expect(page.url().match(/.*\/auth\/signin\?callbackUrl=.*/)).toBeTruthy();
  });

  test('unauthenticated user can login', async ({ page }) => {
    await authSteps.authenticate(page, user.email, auth.USER_DEFAULT_PASSWORD);
    await expect(page.url().match(/.*\/$/)).toBeTruthy();
  });

  test('authenticated user can logout', async ({ page }) => {
    await authSteps.authenticate(page, user.email, auth.USER_DEFAULT_PASSWORD);
    await authSteps.logout(page);
    expect(page.url().match(/.*\/auth\/signin\?callbackUrl=.*/)).toBeTruthy();
  });
});
