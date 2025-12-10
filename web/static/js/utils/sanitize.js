/**
 * HTML sanitization utilities to prevent XSS attacks
 */
const Sanitizer = {
    /**
     * Escape HTML special characters
     */
    escapeHtml(str) {
        if (str === null || str === undefined) return '';
        const div = document.createElement('div');
        div.textContent = String(str);
        return div.innerHTML;
    },
    
    /**
     * Sanitize an object's string values for safe HTML insertion
     */
    sanitizeObject(obj) {
        if (!obj || typeof obj !== 'object') return obj;
        const sanitized = Array.isArray(obj) ? [] : {};
        for (const [key, value] of Object.entries(obj)) {
            if (typeof value === 'string') {
                sanitized[key] = this.escapeHtml(value);
            } else if (typeof value === 'object' && value !== null) {
                sanitized[key] = this.sanitizeObject(value);
            } else {
                sanitized[key] = value;
            }
        }
        return sanitized;
    },
    
    /**
     * Safe template literal tag for HTML
     * Usage: Sanitizer.html`<div>${userInput}</div>`
     */
    html(strings, ...values) {
        return strings.reduce((result, str, i) => {
            const value = i > 0 ? values[i - 1] : '';
            const escaped = typeof value === 'string' ? this.escapeHtml(value) : value;
            return result + (i > 0 ? escaped : '') + str;
        }, '');
    }
};

if (typeof module !== 'undefined' && module.exports) {
    module.exports = Sanitizer;
}

window.Sanitizer = Sanitizer;
