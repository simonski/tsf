import { test, expect } from '@playwright/test';

test.describe('Task Management - Comprehensive Coverage', () => {
  let testProjectId: string;
  let testTaskId: string;

  test.beforeEach(async ({ page }) => {
    // Login as admin
    await page.goto('/');
    await page.locator('#login-username').fill('admin');
    await page.locator('#login-password').fill('admin');
    await page.locator('#login-button').click();
    await expect(page.locator('#board-screen')).toBeVisible({ timeout: 10000 });

    // Get first project ID for testing
    const response = await page.request.get('/api/v1/projects', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });
    const projects = await response.json();
    if (projects && projects.length > 0) {
      testProjectId = projects[0].id;
    }
  });

  test('should list all tasks via API', async ({ request }) => {
    const response = await request.get('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const tasks = await response.json();
    expect(Array.isArray(tasks)).toBeTruthy();
  });

  test('should create a new task via API', async ({ request }) => {
    const response = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Test Task ' + Date.now(),
        type: 'feature',
        description: 'Test task description',
        priority: 'high'
      }
    });

    expect(response.ok()).toBeTruthy();
    const task = await response.json();
    expect(task.title).toContain('Test Task');
    expect(task.type).toBe('feature');
    expect(task.priority).toBe('high');
    testTaskId = task.id;
  });

  test('should get task details via API', async ({ request }) => {
    // First create a task
    const createResponse = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Get Test Task',
        type: 'bug',
        description: 'Task for get test'
      }
    });
    const createdTask = await createResponse.json();

    // Get the task
    const response = await request.get(`/api/v1/tasks/${createdTask.id}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const task = await response.json();
    expect(task.id).toBe(createdTask.id);
    expect(task.title).toBe('Get Test Task');
  });

  test('should update task details via API', async ({ request }) => {
    // Create a task
    const createResponse = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Update Test Task',
        type: 'feature',
        description: 'Original description'
      }
    });
    const createdTask = await createResponse.json();

    // Update the task
    const updateResponse = await request.put(`/api/v1/tasks/${createdTask.id}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        title: 'Updated Task Title',
        description: 'Updated description',
        priority: 'critical'
      }
    });

    expect(updateResponse.ok()).toBeTruthy();
    const updatedTask = await updateResponse.json();
    expect(updatedTask.title).toBe('Updated Task Title');
    expect(updatedTask.description).toBe('Updated description');
    expect(updatedTask.priority).toBe('critical');
  });

  test('should delete a task via API', async ({ request }) => {
    // Create a task
    const createResponse = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Delete Test Task',
        type: 'chore',
        description: 'Task to be deleted'
      }
    });
    const createdTask = await createResponse.json();

    // Delete the task
    const deleteResponse = await request.delete(`/api/v1/tasks/${createdTask.id}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(deleteResponse.status()).toBe(204);
  });

  test('should claim a task via API', async ({ request }) => {
    // Create a task
    const createResponse = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Claim Test Task',
        type: 'feature',
        description: 'Task to be claimed'
      }
    });
    const createdTask = await createResponse.json();

    // Claim the task
    const claimResponse = await request.post(`/api/v1/tasks/${createdTask.id}/claim`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(claimResponse.ok()).toBeTruthy();
    const claimedTask = await claimResponse.json();
    expect(claimedTask.status).toBe('active');
    expect(claimedTask.worker_id).toBeTruthy();
  });

  test('should assign a task to a worker via API', async ({ request }) => {
    // Get admin user ID
    const usersResponse = await request.get('/api/v1/users', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });
    const users = await usersResponse.json();
    const adminUser = users.find((u: any) => u.username === 'admin');

    // Create a task
    const createResponse = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Assign Test Task',
        type: 'feature',
        description: 'Task to be assigned'
      }
    });
    const createdTask = await createResponse.json();

    // Assign the task
    const assignResponse = await request.post(`/api/v1/tasks/${createdTask.id}/assign`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        user_id: adminUser.id
      }
    });

    expect(assignResponse.ok()).toBeTruthy();
    const assignedTask = await assignResponse.json();
    expect(assignedTask.status).toBe('active');
    expect(assignedTask.worker_id).toBe(adminUser.id);
  });

  test('should free a task via API', async ({ request }) => {
    // Create and claim a task
    const createResponse = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Free Test Task',
        type: 'feature',
        description: 'Task to be freed'
      }
    });
    const createdTask = await createResponse.json();

    await request.post(`/api/v1/tasks/${createdTask.id}/claim`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    // Free the task
    const freeResponse = await request.post(`/api/v1/tasks/${createdTask.id}/free`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(freeResponse.ok()).toBeTruthy();
    const freedTask = await freeResponse.json();
    expect(freedTask.status).toBe('idle');
    expect(freedTask.worker_id).toBeNull();
  });

  test('should complete a task via API', async ({ request }) => {
    // Create and claim a task
    const createResponse = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Complete Test Task',
        type: 'feature',
        description: 'Task to be completed'
      }
    });
    const createdTask = await createResponse.json();

    await request.post(`/api/v1/tasks/${createdTask.id}/claim`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    // Complete the task
    const completeResponse = await request.post(`/api/v1/tasks/${createdTask.id}/complete`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        notes: 'Task completed successfully'
      }
    });

    expect(completeResponse.ok()).toBeTruthy();
    const completedTask = await completeResponse.json();
    expect(completedTask.is_complete).toBeTruthy();
    expect(completedTask.completed_at).toBeTruthy();
  });

  test('should get task history via API', async ({ request }) => {
    // Create, claim, and complete a task to generate history
    const createResponse = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'History Test Task',
        type: 'feature',
        description: 'Task for history test'
      }
    });
    const createdTask = await createResponse.json();

    await request.post(`/api/v1/tasks/${createdTask.id}/claim`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    // Get history
    const historyResponse = await request.get(`/api/v1/tasks/${createdTask.id}/history`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(historyResponse.ok()).toBeTruthy();
    const history = await historyResponse.json();
    expect(Array.isArray(history)).toBeTruthy();
    expect(history.length).toBeGreaterThan(0);
  });

  test('should get task dependencies via API', async ({ request }) => {
    // Create two tasks
    const task1Response = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Dependency Task 1',
        type: 'feature',
        description: 'First task'
      }
    });
    const task1 = await task1Response.json();

    const task2Response = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Dependency Task 2',
        type: 'feature',
        description: 'Second task depends on first',
        depends_on_task_id: task1.id
      }
    });
    const task2 = await task2Response.json();

    // Get dependencies
    const depsResponse = await request.get(`/api/v1/tasks/${task2.id}/dependencies`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(depsResponse.ok()).toBeTruthy();
    const deps = await depsResponse.json();
    expect(deps.blocked_by).toBeTruthy();
    expect(deps.blocked_by.id).toBe(task1.id);
  });

  test('should unassign a task via API', async ({ request }) => {
    // Get admin user ID
    const usersResponse = await request.get('/api/v1/users', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });
    const users = await usersResponse.json();
    const adminUser = users.find((u: any) => u.username === 'admin');

    // Create and assign a task
    const createResponse = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Unassign Test Task',
        type: 'feature',
        description: 'Task to be unassigned'
      }
    });
    const createdTask = await createResponse.json();

    await request.post(`/api/v1/tasks/${createdTask.id}/assign`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        user_id: adminUser.id
      }
    });

    // Unassign the task
    const unassignResponse = await request.post(`/api/v1/tasks/${createdTask.id}/unassign`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(unassignResponse.ok()).toBeTruthy();
    const unassignedTask = await unassignResponse.json();
    expect(unassignedTask.status).toBe('idle');
    expect(unassignedTask.worker_id).toBeNull();
  });

  test('should add comment to task via API', async ({ request }) => {
    // Create a task
    const createResponse = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Comment Test Task',
        type: 'feature',
        description: 'Task for comment test'
      }
    });
    const createdTask = await createResponse.json();

    // Add comment
    const commentResponse = await request.post(`/api/v1/tasks/${createdTask.id}/comments`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        text: 'This is a test comment'
      }
    });

    expect(commentResponse.ok()).toBeTruthy();
    const comment = await commentResponse.json();
    expect(comment.text).toBe('This is a test comment');
    expect(comment.author).toBe('admin');
  });

  test('should get task comments via API', async ({ request }) => {
    // Create a task and add a comment
    const createResponse = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Get Comments Test Task',
        type: 'feature',
        description: 'Task for getting comments'
      }
    });
    const createdTask = await createResponse.json();

    await request.post(`/api/v1/tasks/${createdTask.id}/comments`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        text: 'First comment'
      }
    });

    // Get comments
    const commentsResponse = await request.get(`/api/v1/tasks/${createdTask.id}/comments`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(commentsResponse.ok()).toBeTruthy();
    const comments = await commentsResponse.json();
    expect(Array.isArray(comments)).toBeTruthy();
    expect(comments.length).toBeGreaterThan(0);
    expect(comments[0].text).toBe('First comment');
  });

  test('should filter tasks by status via API', async ({ request }) => {
    const response = await request.get('/api/v1/tasks?status=idle', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const tasks = await response.json();
    expect(Array.isArray(tasks)).toBeTruthy();
    tasks.forEach((task: any) => {
      expect(task.status).toBe('idle');
    });
  });

  test('should filter tasks by project via API', async ({ request }) => {
    const response = await request.get(`/api/v1/tasks?project_id=${testProjectId}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const tasks = await response.json();
    expect(Array.isArray(tasks)).toBeTruthy();
    tasks.forEach((task: any) => {
      expect(task.project_id).toBe(testProjectId);
    });
  });

  test('should filter tasks by priority via API', async ({ request }) => {
    const response = await request.get('/api/v1/tasks?priority=high', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const tasks = await response.json();
    expect(Array.isArray(tasks)).toBeTruthy();
  });

  test('should get task status statistics via API', async ({ request }) => {
    const response = await request.get('/api/v1/status', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.ok()).toBeTruthy();
    const status = await response.json();
    expect(status.grand_total).toBeDefined();
    expect(status.by_type_and_status).toBeDefined();
    expect(status.totals_by_type).toBeDefined();
  });

  test('should handle task with labels via API', async ({ request }) => {
    const response = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Labeled Task',
        type: 'feature',
        description: 'Task with labels',
        labels: ['urgent', 'frontend', 'bug-fix']
      }
    });

    expect(response.ok()).toBeTruthy();
    const task = await response.json();
    expect(task.labels).toEqual(['urgent', 'frontend', 'bug-fix']);
  });

  test('should handle task with estimated effort via API', async ({ request }) => {
    const response = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Estimated Task',
        type: 'feature',
        description: 'Task with effort estimate',
        estimated_effort: 8
      }
    });

    expect(response.ok()).toBeTruthy();
    const task = await response.json();
    expect(task.estimated_effort).toBe(8);
  });

  test('should handle task with acceptance criteria via API', async ({ request }) => {
    const response = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Criteria Task',
        type: 'feature',
        description: 'Task with acceptance criteria',
        acceptance_criteria: '- Login successful\n- User redirected to dashboard\n- Session created'
      }
    });

    expect(response.ok()).toBeTruthy();
    const task = await response.json();
    expect(task.acceptance_criteria).toContain('Login successful');
  });

  test('should prevent claiming task blocked by dependency via API', async ({ request }) => {
    // Create two tasks with dependency
    const task1Response = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Blocking Task',
        type: 'feature',
        description: 'This task blocks another'
      }
    });
    const task1 = await task1Response.json();

    const task2Response = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Blocked Task',
        type: 'feature',
        description: 'This task is blocked',
        depends_on_task_id: task1.id
      }
    });
    const task2 = await task2Response.json();

    // Try to claim blocked task
    const claimResponse = await request.post(`/api/v1/tasks/${task2.id}/claim`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(claimResponse.status()).toBe(409); // Conflict
  });

  test('should validate required fields when creating task via API', async ({ request }) => {
    const response = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        title: 'Incomplete Task'
        // Missing project_id, type, and description
      }
    });

    expect(response.status()).toBe(400);
  });

  test('should handle task not found errors via API', async ({ request }) => {
    const response = await request.get('/api/v1/tasks/nonexistent-id', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64')
      }
    });

    expect(response.status()).toBe(404);
  });

  test('should require authentication for task operations', async ({ request }) => {
    const response = await request.get('/api/v1/tasks');
    expect(response.status()).toBe(401);
  });

  test('should update task to remove dependency via API', async ({ request }) => {
    // Create two tasks with dependency
    const task1Response = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Original Dependency',
        type: 'feature',
        description: 'Original blocking task'
      }
    });
    const task1 = await task1Response.json();

    const task2Response = await request.post('/api/v1/tasks', {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        project_id: testProjectId,
        title: 'Dependent Task',
        type: 'feature',
        description: 'Task with dependency to remove',
        depends_on_task_id: task1.id
      }
    });
    const task2 = await task2Response.json();

    // Remove dependency
    const updateResponse = await request.put(`/api/v1/tasks/${task2.id}`, {
      headers: {
        'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
        'Content-Type': 'application/json'
      },
      data: {
        depends_on_task_id: null
      }
    });

    expect(updateResponse.ok()).toBeTruthy();
    const updatedTask = await updateResponse.json();
    expect(updatedTask.depends_on_task_id).toBeNull();
  });
});
