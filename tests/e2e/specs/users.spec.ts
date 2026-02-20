import { test, expect } from '@playwright/test';

test.describe('User Management - Comprehensive Coverage', () => {
  let testUserId: string;
  const testUsername = `testuser_${Date.now()}`;

  test.beforeEach(async ({ page }) => {
    // Login as admin
    await page.goto('/');
    await page.locator('#login-username').fill('admin');
    await page.locator('#login-password').fill('admin');
    await page.locator('#login-button').click();
    await expect(page.locator('#board-screen')).toBeVisible({ timeout: 10000 });
  });

  test('should list all users via API', async ({ request }) => {
    const response = await request.get('/api/v1/users', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const users = await response.json();
    expect(Array.isArray(users)).toBeTruthy();
    expect(users.length).toBeGreaterThan(0);

    // Verify admin user exists
    const adminUser = users.find((u: any) => u.username === 'admin');
    expect(adminUser).toBeTruthy();
  });

  test('should create a new user via API (register)', async ({ request }) => {
    const response = await request.post('/api/v1/users', {
      headers: {
        'Content-Type': 'application/json'
      },
      data: {
        username: testUsername,
        password: 'testpass123',
        type: 'human'
      }
    });

    expect(response.ok()).toBeTruthy();
    const user = await response.json();
    expect(user.username).toBe(testUsername);
    expect(user.type).toBe('human');
    expect(user.password).toBeUndefined(); // Password should not be returned
    testUserId = user.id;
  });

  test('should get user details via API', async ({ request }) => {
    // First create a user
    const createResponse = await request.post('/api/v1/users', {
      headers: {
        'Content-Type': 'application/json'
      },
      data: {
        username: `getuser_${Date.now()}`,
        password: 'testpass123',
        type: 'human'
      }
    });
    const createdUser = await createResponse.json();

    // Get the user (as admin)
    const response = await request.get(`/api/v1/users/${createdUser.id}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const user = await response.json();
    expect(user.id).toBe(createdUser.id);
    expect(user.username).toBe(createdUser.username);
  });

  test('should update user details via API', async ({ request }) => {
    // Create a user
    const createResponse = await request.post('/api/v1/users', {
      headers: {
        'Content-Type': 'application/json'
      },
      data: {
        username: `updateuser_${Date.now()}`,
        password: 'testpass123',
        type: 'human'
      }
    });
    const createdUser = await createResponse.json();

    // Update the user (as admin)
    const updateResponse = await request.put(`/api/v1/users/${createdUser.id}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        type: 'worker'
      }
    });

    expect(updateResponse.ok()).toBeTruthy();
    const updatedUser = await updateResponse.json();
    expect(updatedUser.type).toBe('worker');
  });

  test('should delete a user via API', async ({ request }) => {
    // Create a user
    const createResponse = await request.post('/api/v1/users', {
      headers: {
        'Content-Type': 'application/json'
      },
      data: {
        username: `deleteuser_${Date.now()}`,
        password: 'testpass123',
        type: 'human'
      }
    });
    const createdUser = await createResponse.json();

    // Delete the user (as admin)
    const deleteResponse = await request.delete(`/api/v1/users/${createdUser.id}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(deleteResponse.status()).toBe(204);
  });

  test('should reject duplicate username', async ({ request }) => {
    const username = `duplicate_${Date.now()}`;

    // Create first user
    const response1 = await request.post('/api/v1/users', {
      headers: {
        'Content-Type': 'application/json'
      },
      data: {
        username: username,
        password: 'testpass123',
        type: 'human'
      }
    });
    expect(response1.ok()).toBeTruthy();

    // Try to create user with same username
    const response2 = await request.post('/api/v1/users', {
      headers: {
        'Content-Type': 'application/json'
      },
      data: {
        username: username,
        password: 'testpass123',
        type: 'human'
      }
    });

    expect(response2.status()).toBe(400);
  });

  test('should validate required fields when creating user', async ({ request }) => {
    const response = await request.post('/api/v1/users', {
      headers: {
        'Content-Type': 'application/json'
      },
      data: {
        username: 'incomplete'
        // Missing password
      }
    });

    expect(response.status()).toBe(400);
  });

  test('should handle user not found errors', async ({ request }) => {
    const response = await request.get('/api/v1/users/nonexistent-id', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.status()).toBe(404);
  });

  test('should filter users by type via API', async ({ request }) => {
    const response = await request.get('/api/v1/users?type=human', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const users = await response.json();
    expect(Array.isArray(users)).toBeTruthy();
    users.forEach((user: any) => {
      expect(user.type).toBe('human');
    });
  });

  test('should create worker user via API', async ({ request }) => {
    const response = await request.post('/api/v1/users', {
      headers: {
        'Content-Type': 'application/json'
      },
      data: {
        username: `worker_${Date.now()}`,
        password: 'workerpass123',
        type: 'worker'
      }
    });

    expect(response.ok()).toBeTruthy();
    const user = await response.json();
    expect(user.type).toBe('worker');
  });

  test('should create orchestrator user via API', async ({ request }) => {
    const response = await request.post('/api/v1/users', {
      headers: {
        'Content-Type': 'application/json'
      },
      data: {
        username: `orchestrator_${Date.now()}`,
        password: 'orchpass123',
        type: 'orchestrator'
      }
    });

    expect(response.ok()).toBeTruthy();
    const user = await response.json();
    expect(user.type).toBe('orchestrator');
  });

  test('should authenticate with created user credentials', async ({ request }) => {
    const username = `authuser_${Date.now()}`;
    const password = 'authpass123';

    // Create user
    await request.post('/api/v1/users', {
      headers: {
        'Content-Type': 'application/json'
      },
      data: {
        username: username,
        password: password,
        type: 'human'
      }
    });

    // Test authentication
    const response = await request.get('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from(`${username}:${password}`).toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
  });

  test('should reject authentication with wrong password', async ({ request }) => {
    const response = await request.get('/api/v1/users', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:wrongpassword').toString('base64')
      }
    });

    expect(response.status()).toBe(401);
  });

  test('should not return password in user response', async ({ request }) => {
    const response = await request.get('/api/v1/users', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    const users = await response.json();
    users.forEach((user: any) => {
      expect(user.password).toBeUndefined();
      expect(user.password_hash).toBeUndefined();
    });
  });

  test('should validate username format', async ({ request }) => {
    const response = await request.post('/api/v1/users', {
      headers: {
        'Content-Type': 'application/json'
      },
      data: {
        username: '', // Empty username
        password: 'testpass123',
        type: 'human'
      }
    });

    expect(response.status()).toBe(400);
  });

  test('should handle concurrent user creation', async ({ request }) => {
    const username = `concurrent_${Date.now()}`;

    // Try to create same user twice simultaneously
    const promises = [
      request.post('/api/v1/users', {
        headers: { 'Content-Type': 'application/json' },
        data: { username, password: 'pass1', type: 'human' }
      }),
      request.post('/api/v1/users', {
        headers: { 'Content-Type': 'application/json' },
        data: { username, password: 'pass2', type: 'human' }
      })
    ];

    const results = await Promise.all(promises);

    // One should succeed, one should fail
    const succeeded = results.filter(r => r.ok()).length;
    const failed = results.filter(r => !r.ok()).length;

    expect(succeeded).toBe(1);
    expect(failed).toBe(1);
  });

  test('should prevent non-admin from deleting users', async ({ request }) => {
    // Create a regular user
    const createResponse = await request.post('/api/v1/users', {
      headers: { 'Content-Type': 'application/json' },
      data: {
        username: `regularuser_${Date.now()}`,
        password: 'regularpass',
        type: 'human'
      }
    });
    const regularUser = await createResponse.json();

    // Create another user to attempt deletion
    const targetResponse = await request.post('/api/v1/users', {
      headers: { 'Content-Type': 'application/json' },
      data: {
        username: `targetuser_${Date.now()}`,
        password: 'targetpass',
        type: 'human'
      }
    });
    const targetUser = await targetResponse.json();

    // Try to delete as non-admin (this might return 403 or similar)
    const deleteResponse = await request.delete(`/api/v1/users/${targetUser.id}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from(`${regularUser.username}:regularpass`).toString('base64')
      }
    });

    // Should not be allowed (403 or 401)
    expect([401, 403]).toContain(deleteResponse.status());
  });
});
