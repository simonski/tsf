import { test, expect } from '@playwright/test';

test.describe('Role Management - Comprehensive Coverage', () => {
  let testRoleId: string;
  let testProjectId: string;

  test.beforeEach(async ({ page, request }) => {
    // Login as admin
    await page.goto('/');
    await page.locator('#login-username').fill('admin');
    await page.locator('#login-password').fill('admin');
    await page.locator('#login-button').click();
    await expect(page.locator('#board-screen')).toBeVisible({ timeout: 10000 });

    // Get a project ID for testing project-scoped roles
    const projectsResponse = await request.get('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });
    const projects = await projectsResponse.json();
    if (projects && projects.length > 0) {
      testProjectId = projects[0].id;
    }
  });

  test('should list all roles via API', async ({ request }) => {
    const response = await request.get('/api/v1/roles', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const roles = await response.json();
    expect(Array.isArray(roles)).toBeTruthy();
    expect(roles.length).toBeGreaterThan(0);
  });

  test('should create a system-scoped role via API', async ({ request }) => {
    const response = await request.post('/api/v1/roles', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: `system_role_${Date.now()}`,
        scope: 'system',
        instructions: 'System-level role instructions'
      }
    });

    expect(response.ok()).toBeTruthy();
    const role = await response.json();
    expect(role.name).toContain('system_role');
    expect(role.scope).toBe('system');
    expect(role.instructions).toBe('System-level role instructions');
    testRoleId = role.id;
  });

  test('should create a project-scoped role via API', async ({ request }) => {
    const response = await request.post('/api/v1/roles', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: `project_role_${Date.now()}`,
        scope: 'project',
        project_id: testProjectId,
        instructions: 'Project-specific role instructions'
      }
    });

    expect(response.ok()).toBeTruthy();
    const role = await response.json();
    expect(role.name).toContain('project_role');
    expect(role.scope).toBe('project');
    expect(role.project_id).toBe(testProjectId);
  });

  test('should get role details via API', async ({ request }) => {
    // Create a role
    const createResponse = await request.post('/api/v1/roles', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: `get_test_role_${Date.now()}`,
        scope: 'system',
        instructions: 'Get test instructions'
      }
    });
    const createdRole = await createResponse.json();

    // Get the role
    const response = await request.get(`/api/v1/roles/${createdRole.id}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const role = await response.json();
    expect(role.id).toBe(createdRole.id);
    expect(role.name).toBe(createdRole.name);
  });

  test('should update role details via API', async ({ request }) => {
    // Create a role
    const createResponse = await request.post('/api/v1/roles', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: `update_test_role_${Date.now()}`,
        scope: 'system',
        instructions: 'Original instructions'
      }
    });
    const createdRole = await createResponse.json();

    // Update the role
    const updateResponse = await request.put(`/api/v1/roles/${createdRole.id}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        instructions: 'Updated instructions',
        name: 'updated_role_name'
      }
    });

    expect(updateResponse.ok()).toBeTruthy();
    const updatedRole = await updateResponse.json();
    expect(updatedRole.instructions).toBe('Updated instructions');
    expect(updatedRole.name).toBe('updated_role_name');
  });

  test('should delete a role via API', async ({ request }) => {
    // Create a role
    const createResponse = await request.post('/api/v1/roles', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: `delete_test_role_${Date.now()}`,
        scope: 'system',
        instructions: 'Role to be deleted'
      }
    });
    const createdRole = await createResponse.json();

    // Delete the role
    const deleteResponse = await request.delete(`/api/v1/roles/${createdRole.id}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(deleteResponse.status()).toBe(204);
  });

  test('should filter roles by scope via API', async ({ request }) => {
    const response = await request.get('/api/v1/roles?scope=system', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const roles = await response.json();
    expect(Array.isArray(roles)).toBeTruthy();
    roles.forEach((role: any) => {
      expect(role.scope).toBe('system');
    });
  });

  test('should filter roles by project via API', async ({ request }) => {
    const response = await request.get(`/api/v1/roles?project_id=${testProjectId}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const roles = await response.json();
    expect(Array.isArray(roles)).toBeTruthy();
    roles.forEach((role: any) => {
      if (role.scope === 'project') {
        expect(role.project_id).toBe(testProjectId);
      }
    });
  });

  test('should validate required fields when creating role', async ({ request }) => {
    const response = await request.post('/api/v1/roles', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        instructions: 'Missing name and scope'
      }
    });

    expect(response.status()).toBe(400);
  });

  test('should require project_id for project-scoped roles', async ({ request }) => {
    const response = await request.post('/api/v1/roles', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: 'project_role_no_id',
        scope: 'project',
        instructions: 'Project role without project_id'
      }
    });

    expect(response.status()).toBe(400);
  });

  test('should handle role not found errors', async ({ request }) => {
    const response = await request.get('/api/v1/roles/nonexistent-role-id', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.status()).toBe(404);
  });

  test('should require authentication for role operations', async ({ request }) => {
    const response = await request.get('/api/v1/roles');
    expect(response.status()).toBe(401);
  });

  test('should create role with detailed instructions', async ({ request }) => {
    const detailedInstructions = `
You are a backend developer role.
Your responsibilities include:
1. Building REST APIs
2. Database design and optimization
3. Writing unit and integration tests
4. Code review and documentation

Follow these guidelines:
- Use TypeScript or Go
- Follow RESTful conventions
- Write comprehensive tests
- Document all public APIs
    `.trim();

    const response = await request.post('/api/v1/roles', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: `detailed_role_${Date.now()}`,
        scope: 'system',
        instructions: detailedInstructions
      }
    });

    expect(response.ok()).toBeTruthy();
    const role = await response.json();
    expect(role.instructions).toBe(detailedInstructions);
  });

  test('should prevent duplicate role names in same scope', async ({ request }) => {
    const roleName = `duplicate_role_${Date.now()}`;

    // Create first role
    const response1 = await request.post('/api/v1/roles', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: roleName,
        scope: 'system',
        instructions: 'First role'
      }
    });
    expect(response1.ok()).toBeTruthy();

    // Try to create role with same name
    const response2 = await request.post('/api/v1/roles', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: roleName,
        scope: 'system',
        instructions: 'Second role'
      }
    });

    expect(response2.status()).toBe(400);
  });

  test('should allow same role name in different projects', async ({ request }) => {
    const roleName = `shared_role_name_${Date.now()}`;

    // Create role in first project
    const response1 = await request.post('/api/v1/roles', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: roleName,
        scope: 'project',
        project_id: testProjectId,
        instructions: 'Role in first project'
      }
    });
    expect(response1.ok()).toBeTruthy();

    // Get another project or create one
    const projectResponse = await request.get('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });
    const projects = await projectResponse.json();
    const otherProjectId = projects.find((p: any) => p.id !== testProjectId)?.id || testProjectId;

    // Create role with same name in different project (should be allowed)
    const response2 = await request.post('/api/v1/roles', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: roleName,
        scope: 'project',
        project_id: otherProjectId,
        instructions: 'Role in second project'
      }
    });

    // This should succeed if projects are different, or fail if same project
    if (otherProjectId !== testProjectId) {
      expect(response2.ok()).toBeTruthy();
    }
  });

  test('should verify system roles exist', async ({ request }) => {
    const response = await request.get('/api/v1/roles', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    const roles = await response.json();
    const systemRoles = roles.filter((r: any) => r.scope === 'system');

    // Should have at least some default system roles
    expect(systemRoles.length).toBeGreaterThan(0);

    // Common roles that might exist
    const roleNames = systemRoles.map((r: any) => r.name);
    // Just verify we have some system roles, exact names may vary
    expect(roleNames.length).toBeGreaterThan(0);
  });

  test('should update only specified fields', async ({ request }) => {
    // Create a role
    const createResponse = await request.post('/api/v1/roles', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: `partial_update_role_${Date.now()}`,
        scope: 'system',
        instructions: 'Original instructions'
      }
    });
    const createdRole = await createResponse.json();

    // Update only instructions
    const updateResponse = await request.put(`/api/v1/roles/${createdRole.id}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        instructions: 'New instructions only'
      }
    });

    expect(updateResponse.ok()).toBeTruthy();
    const updatedRole = await updateResponse.json();
    expect(updatedRole.instructions).toBe('New instructions only');
    expect(updatedRole.name).toBe(createdRole.name); // Name unchanged
  });
});
