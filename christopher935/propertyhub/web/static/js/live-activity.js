const LiveActivity = {
    updateInterval: null,
    lastUpdate: null,

    init() {
        this.loadLiveActivity();
        this.loadActiveSessions();
        this.updateInterval = setInterval(() => {
            this.loadLiveActivity();
            this.loadActiveSessions();
        }, 10000);
    },

    async loadLiveActivity() {
        try {
            const response = await fetch('/api/v1/admin/live-activity?minutes=15');
            if (!response.ok) return;
            
            const data = await response.json();
            this.renderLiveActivity(data.activities);
            this.lastUpdate = new Date(data.timestamp);
        } catch (error) {
            console.error('Failed to load live activity:', error);
        }
    },

    async loadActiveSessions() {
        try {
            const response = await fetch('/api/v1/admin/active-sessions');
            if (!response.ok) return;
            
            const data = await response.json();
            this.renderActiveSessions(data.active_sessions);
            this.updateSessionCount(data.count);
        } catch (error) {
            console.error('Failed to load active sessions:', error);
        }
    },

    renderLiveActivity(activities) {
        const container = document.getElementById('liveActivityFeed');
        if (!container) return;

        if (activities.length === 0) {
            container.innerHTML = '<p class="text-center text-gray-500 py-8">No recent activity</p>';
            return;
        }

        container.innerHTML = activities.slice(0, 20).map(activity => {
            const timeAgo = this.timeAgo(new Date(activity.timestamp));
            const icon = this.getEventIcon(activity.event_type);
            const description = this.getEventDescription(activity);

            return `
                <div class="activity-item" data-session="${activity.session_id}">
                    <div class="activity-icon ${this.getEventColor(activity.event_type)}">${icon}</div>
                    <div class="activity-content">
                        <div class="activity-description">${description}</div>
                        <div class="activity-time">${timeAgo}</div>
                    </div>
                </div>
            `;
        }).join('');
    },

    renderActiveSessions(sessions) {
        const container = document.getElementById('activeSessionsList');
        if (!container) return;

        if (sessions.length === 0) {
            container.innerHTML = '<p class="text-center text-gray-500 py-8">No active sessions</p>';
            return;
        }

        container.innerHTML = sessions.map(session => {
            const duration = Math.floor(session.duration_seconds / 60);
            const isActive = session.is_active;

            return `
                <div class="session-card ${isActive ? 'active' : ''}">
                    <div class="session-header">
                        <div class="session-id">Session ${session.session_id.substring(0, 8)}...</div>
                        ${isActive ? '<span class="badge badge-green">🟢 Active</span>' : ''}
                    </div>
                    <div class="session-stats">
                        <div class="stat">
                            <span class="stat-label">Properties Viewed</span>
                            <span class="stat-value">${session.properties_viewed}</span>
                        </div>
                        <div class="stat">
                            <span class="stat-label">Page Views</span>
                            <span class="stat-value">${session.page_views}</span>
                        </div>
                        <div class="stat">
                            <span class="stat-label">Duration</span>
                            <span class="stat-value">${duration}m</span>
                        </div>
                    </div>
                    <button class="btn btn-sm btn-outline" onclick="LiveActivity.viewSessionDetails('${session.session_id}')">
                        View Details
                    </button>
                </div>
            `;
        }).join('');
    },

    updateSessionCount(count) {
        const badges = document.querySelectorAll('.active-sessions-count');
        badges.forEach(badge => {
            badge.textContent = count;
            badge.style.display = count > 0 ? 'inline-block' : 'none';
        });
    },

    getEventIcon(eventType) {
        const icons = {
            'viewed': '👁️',
            'property_viewed': '🏠',
            'saved': '❤️',
            'search': '🔍',
            'inquiry': '✉️',
            'booking': '📅'
        };
        return icons[eventType] || '📊';
    },

    getEventColor(eventType) {
        const colors = {
            'viewed': 'activity-blue',
            'saved': 'activity-red',
            'search': 'activity-purple',
            'inquiry': 'activity-green',
            'booking': 'activity-gold'
        };
        return colors[eventType] || 'activity-gray';
    },

    getEventDescription(activity) {
        if (activity.property_address) {
            return `Viewed <strong>${activity.property_address}</strong>`;
        }
        
        switch(activity.event_type) {
            case 'search':
                return 'Searched for properties';
            case 'saved':
                return 'Saved a property';
            case 'inquiry':
                return 'Submitted inquiry';
            default:
                return activity.event_type.replace('_', ' ');
        }
    },

    timeAgo(date) {
        const seconds = Math.floor((new Date() - date) / 1000);
        
        if (seconds < 60) return 'Just now';
        if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
        if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
        return `${Math.floor(seconds / 86400)}d ago`;
    },

    async viewSessionDetails(sessionId) {
        try {
            const response = await fetch(`/api/v1/admin/session/${sessionId}`);
            if (!response.ok) return;
            
            const data = await response.json();
            this.showSessionModal(data);
        } catch (error) {
            console.error('Failed to load session details:', error);
        }
    },

    showSessionModal(data) {
        alert(`Session Details:\n\nProperties Viewed: ${data.summary.properties_viewed}\nTotal Events: ${data.summary.total_events}\nDuration: ${data.summary.duration_minutes} minutes`);
    },

    destroy() {
        if (this.updateInterval) {
            clearInterval(this.updateInterval);
        }
    }
};

if (window.location.pathname.includes('/admin')) {
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => LiveActivity.init());
    } else {
        LiveActivity.init();
    }
}
