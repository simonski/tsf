import { test, expect } from '@playwright/test';

test.describe('Navigation', () => {
  test.beforeEach(async ({ page }) => {
    // Login first
    await page.goto('/');
    await page.locator('#login-username').fill('admin');
    await page.locator('#login-password').fill('admin');
    await page.locator('#login-button').click();
    await page.waitForURL('**/#board', { timeout: 5000 });
  });

  test('should open and close navigation menu', async ({ page }) => {
    // Menu should be hidden initially
    await expect(page.locator('#side-nav')).not.toHaveClass(/active/);

    // Open menu
    await page.locator('#menu-toggle').click();
    await expect(page.locator('#side-nav')).toHaveClass(/active/);

    // Close menu
    await page.locator('#nav-close').click();
    await expect(page.locator('#side-nav')).not.toHaveClass(/active/);
  });

  test('should navigate to overview', async ({ page }) => {
    await page.locator('#menu-toggle').click();
    await page.locator('#nav-overview').click();
    
    await expect(page.locator('#overview-screen')).toBeVisible();
    await expect(page.locator('#board-screen')).not.toBeVisible();
  });

  test('should navigate to settings', async ({ page }) => {
    await page.locator('#menu-toggle').click();
    await page.locator('#nav-settings').click();
    
    await expect(page.locator('#settings-screen')).toBeVisible();
    await expect(page.locator('#board-screen')).not.toBeVisible();
  });

  test('should logout', async ({ page }) => {
    await page.locator('#menu-toggle').click();
    await page.locator('#nav-logout').click();
    
    // Should return to login page
    await expect(page.locator('#auth-screen')).toBeVisible();
    await expect(page.locator('#board-screen')).not.toBeVisible();
  });
});
