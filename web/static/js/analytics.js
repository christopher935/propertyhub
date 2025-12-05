/**
 * PropertyHub Analytics System
 * Comprehensive analytics tracking, reporting, and business intelligence
 * Integrates with backend analytics API and generates actionable insights
 */

class PropertyHubAnalytics {
    constructor() {
        this.sessionId = this.generateSessionId();
        this.userId = null;
        this.events = [];
        this.batchSize = 50;
        this.flushInterval = 30000; // 30 seconds
        this.retryAttempts = 3;
        this.isOnline = navigator.onLine;
        this.pendingEvents = [];
        
        // Analytics configuration
        this.config = {
            trackPageViews: true,
            trackUserInteractions: true,
            trackPerformance: true,
            trackErrors: true,
            trackBusinessMetrics: true,
            enableHeatmaps: true,
            enableSessionRecording: false,
            anonymizeIP: true,
            respectDNT: true
        };

        this.init();
    }

    init() {
        console.log('📊 PropertyHub Analytics initializing...');
        
        // Check if tracking is allowed
        if (!this.isTrackingAllowed()) {
            console.log('🚫 Analytics tracking disabled by user preference');
            return;
        }

        this.setupEventListeners();
        this.startSession();
        this.setupPerformanceTracking();
        this.setupErrorTracking();
        this.setupBusinessMetricsTracking();
        this.startBatchFlush();
        
        console.log('✅ PropertyHub Analytics initialized');
    }

    // Core Tracking Methods
    trackEvent(eventName, properties = {}, options = {}) {
        if (!this.isTrackingAllowed()) return;

        const event = {
            id: this.generateEventId(),
            sessionId: this.sessionId,
            userId: this.userId,
            event: eventName,
            properties: {
                ...properties,
                timestamp: new Date().toISOString(),
                url: window.location.href,
                referrer: document.referrer,
                userAgent: navigator.userAgent,
                viewport: {
                    width: window.innerWidth,
                    height: window.innerHeight
                },
                screen: {
                    width: screen.width,
                    height: screen.height
                }
            },
            options: options
        };

        // Add user context if available
        if (window.PropertyHubAuth && window.PropertyHubAuth.user) {
            this.userId = window.PropertyHubAuth.user.id;
            event.userId = this.userId;
            event.properties.userRole = window.PropertyHubAuth.user.role;
            event.properties.userPlan = window.PropertyHubAuth.user.plan;
        }

        this.events.push(event);
        
        // Immediate flush for critical events
        if (options.immediate || this.isCriticalEvent(eventName)) {
            this.flushEvents([event]);
        }

        console.log('📈 Event tracked:', eventName, properties);
    }

    trackPageView(page = null, properties = {}) {
        const pageData = {
            page: page || window.location.pathname,
            title: document.title,
            loadTime: this.getPageLoadTime(),
            ...properties
        };

        this.trackEvent('page_view', pageData);
    }

    trackUserAction(action, target, properties = {}) {
        const actionData = {
            action: action,
            target: target,
            targetType: this.getElementType(target),
            ...properties
        };

        this.trackEvent('user_action', actionData);
    }

    // Business Metrics Tracking
    trackPropertyView(propertyId, properties = {}) {
        this.trackEvent('property_viewed', {
            propertyId: propertyId,
            ...properties
        }, { immediate: true });
    }

    trackPropertySearch(searchParams, resultCount) {
        this.trackEvent('property_search', {
            searchParams: searchParams,
            resultCount: resultCount,
            searchType: searchParams.type || 'general'
        });
    }

    trackBookingInitiated(propertyId, bookingData) {
        this.trackEvent('booking_initiated', {
            propertyId: propertyId,
            showingDate: bookingData.showingDate,
            showingTime: bookingData.showingTime,
            showingType: bookingData.showingType,
            attendeeCount: bookingData.attendeeCount
        }, { immediate: true });
    }

    trackBookingCompleted(bookingId, bookingData) {
        this.trackEvent('booking_completed', {
            bookingId: bookingId,
            propertyId: bookingData.propertyId,
            totalAmount: bookingData.totalAmount,
            paymentMethod: bookingData.paymentMethod,
            bookingSource: bookingData.source
        }, { immediate: true });
    }

    trackLeadGenerated(leadData) {
        this.trackEvent('lead_generated', {
            leadId: leadData.id,
            source: leadData.source,
            propertyId: leadData.propertyId,
            leadType: leadData.type,
            contactMethod: leadData.contactMethod
        }, { immediate: true });
    }

    trackLeadConverted(leadId, conversionData) {
        this.trackEvent('lead_converted', {
            leadId: leadId,
            conversionType: conversionData.type,
            conversionValue: conversionData.value,
            timeToConversion: conversionData.timeToConversion
        }, { immediate: true });
    }

    trackMarketDataAccess(dataType, filters = {}) {
        this.trackEvent('market_data_accessed', {
            dataType: dataType,
            filters: filters,
            source: 'HAR_MLS'
        });
    }

    trackReportGenerated(reportType, filters = {}) {
        this.trackEvent('report_generated', {
            reportType: reportType,
            filters: filters,
            exportFormat: filters.format || 'web'
        });
    }

    trackFridayReportViewed(reportDate) {
        this.trackEvent('friday_report_viewed', {
            reportDate: reportDate,
            viewedAt: new Date().toISOString()
        });
    }

    // Performance Tracking
    setupPerformanceTracking() {
        if (!this.config.trackPerformance) return;

        // Track page load performance
        window.addEventListener('load', () => {
            setTimeout(() => {
                const perfData = this.getPerformanceMetrics();
                this.trackEvent('page_performance', perfData);
            }, 0);
        });

        // Track resource loading
        this.trackResourcePerformance();
        
        // Track Core Web Vitals
        this.trackCoreWebVitals();
    }

    getPerformanceMetrics() {
        const navigation = performance.getEntriesByType('navigation')[0];
        if (!navigation) return {};

        return {
            loadTime: navigation.loadEventEnd - navigation.fetchStart,
            domContentLoaded: navigation.domContentLoadedEventEnd - navigation.fetchStart,
            firstPaint: this.getFirstPaint(),
            firstContentfulPaint: this.getFirstContentfulPaint(),
            dnsLookup: navigation.domainLookupEnd - navigation.domainLookupStart,
            tcpConnect: navigation.connectEnd - navigation.connectStart,
            serverResponse: navigation.responseEnd - navigation.requestStart,
            domProcessing: navigation.domComplete - navigation.domLoading,
            redirectCount: navigation.redirectCount,
            transferSize: navigation.transferSize,
            encodedBodySize: navigation.encodedBodySize
        };
    }

    trackResourcePerformance() {
        const observer = new PerformanceObserver((list) => {
            for (const entry of list.getEntries()) {
                if (entry.duration > 1000) { // Track slow resources (>1s)
                    this.trackEvent('slow_resource', {
                        name: entry.name,
                        duration: entry.duration,
                        size: entry.transferSize,
                        type: this.getResourceType(entry.name)
                    });
                }
            }
        });
        
        observer.observe({ entryTypes: ['resource'] });
    }

    trackCoreWebVitals() {
        // Largest Contentful Paint
        new PerformanceObserver((list) => {
            const entries = list.getEntries();
            const lastEntry = entries[entries.length - 1];
            this.trackEvent('core_web_vital', {
                metric: 'LCP',
                value: lastEntry.startTime,
                rating: this.getLCPRating(lastEntry.startTime)
            });
        }).observe({ type: 'largest-contentful-paint', buffered: true });

        // First Input Delay
        new PerformanceObserver((list) => {
            for (const entry of list.getEntries()) {
                this.trackEvent('core_web_vital', {
                    metric: 'FID',
                    value: entry.processingStart - entry.startTime,
                    rating: this.getFIDRating(entry.processingStart - entry.startTime)
                });
            }
        }).observe({ type: 'first-input', buffered: true });

        // Cumulative Layout Shift
        let clsValue = 0;
        new PerformanceObserver((list) => {
            for (const entry of list.getEntries()) {
                if (!entry.hadRecentInput) {
                    clsValue += entry.value;
                }
            }
            this.trackEvent('core_web_vital', {
                metric: 'CLS',
                value: clsValue,
                rating: this.getCLSRating(clsValue)
            });
        }).observe({ type: 'layout-shift', buffered: true });
    }

    // Error Tracking
    setupErrorTracking() {
        if (!this.config.trackErrors) return;

        // JavaScript errors
        window.addEventListener('error', (event) => {
            this.trackError({
                type: 'javascript_error',
                message: event.message,
                filename: event.filename,
                lineno: event.lineno,
                colno: event.colno,
                stack: event.error?.stack
            });
        });

        // Unhandled promise rejections
        window.addEventListener('unhandledrejection', (event) => {
            this.trackError({
                type: 'unhandled_promise_rejection',
                reason: event.reason?.toString(),
                stack: event.reason?.stack
            });
        });

        // Network errors
        this.setupNetworkErrorTracking();
    }

    trackError(errorData) {
        this.trackEvent('error_occurred', {
            ...errorData,
            url: window.location.href,
            userAgent: navigator.userAgent,
            timestamp: new Date().toISOString()
        }, { immediate: true });
    }

    setupNetworkErrorTracking() {
        const originalFetch = window.fetch;
        window.fetch = async (...args) => {
            try {
                const response = await originalFetch.apply(window, args);
                if (!response.ok) {
                    this.trackEvent('api_error', {
                        url: args[0],
                        status: response.status,
                        statusText: response.statusText,
                        method: args[1]?.method || 'GET'
                    });
                }
                return response;
            } catch (error) {
                this.trackEvent('network_error', {
                    url: args[0],
                    error: error.message,
                    method: args[1]?.method || 'GET'
                });
                throw error;
            }
        };
    }

    // Business Intelligence & Reporting
    setupBusinessMetricsTracking() {
        if (!this.config.trackBusinessMetrics) return;

        // Track form interactions
        this.setupFormTracking();
        
        // Track search behavior
        this.setupSearchTracking();
        
        // Track booking funnel
        this.setupBookingFunnelTracking();
        
        // Track user engagement
        this.setupEngagementTracking();
    }

    setupFormTracking() {
        document.addEventListener('submit', (event) => {
            const form = event.target;
            if (form.tagName === 'FORM') {
                this.trackEvent('form_submitted', {
                    formId: form.id,
                    formAction: form.action,
                    formMethod: form.method,
                    fieldCount: form.elements.length,
                    formType: this.getFormType(form)
                });
            }
        });

        // Track form abandonment
        this.setupFormAbandonmentTracking();
    }

    setupFormAbandonmentTracking() {
        const formInteractions = new Map();

        document.addEventListener('input', (event) => {
            const form = event.target.closest('form');
            if (form && !formInteractions.has(form)) {
                formInteractions.set(form, {
                    startTime: Date.now(),
                    formId: form.id,
                    formType: this.getFormType(form)
                });
            }
        });

        window.addEventListener('beforeunload', () => {
            formInteractions.forEach((data, form) => {
                const timeSpent = Date.now() - data.startTime;
                if (timeSpent > 5000) { // Only track if user spent >5s
                    this.trackEvent('form_abandoned', {
                        ...data,
                        timeSpent: timeSpent
                    }, { immediate: true });
                }
            });
        });
    }

    setupSearchTracking() {
        // Track search interactions
        document.addEventListener('input', (event) => {
            if (event.target.type === 'search' || 
                event.target.classList.contains('search-input')) {
                
                clearTimeout(this.searchTimeout);
                this.searchTimeout = setTimeout(() => {
                    this.trackEvent('search_query', {
                        query: event.target.value,
                        searchType: event.target.dataset.searchType || 'general',
                        queryLength: event.target.value.length
                    });
                }, 500);
            }
        });
    }

    setupBookingFunnelTracking() {
        // Track booking funnel steps
        const funnelSteps = [
            'property_viewed',
            'booking_form_opened',
            'booking_details_entered',
            'payment_initiated',
            'booking_completed'
        ];

        // Automatically detect funnel progression
        this.bookingFunnelData = {
            sessionStart: Date.now(),
            steps: []
        };
    }

    setupEngagementTracking() {
        let startTime = Date.now();
        let isActive = true;
        let scrollDepth = 0;

        // Track time on page
        window.addEventListener('beforeunload', () => {
            if (isActive) {
                this.trackEvent('session_duration', {
                    duration: Date.now() - startTime,
                    scrollDepth: scrollDepth,
                    page: window.location.pathname
                }, { immediate: true });
            }
        });

        // Track scroll depth
        window.addEventListener('scroll', () => {
            const currentScroll = Math.round(
                (window.scrollY / (document.body.scrollHeight - window.innerHeight)) * 100
            );
            scrollDepth = Math.max(scrollDepth, currentScroll);
        });

        // Track user activity
        let lastActivity = Date.now();
        const activityEvents = ['mousedown', 'mousemove', 'keypress', 'scroll', 'touchstart'];
        
        activityEvents.forEach(event => {
            document.addEventListener(event, () => {
                lastActivity = Date.now();
                if (!isActive) {
                    isActive = true;
                    startTime = Date.now();
                }
            });
        });

        // Check for inactivity
        setInterval(() => {
            if (Date.now() - lastActivity > 30000 && isActive) { // 30s inactivity
                isActive = false;
                this.trackEvent('user_inactive', {
                    lastActivityTime: new Date(lastActivity).toISOString(),
                    page: window.location.pathname
                });
            }
        }, 10000);
    }

    // Event Processing & Batching
    startBatchFlush() {
        setInterval(() => {
            if (this.events.length > 0) {
                this.flushEvents();
            }
        }, this.flushInterval);

        // Flush on page unload
        window.addEventListener('beforeunload', () => {
            this.flushEvents(null, true);
        });
    }

    async flushEvents(specificEvents = null, isUnloading = false) {
        const eventsToFlush = specificEvents || this.events.splice(0);
        if (eventsToFlush.length === 0) return;

        const batches = this.createBatches(eventsToFlush, this.batchSize);
        
        for (const batch of batches) {
            try {
                await this.sendEventBatch(batch, isUnloading);
            } catch (error) {
                console.error('Failed to send analytics batch:', error);
                if (!isUnloading) {
                    this.pendingEvents.push(...batch);
                }
            }
        }
    }

    async sendEventBatch(events, isUnloading = false) {
        const payload = {
            sessionId: this.sessionId,
            events: events,
            timestamp: new Date().toISOString(),
            userAgent: navigator.userAgent,
            url: window.location.href
        };

        const options = {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(payload)
        };

        // Add auth header if available
        if (window.PropertyHubAuth && window.PropertyHubAuth.token) {
            options.headers.Authorization = `Bearer ${window.PropertyHubAuth.token}`;
        }

        const endpoint = '/api/analytics/events';

        if (isUnloading && navigator.sendBeacon) {
            // Use sendBeacon for reliable delivery on page unload
            return navigator.sendBeacon(endpoint, options.body);
        } else {
            const response = await fetch(endpoint, options);
            if (!response.ok) {
                throw new Error(`Analytics API error: ${response.status}`);
            }
            return response;
        }
    }

    // A/B Testing & Experiments
    trackExperiment(experimentName, variant, properties = {}) {
        this.trackEvent('experiment_exposure', {
            experimentName: experimentName,
            variant: variant,
            ...properties
        });
    }

    trackConversion(goalName, value = null, properties = {}) {
        this.trackEvent('conversion', {
            goalName: goalName,
            value: value,
            ...properties
        }, { immediate: true });
    }

    // Utility Methods
    isTrackingAllowed() {
        // Check Do Not Track header
        if (this.config.respectDNT && navigator.doNotTrack === '1') {
            return false;
        }

        // Check user consent (GDPR/CCPA compliance)
        const consent = localStorage.getItem('analytics_consent');
        return consent === 'granted';
    }

    setTrackingConsent(granted) {
        localStorage.setItem('analytics_consent', granted ? 'granted' : 'denied');
        if (granted && !this.initialized) {
            this.init();
        }
    }

    startSession() {
        this.sessionId = this.generateSessionId();
        this.trackEvent('session_start', {
            sessionId: this.sessionId,
            referrer: document.referrer,
            utm: this.getUTMParameters()
        });
    }

    generateSessionId() {
        return 'session_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }

    generateEventId() {
        return 'event_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    }

    getUTMParameters() {
        const urlParams = new URLSearchParams(window.location.search);
        return {
            source: urlParams.get('utm_source'),
            medium: urlParams.get('utm_medium'),
            campaign: urlParams.get('utm_campaign'),
            term: urlParams.get('utm_term'),
            content: urlParams.get('utm_content')
        };
    }

    getPageLoadTime() {
        const navigation = performance.getEntriesByType('navigation')[0];
        return navigation ? navigation.loadEventEnd - navigation.fetchStart : null;
    }

    getFirstPaint() {
        const paintEntries = performance.getEntriesByType('paint');
        const fpEntry = paintEntries.find(entry => entry.name === 'first-paint');
        return fpEntry ? fpEntry.startTime : null;
    }

    getFirstContentfulPaint() {
        const paintEntries = performance.getEntriesByType('paint');
        const fcpEntry = paintEntries.find(entry => entry.name === 'first-contentful-paint');
        return fcpEntry ? fcpEntry.startTime : null;
    }

    getElementType(element) {
        if (!element) return 'unknown';
        return element.tagName?.toLowerCase() || 'text';
    }

    getFormType(form) {
        const formId = form.id?.toLowerCase() || '';
        const formAction = form.action?.toLowerCase() || '';
        
        if (formId.includes('login') || formAction.includes('login')) return 'login';
        if (formId.includes('register') || formAction.includes('register')) return 'registration';
        if (formId.includes('booking') || formAction.includes('booking')) return 'booking';
        if (formId.includes('contact') || formAction.includes('contact')) return 'contact';
        if (formId.includes('search') || formAction.includes('search')) return 'search';
        
        return 'other';
    }

    getResourceType(url) {
        const extension = url.split('.').pop()?.toLowerCase();
        const typeMap = {
            'js': 'script',
            'css': 'stylesheet',
            'png': 'image',
            'jpg': 'image',
            'jpeg': 'image',
            'gif': 'image',
            'svg': 'image',
            'woff': 'font',
            'woff2': 'font',
            'ttf': 'font'
        };
        return typeMap[extension] || 'other';
    }

    isCriticalEvent(eventName) {
        const criticalEvents = [
            'booking_completed',
            'lead_converted',
            'error_occurred',
            'payment_failed',
            'security_breach'
        ];
        return criticalEvents.includes(eventName);
    }

    createBatches(items, batchSize) {
        const batches = [];
        for (let i = 0; i < items.length; i += batchSize) {
            batches.push(items.slice(i, i + batchSize));
        }
        return batches;
    }

    getLCPRating(value) {
        if (value <= 2500) return 'good';
        if (value <= 4000) return 'needs-improvement';
        return 'poor';
    }

    getFIDRating(value) {
        if (value <= 100) return 'good';
        if (value <= 300) return 'needs-improvement';
        return 'poor';
    }

    getCLSRating(value) {
        if (value <= 0.1) return 'good';
        if (value <= 0.25) return 'needs-improvement';
        return 'poor';
    }

    // Public API for manual tracking
    track(eventName, properties) {
        this.trackEvent(eventName, properties);
    }

    identify(userId, traits) {
        this.userId = userId;
        this.trackEvent('user_identified', {
            userId: userId,
            traits: traits
        });
    }

    page(pageName, properties) {
        this.trackPageView(pageName, properties);
    }

    // Analytics Reporting Methods
    async getAnalyticsReport(reportType, dateRange, filters = {}) {
        try {
            const response = await fetch('/api/analytics/reports', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                },
                body: JSON.stringify({
                    reportType: reportType,
                    dateRange: dateRange,
                    filters: filters
                })
            });

            if (!response.ok) {
                throw new Error(`Report generation failed: ${response.status}`);
            }

            return await response.json();
        } catch (error) {
            console.error('Analytics report error:', error);
            throw error;
        }
    }

    setupEventListeners() {
        // Track page views automatically
        if (this.config.trackPageViews) {
            this.trackPageView();
            
            // Track SPA navigation
            let lastUrl = location.href;
            new MutationObserver(() => {
                const url = location.href;
                if (url !== lastUrl) {
                    lastUrl = url;
                    this.trackPageView();
                }
            }).observe(document, { subtree: true, childList: true });
        }

        // Track user interactions
        if (this.config.trackUserInteractions) {
            document.addEventListener('click', (event) => {
                this.trackUserAction('click', event.target, {
                    text: event.target.textContent?.substring(0, 100),
                    href: event.target.href,
                    className: event.target.className
                });
            });
        }

        // Online/offline status
        window.addEventListener('online', () => {
            this.isOnline = true;
            this.flushEvents(); // Flush pending events
        });

        window.addEventListener('offline', () => {
            this.isOnline = false;
        });
    }
}

// Initialize PropertyHub Analytics when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Check for consent before initializing
    const consent = localStorage.getItem('analytics_consent');
    if (consent === 'granted') {
        window.PropertyHubAnalytics = new PropertyHubAnalytics();
    } else if (consent === null) {
        // Show consent dialog if not already shown
        // This would typically be handled by your consent management system
        console.log('Analytics consent required');
    }
});

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PropertyHubAnalytics;
}
