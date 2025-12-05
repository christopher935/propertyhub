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
                    <div class="activity-icon ${this.getEventColor(activity.event_type)}">
                        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            ${this.getEventSvgPath(activity.event_type)}
                        </svg>
                    </div>
                    <div class="activity-content">
                        <div class="activity-title">${description}</div>
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

    getEventSvgPath(eventType) {
        const paths = {
            'viewed': '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>',
            'property_viewed': '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"></path>',
            'saved': '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"></path>',
            'search': '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>',
            'inquiry': '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path>',
            'booking': '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path>'
        };
        return paths[eventType] || '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>';
    },

    getEventColor(eventType) {
        const colors = {
            'viewed': 'viewed',
            'property_viewed': 'viewed',
            'saved': 'saved',
            'search': 'viewed',
            'inquiry': 'inquiry',
            'booking': 'booking'
        };
        return colors[eventType] || '';
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
