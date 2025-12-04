/**
 * PropertyHub Cookie Consent Banner
 * Uses Scout CSS components + minimal custom styles
 * Just add: <script src="/static/js/cookie-consent.js"></script> before </body>
 */

(function() {
    'use strict';

    // Inject ONLY custom CSS (things Scout doesn't have)
    const style = document.createElement('style');
    style.textContent = `
        /* Cookie Banner - Custom component (not in Scout) */
        .ph-cookie-banner {
            position: fixed;
            bottom: 0;
            left: 0;
            right: 0;
            background: linear-gradient(135deg, var(--navy-primary) 0%, var(--navy-dark) 100%);
            color: white;
            padding: var(--space-6);
            box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.2);
            z-index: 10000;
            animation: slideUp 0.4s ease-out;
            display: none;
        }

        .ph-cookie-banner.show {
            display: block;
        }

        @keyframes slideUp {
            from { transform: translateY(100%); }
            to { transform: translateY(0); }
        }

        .ph-cookie-content {
            max-width: var(--container-xl);
            margin: 0 auto;
            display: flex;
            justify-content: space-between;
            align-items: center;
            gap: var(--space-8);
        }

        .ph-cookie-text {
            display: flex;
            gap: var(--space-4);
            align-items: flex-start;
            flex: 1;
        }

        .ph-cookie-icon {
            font-size: 2rem;
            flex-shrink: 0;
        }

        .ph-cookie-title {
            font-size: 1.125rem;
            font-weight: var(--font-weight-bold);
            margin: 0 0 var(--space-2) 0;
            color: white;
        }

        .ph-cookie-description {
            font-size: var(--font-size-sm);
            line-height: 1.5;
            margin: 0;
            color: rgba(255, 255, 255, 0.9);
        }

        .ph-cookie-link {
            color: var(--gold-primary);
            text-decoration: underline;
            font-weight: var(--font-weight-semibold);
        }

        .ph-cookie-link:hover {
            color: var(--gold-light);
        }

        .ph-cookie-actions {
            display: flex;
            gap: var(--space-3);
            flex-shrink: 0;
        }

        /* Override btn colors for banner context */
        .ph-cookie-banner .btn-primary {
            background: linear-gradient(135deg, var(--gold-primary) 0%, var(--gold-dark) 100%);
            color: white;
            border: none;
        }

        .ph-cookie-banner .btn-primary:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(194, 168, 117, 0.4);
        }

        .ph-cookie-banner .btn-secondary {
            background: transparent;
            color: white;
            border: 2px solid white;
        }

        .ph-cookie-banner .btn-secondary:hover {
            background: rgba(255, 255, 255, 0.1);
            color: white;
        }

        .ph-cookie-banner .btn-settings {
            background: rgba(255, 255, 255, 0.1);
            color: white;
            border: 1px solid rgba(255, 255, 255, 0.3);
            padding: var(--space-3) var(--space-6);
            border-radius: var(--radius-lg);
            font-weight: var(--font-weight-semibold);
            font-size: var(--font-size-sm);
            cursor: pointer;
            transition: all 0.2s;
            white-space: nowrap;
        }

        .ph-cookie-banner .btn-settings:hover {
            background: rgba(255, 255, 255, 0.2);
        }

        /* Modal - Custom component (not in Scout) */
        .ph-cookie-modal {
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            z-index: 10001;
            display: none;
            align-items: center;
            justify-content: center;
            animation: fadeIn 0.3s ease-out;
        }

        .ph-cookie-modal.show {
            display: flex;
        }

        @keyframes fadeIn {
            from { opacity: 0; }
            to { opacity: 1; }
        }

        .ph-modal-overlay {
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: rgba(0, 0, 0, 0.7);
        }

        .ph-modal-content {
            position: relative;
            background: white;
            border-radius: var(--radius-xl);
            max-width: 700px;
            width: 90%;
            max-height: 90vh;
            overflow-y: auto;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            animation: slideIn 0.3s ease-out;
        }

        @keyframes slideIn {
            from { transform: translateY(-50px); opacity: 0; }
            to { transform: translateY(0); opacity: 1; }
        }

        .ph-modal-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: var(--space-6);
            border-bottom: 2px solid var(--gray-200);
        }

        .ph-modal-header h2 {
            margin: 0;
            font-size: 1.5rem;
            color: var(--navy-primary);
            font-weight: var(--font-weight-bold);
        }

        .ph-modal-close {
            background: none;
            border: none;
            font-size: 2rem;
            color: var(--gray-500);
            cursor: pointer;
            padding: 0;
            width: 32px;
            height: 32px;
            display: flex;
            align-items: center;
            justify-content: center;
            border-radius: var(--radius-md);
        }

        .ph-modal-close:hover {
            background: var(--gray-100);
            color: var(--navy-primary);
        }

        .ph-modal-body {
            padding: var(--space-6);
        }

        .ph-modal-intro {
            color: var(--gray-600);
            margin-bottom: var(--space-6);
            line-height: 1.6;
        }

        .ph-cookie-category {
            background: var(--gray-50);
            border: 1px solid var(--gray-200);
            border-radius: var(--radius-lg);
            padding: var(--space-5);
            margin-bottom: var(--space-4);
        }

        .ph-category-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: var(--space-4);
        }

        .ph-category-header h3 {
            margin: 0 0 var(--space-1) 0;
            font-size: 1rem;
            color: var(--navy-primary);
            font-weight: var(--font-weight-bold);
        }

        .ph-category-header p {
            margin: 0;
            font-size: var(--font-size-sm);
            color: var(--gray-600);
        }

        .ph-category-details {
            font-size: var(--font-size-sm);
            color: var(--gray-700);
            line-height: 1.6;
            padding-top: var(--space-4);
            border-top: 1px solid var(--gray-200);
        }

        .ph-category-details strong {
            color: var(--navy-primary);
        }

        /* Toggle Switch - Custom component (not in Scout) */
        .ph-toggle {
            position: relative;
            display: inline-block;
            width: 50px;
            height: 26px;
            flex-shrink: 0;
        }

        .ph-toggle input {
            opacity: 0;
            width: 0;
            height: 0;
        }

        .ph-toggle-slider {
            position: absolute;
            cursor: pointer;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background-color: var(--gray-300);
            transition: 0.3s;
            border-radius: 26px;
        }

        .ph-toggle-slider:before {
            position: absolute;
            content: "";
            height: 20px;
            width: 20px;
            left: 3px;
            bottom: 3px;
            background-color: white;
            transition: 0.3s;
            border-radius: 50%;
        }

        .ph-toggle input:checked + .ph-toggle-slider {
            background-color: #10b981;
        }

        .ph-toggle input:checked + .ph-toggle-slider:before {
            transform: translateX(24px);
        }

        .ph-toggle input:disabled + .ph-toggle-slider {
            opacity: 0.5;
            cursor: not-allowed;
        }

        .ph-modal-footer {
            padding: var(--space-6);
            border-top: 2px solid var(--gray-200);
            display: flex;
            gap: var(--space-4);
            justify-content: flex-end;
        }

        /* Responsive */
        @media (max-width: 768px) {
            .ph-cookie-content {
                flex-direction: column;
                gap: var(--space-4);
            }

            .ph-cookie-actions {
                width: 100%;
                flex-direction: column;
            }

            .ph-cookie-actions button {
                width: 100%;
            }

            .ph-modal-content {
                width: 95%;
                max-height: 95vh;
            }

            .ph-modal-footer {
                flex-direction: column;
            }

            .ph-modal-footer button {
                width: 100%;
            }
        }
    `;
    document.head.appendChild(style);

    // Inject HTML (using Scout CSS button classes)
    const bannerHTML = `
        <div id="phCookieBanner" class="ph-cookie-banner">
            <div class="ph-cookie-content">
                <div class="ph-cookie-text">
                    <div class="ph-cookie-icon">🍪</div>
                    <div>
                        <h3 class="ph-cookie-title">We Value Your Privacy</h3>
                        <p class="ph-cookie-description">
                            We use cookies and tracking technologies to improve your experience, analyze site usage, 
                            and help our agents provide better service through behavioral intelligence. 
                            <a href="/privacy-policy" class="ph-cookie-link">Learn more</a>
                        </p>
                    </div>
                </div>
                <div class="ph-cookie-actions">
                    <button onclick="phCookieConsent.acceptAll()" class="btn-primary">
                        Accept All
                    </button>
                    <button onclick="phCookieConsent.acceptEssential()" class="btn-secondary">
                        Essential Only
                    </button>
                    <button onclick="phCookieConsent.showSettings()" class="btn-settings">
                        Customize
                    </button>
                </div>
            </div>
        </div>

        <div id="phCookieModal" class="ph-cookie-modal">
            <div class="ph-modal-overlay" onclick="phCookieConsent.closeSettings()"></div>
            <div class="ph-modal-content">
                <div class="ph-modal-header">
                    <h2>Cookie Preferences</h2>
                    <button onclick="phCookieConsent.closeSettings()" class="ph-modal-close">&times;</button>
                </div>
                
                <div class="ph-modal-body">
                    <p class="ph-modal-intro">
                        We use different types of cookies to optimize your experience. You can choose which categories to allow.
                    </p>

                    <div class="ph-cookie-category">
                        <div class="ph-category-header">
                            <div>
                                <h3>Essential Cookies</h3>
                                <p>Required for the website to function. Cannot be disabled.</p>
                            </div>
                            <label class="ph-toggle">
                                <input type="checkbox" id="phCookieEssential" checked disabled>
                                <span class="ph-toggle-slider"></span>
                            </label>
                        </div>
                        <div class="ph-category-details">
                            <strong>Purpose:</strong> Authentication, security, basic site functionality<br>
                            <strong>Examples:</strong> Login sessions, CSRF tokens, load balancing
                        </div>
                    </div>

                    <div class="ph-cookie-category">
                        <div class="ph-category-header">
                            <div>
                                <h3>Analytics & Performance</h3>
                                <p>Help us understand how visitors use our website.</p>
                            </div>
                            <label class="ph-toggle">
                                <input type="checkbox" id="phCookieAnalytics" checked>
                                <span class="ph-toggle-slider"></span>
                            </label>
                        </div>
                        <div class="ph-category-details">
                            <strong>Purpose:</strong> Page views, time on site, bounce rate, traffic sources<br>
                            <strong>Tools:</strong> Google Analytics, internal analytics
                        </div>
                    </div>

                    <div class="ph-cookie-category">
                        <div class="ph-category-header">
                            <div>
                                <h3>Behavioral Intelligence</h3>
                                <p>Track your property viewing behavior to provide better service.</p>
                            </div>
                            <label class="ph-toggle">
                                <input type="checkbox" id="phCookieBehavioral" checked>
                                <span class="ph-toggle-slider"></span>
                            </label>
                        </div>
                        <div class="ph-category-details">
                            <strong>Purpose:</strong> Lead scoring, personalized recommendations, agent prioritization<br>
                            <strong>Tracking:</strong> Properties viewed, searches, showings requested, email engagement<br>
                            <strong>Integration:</strong> FollowUpBoss CRM for agent coordination
                        </div>
                    </div>

                    <div class="ph-cookie-category">
                        <div class="ph-category-header">
                            <div>
                                <h3>Marketing & Communication</h3>
                                <p>Used to show you relevant property updates and market information.</p>
                            </div>
                            <label class="ph-toggle">
                                <input type="checkbox" id="phCookieMarketing" checked>
                                <span class="ph-toggle-slider"></span>
                            </label>
                        </div>
                        <div class="ph-category-details">
                            <strong>Purpose:</strong> Email campaigns, property alerts, market updates<br>
                            <strong>Personalization:</strong> Based on your preferences and viewing history
                        </div>
                    </div>
                </div>

                <div class="ph-modal-footer">
                    <button onclick="phCookieConsent.savePreferences()" class="btn-primary">
                        Save Preferences
                    </button>
                    <button onclick="phCookieConsent.acceptAll(); phCookieConsent.closeSettings();" class="btn-primary">
                        Accept All
                    </button>
                </div>
            </div>
        </div>
    `;

    // Wait for DOM to be ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

    function init() {
        document.body.insertAdjacentHTML('beforeend', bannerHTML);
        window.phCookieConsent.init();
    }

    // Cookie Consent Logic
    window.phCookieConsent = {
        init: function() {
            const consent = localStorage.getItem('phCookieConsent');
            if (!consent) {
                setTimeout(() => {
                    document.getElementById('phCookieBanner').classList.add('show');
                }, 1000);
            } else {
                this.applyConsent(JSON.parse(consent));
            }
        },

        acceptAll: function() {
            const consent = {
                essential: true,
                analytics: true,
                behavioral: true,
                marketing: true,
                timestamp: new Date().toISOString()
            };
            this.saveConsent(consent);
            this.hideBanner();
        },

        acceptEssential: function() {
            const consent = {
                essential: true,
                analytics: false,
                behavioral: false,
                marketing: false,
                timestamp: new Date().toISOString()
            };
            this.saveConsent(consent);
            this.hideBanner();
        },

        showSettings: function() {
            document.getElementById('phCookieModal').classList.add('show');
        },

        closeSettings: function() {
            document.getElementById('phCookieModal').classList.remove('show');
        },

        savePreferences: function() {
            const consent = {
                essential: true,
                analytics: document.getElementById('phCookieAnalytics').checked,
                behavioral: document.getElementById('phCookieBehavioral').checked,
                marketing: document.getElementById('phCookieMarketing').checked,
                timestamp: new Date().toISOString()
            };
            this.saveConsent(consent);
            this.closeSettings();
            this.hideBanner();
        },

        saveConsent: function(consent) {
            localStorage.setItem('phCookieConsent', JSON.stringify(consent));
            this.applyConsent(consent);
            
            // Log to backend (if endpoint exists)
            if (typeof fetch !== 'undefined') {
                fetch('/api/cookie-consent', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(consent)
                }).catch(() => {}); // Silent fail if endpoint doesn't exist yet
            }
        },

        applyConsent: function(consent) {
            // Enable/disable analytics
            if (consent.analytics) {
                this.enableAnalytics();
            } else {
                this.disableAnalytics();
            }

            // Enable/disable behavioral tracking
            if (consent.behavioral) {
                this.enableBehavioralTracking();
            } else {
                this.disableBehavioralTracking();
            }

            // Enable/disable marketing
            if (consent.marketing) {
                this.enableMarketing();
            } else {
                this.disableMarketing();
            }

            console.log('Cookie consent applied:', consent);
        },

        hideBanner: function() {
            document.getElementById('phCookieBanner').classList.remove('show');
        },

        enableAnalytics: function() {
            if (window.gtag) {
                gtag('consent', 'update', { 'analytics_storage': 'granted' });
            }
            console.log('Analytics enabled');
        },

        disableAnalytics: function() {
            if (window.gtag) {
                gtag('consent', 'update', { 'analytics_storage': 'denied' });
            }
            console.log('Analytics disabled');
        },

        enableBehavioralTracking: function() {
            window.behavioralTrackingEnabled = true;
            console.log('Behavioral tracking enabled');
        },

        disableBehavioralTracking: function() {
            window.behavioralTrackingEnabled = false;
            console.log('Behavioral tracking disabled');
        },

        enableMarketing: function() {
            if (window.gtag) {
                gtag('consent', 'update', { 'ad_storage': 'granted' });
            }
            console.log('Marketing enabled');
        },

        disableMarketing: function() {
            if (window.gtag) {
                gtag('consent', 'update', { 'ad_storage': 'denied' });
            }
            console.log('Marketing disabled');
        }
    };

    // Global function to revoke consent (for footer link)
    window.revokeCookieConsent = function() {
        localStorage.removeItem('phCookieConsent');
        location.reload();
    };

})();
