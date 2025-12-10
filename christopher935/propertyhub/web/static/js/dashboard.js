function dashboardData() {
    return {
        chartColors: {
            primary: '#1b3559',
            secondary: '#c4a962',
            success: '#10B981',
            warning: '#F59E0B',
            error: '#EF4444',
            info: '#3B82F6'
        },
        
        criticalStats: {
            pendingBookings: 0,
            pendingApplications: 0,
            alerts: 0
        },
        
        keyMetrics: {
            activeLeads: 0,
            activeLeadsTrend: 0,
            conversionRate: 0,
            revenueMTD: 0,
            revenueTrend: 0,
            bookingsWeek: 0,
            confirmedBookings: 0
        },
        
        opportunities: [],
        recentActivity: [],
        
        systemHealth: {
            status: 'healthy',
            uptime_percent: 99.9,
            avg_response_time_ms: 120,
            error_rate_percent: 0.1
        },
        
        sidebarCounts: {
            total_leads: 0,
            hot_leads: 0,
            warm_leads: 0,
            active_properties: 0,
            pending_applications: 0,
            confirmed_bookings: 0,
            closing_pipeline: 0,
            pending_images: 0
        },
        
        userProfile: {
            username: 'Loading...',
            role: 'Loading...',
            initials: 'U'
        },
        
        loadingOpportunities: true,
        searchQuery: '',
        bookingTrendsChart: null,
        leadSourceChart: null,
        
        async init() {
            await this.fetchUserProfile();
            await Promise.all([
                this.fetchCriticalStats(),
                this.fetchKeyMetrics(),
                this.fetchOpportunities(),
                this.fetchRecentActivity(),
                this.fetchChartData(),
                this.fetchSystemHealth()
            ]);
            this.renderCharts();
        },
        
        async fetchUserProfile() {
            try {
                const response = await fetch('/api/admin/settings/profile');
                const data = await response.json();
                this.userProfile = {
                    username: data.username || 'User',
                    role: data.role || 'Unknown',
                    initials: this.getUserInitials(data.username)
                };
            } catch (error) {
                Logger.error('Failed to fetch user profile:', error);
                this.userProfile = { username: 'User', role: 'Unknown', initials: 'U' };
            }
        },
        
        getUserInitials(name) {
            if (!name) return 'U';
            return name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2);
        },
        
        async fetchCriticalStats() {
            try {
                const response = await fetch('/api/stats/critical');
                const data = await response.json();
                this.criticalStats = {
                    pendingBookings: data.pending_bookings || 0,
                    pendingApplications: data.pending_applications || 0,
                    alerts: data.system_alerts || 0
                };
                this.sidebarCounts.pending_applications = data.pending_applications || 0;
            } catch (error) {
                Logger.error('Failed to fetch critical stats:', error);
            }
        },
        
        async fetchKeyMetrics() {
            try {
                const response = await fetch('/api/stats/key-metrics');
                const data = await response.json();
                this.keyMetrics = {
                    activeLeads: data.active_leads || 0,
                    activeLeadsTrend: data.active_leads_trend || 0,
                    conversionRate: data.conversion_rate || 0,
                    revenueMTD: data.revenue_mtd || 0,
                    revenueTrend: data.revenue_trend || 0,
                    bookingsWeek: data.bookings_week || 0,
                    confirmedBookings: data.confirmed_bookings || 0
                };
                this.sidebarCounts.total_leads = data.active_leads || 0;
                this.sidebarCounts.hot_leads = data.hot_leads || 0;
                this.sidebarCounts.warm_leads = data.warm_leads || 0;
                this.sidebarCounts.confirmed_bookings = data.bookings_week || 0;
            } catch (error) {
                Logger.error('Failed to fetch key metrics:', error);
            }
        },
        
        async fetchOpportunities() {
            this.loadingOpportunities = true;
            try {
                const response = await fetch('/api/stats/opportunities?limit=4');
                const data = await response.json();
                this.opportunities = data.opportunities || [];
            } catch (error) {
                Logger.error('Failed to fetch opportunities:', error);
            } finally {
                this.loadingOpportunities = false;
            }
        },
        
        async fetchRecentActivity() {
            try {
                const response = await fetch('/api/stats/activity?limit=5');
                const data = await response.json();
                this.recentActivity = data.activities || [];
            } catch (error) {
                Logger.error('Failed to fetch recent activity:', error);
            }
        },
        
        async fetchChartData() {
            try {
                const [bookingResponse, leadSourceResponse] = await Promise.all([
                    fetch('/api/charts/bookings'),
                    fetch('/api/charts/lead-sources')
                ]);
                
                this.bookingTrendsData = await bookingResponse.json();
                this.leadSourceData = await leadSourceResponse.json();
            } catch (error) {
                Logger.error('Failed to fetch chart data:', error);
            }
        },
        
        async fetchSystemHealth() {
            try {
                const response = await fetch('/api/admin/system/health');
                const data = await response.json();
                this.systemHealth = {
                    status: data.status || 'healthy',
                    uptime_percent: data.uptime_percent || 99.9,
                    avg_response_time_ms: data.avg_response_time_ms || 120,
                    error_rate_percent: data.error_rate_percent || 0.1
                };
                this.sidebarCounts.active_properties = data.active_properties || 0;
                this.sidebarCounts.pending_images = data.pending_images || 0;
                this.sidebarCounts.closing_pipeline = data.closing_pipeline || 0;
            } catch (error) {
                Logger.error('Failed to fetch system health:', error);
            }
        },
        
        renderCharts() {
            this.$nextTick(() => {
                this.renderBookingTrendsChart();
                this.renderLeadSourceChart();
            });
        },
        
        renderBookingTrendsChart() {
            const ctx = this.$refs.bookingTrendsChart;
            if (!ctx) return;
            
            if (this.bookingTrendsChart) {
                this.bookingTrendsChart.destroy();
            }
            
            this.bookingTrendsChart = new Chart(ctx, {
                type: 'line',
                data: {
                    labels: this.bookingTrendsData?.labels || ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
                    datasets: [{
                        label: 'Bookings',
                        data: this.bookingTrendsData?.data || [12, 19, 15, 25, 22, 18, 20],
                        borderColor: this.chartColors.primary,
                        backgroundColor: this.chartColors.primary + '20',
                        tension: 0.4,
                        fill: true
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: true,
                    plugins: {
                        legend: { display: false }
                    },
                    scales: {
                        y: { beginAtZero: true }
                    }
                }
            });
        },
        
        renderLeadSourceChart() {
            const ctx = this.$refs.leadSourceChart;
            if (!ctx) return;
            
            if (this.leadSourceChart) {
                this.leadSourceChart.destroy();
            }
            
            this.leadSourceChart = new Chart(ctx, {
                type: 'doughnut',
                data: {
                    labels: this.leadSourceData?.labels || ['Website', 'Referral', 'Social Media', 'Direct'],
                    datasets: [{
                        data: this.leadSourceData?.data || [45, 25, 20, 10],
                        backgroundColor: [
                            this.chartColors.primary,
                            this.chartColors.secondary,
                            this.chartColors.info,
                            this.chartColors.success
                        ]
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: true,
                    plugins: {
                        legend: {
                            position: 'bottom'
                        }
                    }
                }
            });
        },
        
        get systemHealthClass() {
            if (this.systemHealth.status === 'healthy') return 'healthy';
            if (this.systemHealth.status === 'warning') return 'warning';
            return 'critical';
        },
        
        get systemHealthText() {
            if (this.systemHealth.status === 'healthy') return 'All Systems Operational';
            if (this.systemHealth.status === 'warning') return 'Minor Issues Detected';
            return 'Critical Issues Detected';
        },
        
        async executeAction(oppId, action, email) {
            Logger.log('Executing action:', action, 'for opportunity:', oppId);
        },
        
        getActivityColor(type) {
            const colors = {
                'lead': this.chartColors.info,
                'booking': this.chartColors.success,
                'application': this.chartColors.secondary,
                'message': this.chartColors.primary
            };
            return colors[type] || this.chartColors.info;
        },
        
        getActivityIconPath(type) {
            const paths = {
                'lead': 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z',
                'booking': 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z',
                'application': 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
                'message': 'M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z'
            };
            return paths[type] || paths.lead;
        },
        
        formatTimeAgo(timestamp) {
            const now = new Date();
            const past = new Date(timestamp);
            const diffInSeconds = Math.floor((now - past) / 1000);
            
            if (diffInSeconds < 60) return 'Just now';
            if (diffInSeconds < 3600) return Math.floor(diffInSeconds / 60) + 'm ago';
            if (diffInSeconds < 86400) return Math.floor(diffInSeconds / 3600) + 'h ago';
            return Math.floor(diffInSeconds / 86400) + 'd ago';
        },
        
        formatCurrency(value) {
            return new Intl.NumberFormat('en-US', {
                style: 'currency',
                currency: 'USD',
                minimumFractionDigits: 0,
                maximumFractionDigits: 0
            }).format(value);
        },
        
        performSearch() {
            Logger.log('Searching for:', this.searchQuery);
        }
    };
}

function statCard(metricName) {
    return {
        displayValue: 0,
        targetValue: 0,
        
        init() {
            this.$watch(() => {
                const parent = this.$el.closest('[x-data*="dashboardData"]');
                if (!parent) return 0;
                const data = Alpine.$data(parent);
                
                switch(metricName) {
                    case 'active_leads': return data.keyMetrics.activeLeads;
                    case 'conversion_rate': return data.keyMetrics.conversionRate;
                    case 'revenue_mtd': return data.keyMetrics.revenueMTD;
                    case 'bookings_week': return data.keyMetrics.bookingsWeek;
                    default: return 0;
                }
            }, (newValue) => {
                this.targetValue = newValue;
                this.animateValue();
            });
        },
        
        animateValue() {
            const duration = 1000;
            const start = this.displayValue;
            const end = this.targetValue;
            const startTime = performance.now();
            
            const animate = (currentTime) => {
                const elapsed = currentTime - startTime;
                const progress = Math.min(elapsed / duration, 1);
                
                this.displayValue = start + (end - start) * this.easeOutQuart(progress);
                
                if (progress < 1) {
                    requestAnimationFrame(animate);
                }
            };
            
            requestAnimationFrame(animate);
        },
        
        easeOutQuart(x) {
            return 1 - Math.pow(1 - x, 4);
        }
    };
}

function visitorCounter() {
    return {
        count: 0,
        trend: 0,
        byPage: {},
        hotCount: 0,
        returningCount: 0,
        justChanged: false,
        showBreakdown: false,
        ws: null,
        reconnectAttempts: 0,
        maxReconnectAttempts: 5,
        
        init() {
            this.fetchCount();
            this.connectWebSocket();
        },
        
        connectWebSocket() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${window.location.host}/api/ws`;
            
            try {
                this.ws = new WebSocket(wsUrl);
                
                this.ws.onopen = () => {
                    Logger.log('WebSocket connected');
                    this.reconnectAttempts = 0;
                };
                
                this.ws.onmessage = (event) => {
                    try {
                        const message = JSON.parse(event.data);
                        if (message.type === 'visitor_count') {
                            this.updateCount(message.data);
                        }
                    } catch (e) {
                        Logger.error('Error parsing WebSocket message:', e);
                    }
                };
                
                this.ws.onerror = (error) => {
                    Logger.error('WebSocket error:', error);
                };
                
                this.ws.onclose = () => {
                    Logger.log('WebSocket disconnected');
                    if (this.reconnectAttempts < this.maxReconnectAttempts) {
                        this.reconnectAttempts++;
                        const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
                        setTimeout(() => this.connectWebSocket(), delay);
                    }
                };
            } catch (e) {
                Logger.error('Error creating WebSocket:', e);
            }
        },
        
        updateCount(data) {
            const oldCount = this.count;
            this.count = data.count || 0;
            this.trend = data.trend || 0;
            this.byPage = data.by_page || {};
            this.hotCount = data.hot_count || 0;
            this.returningCount = data.returning_count || 0;
            
            if (oldCount !== this.count) {
                this.justChanged = true;
                setTimeout(() => this.justChanged = false, 1000);
            }
        },
        
        async fetchCount() {
            try {
                const res = await fetch('/api/stats/live');
                const data = await res.json();
                this.updateCount({
                    count: data.active_visitors || 0,
                    trend: data.visitors_trend || 0,
                    by_page: data.visitors_by_page || {},
                    hot_count: data.hot_visitors || 0,
                    returning_count: data.returning_visitors || 0
                });
            } catch (error) {
                Logger.error('Failed to fetch visitor count:', error);
            }
        }
    };
}

function toastManager() {
    return {
        toasts: [],
        nextId: 1,
        
        add(message, type = 'info', duration = 5000) {
            const id = this.nextId++;
            const toast = { id, message, type, visible: true };
            this.toasts.push(toast);
            
            setTimeout(() => {
                this.remove(id);
            }, duration);
        },
        
        remove(id) {
            const index = this.toasts.findIndex(t => t.id === id);
            if (index !== -1) {
                this.toasts[index].visible = false;
                setTimeout(() => {
                    this.toasts.splice(index, 1);
                }, 300);
            }
        },
        
        getToastIcon(type) {
            const icons = {
                success: '✓',
                error: '✕',
                warning: '⚠',
                info: 'ℹ'
            };
            return icons[type] || icons.info;
        }
    };
}

function liveSessionsPanel() {
    return {
        sessions: [],
        loading: true,
        selectedSession: null,
        showJourneyModal: false,
        journeyData: null,
        updateInterval: null,
        
        async init() {
            await this.fetchSessions();
            this.updateInterval = setInterval(() => {
                this.fetchSessions();
            }, 10000);
        },
        
        async fetchSessions() {
            try {
                const response = await fetch('/api/admin/sessions/active');
                const data = await response.json();
                
                const newSessionIds = data.sessions.map(s => s.session_id);
                const oldSessionIds = this.sessions.map(s => s.session_id);
                
                this.sessions = data.sessions.map(session => {
                    const isNew = !oldSessionIds.includes(session.session_id);
                    return { ...session, isNew };
                });
                
                this.loading = false;
            } catch (error) {
                Logger.error('Failed to fetch sessions:', error);
                this.loading = false;
            }
        },
        
        async viewJourney(sessionId) {
            try {
                const response = await fetch(`/api/admin/sessions/${sessionId}/journey`);
                const data = await response.json();
                this.journeyData = data;
                this.showJourneyModal = true;
            } catch (error) {
                Logger.error('Failed to fetch session journey:', error);
            }
        },
        
        closeJourneyModal() {
            this.showJourneyModal = false;
            this.journeyData = null;
        },
        
        identifyLead(session) {
            Logger.log('Identify lead for session:', session.session_id);
        },
        
        sendProperty(session) {
            Logger.log('Send property to:', session.lead_email);
        },
        
        getStatusIcon(category) {
            const icons = {
                'hot': '🔴',
                'warm': '🟡',
                'cold': '🔵'
            };
            return icons[category] || '🔵';
        },
        
        getCategoryLabel(category) {
            const labels = {
                'hot': 'Hot',
                'warm': 'Warm',
                'cold': 'Cold'
            };
            return labels[category] || 'Cold';
        },
        
        getLocationString(location) {
            if (!location) return '';
            const parts = [];
            if (location.city) parts.push(location.city);
            if (location.state) parts.push(location.state);
            return parts.join(', ');
        },
        
        formatDuration(seconds) {
            if (seconds < 60) return `${seconds}s`;
            const minutes = Math.floor(seconds / 60);
            const remainingSeconds = seconds % 60;
            return `${minutes}m ${remainingSeconds}s`;
        },
        
        formatTime(timestamp) {
            const date = new Date(timestamp);
            return date.toLocaleTimeString('en-US', { 
                hour: '2-digit', 
                minute: '2-digit' 
            });
        },
        
        formatEventType(eventType) {
            const types = {
                'viewed': 'Viewed Property',
                'saved': 'Saved Property',
                'inquired': 'Sent Inquiry',
                'applied': 'Submitted Application',
                'searched': 'Searched Properties',
                'session_start': 'Started Session'
            };
            return types[eventType] || eventType;
        },
        
        destroy() {
            if (this.updateInterval) {
                clearInterval(this.updateInterval);
            }
        }
    };
}

function liveActivityFeed() {
    return {
        ws: null,
        events: [],
        activeCount: 0,
        reconnectAttempts: 0,
        maxReconnects: 5,
        reconnectDelay: 1000,
        
        init() {
            this.connect();
        },
        
        connect() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${window.location.host}/api/ws/admin/activity`;
            
            try {
                this.ws = new WebSocket(wsUrl);
                
                this.ws.onopen = () => {
                    Logger.log('✅ Live activity WebSocket connected');
                    this.reconnectAttempts = 0;
                };
                
                this.ws.onmessage = (event) => {
                    try {
                        const message = JSON.parse(event.data);
                        this.handleMessage(message);
                    } catch (e) {
                        Logger.error('Error parsing WebSocket message:', e);
                    }
                };
                
                this.ws.onerror = (error) => {
                    Logger.error('WebSocket error:', error);
                };
                
                this.ws.onclose = () => {
                    Logger.log('WebSocket disconnected');
                    this.reconnect();
                };
            } catch (e) {
                Logger.error('Error creating WebSocket:', e);
                this.reconnect();
            }
        },
        
        handleMessage(message) {
            if (message.type === 'activity_event') {
                this.handleActivityEvent(message.data.event);
            } else if (message.type === 'active_count') {
                this.activeCount = message.data.count || 0;
            }
        },
        
        handleActivityEvent(event) {
            this.events.unshift(event);
            
            if (this.events.length > 50) {
                this.events = this.events.slice(0, 50);
            }
            
            if (event.score >= 70) {
                this.showHighValueNotification(event);
            }
        },
        
        showHighValueNotification(event) {
            if ('Notification' in window && Notification.permission === 'granted') {
                new Notification('High-Value Lead Activity!', {
                    body: event.details,
                    icon: '/static/images/logo.png',
                    tag: 'high-value-lead'
                });
            }
            
            const toastManager = Alpine.$data(document.querySelector('[x-data*="toastManager"]'));
            if (toastManager) {
                toastManager.add(`🔥 ${event.details}`, 'warning', 10000);
            }
        },
        
        reconnect() {
            if (this.reconnectAttempts >= this.maxReconnects) {
                Logger.error('Max reconnection attempts reached');
                return;
            }
            
            this.reconnectAttempts++;
            const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
            
            Logger.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts}/${this.maxReconnects})`);
            
            setTimeout(() => {
                this.connect();
            }, delay);
        },
        
        getActivityIcon(type) {
            const icons = {
                'property_view': '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/></svg>',
                'property_save': '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"/></svg>',
                'inquiry': '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/></svg>',
                'search': '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/></svg>',
                'application': '<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/></svg>'
            };
            return icons[type] || icons['property_view'];
        },
        
        formatTimeAgo(timestamp) {
            const now = new Date();
            const past = new Date(timestamp);
            const diffInSeconds = Math.floor((now - past) / 1000);
            
            if (diffInSeconds < 5) return 'Just now';
            if (diffInSeconds < 60) return `${diffInSeconds}s ago`;
            if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
            if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`;
            return `${Math.floor(diffInSeconds / 86400)}d ago`;
        },
        
        destroy() {
            if (this.ws) {
                this.ws.close();
                this.ws = null;
            }

function sidebarToggle() {
    return {
        sidebarOpen: false,
        
        init() {
            this.$watch('sidebarOpen', (value) => {
                if (value) {
                    document.body.style.overflow = 'hidden';
                } else {
                    document.body.style.overflow = '';
                }
            });
            
            window.addEventListener('resize', () => {
                if (window.innerWidth >= 768) {
                    this.sidebarOpen = false;
                    document.body.style.overflow = '';
                }
            });
        },
        
        toggleSidebar() {
            this.sidebarOpen = !this.sidebarOpen;
        },
        
        closeSidebar() {
            this.sidebarOpen = false;
        }
    };
}
