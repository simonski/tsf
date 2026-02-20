import { test, expect } from '@playwright/test';

const authHeaders = {
  'Authorization': 'Basic ' + Buffer.from('admin:admin').toString('base64'),
  'Content-Type': 'application/json'
};

function uniqueName(prefix: string): string {
  return `${prefix}-${Date.now()}-${Math.floor(Math.random() * 100000)}`;
}

async function createProject(request: any): Promise<string> {
  const response = await request.post('/api/v1/projects', {
    headers: authHeaders,
    data: {
      name: uniqueName('E2E Project'),
      description: 'Project for project file/note/comment e2e tests'
    }
  });
  expect(response.ok()).toBeTruthy();
  const project = await response.json();
  return project.id;
}

async function createTask(request: any, projectId: string): Promise<string> {
  const response = await request.post('/api/v1/tasks', {
    headers: authHeaders,
    data: {
      project_id: projectId,
      title: uniqueName('E2E Task'),
      type: 'task',
      description: 'Task for comment e2e tests'
    }
  });
  expect(response.ok()).toBeTruthy();
  const task = await response.json();
  return task.id;
}

test.describe('Project Files, Notes, and Comments API', () => {
  test('should perform full project file CRUD and validate duplicate names', async ({ request }) => {
    const projectId = await createProject(request);

    const createResponse = await request.post(`/api/v1/projects/${projectId}/files`, {
      headers: authHeaders,
      data: {
        name: 'README.md',
        content: '# hello'
      }
    });
    expect(createResponse.status()).toBe(201);
    const file = await createResponse.json();
    expect(file.name).toBe('README.md');
    expect(file.content).toBe('# hello');

    const listResponse = await request.get(`/api/v1/projects/${projectId}/files`, {
      headers: { 'Authorization': authHeaders.Authorization }
    });
    expect(listResponse.ok()).toBeTruthy();
    const files = await listResponse.json();
    expect(Array.isArray(files)).toBeTruthy();
    expect(files.length).toBe(1);

    const getResponse = await request.get(`/api/v1/projects/${projectId}/files/${file.id}`, {
      headers: { 'Authorization': authHeaders.Authorization }
    });
    expect(getResponse.ok()).toBeTruthy();
    const got = await getResponse.json();
    expect(got.id).toBe(file.id);

    const updateResponse = await request.put(`/api/v1/projects/${projectId}/files/${file.id}`, {
      headers: authHeaders,
      data: {
        content: '# updated'
      }
    });
    expect(updateResponse.ok()).toBeTruthy();
    const updated = await updateResponse.json();
    expect(updated.content).toBe('# updated');

    const duplicateResponse = await request.post(`/api/v1/projects/${projectId}/files`, {
      headers: authHeaders,
      data: {
        name: 'README.md',
        content: 'duplicate'
      }
    });
    expect(duplicateResponse.status()).toBe(400);

    const deleteResponse = await request.delete(`/api/v1/projects/${projectId}/files/${file.id}`, {
      headers: { 'Authorization': authHeaders.Authorization }
    });
    expect(deleteResponse.status()).toBe(204);
  });

  test('should perform full project note CRUD', async ({ request }) => {
    const projectId = await createProject(request);

    const createResponse = await request.post(`/api/v1/projects/${projectId}/notes`, {
      headers: authHeaders,
      data: {
        title: 'Kickoff',
        content: 'Initial decisions'
      }
    });
    expect(createResponse.status()).toBe(201);
    const note = await createResponse.json();
    expect(note.title).toBe('Kickoff');

    const listResponse = await request.get(`/api/v1/projects/${projectId}/notes`, {
      headers: { 'Authorization': authHeaders.Authorization }
    });
    expect(listResponse.ok()).toBeTruthy();
    const notes = await listResponse.json();
    expect(notes.length).toBe(1);

    const updateResponse = await request.put(`/api/v1/projects/${projectId}/notes/${note.id}`, {
      headers: authHeaders,
      data: {
        title: 'Kickoff Updated',
        content: 'Refined plan'
      }
    });
    expect(updateResponse.ok()).toBeTruthy();
    const updated = await updateResponse.json();
    expect(updated.title).toBe('Kickoff Updated');
    expect(updated.content).toBe('Refined plan');

    const deleteResponse = await request.delete(`/api/v1/projects/${projectId}/notes/${note.id}`, {
      headers: { 'Authorization': authHeaders.Authorization }
    });
    expect(deleteResponse.status()).toBe(204);
  });

  test('should run full comment lifecycle on both project and task entities', async ({ request }) => {
    const projectId = await createProject(request);
    const taskId = await createTask(request, projectId);

    const createProjectComment = await request.post('/api/v1/comments', {
      headers: authHeaders,
      data: {
        entity_type: 'project',
        entity_id: projectId,
        text: 'Project comment'
      }
    });
    expect(createProjectComment.status()).toBe(201);
    const projectComment = await createProjectComment.json();
    expect(projectComment.entity_type).toBe('project');

    const createTaskComment = await request.post('/api/v1/comments', {
      headers: authHeaders,
      data: {
        entity_type: 'task',
        entity_id: taskId,
        text: 'Task comment'
      }
    });
    expect(createTaskComment.status()).toBe(201);
    const taskComment = await createTaskComment.json();
    expect(taskComment.entity_type).toBe('task');

    const updateResponse = await request.put(`/api/v1/comments/${taskComment.id}`, {
      headers: authHeaders,
      data: {
        text: 'Task comment updated'
      }
    });
    expect(updateResponse.ok()).toBeTruthy();
    const updated = await updateResponse.json();
    expect(updated.text).toBe('Task comment updated');

    const historyResponse = await request.get(`/api/v1/comments/${taskComment.id}/history`, {
      headers: { 'Authorization': authHeaders.Authorization }
    });
    expect(historyResponse.ok()).toBeTruthy();
    const history = await historyResponse.json();
    expect(Array.isArray(history)).toBeTruthy();
    expect(history.length).toBeGreaterThanOrEqual(2);
    expect(history.some((h: any) => h.action === 'create')).toBeTruthy();
    expect(history.some((h: any) => h.action === 'edit')).toBeTruthy();

    const deleteResponse = await request.delete(`/api/v1/comments/${taskComment.id}`, {
      headers: { 'Authorization': authHeaders.Authorization }
    });
    expect(deleteResponse.status()).toBe(204);

    const listDefaultResponse = await request.get(`/api/v1/comments?entity_type=task&entity_id=${taskId}`, {
      headers: { 'Authorization': authHeaders.Authorization }
    });
    expect(listDefaultResponse.ok()).toBeTruthy();
    const defaultComments = await listDefaultResponse.json();
    expect(defaultComments.length).toBe(0);

    const listDeletedResponse = await request.get(`/api/v1/comments?entity_type=task&entity_id=${taskId}&include_deleted=true`, {
      headers: { 'Authorization': authHeaders.Authorization }
    });
    expect(listDeletedResponse.ok()).toBeTruthy();
    const allComments = await listDeletedResponse.json();
    expect(allComments.length).toBe(1);
    expect(allComments[0].is_deleted).toBeTruthy();

    // keep project comment also visible by default
    const listProjectResponse = await request.get(`/api/v1/comments?entity_type=project&entity_id=${projectId}`, {
      headers: { 'Authorization': authHeaders.Authorization }
    });
    expect(listProjectResponse.ok()).toBeTruthy();
    const projectComments = await listProjectResponse.json();
    expect(projectComments.length).toBe(1);
    expect(projectComments[0].id).toBe(projectComment.id);
  });

  test('should enforce owner-only update/delete for comments', async ({ request }) => {
    const projectId = await createProject(request);

    const createCommentResponse = await request.post('/api/v1/comments', {
      headers: authHeaders,
      data: {
        entity_type: 'project',
        entity_id: projectId,
        text: 'Owner-protected comment'
      }
    });
    expect(createCommentResponse.status()).toBe(201);
    const comment = await createCommentResponse.json();

    const userAuth = {
      'Authorization': 'Basic ' + Buffer.from('user:admin').toString('base64'),
      'Content-Type': 'application/json'
    };

    const nonOwnerUpdate = await request.put(`/api/v1/comments/${comment.id}`, {
      headers: userAuth,
      data: {
        text: 'not allowed'
      }
    });
    expect(nonOwnerUpdate.status()).toBe(403);

    const nonOwnerDelete = await request.delete(`/api/v1/comments/${comment.id}`, {
      headers: { 'Authorization': userAuth.Authorization }
    });
    expect(nonOwnerDelete.status()).toBe(403);
  });
});
