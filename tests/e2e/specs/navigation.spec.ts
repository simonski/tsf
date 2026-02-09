import { test, expect } from '@playwright/test';

test.describe('Navigation', () => {
  test.beforeEach(async ({ page }) => {
    // Login first
    await page.goto('/');
    await page.locator('#login-username').fill('admin');
    await page.locator('#login-password').fill('admin');
    await page.locator('#login-button').click();
    await expect(page.locator('#board-screen')).toBeVisible({ timeout: 10000 });
  });

  test('should open and close navigation menu', async ({ page }) => {
    // Menu should be hidden initially
    await expect(page.locator('#side-nav')).not.toHaveClass(/open/);

    // Open menu
    await page.locator('#menu-toggle').click();
    await expect(page.locator('#side-nav')).toHaveClass(/open/);

    // Close menu
    await page.locator('#nav-close').click();
    await expect(page.locator('#side-nav')).not.toHaveClass(/open/);
  });

  test('should navigate to overview', async ({ page }) => {
    await page.locator('#menu-toggle').click();
    await page.locator('#nav-overview').click();
    
    await expect(page.locator('#overview-screen')).toBeVisible();
    await expect(page.locator('#board-screen')).not.toBeVisible();
  });

  test('should navigate to workers', async ({ page }) => {
    await page.locator('#menu-toggle').click();
    await page.locator('#nav-workers').click();

    await expect(page.locator('#workers-screen')).toBeVisible();
    await expect(page.locator('#board-screen')).not.toBeVisible();
  });

  test.skip('should logout', async ({ page }) => {
    await page.locator('#menu-toggle').click();
    await page.locator('#nav-logout').click();

    // Wait for board screen to disappear first
    await expect(page.locator('#board-screen')).toBeHidden({ timeout: 10000 });

    // Then check auth screen is visible
    await page.waitForSelector('#auth-screen:not(.hidden)', { timeout: 10000 });
    await expect(page.locator('#auth-screen')).toBeVisible();
  });
});
