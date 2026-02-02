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

// Overview Visualization
let overviewScene = null;
let overviewCamera = null;
let overviewRenderer = null;
let overviewAnimationId = null;

async function showOverview() {
    showScreen('overview-screen');
    await initOverviewVisualization();
}

async function initOverviewVisualization() {
    const container = document.getElementById('graph-container');
    
    // Clean up previous instance
    if (overviewAnimationId) {
        cancelAnimationFrame(overviewAnimationId);
    }
    if (overviewRenderer) {
        overviewRenderer.dispose();
        container.innerHTML = '';
    }
    
    // Fetch data
    const [projects, workers] = await Promise.all([
        apiRequest('/projects'),
        apiRequest('/workers')
    ]);
    
    // Update system stats
    const statsContainer = document.getElementById('system-stats');
    statsContainer.innerHTML = `
        <div class="stat-item">
            <div class="stat-label">Projects</div>
            <div class="stat-value">${projects.length}</div>
        </div>
        <div class="stat-item">
            <div class="stat-label">Workers</div>
            <div class="stat-value">${workers.length}</div>
        </div>
        <div class="stat-item">
            <div class="stat-label">Active Workers</div>
            <div class="stat-value">${workers.filter(w => w.status === 'running').length}</div>
        </div>
    `;
    
    // Initialize Three.js scene
    const width = container.clientWidth;
    const height = container.clientHeight || 600;
    
    overviewScene = new THREE.Scene();
    overviewScene.background = new THREE.Color(0xf8fafc);
    
    overviewCamera = new THREE.OrthographicCamera(
        width / -2, width / 2,
        height / 2, height / -2,
        1, 1000
    );
    overviewCamera.position.z = 500;
    
    overviewRenderer = new THREE.WebGLRenderer({ antialias: true });
    overviewRenderer.setSize(width, height);
    container.appendChild(overviewRenderer.domElement);
    
    // Create graph nodes
    const nodes = [];
    const nodeGeometry = new THREE.CircleGeometry(30, 32);
    
    // Central server node
    const serverMaterial = new THREE.MeshBasicMaterial({ color: 0x2563eb });
    const serverNode = new THREE.Mesh(nodeGeometry, serverMaterial);
    serverNode.position.set(0, 0, 0);
    overviewScene.add(serverNode);
    nodes.push({ mesh: serverNode, label: 'Server', type: 'server' });
    
    // Worker nodes in a circle
    const workerRadius = 200;
    workers.forEach((worker, i) => {
        const angle = (i / workers.length) * Math.PI * 2;
        const x = Math.cos(angle) * workerRadius;
        const y = Math.sin(angle) * workerRadius;
        
        const color = worker.status === 'running' ? 0x10b981 : 0x64748b;
        const workerMaterial = new THREE.MeshBasicMaterial({ color });
        const workerNode = new THREE.Mesh(nodeGeometry, workerMaterial);
        workerNode.position.set(x, y, 0);
        overviewScene.add(workerNode);
        nodes.push({ mesh: workerNode, label: worker.name, type: 'worker' });
        
        // Connection line to server
        const points = [
            new THREE.Vector3(0, 0, 0),
            new THREE.Vector3(x, y, 0)
        ];
        const lineGeometry = new THREE.BufferGeometry().setFromPoints(points);
        const lineMaterial = new THREE.LineBasicMaterial({ 
            color: worker.status === 'running' ? 0x10b981 : 0xe2e8f0,
            linewidth: 2
        });
        const line = new THREE.Line(lineGeometry, lineMaterial);
        overviewScene.add(line);
    });
    
    // Project nodes around workers
    const projectRadius = 350;
    projects.forEach((project, i) => {
        const angle = (i / projects.length) * Math.PI * 2;
        const x = Math.cos(angle) * projectRadius;
        const y = Math.sin(angle) * projectRadius;
        
        const projectMaterial = new THREE.MeshBasicMaterial({ color: 0xf59e0b });
        const projectNode = new THREE.Mesh(
            new THREE.CircleGeometry(20, 32),
            projectMaterial
        );
        projectNode.position.set(x, y, 0);
        overviewScene.add(projectNode);
        nodes.push({ mesh: projectNode, label: project.name, type: 'project' });
    });
    
    // Add labels
    nodes.forEach(node => {
        const canvas = document.createElement('canvas');
        const context = canvas.getContext('2d');
        canvas.width = 256;
        canvas.height = 64;
        
        context.fillStyle = '#ffffff';
        context.fillRect(0, 0, canvas.width, canvas.height);
        context.fillStyle = '#1e293b';
        context.font = 'bold 20px Arial';
        context.textAlign = 'center';
        context.textBaseline = 'middle';
        context.fillText(node.label, canvas.width / 2, canvas.height / 2);
        
        const texture = new THREE.CanvasTexture(canvas);
        const spriteMaterial = new THREE.SpriteMaterial({ map: texture });
        const sprite = new THREE.Sprite(spriteMaterial);
        sprite.scale.set(128, 32, 1);
        sprite.position.set(
            node.mesh.position.x,
            node.mesh.position.y - 50,
            0
        );
        overviewScene.add(sprite);
    });
    
    // Animation loop
    function animate() {
        overviewAnimationId = requestAnimationFrame(animate);
        
        // Gentle rotation of worker nodes
        nodes.forEach((node, i) => {
            if (node.type === 'worker') {
                const time = Date.now() * 0.0001;
                node.mesh.position.x = Math.cos(time + i) * workerRadius;
                node.mesh.position.y = Math.sin(time + i) * workerRadius;
            }
        });
        
        overviewRenderer.render(overviewScene, overviewCamera);
    }
    
    animate();
    
    // Handle window resize
    window.addEventListener('resize', () => {
        const newWidth = container.clientWidth;
        const newHeight = container.clientHeight || 600;
        
        overviewCamera.left = newWidth / -2;
        overviewCamera.right = newWidth / 2;
        overviewCamera.top = newHeight / 2;
        overviewCamera.bottom = newHeight / -2;
        overviewCamera.updateProjectionMatrix();
        
        overviewRenderer.setSize(newWidth, newHeight);
    });
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
    
    document.getElementById('nav-overview').addEventListener('click', (e) => {
        e.preventDefault();
        closeNav();
        showOverview();
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
    
    // Overview Screen
    const overviewMenuToggle = document.getElementById('overview-menu-toggle');
    const overviewLogoutButton = document.getElementById('overview-logout-button');
    
    if (overviewMenuToggle) {
        overviewMenuToggle.addEventListener('click', openNav);
    }
    
    if (overviewLogoutButton) {
        overviewLogoutButton.addEventListener('click', logout);
    }
    
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
