import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display login page', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Task Management System');
    await expect(page.locator('#login-form')).toBeVisible();
    await expect(page.locator('#login-username')).toBeVisible();
    await expect(page.locator('#login-password')).toBeVisible();
    await expect(page.locator('#login-button')).toBeVisible();
  });

  test('should show register form when clicking register tab', async ({ page }) => {
    await page.locator('#register-tab').click();
    await expect(page.locator('#register-form')).toBeVisible();
    await expect(page.locator('#login-form')).not.toBeVisible();
  });

  test('should show error on empty login', async ({ page }) => {
    await page.locator('#login-button').click();
    await expect(page.locator('#auth-error')).toContainText('Please enter username and password');
  });

  test('should login with default admin credentials', async ({ page }) => {
    // Fill in login form
    await page.locator('#login-username').fill('admin');
    await page.locator('#login-password').fill('admin');
    await page.locator('#login-button').click();

    // Wait for redirect to board
    await page.waitForURL('**/#board', { timeout: 5000 });
    
    // Verify board is displayed
    await expect(page.locator('#board-screen')).toBeVisible();
    await expect(page.locator('.app-header h1')).toContainText('Task Management');
  });

  test('should show error on invalid credentials', async ({ page }) => {
    await page.locator('#login-username').fill('invalid');
    await page.locator('#login-password').fill('invalid');
    await page.locator('#login-button').click();

    await expect(page.locator('#auth-error')).toBeVisible();
  });

  test('should register new user', async ({ page }) => {
    const username = `testuser_${Date.now()}`;
    const password = 'testpass123';

    await page.locator('#register-tab').click();
    await page.locator('#register-username').fill(username);
    await page.locator('#register-password').fill(password);
    await page.locator('#register-password-confirm').fill(password);
    await page.locator('#register-button').click();

    // Should redirect to board after registration
    await page.waitForURL('**/#board', { timeout: 5000 });
    await expect(page.locator('#board-screen')).toBeVisible();
  });

  test('should validate password confirmation', async ({ page }) => {
    await page.locator('#register-tab').click();
    await page.locator('#register-username').fill('testuser');
    await page.locator('#register-password').fill('password1');
    await page.locator('#register-password-confirm').fill('password2');
    await page.locator('#register-button').click();

    await expect(page.locator('#auth-error')).toContainText('Passwords do not match');
  });

  test('should validate password length', async ({ page }) => {
    await page.locator('#register-tab').click();
    await page.locator('#register-username').fill('testuser');
    await page.locator('#register-password').fill('short');
    await page.locator('#register-password-confirm').fill('short');
    await page.locator('#register-button').click();

    await expect(page.locator('#auth-error')).toContainText('at least 8 characters');
  });
});
