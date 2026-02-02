// API Configuration
const API_BASE = '/api/v1';

// State Management
const state = {
    currentUser: null,
    credentials: null,
    currentProject: null,
    projects: [],
    tasks: []
};

// Utility Functions
function getAuthHeader() {
    if (!state.credentials) return null;
    return 'Basic ' + btoa(state.credentials.username + ':' + state.credentials.password);
}

async function apiRequest(endpoint, options = {}) {
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };
    
    const authHeader = getAuthHeader();
    if (authHeader) {
        headers['Authorization'] = authHeader;
    }
    
    const response = await fetch(API_BASE + endpoint, {
        ...options,
        headers
    });
    
    if (response.status === 401) {
        logout();
        throw new Error('Unauthorized');
    }
    
    if (!response.ok) {
        const error = await response.json().catch(() => ({ error: 'Request failed' }));
        throw new Error(error.error || 'Request failed');
    }
    
    return response.json();
}

function showScreen(screenId) {
    document.querySelectorAll('.screen').forEach(screen => {
        screen.classList.add('hidden');
    });
    document.getElementById(screenId).classList.remove('hidden');
}

function showError(elementId, message) {
    const errorElement = document.getElementById(elementId);
    errorElement.textContent = message;
    errorElement.classList.remove('hidden');
}

function hideError(elementId) {
    document.getElementById(elementId).classList.add('hidden');
}

// Authentication
async function login(username, password) {
    try {
        state.credentials = { username, password };
        const user = await apiRequest('/auth/me');
        state.currentUser = user;
        localStorage.setItem('credentials', JSON.stringify(state.credentials));
        await loadProjects();
        showScreen('board-screen');
        hideError('auth-error');
    } catch (error) {
        state.credentials = null;
        showError('auth-error', error.message);
    }
}

async function register(username, password) {
    try {
        await apiRequest('/auth/register', {
            method: 'POST',
            body: JSON.stringify({
                username: username,
                password: password,
                type: 'human'
            })
        });
        await login(username, password);
    } catch (error) {
        showError('auth-error', error.message);
    }
}

function logout() {
    state.credentials = null;
    state.currentUser = null;
    state.currentProject = null;
    state.projects = [];
    state.tasks = [];
    localStorage.removeItem('credentials');
    showScreen('auth-screen');
}

// Projects
async function loadProjects() {
    try {
        const projects = await apiRequest('/projects');
        state.projects = projects;
        
        const selector = document.getElementById('project-selector');
        selector.innerHTML = '<option value="">Select a project...</option>';
        
        projects.forEach(project => {
            const option = document.createElement('option');
            option.value = project.id;
            option.textContent = project.name;
            selector.appendChild(option);
        });
        
        // Auto-select first project if available
        if (projects.length > 0) {
            selector.value = projects[0].id;
            await loadTasks(projects[0].id);
        }
    } catch (error) {
        console.error('Failed to load projects:', error);
    }
}

async function loadTasks(projectId) {
    if (!projectId) {
        state.currentProject = null;
        state.tasks = [];
        renderBoard();
        return;
    }
    
    try {
        state.currentProject = projectId;
        const tasks = await apiRequest(`/projects/${projectId}/tasks`);
        state.tasks = tasks;
        renderBoard();
    } catch (error) {
        console.error('Failed to load tasks:', error);
        state.tasks = [];
        renderBoard();
    }
}

// Board Rendering
function renderBoard() {
    const statuses = ['todo', 'in_progress', 'blocked', 'completed'];
    
    statuses.forEach(status => {
        const taskList = document.getElementById(`${status}-tasks`);
        taskList.innerHTML = '';
        
        const tasksInStatus = state.tasks.filter(task => task.status === status);
        
        // Update count
        const column = document.querySelector(`.kanban-column[data-status="${status}"]`);
        const countElement = column.querySelector('.task-count');
        countElement.textContent = tasksInStatus.length;
        
        tasksInStatus.forEach(task => {
            const card = createTaskCard(task);
            taskList.appendChild(card);
        });
    });
}

function createTaskCard(task) {
    const card = document.createElement('div');
    card.className = 'task-card';
    card.onclick = () => openTaskModal(task);
    
    const title = document.createElement('div');
    title.className = 'task-card-title';
    title.textContent = task.title;
    
    const description = document.createElement('div');
    description.className = 'task-card-description';
    description.textContent = task.description || 'No description';
    
    const meta = document.createElement('div');
    meta.className = 'task-card-meta';
    
    const priority = document.createElement('span');
    priority.className = `task-priority ${task.priority || 'medium'}`;
    priority.textContent = task.priority || 'medium';
    
    const date = document.createElement('span');
    date.textContent = new Date(task.created_at).toLocaleDateString();
    
    meta.appendChild(priority);
    meta.appendChild(date);
    
    card.appendChild(title);
    card.appendChild(description);
    card.appendChild(meta);
    
    return card;
}

// Task Modal
let currentTask = null;

function openTaskModal(task) {
    currentTask = task;
    
    document.getElementById('task-modal-title').textContent = task.title;
    document.getElementById('task-description').textContent = task.description || 'No description';
    document.getElementById('task-status-select').value = task.status;
    document.getElementById('task-priority').textContent = task.priority || 'medium';
    document.getElementById('task-created').textContent = new Date(task.created_at).toLocaleString();
    document.getElementById('task-assigned').textContent = task.assigned_to || 'Unassigned';
    
    document.getElementById('task-modal').classList.remove('hidden');
}

function closeTaskModal() {
    currentTask = null;
    document.getElementById('task-modal').classList.add('hidden');
}

async function saveTask() {
    if (!currentTask) return;
    
    const newStatus = document.getElementById('task-status-select').value;
    
    if (newStatus !== currentTask.status) {
        try {
            await apiRequest(`/tasks/${currentTask.id}`, {
                method: 'PUT',
                body: JSON.stringify({
                    title: currentTask.title,
                    description: currentTask.description,
                    status: newStatus,
                    priority: currentTask.priority,
                    project_id: currentTask.project_id
                })
            });
            
            currentTask.status = newStatus;
            await loadTasks(state.currentProject);
            closeTaskModal();
        } catch (error) {
            console.error('Failed to update task:', error);
            alert('Failed to update task: ' + error.message);
        }
    } else {
        closeTaskModal();
    }
}

// Event Listeners
document.addEventListener('DOMContentLoaded', () => {
    // Set focus on login username by default
    document.getElementById('login-username').focus();
    
    // Auth Tab Switching
    document.getElementById('login-tab').addEventListener('click', () => {
        document.getElementById('login-tab').classList.add('active');
        document.getElementById('register-tab').classList.remove('active');
        document.getElementById('login-form').classList.remove('hidden');
        document.getElementById('register-form').classList.add('hidden');
        hideError('auth-error');
        // Focus on login username when switching to login tab
        document.getElementById('login-username').focus();
    });
    
    document.getElementById('register-tab').addEventListener('click', () => {
        document.getElementById('register-tab').classList.add('active');
        document.getElementById('login-tab').classList.remove('active');
        document.getElementById('register-form').classList.remove('hidden');
        document.getElementById('login-form').classList.add('hidden');
        hideError('auth-error');
        // Focus on register username when switching to register tab
        document.getElementById('register-username').focus();
    });
    
    // Login
    document.getElementById('login-button').addEventListener('click', async () => {
        const username = document.getElementById('login-username').value.trim();
        const password = document.getElementById('login-password').value;
        
        if (!username || !password) {
            showError('auth-error', 'Please enter username and password');
            return;
        }
        
        await login(username, password);
    });
    
    // Login on Enter key
    document.getElementById('login-password').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            document.getElementById('login-button').click();
        }
    });
    
    // Register
    document.getElementById('register-button').addEventListener('click', async () => {
        const username = document.getElementById('register-username').value.trim();
        const password = document.getElementById('register-password').value;
        const confirmPassword = document.getElementById('register-password-confirm').value;
        
        if (!username || !password) {
            showError('auth-error', 'Please enter username and password');
            return;
        }
        
        if (password !== confirmPassword) {
            showError('auth-error', 'Passwords do not match');
            return;
        }
        
        if (password.length < 8) {
            showError('auth-error', 'Password must be at least 8 characters');
            return;
        }
        
        await register(username, password);
    });
    
    // Register on Enter key
    document.getElementById('register-password-confirm').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            document.getElementById('register-button').click();
        }
    });
    
    // Logout
    document.getElementById('logout-button').addEventListener('click', logout);
    
    // Project Selection
    document.getElementById('project-selector').addEventListener('change', (e) => {
        loadTasks(e.target.value);
    });
    
    // Task Modal
    document.getElementById('close-modal').addEventListener('click', closeTaskModal);
    document.getElementById('cancel-task-button').addEventListener('click', closeTaskModal);
    document.getElementById('save-task-button').addEventListener('click', saveTask);
    
    // Close modal on background click
    document.getElementById('task-modal').addEventListener('click', (e) => {
        if (e.target.id === 'task-modal') {
            closeTaskModal();
        }
    });
    
    // Navigation Menu
    const menuToggle = document.getElementById('menu-toggle');
    const sideNav = document.getElementById('side-nav');
    const navOverlay = document.getElementById('nav-overlay');
    const navClose = document.getElementById('nav-close');
    
    function openNav() {
        sideNav.classList.add('open');
        navOverlay.classList.add('visible');
    }
    
    function closeNav() {
        sideNav.classList.remove('open');
        navOverlay.classList.remove('visible');
    }
    
    menuToggle.addEventListener('click', openNav);
    navClose.addEventListener('click', closeNav);
    navOverlay.addEventListener('click', closeNav);
    
    // Navigation Links
    document.getElementById('nav-projects').addEventListener('click', (e) => {
        e.preventDefault();
        closeNav();
        // Already on projects screen
    });
    
    document.getElementById('nav-users').addEventListener('click', (e) => {
        e.preventDefault();
        closeNav();
        alert('Users management coming soon!');
    });
    
    document.getElementById('nav-settings').addEventListener('click', (e) => {
        e.preventDefault();
        closeNav();
        alert('Settings coming soon!');
    });
    
    document.getElementById('nav-logout').addEventListener('click', (e) => {
        e.preventDefault();
        closeNav();
        logout();
    });
    
    // Auto-login from localStorage
    const savedCredentials = localStorage.getItem('credentials');
    if (savedCredentials) {
        try {
            const creds = JSON.parse(savedCredentials);
            login(creds.username, creds.password);
        } catch (error) {
            localStorage.removeItem('credentials');
        }
    }
});
