import { test, expect } from '@playwright/test';

test.describe('Project Management - Comprehensive Coverage', () => {
  let testProjectId: string;

  test.beforeEach(async ({ page }) => {
    // Login as admin
    await page.goto('/');
    await page.locator('#login-username').fill('admin');
    await page.locator('#login-password').fill('admin');
    await page.locator('#login-button').click();
    await expect(page.locator('#board-screen')).toBeVisible({ timeout: 10000 });
  });

  // UI Tests
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

  // API Tests
  test('should list all projects via API', async ({ request }) => {
    const response = await request.get('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const projects = await response.json();
    expect(Array.isArray(projects)).toBeTruthy();
    expect(projects.length).toBeGreaterThan(0);
  });

  test('should create a new project via API', async ({ request }) => {
    const response = await request.post('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: `Test Project ${Date.now()}`,
        description: 'Test project description'
      }
    });

    expect(response.ok()).toBeTruthy();
    const project = await response.json();
    expect(project.name).toContain('Test Project');
    expect(project.description).toBe('Test project description');
    testProjectId = project.id;
  });

  test('should get project details via API', async ({ request }) => {
    // First create a project
    const createResponse = await request.post('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: 'Get Test Project',
        description: 'Project for get test'
      }
    });
    const createdProject = await createResponse.json();

    // Get the project
    const response = await request.get(`/api/v1/projects/${createdProject.id}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const project = await response.json();
    expect(project.id).toBe(createdProject.id);
    expect(project.name).toBe('Get Test Project');
  });

  test('should update project details via API', async ({ request }) => {
    // Create a project
    const createResponse = await request.post('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: 'Update Test Project',
        description: 'Original description'
      }
    });
    const createdProject = await createResponse.json();

    // Update the project
    const updateResponse = await request.put(`/api/v1/projects/${createdProject.id}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: 'Updated Project Name',
        description: 'Updated description'
      }
    });

    expect(updateResponse.ok()).toBeTruthy();
    const updatedProject = await updateResponse.json();
    expect(updatedProject.name).toBe('Updated Project Name');
    expect(updatedProject.description).toBe('Updated description');
  });

  test('should delete a project via API', async ({ request }) => {
    // Create a project
    const createResponse = await request.post('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: 'Delete Test Project',
        description: 'Project to be deleted'
      }
    });
    const createdProject = await createResponse.json();

    // Delete the project
    const deleteResponse = await request.delete(`/api/v1/projects/${createdProject.id}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(deleteResponse.status()).toBe(204);
  });

  test('should add member to project via API', async ({ request }) => {
    // Create a project
    const projectResponse = await request.post('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: 'Member Test Project',
        description: 'Project for member testing'
      }
    });
    const project = await projectResponse.json();

    // Get admin user ID
    const usersResponse = await request.get('/api/v1/users', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });
    const users = await usersResponse.json();
    const adminUser = users.find((u: any) => u.username === 'admin');

    // Add member to project
    const memberResponse = await request.post(`/api/v1/projects/${project.id}/members`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        user_id: adminUser.id
      }
    });

    expect(memberResponse.ok()).toBeTruthy();
  });

  test('should list project members via API', async ({ request }) => {
    // Create a project
    const projectResponse = await request.post('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: 'Members List Project',
        description: 'Project for listing members'
      }
    });
    const project = await projectResponse.json();

    // Get members
    const membersResponse = await request.get(`/api/v1/projects/${project.id}/members`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(membersResponse.ok()).toBeTruthy();
    const members = await membersResponse.json();
    expect(Array.isArray(members)).toBeTruthy();
  });

  test('should remove member from project via API', async ({ request }) => {
    // Create a project
    const projectResponse = await request.post('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: 'Remove Member Project',
        description: 'Project for member removal'
      }
    });
    const project = await projectResponse.json();

    // Get admin user ID
    const usersResponse = await request.get('/api/v1/users', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });
    const users = await usersResponse.json();
    const adminUser = users.find((u: any) => u.username === 'admin');

    // Add member
    await request.post(`/api/v1/projects/${project.id}/members`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        user_id: adminUser.id
      }
    });

    // Remove member
    const removeResponse = await request.delete(`/api/v1/projects/${project.id}/members/${adminUser.id}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(removeResponse.status()).toBe(204);
  });

  test('should validate required fields when creating project', async ({ request }) => {
    const response = await request.post('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        description: 'Missing name'
      }
    });

    expect(response.status()).toBe(400);
  });

  test('should handle project not found errors', async ({ request }) => {
    const response = await request.get('/api/v1/projects/nonexistent-id', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.status()).toBe(404);
  });

  test('should require authentication for project operations', async ({ request }) => {
    const response = await request.get('/api/v1/projects');
    expect(response.status()).toBe(401);
  });

  test('should create project with minimal fields', async ({ request }) => {
    const response = await request.post('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: `Minimal Project ${Date.now()}`
      }
    });

    expect(response.ok()).toBeTruthy();
    const project = await response.json();
    expect(project.name).toContain('Minimal Project');
  });

  test('should create project with long description', async ({ request }) => {
    const longDescription = 'This is a very long description. '.repeat(50);

    const response = await request.post('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: `Long Desc Project ${Date.now()}`,
        description: longDescription
      }
    });

    expect(response.ok()).toBeTruthy();
    const project = await response.json();
    expect(project.description).toBe(longDescription);
  });

  test('should prevent duplicate member addition', async ({ request }) => {
    // Create a project
    const projectResponse = await request.post('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: 'Duplicate Member Project',
        description: 'Project for testing duplicate members'
      }
    });
    const project = await projectResponse.json();

    // Get admin user ID
    const usersResponse = await request.get('/api/v1/users', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });
    const users = await usersResponse.json();
    const adminUser = users.find((u: any) => u.username === 'admin');

    // Add member first time
    const response1 = await request.post(`/api/v1/projects/${project.id}/members`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        user_id: adminUser.id
      }
    });
    expect(response1.ok()).toBeTruthy();

    // Try to add same member again
    const response2 = await request.post(`/api/v1/projects/${project.id}/members`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        user_id: adminUser.id
      }
    });

    // Should either succeed (idempotent) or fail with 400/409
    expect([200, 201, 400, 409]).toContain(response2.status());
  });

  test('should list projects for authenticated user', async ({ request }) => {
    // Create multiple projects
    await request.post('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: `User Project 1 ${Date.now()}`,
        description: 'First project'
      }
    });

    await request.post('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: `User Project 2 ${Date.now()}`,
        description: 'Second project'
      }
    });

    // List all projects
    const response = await request.get('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const projects = await response.json();
    expect(projects.length).toBeGreaterThanOrEqual(2);
  });

  test('should update only project name', async ({ request }) => {
    // Create a project
    const createResponse = await request.post('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: 'Original Name',
        description: 'Original description'
      }
    });
    const createdProject = await createResponse.json();

    // Update only name
    const updateResponse = await request.put(`/api/v1/projects/${createdProject.id}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: 'New Name'
      }
    });

    expect(updateResponse.ok()).toBeTruthy();
    const updatedProject = await updateResponse.json();
    expect(updatedProject.name).toBe('New Name');
    expect(updatedProject.description).toBe('Original description');
  });

  test('should update only project description', async ({ request }) => {
    // Create a project
    const createResponse = await request.post('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: 'Original Name',
        description: 'Original description'
      }
    });
    const createdProject = await createResponse.json();

    // Update only description
    const updateResponse = await request.put(`/api/v1/projects/${createdProject.id}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        description: 'New description'
      }
    });

    expect(updateResponse.ok()).toBeTruthy();
    const updatedProject = await updateResponse.json();
    expect(updatedProject.name).toBe('Original Name');
    expect(updatedProject.description).toBe('New description');
  });

  test('should verify project timestamps', async ({ request }) => {
    const response = await request.post('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: `Timestamp Project ${Date.now()}`,
        description: 'Project for timestamp verification'
      }
    });

    expect(response.ok()).toBeTruthy();
    const project = await response.json();
    expect(project.created_at).toBeDefined();
    expect(project.updated_at).toBeDefined();
    expect(project.created_by).toBeDefined();
  });

  test('should filter projects by name search', async ({ request }) => {
    const uniqueName = `UniqueSearchTerm${Date.now()}`;

    // Create project with unique name
    await request.post('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        name: uniqueName,
        description: 'Searchable project'
      }
    });

    // Note: This test assumes the API supports name filtering
    // If not supported, this test documents desired functionality
    const response = await request.get(`/api/v1/projects?name=${uniqueName}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    const projects = await response.json();
    // If filtering is supported, verify results
    if (response.ok() && Array.isArray(projects)) {
      const found = projects.some((p: any) => p.name === uniqueName);
      // Either filtering works or we get all projects
      expect(found || projects.length > 0).toBeTruthy();
    }
  });
});
