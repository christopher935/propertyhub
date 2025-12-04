/**
 * PropertyHub Notifications System
 * Real-time notifications, alerts, and messaging system
 * WebSocket, browser notifications, email/SMS integration
 */

class PropertyHubNotifications {
    constructor() {
        this.websocket = null;
        this.notifications = [];
        this.unreadCount = 0;
        this.soundEnabled = true;
        this.browserNotificationsEnabled = false;
        this.preferences = {};
        this.notificationQueue = [];
        this.retryCount = 0;
        this.maxRetries = 3;
        
        // Notification types
        this.notificationTypes = {
            BOOKING_UPDATE: 'booking_update',
            PROPERTY_ALERT: 'property_alert',
            PAYMENT_STATUS: 'payment_status',
            DOCUMENT_READY: 'document_ready',
            DEADLINE_REMINDER: 'deadline_reminder',
            SYSTEM_ALERT: 'system_alert',
            PRICE_ALERT: 'price_alert',
            NEW_MESSAGE: 'new_message',
            INSPECTION_SCHEDULED: 'inspection_scheduled',
            CLOSING_REMINDER: 'closing_reminder',
            FRIDAY_REPORT: 'friday_report',
            TREC_COMPLIANCE: 'trec_compliance'
        };

        // Priority levels
        this.priorityLevels = {
            LOW: 'low',
            MEDIUM: 'medium',
            HIGH: 'high',
            CRITICAL: 'critical'
        };

        // Delivery methods
        this.deliveryMethods = {
            IN_APP: 'in_app',
            BROWSER: 'browser',
            EMAIL: 'email',
            SMS: 'sms',
            PUSH: 'push'
        };

        this.init();
    }

    async init() {
        console.log('🔔 PropertyHub Notifications initializing...');
        
        await this.loadUserPreferences();
        await this.setupWebSocketConnection();
        await this.requestBrowserPermissions();
        
        this.setupNotificationCenter();
        this.loadNotificationHistory();
        this.setupEventListeners();
        this.startPeriodicSync();
        
        console.log('✅ PropertyHub Notifications ready');
    }

    // WebSocket Connection Management
    async setupWebSocketConnection() {
        if (!window.PropertyHubAuth || !window.PropertyHubAuth.isAuthenticated()) {
            console.log('⏳ Waiting for authentication before connecting notifications...');
            setTimeout(() => this.setupWebSocketConnection(), 2000);
            return;
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws/notifications`;
        
        try {
            this.websocket = new WebSocket(wsUrl);
            
            this.websocket.onopen = () => {
                console.log('🔗 Notifications WebSocket connected');
                this.retryCount = 0;
                
                // Send authentication
                this.websocket.send(JSON.stringify({
                    type: 'auth',
                    token: localStorage.getItem('token')
                }));
                
                // Process queued notifications
                this.processNotificationQueue();
                
                this.updateConnectionStatus(true);
            };

            this.websocket.onmessage = (event) => {
                this.handleIncomingNotification(JSON.parse(event.data));
            };

            this.websocket.onclose = () => {
                console.log('🔌 Notifications WebSocket closed');
                this.updateConnectionStatus(false);
                this.scheduleReconnect();
            };

            this.websocket.onerror = (error) => {
                console.error('📡 Notifications WebSocket error:', error);
                this.updateConnectionStatus(false);
            };

        } catch (error) {
            console.error('Failed to connect notifications WebSocket:', error);
            this.scheduleReconnect();
        }
    }

    scheduleReconnect() {
        if (this.retryCount < this.maxRetries) {
            const delay = Math.min(1000 * Math.pow(2, this.retryCount), 30000);
            setTimeout(() => {
                this.retryCount++;
                console.log(`🔄 Reconnecting notifications (attempt ${this.retryCount}/${this.maxRetries})`);
                this.setupWebSocketConnection();
            }, delay);
        }
    }

    updateConnectionStatus(isConnected) {
        const statusEl = document.getElementById('notifications-status');
        if (statusEl) {
            statusEl.className = `notification-status ${isConnected ? 'connected' : 'disconnected'}`;
            statusEl.title = isConnected ? 'Real-time notifications active' : 'Notifications offline';
        }
    }

    // Notification Processing
    handleIncomingNotification(message) {
        switch (message.type) {
            case 'notification':
                this.processNotification(message.data);
                break;
            case 'bulk_notifications':
                message.data.forEach(notification => this.processNotification(notification));
                break;
            case 'notification_read':
                this.markAsRead(message.data.notificationId);
                break;
            case 'preferences_updated':
                this.updatePreferences(message.data);
                break;
            default:
                console.log('Unknown notification message type:', message.type);
        }
    }

    processNotification(notification) {
        // Add to notifications array
        this.notifications.unshift(notification);
        this.unreadCount++;

        // Update UI
        this.updateNotificationBadge();
        this.addNotificationToCenter(notification);

        // Handle based on priority and type
        this.handleNotificationByType(notification);

        // Play sound if enabled
        if (this.shouldPlaySound(notification)) {
            this.playNotificationSound(notification.priority);
        }

        // Show browser notification
        if (this.shouldShowBrowserNotification(notification)) {
            this.showBrowserNotification(notification);
        }

        // Track analytics
        this.trackNotificationReceived(notification);
    }

    handleNotificationByType(notification) {
        switch (notification.type) {
            case this.notificationTypes.BOOKING_UPDATE:
                this.handleBookingNotification(notification);
                break;
            case this.notificationTypes.PROPERTY_ALERT:
                this.handlePropertyAlert(notification);
                break;
            case this.notificationTypes.PAYMENT_STATUS:
                this.handlePaymentNotification(notification);
                break;
            case this.notificationTypes.DEADLINE_REMINDER:
                this.handleDeadlineReminder(notification);
                break;
            case this.notificationTypes.SYSTEM_ALERT:
                this.handleSystemAlert(notification);
                break;
            case this.notificationTypes.FRIDAY_REPORT:
                this.handleFridayReport(notification);
                break;
            case this.notificationTypes.TREC_COMPLIANCE:
                this.handleTRECCompliance(notification);
                break;
        }
    }

    handleBookingNotification(notification) {
        // Show prominent notification for booking updates
        if (notification.priority === this.priorityLevels.HIGH || notification.priority === this.priorityLevels.CRITICAL) {
            this.showModalNotification(notification);
        }

        // Update booking status if on booking page
        if (window.PropertyHubBooking && window.location.pathname.includes('/booking')) {
            window.PropertyHubBooking.refreshBookingStatus();
        }
    }

    handlePropertyAlert(notification) {
        // Handle property-specific alerts (price changes, new listings, etc.)
        if (notification.data.alertType === 'price_drop') {
            this.showPriceDropAlert(notification);
        } else if (notification.data.alertType === 'new_listing') {
            this.showNewListingAlert(notification);
        }
    }

    handlePaymentNotification(notification) {
        // Show payment status updates
        if (notification.data.status === 'failed') {
            this.showPaymentFailedAlert(notification);
        } else if (notification.data.status === 'completed') {
            this.showPaymentSuccessAlert(notification);
        }
    }

    handleDeadlineReminder(notification) {
        // Show deadline reminders with countdown
        this.showDeadlineReminderModal(notification);
    }

    handleSystemAlert(notification) {
        // System-wide alerts
        if (notification.priority === this.priorityLevels.CRITICAL) {
            this.showSystemMaintenanceAlert(notification);
        }
    }

    handleFridayReport(notification) {
        // Friday report notifications
        this.showFridayReportNotification(notification);
    }

    handleTRECCompliance(notification) {
        // TREC compliance reminders
        this.showTRECComplianceAlert(notification);
    }

    // Notification Center UI
    setupNotificationCenter() {
        const notificationCenter = document.getElementById('notification-center');
        if (!notificationCenter) {
            this.createNotificationCenter();
        }

        this.renderNotificationCenter();
    }

    createNotificationCenter() {
        const centerHTML = `
            <div id="notification-center" class="notification-center">
                <div class="notification-header">
                    <h3>Notifications</h3>
                    <div class="notification-actions">
                        <button class="btn btn-sm btn-outline" onclick="PropertyHubNotifications.markAllAsRead()">
                            <i class="fas fa-check-double"></i> Mark All Read
                        </button>
                        <button class="btn btn-sm btn-outline" onclick="PropertyHubNotifications.showPreferences()">
                            <i class="fas fa-cog"></i> Settings
                        </button>
                        <button class="btn btn-sm btn-outline" onclick="PropertyHubNotifications.toggleCenter()">
                            <i class="fas fa-times"></i>
                        </button>
                    </div>
                </div>
                
                <div class="notification-filters">
                    <button class="filter-btn active" onclick="PropertyHubNotifications.filterNotifications('all')">
                        All
                    </button>
                    <button class="filter-btn" onclick="PropertyHubNotifications.filterNotifications('unread')">
                        Unread <span class="unread-badge">${this.unreadCount}</span>
                    </button>
                    <button class="filter-btn" onclick="PropertyHubNotifications.filterNotifications('bookings')">
                        Bookings
                    </button>
                    <button class="filter-btn" onclick="PropertyHubNotifications.filterNotifications('properties')">
                        Properties
                    </button>
                    <button class="filter-btn" onclick="PropertyHubNotifications.filterNotifications('system')">
                        System
                    </button>
                </div>
                
                <div class="notification-list" id="notification-list">
                    <!-- Notifications will be populated here -->
                </div>
            </div>

            <div id="notification-trigger" class="notification-trigger" onclick="PropertyHubNotifications.toggleCenter()">
                <i class="fas fa-bell"></i>
                <span class="notification-badge" id="notification-badge" style="display: none;">0</span>
                <div class="notification-status" id="notifications-status"></div>
            </div>
        `;

        document.body.insertAdjacentHTML('beforeend', centerHTML);
    }

    renderNotificationCenter() {
        const notificationList = document.getElementById('notification-list');
        if (!notificationList) return;

        if (this.notifications.length === 0) {
            notificationList.innerHTML = `
                <div class="no-notifications">
                    <i class="fas fa-bell-slash fa-2x"></i>
                    <h4>No notifications</h4>
                    <p>You're all caught up!</p>
                </div>
            `;
            return;
        }

        notificationList.innerHTML = this.notifications.map(notification => 
            this.renderNotificationItem(notification)
        ).join('');
    }

    renderNotificationItem(notification) {
        const timeAgo = this.formatTimeAgo(notification.timestamp);
        const isUnread = !notification.read;
        
        return `
            <div class="notification-item ${isUnread ? 'unread' : ''} ${notification.priority}" 
                 data-notification-id="${notification.id}"
                 data-type="${notification.type}">
                
                <div class="notification-icon">
                    <i class="${this.getNotificationIcon(notification.type)}"></i>
                </div>
                
                <div class="notification-content">
                    <div class="notification-title">${notification.title}</div>
                    <div class="notification-message">${notification.message}</div>
                    
                    ${notification.actionUrl ? `
                        <div class="notification-actions">
                            <a href="${notification.actionUrl}" class="btn btn-sm btn-primary notification-action">
                                ${notification.actionText || 'View'}
                            </a>
                        </div>
                    ` : ''}
                </div>
                
                <div class="notification-meta">
                    <span class="notification-time">${timeAgo}</span>
                    <div class="notification-controls">
                        ${isUnread ? `
                            <button class="btn btn-xs btn-outline" onclick="PropertyHubNotifications.markAsRead('${notification.id}')">
                                <i class="fas fa-check"></i>
                            </button>
                        ` : ''}
                        <button class="btn btn-xs btn-outline" onclick="PropertyHubNotifications.dismissNotification('${notification.id}')">
                            <i class="fas fa-times"></i>
                        </button>
                    </div>
                </div>
            </div>
        `;
    }

    addNotificationToCenter(notification) {
        const notificationList = document.getElementById('notification-list');
        if (!notificationList) return;

        // Remove "no notifications" message if it exists
        const noNotifications = notificationList.querySelector('.no-notifications');
        if (noNotifications) {
            noNotifications.remove();
        }

        // Add new notification at the top
        const notificationHTML = this.renderNotificationItem(notification);
        notificationList.insertAdjacentHTML('afterbegin', notificationHTML);

        // Animate in
        const newNotificationEl = notificationList.firstElementChild;
        if (newNotificationEl) {
            newNotificationEl.style.opacity = '0';
            newNotificationEl.style.transform = 'translateY(-20px)';
            
            setTimeout(() => {
                newNotificationEl.style.transition = 'all 0.3s ease';
                newNotificationEl.style.opacity = '1';
                newNotificationEl.style.transform = 'translateY(0)';
            }, 100);
        }
    }

    updateNotificationBadge() {
        const badge = document.getElementById('notification-badge');
        const unreadBadges = document.querySelectorAll('.unread-badge');
        
        if (badge) {
            if (this.unreadCount > 0) {
                badge.textContent = this.unreadCount > 99 ? '99+' : this.unreadCount;
                badge.style.display = 'block';
            } else {
                badge.style.display = 'none';
            }
        }

        unreadBadges.forEach(unreadBadge => {
            unreadBadge.textContent = this.unreadCount;
        });

        // Update page title with unread count
        this.updatePageTitle();
    }

    updatePageTitle() {
        const originalTitle = document.title.replace(/^\(\d+\) /, '');
        
        if (this.unreadCount > 0) {
            document.title = `(${this.unreadCount}) ${originalTitle}`;
        } else {
            document.title = originalTitle;
        }
    }

    // Browser Notifications
    async requestBrowserPermissions() {
        if ('Notification' in window) {
            const permission = await Notification.requestPermission();
            this.browserNotificationsEnabled = permission === 'granted';
            
            if (this.browserNotificationsEnabled) {
                console.log('✅ Browser notifications enabled');
            }
        }
    }

    showBrowserNotification(notification) {
        if (!this.browserNotificationsEnabled || document.visibilityState === 'visible') {
            return;
        }

        const browserNotification = new Notification(notification.title, {
            body: notification.message,
            icon: '/static/images/logo-notification.png',
            badge: '/static/images/badge-notification.png',
            tag: notification.id,
            requireInteraction: notification.priority === this.priorityLevels.CRITICAL,
            actions: notification.actionUrl ? [{
                action: 'view',
                title: notification.actionText || 'View'
            }] : []
        });

        browserNotification.onclick = () => {
            window.focus();
            if (notification.actionUrl) {
                window.location.href = notification.actionUrl;
            }
            browserNotification.close();
        };

        // Auto-close after 10 seconds for non-critical notifications
        if (notification.priority !== this.priorityLevels.CRITICAL) {
            setTimeout(() => {
                browserNotification.close();
            }, 10000);
        }
    }

    // Sound Notifications
    playNotificationSound(priority) {
        if (!this.soundEnabled) return;

        const soundMap = {
            [this.priorityLevels.LOW]: '/static/sounds/notification-low.mp3',
            [this.priorityLevels.MEDIUM]: '/static/sounds/notification-medium.mp3',
            [this.priorityLevels.HIGH]: '/static/sounds/notification-high.mp3',
            [this.priorityLevels.CRITICAL]: '/static/sounds/notification-critical.mp3'
        };

        const soundFile = soundMap[priority] || soundMap[this.priorityLevels.MEDIUM];
        
        try {
            const audio = new Audio(soundFile);
            audio.volume = 0.5;
            audio.play().catch(error => {
                console.log('Could not play notification sound:', error);
            });
        } catch (error) {
            console.error('Error playing notification sound:', error);
        }
    }

    // Modal Notifications
    showModalNotification(notification) {
        const modal = document.createElement('div');
        modal.className = 'notification-modal-overlay';
        modal.innerHTML = `
            <div class="notification-modal ${notification.priority}">
                <div class="notification-modal-header">
                    <div class="notification-modal-icon">
                        <i class="${this.getNotificationIcon(notification.type)}"></i>
                    </div>
                    <h3>${notification.title}</h3>
                    <button class="notification-modal-close" onclick="this.closest('.notification-modal-overlay').remove()">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                
                <div class="notification-modal-content">
                    <p>${notification.message}</p>
                    
                    ${notification.data && notification.data.details ? `
                        <div class="notification-details">
                            <h4>Details:</h4>
                            <ul>
                                ${Object.entries(notification.data.details).map(([key, value]) => 
                                    `<li><strong>${key}:</strong> ${value}</li>`
                                ).join('')}
                            </ul>
                        </div>
                    ` : ''}
                </div>
                
                <div class="notification-modal-actions">
                    ${notification.actionUrl ? `
                        <a href="${notification.actionUrl}" class="btn btn-primary">
                            ${notification.actionText || 'View Details'}
                        </a>
                    ` : ''}
                    <button class="btn btn-secondary" onclick="this.closest('.notification-modal-overlay').remove()">
                        Dismiss
                    </button>
                </div>
            </div>
        `;

        document.body.appendChild(modal);

        // Auto-remove after 30 seconds for non-critical notifications
        if (notification.priority !== this.priorityLevels.CRITICAL) {
            setTimeout(() => {
                if (modal.parentElement) {
                    modal.remove();
                }
            }, 30000);
        }
    }

    // Specialized Alert Handlers
    showPriceDropAlert(notification) {
        const property = notification.data.property;
        const oldPrice = notification.data.oldPrice;
        const newPrice = notification.data.newPrice;
        const savings = oldPrice - newPrice;

        this.showToastNotification({
            title: 'Price Drop Alert!',
            message: `${property.address} dropped by $${savings.toLocaleString()}`,
            type: 'success',
            action: {
                text: 'View Property',
                url: `/properties/${property.id}`
            },
            duration: 8000
        });
    }

    showNewListingAlert(notification) {
        const property = notification.data.property;
        
        this.showToastNotification({
            title: 'New Listing Match!',
            message: `New property matching your saved search: ${property.address}`,
            type: 'info',
            action: {
                text: 'View Property',
                url: `/properties/${property.id}`
            },
            duration: 10000
        });
    }

    showFridayReportNotification(notification) {
        this.showToastNotification({
            title: 'Friday Report Ready',
            message: 'Your weekly PropertyHub analytics report is available',
            type: 'info',
            action: {
                text: 'View Report',
                url: '/analytics/friday-report'
            },
            duration: 0 // Persistent until dismissed
        });
    }

    showTRECComplianceAlert(notification) {
        this.showModalNotification({
            ...notification,
            priority: this.priorityLevels.HIGH,
            type: 'trec_compliance'
        });
    }

    showToastNotification(toast) {
        const toastElement = document.createElement('div');
        toastElement.className = `notification-toast ${toast.type}`;
        toastElement.innerHTML = `
            <div class="toast-content">
                <h4>${toast.title}</h4>
                <p>${toast.message}</p>
                ${toast.action ? `
                    <a href="${toast.action.url}" class="toast-action">${toast.action.text}</a>
                ` : ''}
            </div>
            <button class="toast-close" onclick="this.parentElement.remove()">
                <i class="fas fa-times"></i>
            </button>
        `;

        const container = document.getElementById('toast-container') || this.createToastContainer();
        container.appendChild(toastElement);

        // Animate in
        setTimeout(() => {
            toastElement.classList.add('show');
        }, 100);

        // Auto-remove if duration is set
        if (toast.duration > 0) {
            setTimeout(() => {
                if (toastElement.parentElement) {
                    toastElement.classList.remove('show');
                    setTimeout(() => toastElement.remove(), 300);
                }
            }, toast.duration);
        }
    }

    createToastContainer() {
        const container = document.createElement('div');
        container.id = 'toast-container';
        container.className = 'toast-container';
        document.body.appendChild(container);
        return container;
    }

    // Preferences Management
    async loadUserPreferences() {
        try {
            const response = await fetch('/api/notifications/preferences', {
                headers: {
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                }
            });

            if (response.ok) {
                const prefs = await response.json();
                this.preferences = prefs;
                this.soundEnabled = prefs.soundEnabled !== false;
                this.browserNotificationsEnabled = prefs.browserNotifications && Notification.permission === 'granted';
            }
        } catch (error) {
            console.error('Error loading notification preferences:', error);
            this.loadDefaultPreferences();
        }
    }

    loadDefaultPreferences() {
        this.preferences = {
            soundEnabled: true,
            browserNotifications: true,
            emailNotifications: true,
            smsNotifications: false,
            types: {
                [this.notificationTypes.BOOKING_UPDATE]: { enabled: true, methods: ['in_app', 'email'] },
                [this.notificationTypes.PROPERTY_ALERT]: { enabled: true, methods: ['in_app', 'browser'] },
                [this.notificationTypes.PAYMENT_STATUS]: { enabled: true, methods: ['in_app', 'email'] },
                [this.notificationTypes.DEADLINE_REMINDER]: { enabled: true, methods: ['in_app', 'email', 'sms'] },
                [this.notificationTypes.SYSTEM_ALERT]: { enabled: true, methods: ['in_app', 'browser'] },
                [this.notificationTypes.FRIDAY_REPORT]: { enabled: true, methods: ['in_app', 'email'] }
            }
        };
    }

    // Notification Actions
    async markAsRead(notificationId) {
        const notification = this.notifications.find(n => n.id === notificationId);
        if (notification && !notification.read) {
            notification.read = true;
            this.unreadCount--;
            
            // Update UI
            this.updateNotificationBadge();
            const notificationEl = document.querySelector(`[data-notification-id="${notificationId}"]`);
            if (notificationEl) {
                notificationEl.classList.remove('unread');
            }

            // Notify server
            try {
                await fetch(`/api/notifications/${notificationId}/read`, {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                    }
                });
            } catch (error) {
                console.error('Error marking notification as read:', error);
            }
        }
    }

    async markAllAsRead() {
        const unreadNotifications = this.notifications.filter(n => !n.read);
        
        unreadNotifications.forEach(notification => {
            notification.read = true;
        });
        
        this.unreadCount = 0;
        this.updateNotificationBadge();
        
        // Update UI
        document.querySelectorAll('.notification-item.unread').forEach(el => {
            el.classList.remove('unread');
        });

        // Notify server
        try {
            await fetch('/api/notifications/mark-all-read', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                }
            });
        } catch (error) {
            console.error('Error marking all notifications as read:', error);
        }
    }

    async dismissNotification(notificationId) {
        this.notifications = this.notifications.filter(n => n.id !== notificationId);
        
        const notificationEl = document.querySelector(`[data-notification-id="${notificationId}"]`);
        if (notificationEl) {
            notificationEl.remove();
        }

        // Update count if it was unread
        const notification = this.notifications.find(n => n.id === notificationId);
        if (notification && !notification.read) {
            this.unreadCount--;
            this.updateNotificationBadge();
        }

        try {
            await fetch(`/api/notifications/${notificationId}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                }
            });
        } catch (error) {
            console.error('Error dismissing notification:', error);
        }
    }

    // Event Listeners
    setupEventListeners() {
        // Click outside to close notification center
        document.addEventListener('click', (e) => {
            const center = document.getElementById('notification-center');
            const trigger = document.getElementById('notification-trigger');
            
            if (center && center.classList.contains('open') &&
                !center.contains(e.target) && !trigger.contains(e.target)) {
                this.closeCenter();
            }
        });

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.closeCenter();
            }
        });

        // Page visibility changes
        document.addEventListener('visibilitychange', () => {
            if (document.visibilityState === 'visible') {
                this.syncNotifications();
            }
        });

        // Notification clicks
        document.addEventListener('click', (e) => {
            if (e.target.matches('.notification-item') || e.target.closest('.notification-item')) {
                const notificationEl = e.target.closest('.notification-item');
                const notificationId = notificationEl.dataset.notificationId;
                
                if (notificationEl.classList.contains('unread')) {
                    this.markAsRead(notificationId);
                }
            }
        });
    }

    // Utility Methods
    shouldPlaySound(notification) {
        return this.soundEnabled && 
               this.preferences.types?.[notification.type]?.soundEnabled !== false &&
               document.visibilityState !== 'visible';
    }

    shouldShowBrowserNotification(notification) {
        return this.browserNotificationsEnabled &&
               this.preferences.types?.[notification.type]?.methods?.includes('browser') &&
               document.visibilityState !== 'visible';
    }

    getNotificationIcon(type) {
        const iconMap = {
            [this.notificationTypes.BOOKING_UPDATE]: 'fas fa-calendar-check',
            [this.notificationTypes.PROPERTY_ALERT]: 'fas fa-home',
            [this.notificationTypes.PAYMENT_STATUS]: 'fas fa-credit-card',
            [this.notificationTypes.DOCUMENT_READY]: 'fas fa-file-alt',
            [this.notificationTypes.DEADLINE_REMINDER]: 'fas fa-clock',
            [this.notificationTypes.SYSTEM_ALERT]: 'fas fa-exclamation-triangle',
            [this.notificationTypes.PRICE_ALERT]: 'fas fa-tag',
            [this.notificationTypes.NEW_MESSAGE]: 'fas fa-envelope',
            [this.notificationTypes.INSPECTION_SCHEDULED]: 'fas fa-search',
            [this.notificationTypes.CLOSING_REMINDER]: 'fas fa-key',
            [this.notificationTypes.FRIDAY_REPORT]: 'fas fa-chart-bar',
            [this.notificationTypes.TREC_COMPLIANCE]: 'fas fa-shield-alt'
        };

        return iconMap[type] || 'fas fa-bell';
    }

    formatTimeAgo(timestamp) {
        const now = new Date();
        const time = new Date(timestamp);
        const diffInSeconds = Math.floor((now - time) / 1000);

        if (diffInSeconds < 60) return 'Just now';
        if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
        if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`;
        if (diffInSeconds < 2592000) return `${Math.floor(diffInSeconds / 86400)}d ago`;
        return time.toLocaleDateString();
    }

    trackNotificationReceived(notification) {
        if (window.PropertyHubAnalytics) {
            window.PropertyHubAnalytics.trackEvent('notification_received', {
                type: notification.type,
                priority: notification.priority,
                source: notification.source
            });
        }
    }

    async loadNotificationHistory() {
        try {
            const response = await fetch('/api/notifications', {
                headers: {
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                }
            });

            if (response.ok) {
                const data = await response.json();
                this.notifications = data.notifications || [];
                this.unreadCount = this.notifications.filter(n => !n.read).length;
                
                this.updateNotificationBadge();
                this.renderNotificationCenter();
            }
        } catch (error) {
            console.error('Error loading notification history:', error);
        }
    }

    startPeriodicSync() {
        // Sync notifications every 30 seconds when WebSocket is disconnected
        setInterval(() => {
            if (!this.websocket || this.websocket.readyState !== WebSocket.OPEN) {
                this.syncNotifications();
            }
        }, 30000);
    }

    async syncNotifications() {
        // Fallback sync when WebSocket is unavailable
        try {
            const lastNotificationTime = this.notifications.length > 0 ? 
                this.notifications[0].timestamp : new Date(Date.now() - 3600000).toISOString();

            const response = await fetch(`/api/notifications/sync?since=${lastNotificationTime}`, {
                headers: {
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                }
            });

            if (response.ok) {
                const data = await response.json();
                data.notifications.forEach(notification => {
                    this.processNotification(notification);
                });
            }
        } catch (error) {
            console.error('Error syncing notifications:', error);
        }
    }

    processNotificationQueue() {
        while (this.notificationQueue.length > 0) {
            const notification = this.notificationQueue.shift();
            this.processNotification(notification);
        }
    }

    // Public API Methods
    toggleCenter() {
        const center = document.getElementById('notification-center');
        if (center) {
            center.classList.toggle('open');
        }
    }

    closeCenter() {
        const center = document.getElementById('notification-center');
        if (center) {
            center.classList.remove('open');
        }
    }

    filterNotifications(filter) {
        // Update active filter button
        document.querySelectorAll('.filter-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        event.target.classList.add('active');

        // Filter notifications
        let filteredNotifications = this.notifications;

        switch (filter) {
            case 'unread':
                filteredNotifications = this.notifications.filter(n => !n.read);
                break;
            case 'bookings':
                filteredNotifications = this.notifications.filter(n => 
                    n.type === this.notificationTypes.BOOKING_UPDATE ||
                    n.type === this.notificationTypes.INSPECTION_SCHEDULED ||
                    n.type === this.notificationTypes.CLOSING_REMINDER
                );
                break;
            case 'properties':
                filteredNotifications = this.notifications.filter(n => 
                    n.type === this.notificationTypes.PROPERTY_ALERT ||
                    n.type === this.notificationTypes.PRICE_ALERT
                );
                break;
            case 'system':
                filteredNotifications = this.notifications.filter(n => 
                    n.type === this.notificationTypes.SYSTEM_ALERT ||
                    n.type === this.notificationTypes.FRIDAY_REPORT
                );
                break;
        }

        // Re-render with filtered notifications
        const notificationList = document.getElementById('notification-list');
        if (notificationList) {
            notificationList.innerHTML = filteredNotifications.map(notification => 
                this.renderNotificationItem(notification)
            ).join('');
        }
    }

    showPreferences() {
        // Show notification preferences modal
        console.log('Showing notification preferences');
    }

    // Send notification (for testing or admin use)
    async sendNotification(notification) {
        try {
            const response = await fetch('/api/notifications/send', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                },
                body: JSON.stringify(notification)
            });

            return response.ok;
        } catch (error) {
            console.error('Error sending notification:', error);
            return false;
        }
    }
}

// Initialize PropertyHub Notifications when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.PropertyHubNotifications = new PropertyHubNotifications();
});

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PropertyHubNotifications;
}
