/**
 * PropertyHub Authentication System
 * Complete authentication, MFA, and security management
 * Enterprise-grade auth with JWT, MFA, RBAC, and session management
 */

class PropertyHubAuth {
    constructor() {
        this.token = localStorage.getItem('token');
        this.refreshToken = localStorage.getItem('refreshToken');
        this.user = JSON.parse(localStorage.getItem('user') || 'null');
        this.mfaSetup = JSON.parse(localStorage.getItem('mfaSetup') || 'false');
        this.sessionTimeout = 30 * 60 * 1000; // 30 minutes
        this.refreshInterval = null;
        this.sessionTimer = null;
        this.loginAttempts = 0;
        this.maxLoginAttempts = 5;
        this.lockoutDuration = 15 * 60 * 1000; // 15 minutes
        
        this.init();
    }

    init() {
        Logger.log('🔐 PropertyHub Authentication initializing...');
        
        this.setupTokenRefresh();
        this.setupSessionTimeout();
        this.setupAuthHeaders();
        this.checkAuthStatus();
        this.setupEventListeners();
        
        Logger.log('✅ PropertyHub Authentication initialized');
    }

    // Authentication Core Methods
    async login(credentials) {
        try {
            // Check if account is locked
            if (this.isAccountLocked()) {
                throw new Error(`Account locked. Try again in ${this.getRemainingLockoutTime()} minutes.`);
            }

            const response = await fetch('/api/auth/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Requested-With': 'XMLHttpRequest'
                },
                body: JSON.stringify({
                    username: credentials.username,
                    password: credentials.password,
                    rememberMe: credentials.rememberMe || false
                })
            });

            const data = await response.json();

            if (!response.ok) {
                this.handleLoginFailure();
                throw new Error(data.message || 'Login failed');
            }

            // Reset login attempts on successful login
            this.resetLoginAttempts();

            // Check if MFA is required
            if (data.requiresMFA) {
                return {
                    success: false,
                    requiresMFA: true,
                    tempToken: data.tempToken,
                    mfaMethods: data.mfaMethods
                };
            }

            // Complete login process
            await this.completeLogin(data);
            
            return {
                success: true,
                user: this.user,
                redirectUrl: data.redirectUrl || '/dashboard'
            };

        } catch (error) {
            Logger.error('Login error:', error);
            this.showAuthError(error.message);
            return {
                success: false,
                error: error.message
            };
        }
    }

    async verifyMFA(mfaCode, tempToken, method = 'totp') {
        try {
            const response = await fetch('/api/auth/verify-mfa', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${tempToken}`
                },
                body: JSON.stringify({
                    code: mfaCode,
                    method: method
                })
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || 'MFA verification failed');
            }

            // Complete login after MFA verification
            await this.completeLogin(data);

            return {
                success: true,
                user: this.user,
                redirectUrl: data.redirectUrl || '/dashboard'
            };

        } catch (error) {
            Logger.error('MFA verification error:', error);
            this.showAuthError(error.message);
            return {
                success: false,
                error: error.message
            };
        }
    }

    async completeLogin(authData) {
        // Store authentication data
        this.token = authData.token;
        this.refreshToken = authData.refreshToken;
        this.user = authData.user;
        
        localStorage.setItem('token', this.token);
        localStorage.setItem('refreshToken', this.refreshToken);
        localStorage.setItem('user', JSON.stringify(this.user));
        localStorage.setItem('loginTime', Date.now().toString());

        // Setup auth headers for future requests
        this.setupAuthHeaders();
        
        // Start token refresh and session management
        this.setupTokenRefresh();
        this.setupSessionTimeout();
        
        // Track login analytics
        this.trackAuthEvent('login', {
            userId: this.user.id,
            timestamp: new Date().toISOString(),
            userAgent: navigator.userAgent
        });

        // Notify other tabs of login
        this.broadcastAuthChange('login');
        
        Logger.log('✅ Login completed successfully');
    }

    async logout(reason = 'user_initiated') {
        try {
            // Notify server of logout
            if (this.token) {
                await fetch('/api/auth/logout', {
                    method: 'POST',
                    headers: {
                        'Authorization': `Bearer ${this.token}`,
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({ reason })
                });
            }
        } catch (error) {
            Logger.error('Logout request failed:', error);
        }

        // Clear local storage
        this.clearAuthData();
        
        // Track logout analytics
        this.trackAuthEvent('logout', {
            userId: this.user?.id,
            reason: reason,
            timestamp: new Date().toISOString()
        });

        // Notify other tabs of logout
        this.broadcastAuthChange('logout');
        
        // Redirect to login
        if (reason !== 'token_expired') {
            window.location.href = '/login';
        }
        
        Logger.log('🚪 User logged out:', reason);
    }

    async register(userData) {
        try {
            const response = await fetch('/api/auth/register', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(userData)
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || 'Registration failed');
            }

            // Track registration analytics
            this.trackAuthEvent('register', {
                userId: data.user.id,
                timestamp: new Date().toISOString()
            });

            return {
                success: true,
                message: data.message,
                requiresVerification: data.requiresVerification
            };

        } catch (error) {
            Logger.error('Registration error:', error);
            this.showAuthError(error.message);
            return {
                success: false,
                error: error.message
            };
        }
    }

    async forgotPassword(email) {
        try {
            const response = await fetch('/api/auth/forgot-password', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ email })
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || 'Password reset failed');
            }

            return {
                success: true,
                message: data.message
            };

        } catch (error) {
            Logger.error('Password reset error:', error);
            return {
                success: false,
                error: error.message
            };
        }
    }

    async resetPassword(token, newPassword) {
        try {
            const response = await fetch('/api/auth/reset-password', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    token: token,
                    password: newPassword
                })
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || 'Password reset failed');
            }

            return {
                success: true,
                message: data.message
            };

        } catch (error) {
            Logger.error('Password reset error:', error);
            return {
                success: false,
                error: error.message
            };
        }
    }

    // Multi-Factor Authentication
    async setupMFA(method = 'totp') {
        try {
            const response = await fetch('/api/auth/mfa/setup', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ method })
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || 'MFA setup failed');
            }

            return {
                success: true,
                qrCode: data.qrCode,
                secret: data.secret,
                backupCodes: data.backupCodes
            };

        } catch (error) {
            Logger.error('MFA setup error:', error);
            return {
                success: false,
                error: error.message
            };
        }
    }

    async confirmMFASetup(code) {
        try {
            const response = await fetch('/api/auth/mfa/confirm', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ code })
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || 'MFA confirmation failed');
            }

            // Update local MFA status
            this.mfaSetup = true;
            localStorage.setItem('mfaSetup', 'true');

            // Update user object
            if (this.user) {
                this.user.mfaEnabled = true;
                localStorage.setItem('user', JSON.stringify(this.user));
            }

            return {
                success: true,
                message: data.message
            };

        } catch (error) {
            Logger.error('MFA confirmation error:', error);
            return {
                success: false,
                error: error.message
            };
        }
    }

    async disableMFA(password) {
        try {
            const response = await fetch('/api/auth/mfa/disable', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${this.token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ password })
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || 'MFA disable failed');
            }

            // Update local MFA status
            this.mfaSetup = false;
            localStorage.setItem('mfaSetup', 'false');

            // Update user object
            if (this.user) {
                this.user.mfaEnabled = false;
                localStorage.setItem('user', JSON.stringify(this.user));
            }

            return {
                success: true,
                message: data.message
            };

        } catch (error) {
            Logger.error('MFA disable error:', error);
            return {
                success: false,
                error: error.message
            };
        }
    }

    // Token Management
    async refreshAuthToken() {
        try {
            if (!this.refreshToken) {
                throw new Error('No refresh token available');
            }

            const response = await fetch('/api/auth/refresh', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    refreshToken: this.refreshToken
                })
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.message || 'Token refresh failed');
            }

            // Update tokens
            this.token = data.token;
            if (data.refreshToken) {
                this.refreshToken = data.refreshToken;
                localStorage.setItem('refreshToken', this.refreshToken);
            }
            
            localStorage.setItem('token', this.token);
            this.setupAuthHeaders();

            Logger.log('🔄 Token refreshed successfully');
            return true;

        } catch (error) {
            Logger.error('Token refresh failed:', error);
            this.logout('token_expired');
            return false;
        }
    }

    setupTokenRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
        }

        if (this.token) {
            // Refresh token every 25 minutes (5 minutes before expiry)
            this.refreshInterval = setInterval(() => {
                this.refreshAuthToken();
            }, 25 * 60 * 1000);
        }
    }

    setupSessionTimeout() {
        if (this.sessionTimer) {
            clearTimeout(this.sessionTimer);
        }

        if (this.token) {
            this.sessionTimer = setTimeout(() => {
                this.logout('session_timeout');
            }, this.sessionTimeout);
        }
    }

    resetSessionTimeout() {
        this.setupSessionTimeout();
    }

    // Authentication Status
    isAuthenticated() {
        return !!this.token && !!this.user;
    }

    hasRole(role) {
        return this.user && this.user.roles && this.user.roles.includes(role);
    }

    hasPermission(permission) {
        return this.user && this.user.permissions && this.user.permissions.includes(permission);
    }

    isAdmin() {
        return this.hasRole('admin') || this.hasRole('super_admin');
    }

    isMFAEnabled() {
        return this.user && this.user.mfaEnabled;
    }

    // Account Security
    isAccountLocked() {
        const lockoutTime = localStorage.getItem('accountLockoutTime');
        if (!lockoutTime) return false;
        
        const lockTime = parseInt(lockoutTime);
        return Date.now() < lockTime;
    }

    getRemainingLockoutTime() {
        const lockoutTime = localStorage.getItem('accountLockoutTime');
        if (!lockoutTime) return 0;
        
        const remaining = parseInt(lockoutTime) - Date.now();
        return Math.ceil(remaining / (60 * 1000)); // Return minutes
    }

    handleLoginFailure() {
        this.loginAttempts++;
        
        if (this.loginAttempts >= this.maxLoginAttempts) {
            const lockoutTime = Date.now() + this.lockoutDuration;
            localStorage.setItem('accountLockoutTime', lockoutTime.toString());
            localStorage.setItem('loginAttempts', '0');
            this.loginAttempts = 0;
        } else {
            localStorage.setItem('loginAttempts', this.loginAttempts.toString());
        }
    }

    resetLoginAttempts() {
        this.loginAttempts = 0;
        localStorage.removeItem('loginAttempts');
        localStorage.removeItem('accountLockoutTime');
    }

    // Password Validation
    validatePassword(password) {
        const requirements = {
            minLength: password.length >= 8,
            hasUppercase: /[A-Z]/.test(password),
            hasLowercase: /[a-z]/.test(password),
            hasNumbers: /\d/.test(password),
            hasSpecialChar: /[!@#$%^&*(),.?":{}|<>]/.test(password)
        };

        const isValid = Object.values(requirements).every(req => req);
        
        return {
            isValid,
            requirements,
            strength: this.calculatePasswordStrength(password)
        };
    }

    calculatePasswordStrength(password) {
        let strength = 0;
        
        if (password.length >= 8) strength++;
        if (password.length >= 12) strength++;
        if (/[A-Z]/.test(password)) strength++;
        if (/[a-z]/.test(password)) strength++;
        if (/\d/.test(password)) strength++;
        if (/[!@#$%^&*(),.?":{}|<>]/.test(password)) strength++;
        if (password.length >= 16) strength++;

        if (strength <= 2) return 'weak';
        if (strength <= 4) return 'medium';
        if (strength <= 5) return 'strong';
        return 'very-strong';
    }

    // Utility Methods
    setupAuthHeaders() {
        // Set default headers for fetch requests
        const originalFetch = window.fetch;
        window.fetch = (...args) => {
            if (args[1] && args[1].headers && this.token) {
                if (!args[1].headers.Authorization) {
                    args[1].headers.Authorization = `Bearer ${this.token}`;
                }
            }
            return originalFetch.apply(window, args);
        };
    }

    clearAuthData() {
        this.token = null;
        this.refreshToken = null;
        this.user = null;
        this.mfaSetup = false;

        localStorage.removeItem('token');
        localStorage.removeItem('refreshToken');
        localStorage.removeItem('user');
        localStorage.removeItem('mfaSetup');
        localStorage.removeItem('loginTime');

        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }

        if (this.sessionTimer) {
            clearTimeout(this.sessionTimer);
            this.sessionTimer = null;
        }
    }

    checkAuthStatus() {
        // Check if user should be logged out due to inactivity
        const loginTime = localStorage.getItem('loginTime');
        if (loginTime && this.token) {
            const timeSinceLogin = Date.now() - parseInt(loginTime);
            if (timeSinceLogin > this.sessionTimeout) {
                this.logout('session_timeout');
                return;
            }
        }

        // Check if we're on a protected route without authentication
        const protectedRoutes = ['/dashboard', '/properties', '/bookings', '/analytics', '/admin'];
        const currentPath = window.location.pathname;
        
        if (protectedRoutes.some(route => currentPath.startsWith(route)) && !this.isAuthenticated()) {
            window.location.href = '/login?redirect=' + encodeURIComponent(currentPath);
        }

        // Redirect authenticated users away from auth pages
        const authPages = ['/login', '/register', '/forgot-password'];
        if (authPages.includes(currentPath) && this.isAuthenticated()) {
            window.location.href = '/dashboard';
        }
    }

    setupEventListeners() {
        // Listen for storage changes (other tabs)
        window.addEventListener('storage', (e) => {
            if (e.key === 'authChange') {
                const change = JSON.parse(e.newValue);
                if (change.action === 'logout') {
                    this.clearAuthData();
                    window.location.href = '/login';
                }
            }
        });

        // Reset session timeout on user activity
        const resetTimeout = () => this.resetSessionTimeout();
        document.addEventListener('click', resetTimeout);
        document.addEventListener('keypress', resetTimeout);
        document.addEventListener('scroll', resetTimeout);

        // Handle page visibility changes
        document.addEventListener('visibilitychange', () => {
            if (document.visibilityState === 'visible' && this.isAuthenticated()) {
                this.refreshAuthToken();
            }
        });
    }

    broadcastAuthChange(action) {
        localStorage.setItem('authChange', JSON.stringify({
            action,
            timestamp: Date.now()
        }));
        localStorage.removeItem('authChange');
    }

    trackAuthEvent(event, data) {
        // Track authentication events for analytics
        if (window.PropertyHubAnalytics) {
            window.PropertyHubAnalytics.trackEvent(`auth_${event}`, data);
        }
    }

    showAuthError(message) {
        // Show error notification
        if (window.PropertyHubDashboard) {
            window.PropertyHubDashboard.showErrorAlert(message);
        } else {
            // Fallback error display
            const errorDiv = document.getElementById('auth-error');
            if (errorDiv) {
                errorDiv.textContent = message;
                errorDiv.style.display = 'block';
            } else {
                alert(message);
            }
        }
    }

    // Public API Methods for Forms
    async handleLoginForm(formData) {
        const form = formData.target || formData;
        const formDataObj = new FormData(form);
        
        const credentials = {
            username: formDataObj.get('username'),
            password: formDataObj.get('password'),
            rememberMe: formDataObj.get('rememberMe') === 'on'
        };

        const loginBtn = form.querySelector('button[type="submit"]');
        if (loginBtn) {
            loginBtn.disabled = true;
            loginBtn.textContent = 'Signing In...';
        }

        try {
            const result = await this.login(credentials);
            
            if (result.success) {
                // Redirect on successful login
                const urlParams = new URLSearchParams(window.location.search);
                const redirectUrl = urlParams.get('redirect') || result.redirectUrl;
                window.location.href = redirectUrl;
            } else if (result.requiresMFA) {
                // Show MFA form
                this.showMFAForm(result.tempToken, result.mfaMethods);
            }
            
            return result;
        } finally {
            if (loginBtn) {
                loginBtn.disabled = false;
                loginBtn.textContent = 'Sign In';
            }
        }
    }

    async handleMFAForm(formData, tempToken) {
        const form = formData.target || formData;
        const formDataObj = new FormData(form);
        
        const mfaCode = formDataObj.get('mfaCode');
        const method = formDataObj.get('method') || 'totp';

        const result = await this.verifyMFA(mfaCode, tempToken, method);
        
        if (result.success) {
            const urlParams = new URLSearchParams(window.location.search);
            const redirectUrl = urlParams.get('redirect') || result.redirectUrl;
            window.location.href = redirectUrl;
        }
        
        return result;
    }

    showMFAForm(tempToken, methods) {
        // Create and show MFA verification form
        const mfaHtml = `
            <div class="mfa-overlay">
                <div class="mfa-form-container">
                    <h2>Multi-Factor Authentication</h2>
                    <p>Please enter your authentication code to continue</p>
                    <form id="mfa-form" onsubmit="return PropertyHubAuth.handleMFASubmit(event, '${tempToken}')">
                        <div class="form-group">
                            <label for="mfaCode">Authentication Code</label>
                            <input type="text" id="mfaCode" name="mfaCode" required 
                                   maxlength="6" placeholder="000000" autocomplete="one-time-code">
                        </div>
                        <button type="submit" class="btn btn-primary">Verify Code</button>
                    </form>
                </div>
            </div>
        `;
        
        document.body.insertAdjacentHTML('beforeend', mfaHtml);
    }

    async handleMFASubmit(event, tempToken) {
        event.preventDefault();
        const result = await this.handleMFAForm(event, tempToken);
        
        if (!result.success) {
            const errorDiv = document.querySelector('#mfa-form .error');
            if (errorDiv) {
                errorDiv.textContent = result.error;
            } else {
                const form = document.getElementById('mfa-form');
                form.insertAdjacentHTML('afterbegin', `<div class="error">${result.error}</div>`);
            }
        }
        
        return false;
    }
}

// Initialize PropertyHub Auth when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.PropertyHubAuth = new PropertyHubAuth();
});

// Expose methods for form handling
window.PropertyHubAuth = {
    handleMFASubmit: (event, tempToken) => {
        return window.PropertyHubAuth.handleMFASubmit(event, tempToken);
    }
};

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PropertyHubAuth;
}
