/**
 * PropertyHub Alpine.js Component Library
 * Reusable components for tiered loading, attention tracking, and AI decision making
 */

function tieredDataFetcher(apiBasePath, config = {}) {
    return {
        data: {
            live: {},
            hot: {},
            warm: {},
            daily: {},
            derived: {}
        },
        
        get metrics() {
            return {
                total_leads: this.data.warm?.total_leads || 0,
                active_properties: this.data.hot?.pending_bookings || 0,
                pending_applications: this.data.hot?.pending_applications || 0,
                cold_leads: this.data.warm?.cold_leads || 0,
                showings_this_week: this.data.live?.current_showings || 0,
                pending_approvals: this.data.hot?.pending_approvals || 0,
                active_leads: this.data.warm?.total_leads || 0,
                conversion_rate: this.data.warm?.conversion_rate || 0,
                avg_behavioral_score: this.data.hot?.behavioral_scores?.average_score || 0,
                hot_lead_count: this.data.hot?.behavioral_scores?.high_score_count || 0,
            };
        },
        
        intervals: {},
        config: {
            liveInterval: config.liveInterval || 15000,
            hotInterval: config.hotInterval || 300000,
            warmInterval: config.warmInterval || 3600000,
            derivedInterval: config.derivedInterval || 1000,
            ...config
        },
        
        async loadAllTiers() {
            Logger.log('📊 Loading all data tiers...');
            await Promise.allSettled([
                this.loadTier('live', `${apiBasePath}/live`),
                this.loadTier('hot', `${apiBasePath}/hot`),
                this.loadTier('warm', `${apiBasePath}/warm`),
                this.loadTier('daily', `${apiBasePath}/daily`)
            ]);
            Logger.log('✅ All tiers loaded');
        },
        
        async loadTier(tier, endpoint) {
            try {
                const response = await fetch(endpoint);
                if (!response.ok) {
                    Logger.warn(`⚠️ Failed to fetch ${tier}: ${response.status}`);
                    return;
                }
                const newData = await response.json();
                this.data[tier] = newData;
                if (this.onDataUpdate) this.onDataUpdate(tier, newData);
            } catch (error) {
                Logger.error(`❌ Error loading ${tier}:`, error);
            }
        },
        
        startPolling() {
            this.intervals.live = setInterval(() => this.loadTier('live', `${apiBasePath}/live`), this.config.liveInterval);
            this.intervals.hot = setInterval(() => this.loadTier('hot', `${apiBasePath}/hot`), this.config.hotInterval);
            this.intervals.warm = setInterval(() => this.loadTier('warm', `${apiBasePath}/warm`), this.config.warmInterval);
            Logger.log('✅ Tiered polling started');
        },
        
        startDerivedCalculations() {
            this.intervals.derived = setInterval(() => {
                this.data.derived = this.config.customDerivedCalculations ? 
                    this.config.customDerivedCalculations(this.data) : {};
            }, this.config.derivedInterval);
        },
        
        destroy() {
            Object.values(this.intervals).forEach(interval => clearInterval(interval));
        }
    }
}

function attentionTracker() {
    return {
        userAttention: null,
        updateQueue: [],
        shouldQueueUpdate() { return false; },
        queueUpdate(tier, data) { this.updateQueue.push({ tier, data }); },
        applyQueuedUpdates() { this.updateQueue = []; }
    }
}

function animationManager() {
    return {
        animationState: 'idle',
        async executeAction(action, callback) {
            if (callback) await callback();
        }
    }
}

function aiInsightsManager() {
    return {
        insights: [],
        topInsight: null,
        
        updateInsights() {
            const opps = this.data.hot?.top_opportunities || [];
            this.insights = opps.map(opp => ({
                id: opp.id,
                priority: opp.priority || 50,
                type: opp.type === 'hot_lead' ? 'success' : 'warning',
                message: this.formatOpportunityMessage(opp),
                actions: (opp.action_sequence || []).map(action => ({
                    label: action.description || 'Take Action',
                    endpoint: `/admin/intelligence/opportunity/${opp.id}/execute`
                }))
            }));
            this.topInsight = this.insights[0] || null;
        },
        
        formatOpportunityMessage(opp) {
            let msg = `<strong>${opp.lead_name || 'Lead'}</strong> `;
            if (opp.type === 'hot_lead') {
                msg += `is a hot lead with ${Math.round(opp.conversion_probability || 0)}% conversion probability.`;
            } else if (opp.type === 'high_view_no_contact') {
                msg += `has viewed multiple properties but hasn't been contacted.`;
            } else {
                msg += opp.context || 'requires attention.';
            }
            return msg;
        },
        
        async handleAction(action) {
            Logger.log('🎯 Executing action:', action.label);
            await this.executeAction(action, async () => {
                await this.loadTier('hot', `${this.config.apiBasePath || '/api/stats'}/hot`);
                this.updateInsights();
            });
        },
        
        handleOpportunityAction(opp) {
            Logger.log('Taking action on opportunity:', opp.lead_name);
        }
    }
}

function formatNumber(value, decimals = 0) {
    if (value === null || value === undefined || isNaN(value)) return '0';
    return Number(value).toLocaleString('en-US', {
        minimumFractionDigits: decimals,
        maximumFractionDigits: decimals
    });
}

function formatPercent(value, decimals = 1) {
    if (value === null || value === undefined || isNaN(value)) return '0%';
    return `${Math.abs(value).toFixed(decimals)}%`;
}

function formatCurrency(value, decimals = 0) {
    if (value === null || value === undefined || isNaN(value)) return '$0';
    return '$' + Number(value).toLocaleString('en-US', {
        minimumFractionDigits: decimals,
        maximumFractionDigits: decimals
    });
}

function formatTimeAgo(timestamp) {
    if (!timestamp) return 'Never';
    const now = new Date();
    const time = new Date(timestamp);
    const diffInSeconds = Math.floor((now - time) / 1000);
    if (diffInSeconds < 60) return 'Just now';
    if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
    if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`;
    return `${Math.floor(diffInSeconds / 86400)}d ago`;
}

Logger.log('✅ PropertyHub Alpine.js Component Library loaded');

// Count-up animation helper
function animateNumber(element, targetValue, duration = 800) {
    const startValue = parseInt(element.textContent) || 0;
    if (startValue === targetValue) return;
    
    const startTime = Date.now();
    const range = targetValue - startValue;
    
    function update() {
        const elapsed = Date.now() - startTime;
        const progress = Math.min(elapsed / duration, 1);
        const easeOut = 1 - Math.pow(1 - progress, 3);
        const currentValue = Math.round(startValue + (range * easeOut));
        
        element.textContent = formatNumber(currentValue);
        
        if (progress < 1) {
            requestAnimationFrame(update);
        }
    }
    
    requestAnimationFrame(update);
}
