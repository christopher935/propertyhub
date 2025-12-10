/**
 * PropertyHub Admin System
 * Complete administrative interface for system management
 * User management, system monitoring, configuration, and oversight
 */

class PropertyHubAdmin {
    constructor() {
        this.currentUser = null;
        this.systemStatus = {};
        this.activeConnections = new Map();
        this.auditLog = [];
        this.pendingApprovals = [];
        this.systemAlerts = [];
        
        // Admin permissions
        this.permissions = {
            userManagement: false,
            systemConfig: false,
            dataExport: false,
            auditAccess: false,
            systemMonitoring: false,
            propertyOverride: false,
            financialAccess: false,
            integrationManagement: false
        };

        this.init();
    }

    async init() {
        console.log('⚙️ PropertyHub Admin System initializing...');
        
        // Verify admin access
        if (!await this.verifyAdminAccess()) {
            console.error('❌ Admin access denied');
            window.location.href = '/dashboard';
            return;
        }

        await this.loadAdminData();
        this.setupAdminInterface();
        this.startSystemMonitoring();
        this.setupEventListeners();
        
        console.log('✅ PropertyHub Admin System ready');
    }

    // Authentication & Authorization
    async verifyAdminAccess() {
        try {
            if (!window.PropertyHubAuth || !window.PropertyHubAuth.isAuthenticated()) {
                return false;
            }

            const response = await fetch('/api/admin/verify', {
                headers: {
                    'Authorization': `Bearer ${window.PropertyHubAuth.token}`
                }
            });

            if (response.ok) {
                const data = await response.json();
                this.currentUser = data.user;
                this.permissions = data.permissions;
                return true;
            }
            
            return false;
        } catch (error) {
            console.error('Admin verification failed:', error);
            return false;
        }
    }

    // User Management
    async loadUserManagement() {
        if (!this.permissions.userManagement) {
            this.showAccessDenied('User Management');
            return;
        }

        try {
            const response = await fetch('/api/admin/users', {
                headers: { 'Authorization': `Bearer ${window.PropertyHubAuth.token}` }
            });
            
            const data = await response.json();
            this.renderUserManagement(data);
        } catch (error) {
            this.showError('Failed to load user management data', error);
        }
    }

    renderUserManagement(data) {
        const container = document.getElementById('admin-content');
        container.innerHTML = `
            <div class="admin-section">
                <div class="section-header">
                    <h2>User Management</h2>
                    <div class="section-actions">
                        <button class="btn btn-primary" onclick="PropertyHubAdmin.showCreateUserModal()">
                            <i class="fas fa-user-plus"></i> Add User
                        </button>
                        <button class="btn btn-secondary" onclick="PropertyHubAdmin.exportUserData()">
                            <i class="fas fa-download"></i> Export Users
                        </button>
                    </div>
                </div>

                <div class="user-stats">
                    <div class="stat-card">
                        <h3>${data.totalUsers}</h3>
                        <p>Total Users</p>
                    </div>
                    <div class="stat-card">
                        <h3>${data.activeUsers}</h3>
                        <p>Active Users</p>
                    </div>
                    <div class="stat-card">
                        <h3>${data.adminUsers}</h3>
                        <p>Admin Users</p>
                    </div>
                    <div class="stat-card">
                        <h3>${data.pendingUsers}</h3>
                        <p>Pending Approval</p>
                    </div>
                </div>

                <div class="user-filters">
                    <input type="text" id="user-search" placeholder="Search users..." class="form-control">
                    <select id="role-filter" class="form-control">
                        <option value="">All Roles</option>
                        <option value="admin">Admin</option>
                        <option value="manager">Manager</option>
                        <option value="agent">Agent</option>
                        <option value="user">User</option>
                    </select>
                    <select id="status-filter" class="form-control">
                        <option value="">All Status</option>
                        <option value="active">Active</option>
                        <option value="inactive">Inactive</option>
                        <option value="suspended">Suspended</option>
                        <option value="pending">Pending</option>
                    </select>
                </div>

                <div class="user-table-container">
                    <table class="admin-table" id="users-table">
                        <thead>
                            <tr>
                                <th>User</th>
                                <th>Email</th>
                                <th>Role</th>
                                <th>Status</th>
                                <th>Last Login</th>
                                <th>Properties</th>
                                <th>Actions</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${data.users.map(user => this.renderUserRow(user)).join('')}
                        </tbody>
                    </table>
                </div>
            </div>
        `;

        this.setupUserManagementEvents();
    }

    renderUserRow(user) {
        const safe = Sanitizer.sanitizeObject(user);
        return `
            <tr data-user-id="${safe.id}">
                <td>
                    <div class="user-info">
                        <img src="${safe.avatar || '/static/images/default-avatar.png'}" alt="${safe.name}" class="user-avatar">
                        <div>
                            <strong>${safe.name}</strong>
                            <br><small>ID: ${safe.id}</small>
                        </div>
                    </div>
                </td>
                <td>${safe.email}</td>
                <td>
                    <span class="role-badge ${safe.role}">${safe.role}</span>
                </td>
                <td>
                    <span class="status-indicator ${safe.status}">${safe.status}</span>
                </td>
                <td>${user.lastLogin ? new Date(user.lastLogin).toLocaleString() : 'Never'}</td>
                <td>${user.propertyCount || 0}</td>
                <td>
                    <div class="action-buttons">
                        <button class="btn btn-sm btn-outline" onclick="PropertyHubAdmin.viewUser('${safe.id}')">
                            <i class="fas fa-eye"></i>
                        </button>
                        <button class="btn btn-sm btn-outline" onclick="PropertyHubAdmin.editUser('${safe.id}')">
                            <i class="fas fa-edit"></i>
                        </button>
                        <button class="btn btn-sm btn-outline" onclick="PropertyHubAdmin.toggleUserStatus('${safe.id}', '${safe.status}')">
                            <i class="fas fa-${user.status === 'active' ? 'pause' : 'play'}"></i>
                        </button>
                        <button class="btn btn-sm btn-danger" onclick="PropertyHubAdmin.deleteUser('${safe.id}')">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </td>
            </tr>
        `;
    }

    // System Monitoring
    async loadSystemMonitoring() {
        if (!this.permissions.systemMonitoring) {
            this.showAccessDenied('System Monitoring');
            return;
        }

        try {
            const response = await fetch('/api/admin/system/status', {
                headers: { 'Authorization': `Bearer ${window.PropertyHubAuth.token}` }
            });
            
            const data = await response.json();
            this.systemStatus = data;
            this.renderSystemMonitoring(data);
        } catch (error) {
            this.showError('Failed to load system monitoring data', error);
        }
    }

    renderSystemMonitoring(data) {
        const container = document.getElementById('admin-content');
        container.innerHTML = `
            <div class="admin-section">
                <div class="section-header">
                    <h2>System Monitoring</h2>
                    <div class="section-actions">
                        <button class="btn btn-primary" onclick="PropertyHubAdmin.refreshSystemStatus()">
                            <i class="fas fa-sync"></i> Refresh
                        </button>
                        <button class="btn btn-secondary" onclick="PropertyHubAdmin.downloadSystemReport()">
                            <i class="fas fa-file-download"></i> System Report
                        </button>
                    </div>
                </div>

                <div class="system-overview">
                    <div class="system-health ${data.overallHealth}">
                        <h3>System Health: ${data.overallHealth.toUpperCase()}</h3>
                        <p>Last updated: ${new Date(data.lastUpdated).toLocaleString()}</p>
                    </div>
                </div>

                <div class="monitoring-grid">
                    <div class="monitor-card">
                        <h3>Server Performance</h3>
                        <div class="metric">
                            <label>CPU Usage:</label>
                            <div class="progress-bar">
                                <div class="progress-fill" style="width: ${data.server.cpuUsage}%"></div>
                            </div>
                            <span>${data.server.cpuUsage}%</span>
                        </div>
                        <div class="metric">
                            <label>Memory Usage:</label>
                            <div class="progress-bar">
                                <div class="progress-fill" style="width: ${data.server.memoryUsage}%"></div>
                            </div>
                            <span>${data.server.memoryUsage}%</span>
                        </div>
                        <div class="metric">
                            <label>Disk Usage:</label>
                            <div class="progress-bar">
                                <div class="progress-fill" style="width: ${data.server.diskUsage}%"></div>
                            </div>
                            <span>${data.server.diskUsage}%</span>
                        </div>
                    </div>

                    <div class="monitor-card">
                        <h3>Database Status</h3>
                        <div class="status-item">
                            <label>Connection Status:</label>
                            <span class="status-indicator ${data.database.status}">${data.database.status}</span>
                        </div>
                        <div class="metric">
                            <label>Active Connections:</label>
                            <span>${data.database.activeConnections}</span>
                        </div>
                        <div class="metric">
                            <label>Query Response Time:</label>
                            <span>${data.database.avgResponseTime}ms</span>
                        </div>
                        <div class="metric">
                            <label>Database Size:</label>
                            <span>${this.formatFileSize(data.database.size)}</span>
                        </div>
                    </div>

                    <div class="monitor-card">
                        <h3>Application Metrics</h3>
                        <div class="metric">
                            <label>Active Sessions:</label>
                            <span>${data.app.activeSessions}</span>
                        </div>
                        <div class="metric">
                            <label>Requests/Minute:</label>
                            <span>${data.app.requestsPerMinute}</span>
                        </div>
                        <div class="metric">
                            <label>Error Rate:</label>
                            <span class="${data.app.errorRate > 5 ? 'high-error' : ''}">${data.app.errorRate}%</span>
                        </div>
                        <div class="metric">
                            <label>Uptime:</label>
                            <span>${this.formatUptime(data.app.uptime)}</span>
                        </div>
                    </div>

                    <div class="monitor-card">
                        <h3>External Integrations</h3>
                        ${Object.entries(data.integrations).map(([name, status]) => {
                            const safeName = Sanitizer.escapeHtml(name);
                            const safeStatus = Sanitizer.sanitizeObject(status);
                            return `
                            <div class="status-item">
                                <label>${safeName}:</label>
                                <span class="status-indicator ${safeStatus.status}">${safeStatus.status}</span>
                                <small>${status.lastCheck ? new Date(status.lastCheck).toLocaleString() : 'Never'}</small>
                            </div>
                        `;}).join('')}
                    </div>
                </div>

                <div class="system-logs">
                    <h3>Recent System Logs</h3>
                    <div class="log-container">
                        ${data.recentLogs.map(log => {
                            const safeLog = Sanitizer.sanitizeObject(log);
                            return `
                            <div class="log-entry ${safeLog.level}">
                                <span class="log-time">${new Date(log.timestamp).toLocaleString()}</span>
                                <span class="log-level">${safeLog.level}</span>
                                <span class="log-message">${safeLog.message}</span>
                            </div>
                        `;}).join('')}
                    </div>
                </div>
            </div>
        `;
    }

    // Configuration Management
    async loadConfigurationManagement() {
        if (!this.permissions.systemConfig) {
            this.showAccessDenied('System Configuration');
            return;
        }

        try {
            const response = await fetch('/api/admin/config', {
                headers: { 'Authorization': `Bearer ${window.PropertyHubAuth.token}` }
            });
            
            const data = await response.json();
            this.renderConfigurationManagement(data);
        } catch (error) {
            this.showError('Failed to load configuration data', error);
        }
    }

    renderConfigurationManagement(data) {
        const container = document.getElementById('admin-content');
        container.innerHTML = `
            <div class="admin-section">
                <div class="section-header">
                    <h2>System Configuration</h2>
                    <div class="section-actions">
                        <button class="btn btn-success" onclick="PropertyHubAdmin.saveAllConfigurations()">
                            <i class="fas fa-save"></i> Save All Changes
                        </button>
                        <button class="btn btn-warning" onclick="PropertyHubAdmin.resetToDefaults()">
                            <i class="fas fa-undo"></i> Reset to Defaults
                        </button>
                    </div>
                </div>

                <div class="config-tabs">
                    <button class="tab-button active" onclick="PropertyHubAdmin.showConfigTab('general')">General</button>
                    <button class="tab-button" onclick="PropertyHubAdmin.showConfigTab('integrations')">Integrations</button>
                    <button class="tab-button" onclick="PropertyHubAdmin.showConfigTab('security')">Security</button>
                    <button class="tab-button" onclick="PropertyHubAdmin.showConfigTab('notifications')">Notifications</button>
                    <button class="tab-button" onclick="PropertyHubAdmin.showConfigTab('analytics')">Analytics</button>
                </div>

                <div id="config-general" class="config-panel active">
                    <h3>General Settings</h3>
                    <div class="config-form">
                        <div class="form-group">
                            <label>Application Name</label>
                            <input type="text" class="form-control" value="${data.general.appName}" data-config="general.appName">
                        </div>
                        <div class="form-group">
                            <label>Default Timezone</label>
                            <select class="form-control" data-config="general.timezone">
                                ${this.renderTimezoneOptions(data.general.timezone)}
                            </select>
                        </div>
                        <div class="form-group">
                            <label>Default Currency</label>
                            <select class="form-control" data-config="general.currency">
                                <option value="USD" ${data.general.currency === 'USD' ? 'selected' : ''}>USD</option>
                                <option value="CAD" ${data.general.currency === 'CAD' ? 'selected' : ''}>CAD</option>
                            </select>
                        </div>
                        <div class="form-group">
                            <label>Maximum Upload Size (MB)</label>
                            <input type="number" class="form-control" value="${data.general.maxUploadSize}" data-config="general.maxUploadSize">
                        </div>
                    </div>
                </div>

                <div id="config-integrations" class="config-panel">
                    <h3>Integration Settings</h3>
                    <div class="config-form">
                        <!-- HAR MLS Integration removed - HAR blocked access -->

                        <div class="integration-section">
                            <h4>Follow Up Boss Integration</h4>
                            <div class="form-group">
                                <label>API Endpoint</label>
                                <input type="text" class="form-control" value="${data.integrations.fub.endpoint}" data-config="integrations.fub.endpoint">
                            </div>
                            <div class="form-group">
                                <label>Sync Frequency (minutes)</label>
                                <input type="number" class="form-control" value="${data.integrations.fub.syncFrequency}" data-config="integrations.fub.syncFrequency">
                            </div>
                            <div class="form-group">
                                <label class="checkbox-label">
                                    <input type="checkbox" ${data.integrations.fub.enabled ? 'checked' : ''} data-config="integrations.fub.enabled">
                                    Enable FUB Integration
                                </label>
                            </div>
                        </div>
                    </div>
                </div>

                <div id="config-security" class="config-panel">
                    <h3>Security Settings</h3>
                    <div class="config-form">
                        <div class="form-group">
                            <label>Session Timeout (minutes)</label>
                            <input type="number" class="form-control" value="${data.security.sessionTimeout}" data-config="security.sessionTimeout">
                        </div>
                        <div class="form-group">
                            <label>Max Login Attempts</label>
                            <input type="number" class="form-control" value="${data.security.maxLoginAttempts}" data-config="security.maxLoginAttempts">
                        </div>
                        <div class="form-group">
                            <label>Password Min Length</label>
                            <input type="number" class="form-control" value="${data.security.passwordMinLength}" data-config="security.passwordMinLength">
                        </div>
                        <div class="form-group">
                            <label class="checkbox-label">
                                <input type="checkbox" ${data.security.requireMFA ? 'checked' : ''} data-config="security.requireMFA">
                                Require MFA for Admin Users
                            </label>
                        </div>
                        <div class="form-group">
                            <label class="checkbox-label">
                                <input type="checkbox" ${data.security.auditLogging ? 'checked' : ''} data-config="security.auditLogging">
                                Enable Audit Logging
                            </label>
                        </div>
                    </div>
                </div>

                <div id="config-notifications" class="config-panel">
                    <h3>Notification Settings</h3>
                    <div class="config-form">
                        <div class="form-group">
                            <label>Email From Address</label>
                            <input type="email" class="form-control" value="${data.notifications.fromEmail}" data-config="notifications.fromEmail">
                        </div>
                        <div class="form-group">
                            <label>Email From Name</label>
                            <input type="text" class="form-control" value="${data.notifications.fromName}" data-config="notifications.fromName">
                        </div>
                        <div class="form-group">
                            <label class="checkbox-label">
                                <input type="checkbox" ${data.notifications.bookingAlerts ? 'checked' : ''} data-config="notifications.bookingAlerts">
                                Send Booking Alerts
                            </label>
                        </div>
                        <div class="form-group">
                            <label class="checkbox-label">
                                <input type="checkbox" ${data.notifications.systemAlerts ? 'checked' : ''} data-config="notifications.systemAlerts">
                                Send System Alerts
                            </label>
                        </div>
                    </div>
                </div>

                <div id="config-analytics" class="config-panel">
                    <h3>Analytics Settings</h3>
                    <div class="config-form">
                        <div class="form-group">
                            <label>Data Retention Period (days)</label>
                            <input type="number" class="form-control" value="${data.analytics.retentionDays}" data-config="analytics.retentionDays">
                        </div>
                        <div class="form-group">
                            <label class="checkbox-label">
                                <input type="checkbox" ${data.analytics.trackUserBehavior ? 'checked' : ''} data-config="analytics.trackUserBehavior">
                                Track User Behavior
                            </label>
                        </div>
                        <div class="form-group">
                            <label class="checkbox-label">
                                <input type="checkbox" ${data.analytics.fridayReports ? 'checked' : ''} data-config="analytics.fridayReports">
                                Enable Friday Reports
                            </label>
                        </div>
                    </div>
                </div>
            </div>
        `;

        this.setupConfigurationEvents();
    }

    // Property Management Oversight
    async loadPropertyOversight() {
        if (!this.permissions.propertyOverride) {
            this.showAccessDenied('Property Management');
            return;
        }

        try {
            const response = await fetch('/api/admin/properties/oversight', {
                headers: { 'Authorization': `Bearer ${window.PropertyHubAuth.token}` }
            });
            
            const data = await response.json();
            this.renderPropertyOversight(data);
        } catch (error) {
            this.showError('Failed to load property oversight data', error);
        }
    }

    // Audit & Compliance
    async loadAuditLog() {
        if (!this.permissions.auditAccess) {
            this.showAccessDenied('Audit Log');
            return;
        }

        try {
            const response = await fetch('/api/admin/audit/log', {
                headers: { 'Authorization': `Bearer ${window.PropertyHubAuth.token}` }
            });
            
            const data = await response.json();
            this.renderAuditLog(data);
        } catch (error) {
            this.showError('Failed to load audit log', error);
        }
    }

    renderAuditLog(data) {
        const container = document.getElementById('admin-content');
        container.innerHTML = `
            <div class="admin-section">
                <div class="section-header">
                    <h2>Audit Log</h2>
                    <div class="section-actions">
                        <button class="btn btn-primary" onclick="PropertyHubAdmin.exportAuditLog()">
                            <i class="fas fa-download"></i> Export Log
                        </button>
                        <button class="btn btn-secondary" onclick="PropertyHubAdmin.clearOldAuditEntries()">
                            <i class="fas fa-trash-alt"></i> Clear Old Entries
                        </button>
                    </div>
                </div>

                <div class="audit-filters">
                    <input type="date" id="audit-date-from" class="form-control">
                    <input type="date" id="audit-date-to" class="form-control">
                    <select id="audit-action-filter" class="form-control">
                        <option value="">All Actions</option>
                        <option value="login">Login</option>
                        <option value="logout">Logout</option>
                        <option value="create">Create</option>
                        <option value="update">Update</option>
                        <option value="delete">Delete</option>
                        <option value="config_change">Config Change</option>
                    </select>
                    <select id="audit-user-filter" class="form-control">
                        <option value="">All Users</option>
                        ${data.users.map(user => {
                            const safeUser = Sanitizer.sanitizeObject(user);
                            return `<option value="${safeUser.id}">${safeUser.name}</option>`;
                        }).join('')}
                    </select>
                    <button class="btn btn-primary" onclick="PropertyHubAdmin.filterAuditLog()">Filter</button>
                </div>

                <div class="audit-table-container">
                    <table class="admin-table">
                        <thead>
                            <tr>
                                <th>Timestamp</th>
                                <th>User</th>
                                <th>Action</th>
                                <th>Resource</th>
                                <th>Details</th>
                                <th>IP Address</th>
                                <th>Status</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${data.auditEntries.map(entry => {
                                const safeEntry = Sanitizer.sanitizeObject(entry);
                                return `
                                <tr>
                                    <td>${new Date(entry.timestamp).toLocaleString()}</td>
                                    <td>${safeEntry.userName}</td>
                                    <td>
                                        <span class="action-badge ${safeEntry.action}">${safeEntry.action}</span>
                                    </td>
                                    <td>${safeEntry.resource}</td>
                                    <td>
                                        <button class="btn btn-sm btn-outline" onclick="PropertyHubAdmin.showAuditDetails('${safeEntry.id}')">
                                            View Details
                                        </button>
                                    </td>
                                    <td>${safeEntry.ipAddress}</td>
                                    <td>
                                        <span class="status-indicator ${safeEntry.status}">${safeEntry.status}</span>
                                    </td>
                                </tr>
                            `;}).join('')}
                        </tbody>
                    </table>
                </div>
            </div>
        `;
    }

    // Event Handlers
    setupUserManagementEvents() {
        // User search
        document.getElementById('user-search').addEventListener('input', (e) => {
            this.filterUsers('search', e.target.value);
        });

        // Filter events
        document.getElementById('role-filter').addEventListener('change', (e) => {
            this.filterUsers('role', e.target.value);
        });

        document.getElementById('status-filter').addEventListener('change', (e) => {
            this.filterUsers('status', e.target.value);
        });
    }

    setupConfigurationEvents() {
        // Track configuration changes
        document.querySelectorAll('[data-config]').forEach(input => {
            input.addEventListener('change', (e) => {
                this.markConfigurationChanged(e.target.dataset.config, e.target.value || e.target.checked);
            });
        });
    }

    setupEventListeners() {
        // Admin navigation
        document.addEventListener('click', (e) => {
            if (e.target.matches('.admin-nav-item')) {
                this.navigateToSection(e.target.dataset.section);
            }
        });

        // Modal events
        document.addEventListener('click', (e) => {
            if (e.target.matches('.modal-close') || e.target.matches('.modal-backdrop')) {
                this.closeModal();
            }
        });

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.ctrlKey || e.metaKey) {
                switch (e.key) {
                    case 's':
                        e.preventDefault();
                        this.saveAllConfigurations();
                        break;
                    case 'r':
                        e.preventDefault();
                        this.refreshCurrentSection();
                        break;
                }
            }
        });
    }

    // System Monitoring
    startSystemMonitoring() {
        // Update system status every 30 seconds
        setInterval(() => {
            this.updateSystemStatus();
        }, 30000);

        // Check for alerts every 10 seconds
        setInterval(() => {
            this.checkSystemAlerts();
        }, 10000);
    }

    async updateSystemStatus() {
        try {
            const response = await fetch('/api/admin/system/status', {
                headers: { 'Authorization': `Bearer ${window.PropertyHubAuth.token}` }
            });
            
            const data = await response.json();
            this.systemStatus = data;
            this.updateSystemStatusDisplay(data);
        } catch (error) {
            console.error('Failed to update system status:', error);
        }
    }

    async checkSystemAlerts() {
        try {
            const response = await fetch('/api/admin/system/alerts', {
                headers: { 'Authorization': `Bearer ${window.PropertyHubAuth.token}` }
            });
            
            const data = await response.json();
            
            if (data.alerts && data.alerts.length > 0) {
                data.alerts.forEach(alert => {
                    if (!this.systemAlerts.find(a => a.id === alert.id)) {
                        this.systemAlerts.push(alert);
                        this.showSystemAlert(alert);
                    }
                });
            }
        } catch (error) {
            console.error('Failed to check system alerts:', error);
        }
    }

    // Utility Methods
    async loadAdminData() {
        // Load initial admin data
        const dataPromises = [
            this.loadSystemStatus(),
            this.loadPendingApprovals(),
            this.loadRecentActivity()
        ];

        await Promise.allSettled(dataPromises);
    }

    setupAdminInterface() {
        // Setup admin sidebar navigation
        this.renderAdminNavigation();
        
        // Load default section
        this.navigateToSection('dashboard');
    }

    renderAdminNavigation() {
        const navigation = document.getElementById('admin-navigation');
        if (!navigation) return;

        navigation.innerHTML = `
            <div class="admin-nav">
                <div class="nav-item admin-nav-item ${this.permissions.systemMonitoring ? '' : 'disabled'}" data-section="dashboard">
                    <i class="fas fa-tachometer-alt"></i> Dashboard
                </div>
                <div class="nav-item admin-nav-item ${this.permissions.userManagement ? '' : 'disabled'}" data-section="users">
                    <i class="fas fa-users"></i> User Management
                </div>
                <div class="nav-item admin-nav-item ${this.permissions.propertyOverride ? '' : 'disabled'}" data-section="properties">
                    <i class="fas fa-home"></i> Property Oversight
                </div>
                <div class="nav-item admin-nav-item ${this.permissions.systemConfig ? '' : 'disabled'}" data-section="config">
                    <i class="fas fa-cog"></i> Configuration
                </div>
                <div class="nav-item admin-nav-item ${this.permissions.systemMonitoring ? '' : 'disabled'}" data-section="monitoring">
                    <i class="fas fa-chart-line"></i> System Monitoring
                </div>
                <div class="nav-item admin-nav-item ${this.permissions.auditAccess ? '' : 'disabled'}" data-section="audit">
                    <i class="fas fa-clipboard-list"></i> Audit Log
                </div>
                <div class="nav-item admin-nav-item ${this.permissions.integrationManagement ? '' : 'disabled'}" data-section="integrations">
                    <i class="fas fa-plug"></i> Integrations
                </div>
            </div>
        `;
    }

    navigateToSection(section) {
        // Update active navigation
        document.querySelectorAll('.admin-nav-item').forEach(item => {
            item.classList.remove('active');
        });
        document.querySelector(`[data-section="${section}"]`)?.classList.add('active');

        // Load section content
        switch (section) {
            case 'dashboard':
                this.loadSystemMonitoring();
                break;
            case 'users':
                this.loadUserManagement();
                break;
            case 'properties':
                this.loadPropertyOversight();
                break;
            case 'config':
                this.loadConfigurationManagement();
                break;
            case 'monitoring':
                this.loadSystemMonitoring();
                break;
            case 'audit':
                this.loadAuditLog();
                break;
            case 'integrations':
                this.loadIntegrationManagement();
                break;
        }
    }

    showAccessDenied(feature) {
        const container = document.getElementById('admin-content');
        const safeFeature = Sanitizer.escapeHtml(feature);
        container.innerHTML = `
            <div class="access-denied">
                <i class="fas fa-lock fa-4x"></i>
                <h2>Access Denied</h2>
                <p>You don't have permission to access ${safeFeature}.</p>
                <p>Contact your system administrator to request access.</p>
            </div>
        `;
    }

    showError(message, error) {
        console.error(message, error);
        const container = document.getElementById('admin-content');
        const safeMessage = Sanitizer.escapeHtml(message);
        const safeErrorMessage = Sanitizer.escapeHtml(error?.message || 'An unexpected error occurred.');
        container.innerHTML = `
            <div class="error-message">
                <i class="fas fa-exclamation-triangle fa-2x"></i>
                <h3>${safeMessage}</h3>
                <p>${safeErrorMessage}</p>
                <button class="btn btn-primary" onclick="location.reload()">Retry</button>
            </div>
        `;
    }

    showSystemAlert(alert) {
        const safeAlert = Sanitizer.sanitizeObject(alert);
        const alertElement = document.createElement('div');
        alertElement.className = `system-alert ${safeAlert.severity}`;
        alertElement.innerHTML = `
            <div class="alert-content">
                <h4>${safeAlert.title}</h4>
                <p>${safeAlert.message}</p>
                <span class="alert-time">${new Date(alert.timestamp).toLocaleString()}</span>
            </div>
            <button class="alert-dismiss" onclick="this.parentElement.remove()">×</button>
        `;

        document.body.appendChild(alertElement);

        // Auto-remove after 10 seconds for non-critical alerts
        if (alert.severity !== 'critical') {
            setTimeout(() => {
                if (alertElement.parentElement) {
                    alertElement.remove();
                }
            }, 10000);
        }
    }

    formatFileSize(bytes) {
        const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
        if (bytes === 0) return '0 Bytes';
        const i = parseInt(Math.floor(Math.log(bytes) / Math.log(1024)));
        return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
    }

    formatUptime(seconds) {
        const days = Math.floor(seconds / 86400);
        const hours = Math.floor((seconds % 86400) / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        
        return `${days}d ${hours}h ${minutes}m`;
    }

    // Public API Methods
    async viewUser(userId) {
        // Implementation for viewing user details
        console.log('Viewing user:', userId);
    }

    async editUser(userId) {
        // Implementation for editing user
        console.log('Editing user:', userId);
    }

    async deleteUser(userId) {
        // Implementation for deleting user
        console.log('Deleting user:', userId);
    }

    async toggleUserStatus(userId, currentStatus) {
        // Implementation for toggling user status
        console.log('Toggling user status:', userId, currentStatus);
    }

    async exportUserData() {
        // Implementation for exporting user data
        console.log('Exporting user data');
    }

    async saveAllConfigurations() {
        // Implementation for saving all configurations
        console.log('Saving all configurations');
    }

    async resetToDefaults() {
        // Implementation for resetting to defaults
        console.log('Resetting to defaults');
    }

    showConfigTab(tabName) {
        // Hide all panels
        document.querySelectorAll('.config-panel').forEach(panel => {
            panel.classList.remove('active');
        });
        
        // Show selected panel
        document.getElementById(`config-${tabName}`).classList.add('active');
        
        // Update tab buttons
        document.querySelectorAll('.tab-button').forEach(button => {
            button.classList.remove('active');
        });
        event.target.classList.add('active');
    }
}

// Initialize PropertyHub Admin when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Only initialize if on admin page
    if (window.location.pathname.startsWith('/admin')) {
        window.PropertyHubAdmin = new PropertyHubAdmin();
    }
});

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PropertyHubAdmin;
}
