import { test, expect } from '@playwright/test';

test.describe('Configuration Management - Comprehensive Coverage', () => {
  test.beforeEach(async ({ page }) => {
    // Login as admin
    await page.goto('/');
    await page.locator('#login-username').fill('admin');
    await page.locator('#login-password').fill('admin');
    await page.locator('#login-button').click();
    await expect(page.locator('#board-screen')).toBeVisible({ timeout: 10000 });
  });

  test('should list all configuration via API', async ({ request }) => {
    const response = await request.get('/api/v1/config', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const config = await response.json();
    expect(Array.isArray(config)).toBeTruthy();
  });

  test('should set a configuration value via API', async ({ request }) => {
    const testKey = `test.key.${Date.now()}`;
    const testValue = 'test_value_123';

    const response = await request.put(`/api/v1/config/${testKey}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        value: testValue
      }
    });

    expect(response.ok()).toBeTruthy();
    const result = await response.json();
    expect(result.key).toBe(testKey);
    expect(result.value).toBe(testValue);
  });

  test('should get a specific configuration value via API', async ({ request }) => {
    const testKey = `test.get.${Date.now()}`;
    const testValue = 'get_test_value';

    // First set the value
    await request.put(`/api/v1/config/${testKey}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        value: testValue
      }
    });

    // Get the value
    const response = await request.get(`/api/v1/config/${testKey}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const config = await response.json();
    expect(config.key).toBe(testKey);
    expect(config.value).toBe(testValue);
  });

  test('should update a configuration value via API', async ({ request }) => {
    const testKey = `test.update.${Date.now()}`;
    const originalValue = 'original_value';
    const updatedValue = 'updated_value';

    // Set initial value
    await request.put(`/api/v1/config/${testKey}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        value: originalValue
      }
    });

    // Update the value
    const updateResponse = await request.put(`/api/v1/config/${testKey}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        value: updatedValue
      }
    });

    expect(updateResponse.ok()).toBeTruthy();
    const updatedConfig = await updateResponse.json();
    expect(updatedConfig.value).toBe(updatedValue);
  });

  test('should delete a configuration value via API', async ({ request }) => {
    const testKey = `test.delete.${Date.now()}`;

    // Set value
    await request.put(`/api/v1/config/${testKey}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        value: 'to_be_deleted'
      }
    });

    // Delete the value
    const deleteResponse = await request.delete(`/api/v1/config/${testKey}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(deleteResponse.status()).toBe(204);

    // Verify it's deleted
    const getResponse = await request.get(`/api/v1/config/${testKey}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(getResponse.status()).toBe(404);
  });

  test('should set orchestrator interval configuration', async ({ request }) => {
    const response = await request.put('/api/v1/config/orchestrator.interval', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        value: '60s'
      }
    });

    expect(response.ok()).toBeTruthy();
    const config = await response.json();
    expect(config.key).toBe('orchestrator.interval');
    expect(config.value).toBe('60s');
  });

  test('should set worker interval configuration', async ({ request }) => {
    const response = await request.put('/api/v1/config/worker.interval', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        value: '10s'
      }
    });

    expect(response.ok()).toBeTruthy();
    const config = await response.json();
    expect(config.key).toBe('worker.interval');
    expect(config.value).toBe('10s');
  });

  test('should handle nested configuration keys', async ({ request }) => {
    const nestedKey = `app.feature.sub.setting.${Date.now()}`;

    const response = await request.put(`/api/v1/config/${nestedKey}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        value: 'nested_value'
      }
    });

    expect(response.ok()).toBeTruthy();
    const config = await response.json();
    expect(config.key).toBe(nestedKey);
  });

  test('should validate required fields when setting config', async ({ request }) => {
    const response = await request.put('/api/v1/config/test.key', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        // Missing value
      }
    });

    expect(response.status()).toBe(400);
  });

  test('should validate value is provided when setting config', async ({ request }) => {
    const response = await request.put('/api/v1/config/test.some.key', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        // Missing value - the value field should be required
      }
    });

    expect(response.status()).toBe(400);
  });

  test('should handle config not found errors', async ({ request }) => {
    const response = await request.get('/api/v1/config/nonexistent.key.12345', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.status()).toBe(404);
  });

  test('should require authentication for config operations', async ({ request }) => {
    const response = await request.get('/api/v1/config');
    expect(response.status()).toBe(401);
  });

  test('should handle numeric configuration values', async ({ request }) => {
    const testKey = `test.numeric.${Date.now()}`;

    const response = await request.put(`/api/v1/config/${testKey}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        value: '12345'
      }
    });

    expect(response.ok()).toBeTruthy();
    const config = await response.json();
    expect(config.value).toBe('12345');
  });

  test('should handle boolean-like configuration values', async ({ request }) => {
    const testKey = `test.boolean.${Date.now()}`;

    const response = await request.put(`/api/v1/config/${testKey}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        value: 'true'
      }
    });

    expect(response.ok()).toBeTruthy();
    const config = await response.json();
    expect(config.value).toBe('true');
  });

  test('should handle JSON string configuration values', async ({ request }) => {
    const testKey = `test.json.${Date.now()}`;
    const jsonValue = JSON.stringify({ setting1: 'value1', setting2: 'value2' });

    const response = await request.put(`/api/v1/config/${testKey}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        value: jsonValue
      }
    });

    expect(response.ok()).toBeTruthy();
    const config = await response.json();
    expect(config.value).toBe(jsonValue);
  });

  test('should list all config and verify structure', async ({ request }) => {
    // Set a few known values first
    const timestamp = Date.now();
    await request.put(`/api/v1/config/test.a.${timestamp}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: { value: 'value_a' }
    });

    await request.put(`/api/v1/config/test.b.${timestamp}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: { value: 'value_b' }
    });

    // List all config
    const response = await request.get('/api/v1/config', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const configs = await response.json();
    expect(Array.isArray(configs)).toBeTruthy();

    // Verify each config has required fields
    configs.forEach((config: any) => {
      expect(config.key).toBeDefined();
      expect(config.value).toBeDefined();
      expect(config.created_at).toBeDefined();
      expect(config.updated_at).toBeDefined();
    });
  });

  test('should update existing config value idempotently', async ({ request }) => {
    const testKey = `test.idempotent.${Date.now()}`;

    // Set initial value
    await request.put(`/api/v1/config/${testKey}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        value: 'value1'
      }
    });

    // Set same key to new value (should update)
    const response2 = await request.put(`/api/v1/config/${testKey}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        value: 'value2'
      }
    });

    expect(response2.ok()).toBeTruthy();
    const config = await response2.json();
    expect(config.value).toBe('value2');

    // Verify via GET
    const getResponse = await request.get(`/api/v1/config/${testKey}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    const retrieved = await getResponse.json();
    expect(retrieved.value).toBe('value2');
  });

  test('should handle special characters in config keys', async ({ request }) => {
    const testKey = `test.special-chars_123.${Date.now()}`;

    const response = await request.put(`/api/v1/config/${testKey}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        value: 'special_value'
      }
    });

    expect(response.ok()).toBeTruthy();
    const config = await response.json();
    expect(config.key).toBe(testKey);
  });

  test('should handle empty string config values', async ({ request }) => {
    const testKey = `test.empty.${Date.now()}`;

    const response = await request.put(`/api/v1/config/${testKey}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        value: ''
      }
    });

    expect(response.ok()).toBeTruthy();
    const config = await response.json();
    expect(config.value).toBe('');
  });

  test('should handle whitespace-only config values', async ({ request }) => {
    const testKey = `test.whitespace.${Date.now()}`;

    const response = await request.put(`/api/v1/config/${testKey}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        value: '   '
      }
    });

    expect(response.ok()).toBeTruthy();
    const config = await response.json();
    expect(config.value).toBe('   ');
  });
});
