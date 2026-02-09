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

// Custom Dialog System
const Dialog = {
    show(options) {
        return new Promise((resolve) => {
            const overlay = document.getElementById('custom-dialog');
            const icon = document.getElementById('dialog-icon');
            const title = document.getElementById('dialog-title');
            const message = document.getElementById('dialog-message');
            const input = document.getElementById('dialog-input');
            const buttons = document.getElementById('dialog-buttons');

            // Set icon
            icon.className = 'dialog-icon ' + (options.type || 'info');
            const icons = {
                success: '✓',
                error: '✕',
                warning: '⚠',
                info: 'ℹ',
                question: '?'
            };
            icon.textContent = icons[options.type] || icons.info;

            // Set content
            title.textContent = options.title || '';
            message.textContent = options.message || '';

            // Handle input
            if (options.input) {
                input.classList.remove('hidden');
                input.value = options.inputValue || '';
                input.placeholder = options.inputPlaceholder || '';
                setTimeout(() => input.focus(), 100);
            } else {
                input.classList.add('hidden');
            }

            // Create buttons
            buttons.innerHTML = '';
            const buttonConfigs = options.buttons || [{ text: 'OK', primary: true }];
            
            buttonConfigs.forEach((btn, index) => {
                const button = document.createElement('button');
                button.textContent = btn.text;
                button.className = btn.danger ? 'dialog-btn-danger' : 
                                   btn.primary ? 'dialog-btn-primary' : 'dialog-btn-secondary';
                button.onclick = () => {
                    overlay.classList.add('hidden');
                    if (options.input) {
                        resolve(btn.primary || btn.danger ? input.value : null);
                    } else {
                        resolve(btn.primary || btn.danger ? true : false);
                    }
                };
                buttons.appendChild(button);
            });

            // Handle Enter key for input
            if (options.input) {
                input.onkeypress = (e) => {
                    if (e.key === 'Enter') {
                        overlay.classList.add('hidden');
                        resolve(input.value);
                    }
                };
            }

            // Show dialog
            overlay.classList.remove('hidden');
        });
    },

    alert(title, message, type = 'info') {
        return this.show({
            type,
            title,
            message,
            buttons: [{ text: 'OK', primary: true }]
        });
    },

    success(title, message) {
        return this.alert(title, message, 'success');
    },

    error(title, message) {
        return this.alert(title, message, 'error');
    },

    confirm(title, message, options = {}) {
        return this.show({
            type: options.type || 'question',
            title,
            message,
            buttons: [
                { text: options.cancelText || 'Cancel', primary: false },
                { text: options.confirmText || 'Confirm', primary: !options.danger, danger: options.danger }
            ]
        });
    },

    prompt(title, message, defaultValue = '') {
        return this.show({
            type: 'info',
            title,
            message,
            input: true,
            inputValue: defaultValue,
            buttons: [
                { text: 'Cancel', primary: false },
                { text: 'OK', primary: true }
            ]
        });
    }
};

// Utility Functions
function getAuthHeader() {
    // Check for JWT token first (from passkey or token-based login)
    const token = localStorage.getItem('authToken');
    if (token) {
        return 'Bearer ' + token;
    }
    
    // Fall back to Basic auth
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
    
    if (response.status === 401 && !options.skipAutoLogout) {
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
    
    // Save current screen to localStorage (except auth screen)
    if (screenId !== 'auth-screen') {
        localStorage.setItem('lastScreen', screenId);
    }
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
        
        // Show/hide config link based on admin status
        const configNav = document.getElementById('nav-config');
        if (configNav) {
            if (user.username === 'admin') {
                configNav.classList.remove('hidden');
            } else {
                configNav.classList.add('hidden');
            }
        }
        
        await loadProjects();
        
        // Restore last viewed screen or default to board-screen
        const lastScreen = localStorage.getItem('lastScreen');
        if (lastScreen && document.getElementById(lastScreen)) {
            showScreen(lastScreen);
            
            // If it was overview screen, reinitialize visualization
            if (lastScreen === 'overview-screen') {
                await initOverviewVisualization();
            }
        } else {
            showScreen('board-screen');
        }
        
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
    localStorage.removeItem('lastScreen');
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

// Passkey Authentication
function bufferToBase64url(buffer) {
    const bytes = new Uint8Array(buffer);
    let str = '';
    for (const byte of bytes) {
        str += String.fromCharCode(byte);
    }
    return btoa(str).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

function base64urlToBuffer(base64url) {
    const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/');
    const binary = atob(base64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
        bytes[i] = binary.charCodeAt(i);
    }
    return bytes.buffer;
}

async function registerPasskey(deviceName = null) {
    try {
        // Check WebAuthn support
        if (!window.PublicKeyCredential) {
            throw new Error('Passkeys are not supported on this device');
        }

        // Begin registration
        const beginResponse = await apiRequest('/auth/passkey/register/begin', {
            method: 'POST',
            body: JSON.stringify({
                username: state.currentUser.username
            })
        });

        // Convert challenge and user ID from base64url
        const publicKey = beginResponse.options.publicKey;
        publicKey.challenge = base64urlToBuffer(publicKey.challenge);
        publicKey.user.id = base64urlToBuffer(publicKey.user.id);

        // Create credential
        const credential = await navigator.credentials.create({ publicKey });

        // Finish registration
        await apiRequest('/auth/passkey/register/finish', {
            method: 'POST',
            body: JSON.stringify({
                session_id: beginResponse.session_id,
                credential: {
                    id: credential.id,
                    rawId: bufferToBase64url(credential.rawId),
                    type: credential.type,
                    response: {
                        clientDataJSON: bufferToBase64url(credential.response.clientDataJSON),
                        attestationObject: bufferToBase64url(credential.response.attestationObject)
                    }
                },
                device_name: deviceName || `Device ${new Date().toLocaleDateString()}`
            })
        });

        return true;
    } catch (error) {
        console.error('Passkey registration error:', error);
        throw error;
    }
}

async function authenticateWithPasskey(username = null) {
    try {
        // Check WebAuthn support
        if (!window.PublicKeyCredential) {
            throw new Error('Passkeys are not supported on this device');
        }

        // Begin authentication
        const beginResponse = await apiRequest('/auth/passkey/authenticate/begin', {
            method: 'POST',
            body: JSON.stringify({ username }),
            skipAutoLogout: true
        });

        // Convert challenge from base64url
        const publicKey = beginResponse.options.publicKey;
        publicKey.challenge = base64urlToBuffer(publicKey.challenge);
        if (publicKey.allowCredentials) {
            publicKey.allowCredentials = publicKey.allowCredentials.map(cred => ({
                ...cred,
                id: base64urlToBuffer(cred.id)
            }));
        }

        // Get credential
        const credential = await navigator.credentials.get({ publicKey });

        // Finish authentication
        const loginResponse = await apiRequest('/auth/passkey/authenticate/finish', {
            method: 'POST',
            body: JSON.stringify({
                session_id: beginResponse.session_id,
                credential: {
                    id: credential.id,
                    rawId: bufferToBase64url(credential.rawId),
                    type: credential.type,
                    response: {
                        clientDataJSON: bufferToBase64url(credential.response.clientDataJSON),
                        authenticatorData: bufferToBase64url(credential.response.authenticatorData),
                        signature: bufferToBase64url(credential.response.signature),
                        userHandle: credential.response.userHandle ? bufferToBase64url(credential.response.userHandle) : null
                    }
                }
            }),
            skipAutoLogout: true
        });

        // Store token and user info
        state.currentUser = loginResponse.user;
        localStorage.setItem('authToken', loginResponse.token);
        console.log('Passkey login successful, token stored:', loginResponse.token.substring(0, 20) + '...');
        await loadProjects();
        showScreen('board-screen');
        return true;
    } catch (error) {
        console.error('Passkey authentication error:', error);
        throw error;
    }
}

async function loadPasskeys() {
    try {
        const passkeys = await apiRequest('/auth/passkey/list');
        const container = document.getElementById('passkey-list');
        
        if (!passkeys || passkeys.length === 0) {
            container.innerHTML = '<div class="passkey-empty">No passkeys registered</div>';
            return;
        }

        container.innerHTML = passkeys.map(pk => `
            <div class="passkey-item">
                <div class="passkey-info">
                    <div class="passkey-name">${pk.device_name || 'Unnamed Device'}</div>
                    <div class="passkey-meta">
                        Created: ${new Date(pk.created_at).toLocaleDateString()}
                        ${pk.last_used_at ? `• Last used: ${new Date(pk.last_used_at).toLocaleDateString()}` : ''}
                    </div>
                </div>
                <div class="passkey-actions">
                    <button onclick="deletePasskey('${pk.id}')">Delete</button>
                </div>
            </div>
        `).join('');
    } catch (error) {
        showError('passkey-error', error.message);
    }
}

async function deletePasskey(passkeyId) {
    const confirmed = await Dialog.confirm(
        'Delete Passkey',
        'Are you sure you want to delete this passkey? You will no longer be able to sign in with it.',
        { danger: true, confirmText: 'Delete' }
    );
    
    if (!confirmed) {
        return;
    }

    try {
        await apiRequest(`/auth/passkey?id=${passkeyId}`, {
            method: 'DELETE'
        });
        await loadPasskeys();
    } catch (error) {
        showError('passkey-error', error.message);
    }
}

function showSettings() {
    showScreen('settings-screen');
    loadPasskeys();
}

// Config Management (Admin Only)
async function showConfig() {
    if (!state.currentUser || state.currentUser.username !== 'admin') {
        await Dialog.error('Access Denied', 'Only administrators can access configuration.');
        return;
    }
    showScreen('config-screen');
    await loadConfigList();
}

async function loadConfigList() {
    const tbody = document.getElementById('config-table-body');
    hideError('config-error');
    
    try {
        const configs = await apiRequest('/config');
        
        if (!configs || configs.length === 0) {
            tbody.innerHTML = '<tr><td colspan="4" class="loading-row">No configuration entries found.</td></tr>';
            return;
        }
        
        tbody.innerHTML = '';
        configs.forEach(cfg => {
            const row = document.createElement('tr');
            
            const keyCell = document.createElement('td');
            keyCell.innerHTML = `<div class="config-key">${escapeHtml(cfg.key)}</div>`;
            row.appendChild(keyCell);
            
            const valueCell = document.createElement('td');
            valueCell.innerHTML = `<div class="config-value">${escapeHtml(cfg.value)}</div>`;
            row.appendChild(valueCell);
            
            const descCell = document.createElement('td');
            descCell.innerHTML = `<div class="config-description">${cfg.description ? escapeHtml(cfg.description) : '<em>No description</em>'}</div>`;
            row.appendChild(descCell);
            
            const actionsCell = document.createElement('td');
            const actionsDiv = document.createElement('div');
            actionsDiv.className = 'config-actions';
            
            const editBtn = document.createElement('button');
            editBtn.className = 'edit-config-btn';
            editBtn.textContent = 'Edit';
            editBtn.addEventListener('click', () => editConfig(cfg.key));
            
            const deleteBtn = document.createElement('button');
            deleteBtn.className = 'delete-config-btn';
            deleteBtn.textContent = 'Delete';
            deleteBtn.addEventListener('click', () => deleteConfig(cfg.key));
            
            actionsDiv.appendChild(editBtn);
            actionsDiv.appendChild(deleteBtn);
            actionsCell.appendChild(actionsDiv);
            row.appendChild(actionsCell);
            
            tbody.appendChild(row);
        });
    } catch (error) {
        showError('config-error', error.message);
        tbody.innerHTML = '<tr><td colspan="4" class="loading-row">Failed to load configuration.</td></tr>';
    }
}

async function addConfig() {
    const key = await Dialog.prompt('New Config Key', 'Enter the configuration key:');
    if (!key || !key.trim()) return;
    
    const value = await Dialog.prompt('Config Value', 'Enter the value:');
    if (!value || !value.trim()) return;
    
    const description = await Dialog.prompt('Description (Optional)', 'Enter a description:');
    
    try {
        await apiRequest(`/config/${encodeURIComponent(key)}`, {
            method: 'PUT',
            body: JSON.stringify({
                value: value,
                description: description && description.trim() ? description : null
            })
        });
        await Dialog.success('Config Added', `Configuration "${key}" has been created.`);
        await loadConfigList();
    } catch (error) {
        await Dialog.error('Failed to Add Config', error.message);
    }
}

async function editConfig(key) {
    try {
        const cfg = await apiRequest(`/config/${encodeURIComponent(key)}`);
        
        const newValue = await Dialog.prompt('Edit Value', `Enter new value for "${key}":`, cfg.value);
        if (newValue === null) return;
        
        const newDescription = await Dialog.prompt('Edit Description', 'Enter new description (optional):', cfg.description || '');
        
        await apiRequest(`/config/${encodeURIComponent(key)}`, {
            method: 'PUT',
            body: JSON.stringify({
                value: newValue,
                description: newDescription && newDescription.trim() ? newDescription : null
            })
        });
        await Dialog.success('Config Updated', `Configuration "${key}" has been updated.`);
        await loadConfigList();
    } catch (error) {
        await Dialog.error('Failed to Update Config', error.message);
    }
}

async function deleteConfig(key) {
    const confirmed = await Dialog.confirm(
        'Delete Configuration',
        `Are you sure you want to delete the configuration key "${key}"?`,
        'warning'
    );
    
    if (!confirmed) return;
    
    try {
        await apiRequest(`/config/${encodeURIComponent(key)}`, {
            method: 'DELETE'
        });
        await Dialog.success('Config Deleted', `Configuration "${key}" has been deleted.`);
        await loadConfigList();
    } catch (error) {
        await Dialog.error('Failed to Delete Config', error.message);
    }
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
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
        const tasks = await apiRequest(`/tasks?project_id=${projectId}`);
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
    // Ensure tasks is an array
    if (!state.tasks) {
        state.tasks = [];
    }
    
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
            Dialog.error('Update Failed', 'Failed to update task: ' + error.message);
        }
    } else {
        closeTaskModal();
    }
}

// Session Management
async function initializeSession() {
    const token = localStorage.getItem('authToken');
    if (!token) {
        return; // No token, stay on login screen
    }
    
    try {
        // Try to load user info by fetching projects (validates token)
        const projects = await apiRequest('/projects');
        
        // Token is valid, fetch current user info
        const users = await apiRequest('/users');
        if (users && users.length > 0) {
            // Find current user by decoding the JWT token
            const payload = JSON.parse(atob(token.split('.')[1]));
            state.currentUser = users.find(u => u.id === payload.user_id);
            
            if (state.currentUser) {
                state.projects = projects;
                showScreen('board-screen');
                await loadProjects();
            }
        }
    } catch (error) {
        // Token is invalid or expired, clear it
        console.log('Session restore failed:', error.message);
        localStorage.removeItem('authToken');
        state.currentUser = null;
    }
}

// Event Listeners
document.addEventListener('DOMContentLoaded', () => {
    // Try to restore session from JWT token
    initializeSession();
    
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
    
    // Passkey Login
    document.getElementById('passkey-login-button').addEventListener('click', async () => {
        try {
            const username = document.getElementById('login-username').value.trim();
            await authenticateWithPasskey(username || null);
            hideError('auth-error');
        } catch (error) {
            showError('auth-error', error.message);
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
        Dialog.alert('Coming Soon', 'Users management will be available in a future update.', 'info');
    });
    
    document.getElementById('nav-config').addEventListener('click', (e) => {
        e.preventDefault();
        closeNav();
        showConfig();
    });
    
    document.getElementById('nav-settings').addEventListener('click', (e) => {
        e.preventDefault();
        closeNav();
        showSettings();
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
    
    // Settings Screen
    const settingsMenuToggle = document.getElementById('settings-menu-toggle');
    const settingsLogoutButton = document.getElementById('settings-logout-button');
    const registerPasskeyButton = document.getElementById('register-passkey-button');
    
    if (settingsMenuToggle) {
        settingsMenuToggle.addEventListener('click', openNav);
    }
    
    if (settingsLogoutButton) {
        settingsLogoutButton.addEventListener('click', logout);
    }
    
    if (registerPasskeyButton) {
        registerPasskeyButton.addEventListener('click', async () => {
            const deviceName = await Dialog.prompt(
                'Name Your Passkey',
                'Enter a name to identify this passkey (optional):'
            );
            try {
                await registerPasskey(deviceName || '');
                await Dialog.success('Passkey Added', 'Your passkey has been registered successfully. You can now use it to sign in.');
                await loadPasskeys();
                hideError('passkey-error');
            } catch (error) {
                showError('passkey-error', error.message);
            }
        });
    }
    
    // Config Screen
    const configMenuToggle = document.getElementById('config-menu-toggle');
    const configLogoutButton = document.getElementById('config-logout-button');
    const addConfigButton = document.getElementById('add-config-button');
    
    if (configMenuToggle) {
        configMenuToggle.addEventListener('click', openNav);
    }
    
    if (configLogoutButton) {
        configLogoutButton.addEventListener('click', logout);
    }
    
    if (addConfigButton) {
        addConfigButton.addEventListener('click', addConfig);
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
