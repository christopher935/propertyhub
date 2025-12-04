/**
 * PropertyHub Dashboard - Real-Time Enterprise Dashboard
 * Live updates, real-time analytics, and interactive widgets
 * Integrates with PropertyHub backend for comprehensive property management
 */

class PropertyHubDashboard {
    constructor() {
        this.websocket = null;
        this.updateIntervals = new Map();
        this.widgets = new Map();
        this.alerts = [];
        this.connectionRetryCount = 0;
        this.maxRetries = 5;
        this.retryDelay = 5000;
        
        // Dashboard configuration
        this.config = {
            updateFrequency: 30000, // 30 seconds
            criticalUpdateFrequency: 5000, // 5 seconds
            websocketReconnectDelay: 3000,
            maxConcurrentRequests: 10
        };

        this.init();
    }

    async init() {
        console.log('🏢 PropertyHub Dashboard initializing...');
        
        try {
            await this.setupWebSocket();
            await this.loadDashboardData();
            this.setupWidgets();
            this.setupEventListeners();
            this.startRealTimeUpdates();
            this.setupKeyboardShortcuts();
            
            console.log('✅ PropertyHub Dashboard initialized successfully');
        } catch (error) {
            console.error('❌ Dashboard initialization failed:', error);
            this.showErrorAlert('Dashboard initialization failed. Some features may not work properly.');
        }
    }

    // WebSocket Connection Management
    async setupWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws/dashboard`;
        
        try {
            this.websocket = new WebSocket(wsUrl);
            
            this.websocket.onopen = () => {
                console.log('🔗 WebSocket connected for real-time updates');
                this.connectionRetryCount = 0;
                this.updateConnectionStatus(true);
                
                // Send authentication
                this.websocket.send(JSON.stringify({
                    type: 'auth',
                    token: localStorage.getItem('token')
                }));
            };

            this.websocket.onmessage = (event) => {
                this.handleWebSocketMessage(JSON.parse(event.data));
            };

            this.websocket.onclose = () => {
                console.log('🔌 WebSocket connection closed');
                this.updateConnectionStatus(false);
                this.scheduleReconnect();
            };

            this.websocket.onerror = (error) => {
                console.error('📡 WebSocket error:', error);
                this.updateConnectionStatus(false);
            };

        } catch (error) {
            console.error('Failed to establish WebSocket connection:', error);
            this.updateConnectionStatus(false);
        }
    }

    handleWebSocketMessage(message) {
        switch (message.type) {
            case 'property_update':
                this.updatePropertyWidget(message.data);
                break;
            case 'booking_update':
                this.updateBookingWidget(message.data);
                break;
            case 'revenue_update':
                this.updateRevenueWidget(message.data);
                break;
            case 'alert':
                this.showAlert(message.data);
                break;
            case 'market_data_update':
                this.updateMarketDataWidget(message.data);
                break;
            case 'maintenance_update':
                this.updateMaintenanceWidget(message.data);
                break;
            case 'lead_update':
                this.updateLeadWidget(message.data);
                break;
            default:
                console.log('Unknown WebSocket message type:', message.type);
        }
    }

    scheduleReconnect() {
        if (this.connectionRetryCount < this.maxRetries) {
            setTimeout(() => {
                this.connectionRetryCount++;
                console.log(`🔄 Reconnecting WebSocket (attempt ${this.connectionRetryCount}/${this.maxRetries})`);
                this.setupWebSocket();
            }, this.config.websocketReconnectDelay);
        } else {
            console.error('❌ WebSocket max reconnection attempts reached');
            this.showErrorAlert('Real-time updates disabled. Please refresh the page.');
        }
    }

    updateConnectionStatus(isConnected) {
        const statusEl = document.getElementById('connection-status');
        if (statusEl) {
            statusEl.className = isConnected ? 
                'connection-status connected' : 
                'connection-status disconnected';
            statusEl.textContent = isConnected ? 'Connected' : 'Disconnected';
        }
    }

    // Dashboard Data Loading
    async loadDashboardData() {
        console.log('📊 Loading dashboard data...');
        
        const dataPromises = [
            this.loadPropertySummary(),
            this.loadBookingSummary(),
            this.loadRevenueSummary(),
            this.loadMarketData(),
            this.loadRecentActivity(),
            this.loadAlerts(),
            this.loadMaintenanceRequests(),
            this.loadLeadSummary(),
            this.loadUpcomingTasks()
        ];

        try {
            await Promise.allSettled(dataPromises);
            console.log('✅ Dashboard data loaded successfully');
        } catch (error) {
            console.error('❌ Error loading dashboard data:', error);
            this.showErrorAlert('Failed to load some dashboard data');
        }
    }

    async loadPropertySummary() {
        try {
            const response = await this.apiCall('/api/dashboard/properties');
            const data = await response.json();
            this.updatePropertySummaryWidget(data);
        } catch (error) {
            console.error('Error loading property summary:', error);
        }
    }

    async loadBookingSummary() {
        try {
            const response = await this.apiCall('/api/dashboard/bookings');
            const data = await response.json();
            this.updateBookingSummaryWidget(data);
        } catch (error) {
            console.error('Error loading booking summary:', error);
        }
    }

    async loadRevenueSummary() {
        try {
            const response = await this.apiCall('/api/dashboard/revenue');
            const data = await response.json();
            this.updateRevenueSummaryWidget(data);
        } catch (error) {
            console.error('Error loading revenue summary:', error);
        }
    }

    async loadMarketData() {
        try {
            const response = await this.apiCall('/api/dashboard/market-data');
            const data = await response.json();
            this.updateMarketDataSummary(data);
        } catch (error) {
            console.error('Error loading market data:', error);
        }
    }

    async loadRecentActivity() {
        try {
            const response = await this.apiCall('/api/dashboard/recent-activity');
            const data = await response.json();
            this.updateRecentActivityWidget(data);
        } catch (error) {
            console.error('Error loading recent activity:', error);
        }
    }

    async loadAlerts() {
        try {
            const response = await this.apiCall('/api/dashboard/alerts');
            const data = await response.json();
            this.updateAlertsWidget(data);
        } catch (error) {
            console.error('Error loading alerts:', error);
        }
    }

    async loadMaintenanceRequests() {
        try {
            const response = await this.apiCall('/api/dashboard/maintenance');
            const data = await response.json();
            this.updateMaintenanceWidget(data);
        } catch (error) {
            console.error('Error loading maintenance requests:', error);
        }
    }

    async loadLeadSummary() {
        try {
            const response = await this.apiCall('/api/dashboard/leads');
            const data = await response.json();
            this.updateLeadSummaryWidget(data);
        } catch (error) {
            console.error('Error loading lead summary:', error);
        }
    }

    async loadUpcomingTasks() {
        try {
            const response = await this.apiCall('/api/dashboard/upcoming-tasks');
            const data = await response.json();
            this.updateUpcomingTasksWidget(data);
        } catch (error) {
            console.error('Error loading upcoming tasks:', error);
        }
    }

    // Widget Update Methods
    updatePropertySummaryWidget(data) {
        const widget = document.getElementById('property-summary-widget');
        if (!widget) return;

        widget.innerHTML = `
            <div class="widget-header">
                <h3>Property Portfolio</h3>
                <div class="widget-actions">
                    <button onclick="window.location.href='/properties'" class="btn btn-sm btn-primary">
                        View All
                    </button>
                </div>
            </div>
            <div class="widget-content">
                <div class="stats-grid">
                    <div class="stat-item">
                        <div class="stat-value">${data.totalProperties}</div>
                        <div class="stat-label">Total Properties</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">${data.occupiedUnits}</div>
                        <div class="stat-label">Occupied Units</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">${data.vacantUnits}</div>
                        <div class="stat-label">Vacant Units</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">${data.occupancyRate}%</div>
                        <div class="stat-label">Occupancy Rate</div>
                    </div>
                </div>
                <div class="property-list">
                    ${data.recentProperties.map(property => `
                        <div class="property-item">
                            <div class="property-info">
                                <h4>${property.address}</h4>
                                <p>${property.city}, ${property.state} ${property.zipCode}</p>
                            </div>
                            <div class="property-status ${property.status.toLowerCase()}">
                                ${property.status}
                            </div>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }

    updateBookingSummaryWidget(data) {
        const widget = document.getElementById('booking-summary-widget');
        if (!widget) return;

        widget.innerHTML = `
            <div class="widget-header">
                <h3>Bookings Overview</h3>
                <div class="widget-actions">
                    <button onclick="window.location.href='/bookings'" class="btn btn-sm btn-primary">
                        Manage Bookings
                    </button>
                </div>
            </div>
            <div class="widget-content">
                <div class="stats-grid">
                    <div class="stat-item">
                        <div class="stat-value">${data.activeBookings}</div>
                        <div class="stat-label">Active Bookings</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">${data.pendingBookings}</div>
                        <div class="stat-label">Pending Approval</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">${data.todayCheckIns}</div>
                        <div class="stat-label">Today's Check-ins</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">${data.todayCheckOuts}</div>
                        <div class="stat-label">Today's Check-outs</div>
                    </div>
                </div>
                <div class="recent-bookings">
                    <h4>Recent Bookings</h4>
                    ${data.recentBookings.map(booking => `
                        <div class="booking-item">
                            <div class="booking-info">
                                <h5>${booking.guestName}</h5>
                                <p>${booking.propertyAddress}</p>
                                <p>${booking.checkInDate} - ${booking.checkOutDate}</p>
                            </div>
                            <div class="booking-status ${booking.status.toLowerCase()}">
                                ${booking.status}
                            </div>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }

    updateRevenueSummaryWidget(data) {
        const widget = document.getElementById('revenue-summary-widget');
        if (!widget) return;

        widget.innerHTML = `
            <div class="widget-header">
                <h3>Revenue Analytics</h3>
                <div class="widget-actions">
                    <button onclick="window.location.href='/analytics'" class="btn btn-sm btn-primary">
                        View Reports
                    </button>
                </div>
            </div>
            <div class="widget-content">
                <div class="stats-grid">
                    <div class="stat-item">
                        <div class="stat-value">$${data.monthlyRevenue.toLocaleString()}</div>
                        <div class="stat-label">Monthly Revenue</div>
                        <div class="stat-change ${data.monthlyChange >= 0 ? 'positive' : 'negative'}">
                            ${data.monthlyChange >= 0 ? '+' : ''}${data.monthlyChange}%
                        </div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">$${data.yearlyRevenue.toLocaleString()}</div>
                        <div class="stat-label">Yearly Revenue</div>
                        <div class="stat-change ${data.yearlyChange >= 0 ? 'positive' : 'negative'}">
                            ${data.yearlyChange >= 0 ? '+' : ''}${data.yearlyChange}%
                        </div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">$${data.avgDailyRate}</div>
                        <div class="stat-label">Avg Daily Rate</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">${data.revPAR}</div>
                        <div class="stat-label">RevPAR</div>
                    </div>
                </div>
                <canvas id="revenue-trend-chart" width="400" height="200"></canvas>
            </div>
        `;

        // Update revenue trend chart if charts library is available
        if (window.PropertyHubCharts) {
            this.updateRevenueTrendChart(data.trendData);
        }
    }

    updateMarketDataSummary(data) {
        const widget = document.getElementById('market-data-widget');
        if (!widget) return;

        widget.innerHTML = `
            <div class="widget-header">
                <h3>Market Intelligence</h3>
                <div class="widget-actions">
                    <button onclick="this.refreshMarketData()" class="btn btn-sm btn-secondary">
                        Refresh Data
                    </button>
                </div>
            </div>
            <div class="widget-content">
                <div class="market-stats">
                    <div class="stat-item">
                        <div class="stat-value">$${data.averageRent}</div>
                        <div class="stat-label">Market Avg Rent</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">${data.marketOccupancy}%</div>
                        <div class="stat-label">Market Occupancy</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">${data.daysOnMarket}</div>
                        <div class="stat-label">Avg Days on Market</div>
                    </div>
                </div>
                <div class="market-insights">
                    <h4>Market Insights</h4>
                    <ul>
                        ${data.insights.map(insight => `<li>${insight}</li>`).join('')}
                    </ul>
                </div>
                <div class="last-updated">
                    Last updated: ${new Date(data.lastUpdated).toLocaleString()}
                </div>
            </div>
        `;
    }

    updateRecentActivityWidget(data) {
        const widget = document.getElementById('recent-activity-widget');
        if (!widget) return;

        widget.innerHTML = `
            <div class="widget-header">
                <h3>Recent Activity</h3>
            </div>
            <div class="widget-content">
                <div class="activity-feed">
                    ${data.activities.map(activity => `
                        <div class="activity-item">
                            <div class="activity-icon ${activity.type}">
                                <i class="${this.getActivityIcon(activity.type)}"></i>
                            </div>
                            <div class="activity-content">
                                <p>${activity.description}</p>
                                <span class="activity-time">${this.formatTimeAgo(activity.timestamp)}</span>
                            </div>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }

    updateAlertsWidget(data) {
        const widget = document.getElementById('alerts-widget');
        if (!widget) return;

        widget.innerHTML = `
            <div class="widget-header">
                <h3>Alerts & Notifications</h3>
                <div class="alert-count ${data.criticalCount > 0 ? 'has-critical' : ''}">
                    ${data.totalCount}
                </div>
            </div>
            <div class="widget-content">
                <div class="alert-list">
                    ${data.alerts.map(alert => `
                        <div class="alert-item ${alert.priority}">
                            <div class="alert-content">
                                <h5>${alert.title}</h5>
                                <p>${alert.message}</p>
                                <span class="alert-time">${this.formatTimeAgo(alert.timestamp)}</span>
                            </div>
                            <div class="alert-actions">
                                <button onclick="this.dismissAlert('${alert.id}')" class="btn btn-sm">
                                    Dismiss
                                </button>
                            </div>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }

    updateMaintenanceWidget(data) {
        const widget = document.getElementById('maintenance-widget');
        if (!widget) return;

        widget.innerHTML = `
            <div class="widget-header">
                <h3>Maintenance Requests</h3>
                <div class="widget-actions">
                    <button onclick="window.location.href='/maintenance'" class="btn btn-sm btn-primary">
                        View All
                    </button>
                </div>
            </div>
            <div class="widget-content">
                <div class="stats-grid">
                    <div class="stat-item">
                        <div class="stat-value">${data.openRequests}</div>
                        <div class="stat-label">Open Requests</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">${data.urgentRequests}</div>
                        <div class="stat-label">Urgent</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">${data.avgResponseTime}h</div>
                        <div class="stat-label">Avg Response Time</div>
                    </div>
                </div>
                <div class="maintenance-list">
                    ${data.recentRequests.map(request => `
                        <div class="maintenance-item">
                            <div class="maintenance-info">
                                <h5>${request.title}</h5>
                                <p>${request.property}</p>
                                <span class="request-time">${this.formatTimeAgo(request.createdAt)}</span>
                            </div>
                            <div class="priority ${request.priority}">
                                ${request.priority}
                            </div>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }

    updateLeadSummaryWidget(data) {
        const widget = document.getElementById('lead-summary-widget');
        if (!widget) return;

        widget.innerHTML = `
            <div class="widget-header">
                <h3>Lead Management</h3>
                <div class="widget-actions">
                    <button onclick="window.location.href='/leads'" class="btn btn-sm btn-primary">
                        View CRM
                    </button>
                </div>
            </div>
            <div class="widget-content">
                <div class="stats-grid">
                    <div class="stat-item">
                        <div class="stat-value">${data.newLeads}</div>
                        <div class="stat-label">New Leads Today</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">${data.hotLeads}</div>
                        <div class="stat-label">Hot Leads</div>
                    </div>
                    <div class="stat-item">
                        <div class="stat-value">${data.conversionRate}%</div>
                        <div class="stat-label">Conversion Rate</div>
                    </div>
                </div>
                <div class="lead-sources">
                    <h4>Top Lead Sources</h4>
                    ${data.leadSources.map(source => `
                        <div class="source-item">
                            <span class="source-name">${source.name}</span>
                            <span class="source-count">${source.count}</span>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }

    updateUpcomingTasksWidget(data) {
        const widget = document.getElementById('upcoming-tasks-widget');
        if (!widget) return;

        widget.innerHTML = `
            <div class="widget-header">
                <h3>Upcoming Tasks</h3>
            </div>
            <div class="widget-content">
                <div class="task-list">
                    ${data.tasks.map(task => `
                        <div class="task-item">
                            <div class="task-content">
                                <h5>${task.title}</h5>
                                <p>${task.description}</p>
                                <span class="task-due">Due: ${new Date(task.dueDate).toLocaleDateString()}</span>
                            </div>
                            <div class="task-priority ${task.priority}">
                                ${task.priority}
                            </div>
                        </div>
                    `).join('')}
                </div>
            </div>
        `;
    }

    // Real-time Updates
    startRealTimeUpdates() {
        // Critical widgets update every 5 seconds
        this.updateIntervals.set('critical', setInterval(() => {
            this.updateCriticalWidgets();
        }, this.config.criticalUpdateFrequency));

        // Regular widgets update every 30 seconds
        this.updateIntervals.set('regular', setInterval(() => {
            this.updateRegularWidgets();
        }, this.config.updateFrequency));

        // Market data updates every 5 minutes
        this.updateIntervals.set('market', setInterval(() => {
            this.loadMarketData();
        }, 300000));
    }

    async updateCriticalWidgets() {
        const criticalUpdates = [
            this.loadBookingSummary(),
            this.loadAlerts(),
            this.loadRecentActivity()
        ];

        try {
            await Promise.allSettled(criticalUpdates);
        } catch (error) {
            console.error('Error updating critical widgets:', error);
        }
    }

    async updateRegularWidgets() {
        const regularUpdates = [
            this.loadPropertySummary(),
            this.loadRevenueSummary(),
            this.loadMaintenanceRequests(),
            this.loadLeadSummary(),
            this.loadUpcomingTasks()
        ];

        try {
            await Promise.allSettled(regularUpdates);
        } catch (error) {
            console.error('Error updating regular widgets:', error);
        }
    }

    // Utility Methods
    async apiCall(endpoint, options = {}) {
        const defaultOptions = {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`,
                'Content-Type': 'application/json'
            }
        };

        const response = await fetch(endpoint, { ...defaultOptions, ...options });
        
        if (!response.ok) {
            throw new Error(`API call failed: ${response.status} ${response.statusText}`);
        }
        
        return response;
    }

    getActivityIcon(type) {
        const iconMap = {
            booking: 'fas fa-calendar-check',
            property: 'fas fa-home',
            revenue: 'fas fa-dollar-sign',
            maintenance: 'fas fa-wrench',
            lead: 'fas fa-user-plus',
            alert: 'fas fa-exclamation-triangle'
        };
        return iconMap[type] || 'fas fa-info-circle';
    }

    formatTimeAgo(timestamp) {
        const now = new Date();
        const time = new Date(timestamp);
        const diffInSeconds = Math.floor((now - time) / 1000);

        if (diffInSeconds < 60) return 'Just now';
        if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
        if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`;
        return `${Math.floor(diffInSeconds / 86400)}d ago`;
    }

    showAlert(alert) {
        this.alerts.push(alert);
        this.displayNotification(alert);
        this.loadAlerts(); // Refresh alerts widget
    }

    showErrorAlert(message) {
        this.displayNotification({
            type: 'error',
            title: 'Error',
            message: message
        });
    }

    displayNotification(alert) {
        // Create notification toast
        const toast = document.createElement('div');
        toast.className = `notification-toast ${alert.type || 'info'}`;
        toast.innerHTML = `
            <div class="toast-content">
                <h4>${alert.title}</h4>
                <p>${alert.message}</p>
            </div>
            <button class="toast-close" onclick="this.parentElement.remove()">×</button>
        `;

        const container = document.getElementById('notification-container') || this.createNotificationContainer();
        container.appendChild(toast);

        // Auto-remove after 5 seconds
        setTimeout(() => {
            if (toast.parentElement) {
                toast.remove();
            }
        }, 5000);
    }

    createNotificationContainer() {
        const container = document.createElement('div');
        container.id = 'notification-container';
        container.className = 'notification-container';
        document.body.appendChild(container);
        return container;
    }

    dismissAlert(alertId) {
        fetch(`/api/alerts/${alertId}/dismiss`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        }).then(() => {
            this.loadAlerts(); // Refresh alerts widget
        });
    }

    async refreshMarketData() {
        const button = event.target;
        button.disabled = true;
        button.textContent = 'Refreshing...';
        
        try {
            await this.loadMarketData();
            button.textContent = 'Refresh Data';
        } catch (error) {
            button.textContent = 'Refresh Failed';
            console.error('Market data refresh failed:', error);
        } finally {
            button.disabled = false;
            setTimeout(() => {
                if (button.textContent !== 'Refresh Data') {
                    button.textContent = 'Refresh Data';
                }
            }, 3000);
        }
    }

    setupEventListeners() {
        // Dashboard refresh
        document.addEventListener('keydown', (e) => {
            if (e.key === 'F5' || (e.ctrlKey && e.key === 'r')) {
                e.preventDefault();
                this.refreshDashboard();
            }
        });

        // Window focus/blur for pausing updates
        window.addEventListener('focus', () => {
            this.startRealTimeUpdates();
        });

        window.addEventListener('blur', () => {
            // Optionally pause updates when window is not focused
            // this.pauseUpdates();
        });
    }

    setupKeyboardShortcuts() {
        document.addEventListener('keydown', (e) => {
            if (e.ctrlKey || e.metaKey) {
                switch (e.key) {
                    case 'd':
                        e.preventDefault();
                        this.refreshDashboard();
                        break;
                    case 'p':
                        e.preventDefault();
                        window.location.href = '/properties';
                        break;
                    case 'b':
                        e.preventDefault();
                        window.location.href = '/bookings';
                        break;
                    case 'a':
                        e.preventDefault();
                        window.location.href = '/analytics';
                        break;
                }
            }
        });
    }

    async refreshDashboard() {
        console.log('🔄 Refreshing dashboard...');
        const refreshButton = document.getElementById('dashboard-refresh');
        if (refreshButton) {
            refreshButton.disabled = true;
            refreshButton.textContent = 'Refreshing...';
        }

        try {
            await this.loadDashboardData();
            this.displayNotification({
                type: 'success',
                title: 'Dashboard Refreshed',
                message: 'All data has been updated successfully'
            });
        } catch (error) {
            this.showErrorAlert('Failed to refresh dashboard data');
        } finally {
            if (refreshButton) {
                refreshButton.disabled = false;
                refreshButton.textContent = 'Refresh Dashboard';
            }
        }
    }

    setupWidgets() {
        // Initialize any interactive widgets
        this.setupDraggableWidgets();
        this.setupResizableWidgets();
    }

    setupDraggableWidgets() {
        // Make widgets draggable for customization
        const widgets = document.querySelectorAll('.dashboard-widget');
        // Implement drag and drop functionality if needed
    }

    setupResizableWidgets() {
        // Make widgets resizable
        const widgets = document.querySelectorAll('.dashboard-widget');
        // Implement resize functionality if needed
    }

    destroy() {
        // Cleanup intervals
        for (const [name, interval] of this.updateIntervals) {
            clearInterval(interval);
        }
        this.updateIntervals.clear();

        // Close WebSocket
        if (this.websocket) {
            this.websocket.close();
        }

        // Clear alerts
        this.alerts = [];
    }
}

// Initialize PropertyHub Dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.PropertyHubDashboard = new PropertyHubDashboard();
});

// Cleanup on page unload
window.addEventListener('beforeunload', () => {
    if (window.PropertyHubDashboard) {
        window.PropertyHubDashboard.destroy();
    }
});

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PropertyHubDashboard;
}
