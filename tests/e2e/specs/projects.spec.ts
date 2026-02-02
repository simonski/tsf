import { test, expect } from '@playwright/test';

test.describe('Project Management', () => {
  test.beforeEach(async ({ page }) => {
    // Login
    await page.goto('/');
    await page.locator('#login-username').fill('admin');
    await page.locator('#login-password').fill('admin');
    await page.locator('#login-button').click();
    await page.waitForURL('**/#board', { timeout: 5000 });
  });

  test('should display project selector', async ({ page }) => {
    await expect(page.locator('#project-selector')).toBeVisible();
  });

  test('should load projects in selector', async ({ page }) => {
    const selector = page.locator('#project-selector');
    await expect(selector).toBeVisible();
    
    // Wait for projects to load
    await page.waitForTimeout(1000);
    
    const options = await selector.locator('option').count();
    expect(options).toBeGreaterThan(0);
  });

  test('should switch projects', async ({ page }) => {
    const selector = page.locator('#project-selector');
    
    // Select first project
    await selector.selectOption({ index: 1 });
    
    // Wait for tasks to load
    await page.waitForTimeout(500);
    
    // Board should be visible
    await expect(page.locator('.kanban-board')).toBeVisible();
  });
});
