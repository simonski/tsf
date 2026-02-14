import { test, expect } from '@playwright/test';

test.describe('Comprehensive Navigation Tests', () => {
  let javascriptErrors: string[] = [];

  test.beforeEach(async ({ page }) => {
    // Only capture actual JavaScript errors (exceptions), not console.error/warn
    javascriptErrors = [];

    // Capture uncaught page errors (actual exceptions)
    page.on('pageerror', error => {
      javascriptErrors.push(`Page error: ${error.message}`);
    });

    // Login first
    await page.goto('/');
    await page.locator('#login-username').fill('admin');
    await page.locator('#login-password').fill('admin');
    await page.locator('#login-button').click();
    await expect(page.locator('#board-screen')).toBeVisible({ timeout: 10000 });
  });

  test.afterEach(async () => {
    // Check for actual JavaScript errors (exceptions) after each test
    if (javascriptErrors.length > 0) {
      console.log('JavaScript exceptions detected:', javascriptErrors);
    }
    expect(javascriptErrors, 'No JavaScript exceptions should occur').toHaveLength(0);
  });

  test('should navigate to Overview without errors', async ({ page }) => {
    await page.locator('#menu-toggle').click();
    await page.locator('#nav-overview').click();

    await expect(page.locator('#overview-screen')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#board-screen')).not.toBeVisible();
  });

  test('should navigate to Projects without errors', async ({ page }) => {
    await page.locator('#menu-toggle').click();
    await page.locator('#nav-projects').click();

    await expect(page.locator('#projects-screen')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#board-screen')).not.toBeVisible();
  });

  test('should navigate to Kanban without errors', async ({ page }) => {
    await page.locator('#menu-toggle').click();
    await page.locator('#nav-kanban').click();

    await expect(page.locator('#board-screen')).toBeVisible({ timeout: 5000 });
  });

  test('should navigate to Workers without errors', async ({ page }) => {
    await page.locator('#menu-toggle').click();
    await page.locator('#nav-workers').click();

    await expect(page.locator('#workers-screen')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#board-screen')).not.toBeVisible();
  });

  test('should navigate to Orchestrators without errors', async ({ page }) => {
    await page.locator('#menu-toggle').click();
    await page.locator('#nav-orchestrators').click();

    await expect(page.locator('#orchestrators-screen')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#board-screen')).not.toBeVisible();
  });

  test('should navigate to Activity without errors', async ({ page }) => {
    await page.locator('#menu-toggle').click();
    await page.locator('#nav-activity').click();

    await expect(page.locator('#activity-screen')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#board-screen')).not.toBeVisible();
  });

  test('should handle Users link gracefully (even if screen missing)', async ({ page }) => {
    // Users link is hidden by default, but let's test if admin can access it
    await page.locator('#menu-toggle').click();

    // Check if users link exists and is visible
    const usersLink = page.locator('#nav-users');
    const isVisible = await usersLink.isVisible();

    if (isVisible) {
      await usersLink.click();
      // Should either show users screen or fallback to board gracefully
      await page.waitForTimeout(1000);

      // App should still be functional (at least one screen visible)
      const screens = await page.locator('.screen:not(.hidden)').count();
      expect(screens).toBeGreaterThan(0);
    }
  });

  test('should handle Config link gracefully (even if screen missing)', async ({ page }) => {
    // Config link is hidden by default, but let's test if admin can access it
    await page.locator('#menu-toggle').click();

    // Check if config link exists and is visible
    const configLink = page.locator('#nav-config');
    const isVisible = await configLink.isVisible();

    if (isVisible) {
      await configLink.click();
      await page.waitForTimeout(1000);

      // Should show config screen
      await expect(page.locator('#config-screen')).toBeVisible({ timeout: 5000 });
      await expect(page.locator('#board-screen')).not.toBeVisible();
    }
  });

  test('should handle logo click without errors', async ({ page }) => {
    // Navigate away from board first
    await page.locator('#menu-toggle').click();
    await page.locator('#nav-overview').click();
    await expect(page.locator('#overview-screen')).toBeVisible();

    // Try to click logo if visible, different screens have different header structures
    const logoLink = page.locator('a.logo-link').first();
    if (await logoLink.isVisible()) {
      await logoLink.click();
      await page.waitForTimeout(500);
    }

    // Should have at least one visible screen
    const visibleScreens = await page.locator('.screen:not(.hidden)').count();
    expect(visibleScreens).toBeGreaterThan(0);
  });

  test('should open and close menu without errors', async ({ page }) => {
    // Open menu
    await page.locator('#menu-toggle').click();
    await expect(page.locator('#side-nav')).toHaveClass(/open/);

    // Close menu with close button
    await page.locator('#nav-close').click();
    await expect(page.locator('#side-nav')).not.toHaveClass(/open/);

    // Open again
    await page.locator('#menu-toggle').click();
    await expect(page.locator('#side-nav')).toHaveClass(/open/);

    // Close with overlay
    await page.locator('#nav-overlay').click();
    await expect(page.locator('#side-nav')).not.toHaveClass(/open/);
  });

  test('should switch between multiple screens without errors', async ({ page }) => {
    const navigationSequence = [
      { link: '#nav-overview', screen: '#overview-screen' },
      { link: '#nav-workers', screen: '#workers-screen' },
      { link: '#nav-kanban', screen: '#board-screen' },
      { link: '#nav-projects', screen: '#projects-screen' },
      { link: '#nav-orchestrators', screen: '#orchestrators-screen' },
      { link: '#nav-activity', screen: '#activity-screen' },
    ];

    for (const nav of navigationSequence) {
      // Find the visible hamburger menu (filter out hidden screens)
      const menuButton = page.locator('.screen:not(.hidden) button.hamburger-menu');
      await menuButton.waitFor({ state: 'visible', timeout: 10000 });
      await menuButton.click();
      await page.locator(nav.link).click();
      await expect(page.locator(nav.screen)).toBeVisible({ timeout: 5000 });
    }
  });

  test('should handle rapid navigation clicks without errors', async ({ page }) => {
    // Navigate between screens quickly
    await page.locator('#menu-toggle').click();
    await page.locator('#nav-overview').click();
    await expect(page.locator('#overview-screen')).toBeVisible();

    // Wait for hamburger menu from visible screen only
    const menuButton1 = page.locator('.screen:not(.hidden) button.hamburger-menu');
    await menuButton1.waitFor({ state: 'visible', timeout: 10000 });
    await menuButton1.click();
    await page.locator('#nav-workers').click();
    await expect(page.locator('#workers-screen')).toBeVisible();

    // Wait for hamburger menu from visible screen only
    const menuButton2 = page.locator('.screen:not(.hidden) button.hamburger-menu');
    await menuButton2.waitFor({ state: 'visible', timeout: 10000 });
    await menuButton2.click();
    await page.locator('#nav-kanban').click();
    await expect(page.locator('#board-screen')).toBeVisible();

    // App should still be functional
    const visibleScreens = await page.locator('.screen:not(.hidden)').count();
    expect(visibleScreens).toBe(1);
  });
});
