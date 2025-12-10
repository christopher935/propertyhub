class AdminNotificationManager {
    constructor() {
        this.socket = null;
        this.notifications = [];
        this.unreadCount = 0;
        this.retryCount = 0;
        this.maxRetries = 5;
        this.retryDelay = 3000;
        this.bellElement = null;
    }

    async init() {
        Logger.log('🔔 Initializing admin notifications...');
        
        await this.loadNotifications();
        
        this.connect();
        
        this.setupEventListeners();
        
        Logger.log('✅ Admin notifications ready');
    }

    connect() {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            return;
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/api/notifications/ws`;
        
        Logger.log('🔌 Connecting to notification WebSocket...');
        
        try {
            this.socket = new WebSocket(wsUrl);
            
            this.socket.onopen = () => {
                Logger.log('✅ Notification WebSocket connected');
                this.retryCount = 0;
            };

            this.socket.onmessage = (event) => {
                const notification = JSON.parse(event.data);
                this.handleNotification(notification);
            };

            this.socket.onclose = () => {
                Logger.log('🔌 Notification WebSocket closed');
                this.scheduleReconnect();
            };

            this.socket.onerror = (error) => {
                console.error('❌ Notification WebSocket error:', error);
            };

        } catch (error) {
            console.error('❌ Failed to create WebSocket:', error);
            this.scheduleReconnect();
        }
    }

    scheduleReconnect() {
        if (this.retryCount < this.maxRetries) {
            const delay = this.retryDelay * Math.pow(2, this.retryCount);
            Logger.log(`🔄 Reconnecting in ${delay/1000}s... (attempt ${this.retryCount + 1}/${this.maxRetries})`);
            
            setTimeout(() => {
                this.retryCount++;
                this.connect();
            }, delay);
        } else {
            console.error('❌ Max reconnection attempts reached. Falling back to polling.');
            this.startPolling();
        }
    }

    startPolling() {
        setInterval(() => {
            this.loadNotifications();
        }, 30000);
    }

    async loadNotifications() {
        try {
            const response = await fetch('/api/notifications?limit=50');
            if (response.ok) {
                const data = await response.json();
                this.notifications = data.notifications || [];
                this.updateUnreadCount();
                this.updateUI();
            }
        } catch (error) {
            console.error('❌ Error loading notifications:', error);
        }
    }

    handleNotification(notification) {
        Logger.log('📢 New notification:', notification.title);
        
        this.notifications.unshift(notification);
        
        if (this.notifications.length > 100) {
            this.notifications = this.notifications.slice(0, 100);
        }
        
        this.updateUnreadCount();
        this.updateUI();
        
        this.showToast(notification);
        
        if (notification.priority === 'high' || notification.priority === 'urgent') {
            this.playSound();
        }
        
        if ('Notification' in window && Notification.permission === 'granted' && document.hidden) {
            new Notification(notification.title, {
                body: notification.message,
                icon: '/static/images/logo-notification.png',
                tag: notification.id,
            });
        }
    }

    updateUnreadCount() {
        this.unreadCount = this.notifications.filter(n => !n.read && !n.read_at).length;
    }

    updateUI() {
        if (this.bellElement) {
            const event = new CustomEvent('notification-update', {
                detail: {
                    notifications: this.notifications.slice(0, 5),
                    unreadCount: this.unreadCount
                }
            });
            this.bellElement.dispatchEvent(event);
        }

        if (window.Alpine && this.bellElement) {
            const alpineData = Alpine.$data(this.bellElement);
            if (alpineData) {
                alpineData.notifications = this.notifications.slice(0, 5);
                alpineData.unreadCount = this.unreadCount;
            }
        }
    }

    showToast(notification) {
        const toast = document.createElement('div');
        toast.className = `admin-notification-toast priority-${notification.priority}`;
        toast.innerHTML = `
            <div class="toast-icon type-${notification.type}"></div>
            <div class="toast-content">
                <div class="toast-title">${this.escapeHtml(notification.title)}</div>
                <div class="toast-message">${this.escapeHtml(notification.message)}</div>
            </div>
            <button class="toast-close" onclick="this.parentElement.remove()">
                <svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                </svg>
            </button>
        `;

        const container = this.getToastContainer();
        container.appendChild(toast);

        setTimeout(() => {
            toast.classList.add('show');
        }, 100);

        setTimeout(() => {
            toast.classList.remove('show');
            setTimeout(() => toast.remove(), 300);
        }, 5000);
    }

    getToastContainer() {
        let container = document.getElementById('admin-toast-container');
        if (!container) {
            container = document.createElement('div');
            container.id = 'admin-toast-container';
            container.className = 'admin-toast-container';
            document.body.appendChild(container);
        }
        return container;
    }

    playSound() {
        try {
            const audio = new Audio('/static/sounds/notification.mp3');
            audio.volume = 0.3;
            audio.play().catch(() => {});
        } catch (error) {}
    }

    async markAsRead(id) {
        try {
            const response = await fetch(`/api/notifications/${id}/read`, {
                method: 'PUT'
            });

            if (response.ok) {
                const notification = this.notifications.find(n => n.id === id);
                if (notification) {
                    notification.read = true;
                    notification.read_at = new Date().toISOString();
                }
                this.updateUnreadCount();
                this.updateUI();
            }
        } catch (error) {
            console.error('❌ Error marking notification as read:', error);
        }
    }

    async markAllAsRead() {
        try {
            const response = await fetch('/api/notifications/read-all', {
                method: 'PUT'
            });

            if (response.ok) {
                this.notifications.forEach(n => {
                    n.read = true;
                    n.read_at = new Date().toISOString();
                });
                this.updateUnreadCount();
                this.updateUI();
            }
        } catch (error) {
            console.error('❌ Error marking all as read:', error);
        }
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

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    setupEventListeners() {
        document.addEventListener('DOMContentLoaded', () => {
            this.bellElement = document.querySelector('.notification-bell');
            
            if (this.bellElement) {
                this.bellElement.addEventListener('notification-update', () => {});
            }
        });

        if (document.readyState === 'complete') {
            this.bellElement = document.querySelector('.notification-bell');
        }

        if ('Notification' in window && Notification.permission === 'default') {
            Notification.requestPermission();
        }
    }
}

if (typeof window !== 'undefined') {
    window.adminNotifications = new AdminNotificationManager();
    
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => {
            window.adminNotifications.init();
        });
    } else {
        window.adminNotifications.init();
    }
}
