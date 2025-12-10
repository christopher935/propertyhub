/**
 * Debug logger that only logs in development mode
 * Prevents console output in production while preserving debugging capability
 */
const Logger = {
    isDev: window.location.hostname === 'localhost' || 
           window.location.hostname === '127.0.0.1' ||
           window.location.search.includes('debug=true'),
    
    log(...args) {
        if (this.isDev) console.log(...args);
    },
    
    warn(...args) {
        if (this.isDev) console.warn(...args);
    },
    
    error(...args) {
        console.error(...args);
    },
    
    info(...args) {
        if (this.isDev) console.info(...args);
    },
    
    debug(...args) {
        if (this.isDev) console.debug(...args);
    }
};

window.Logger = Logger;
