class ModalManager {
    constructor() {
        this.createModalContainer();
    }
    
    createModalContainer() {
        if (document.getElementById('modal-container')) return;
        
        const container = document.createElement('div');
        container.id = 'modal-container';
        container.innerHTML = `
            <div class="modal-overlay" onclick="modalManager.close()"></div>
            <div class="modal-content">
                <div class="modal-header">
                    <h3 class="modal-title"></h3>
                    <button class="modal-close" onclick="modalManager.close()">&times;</button>
                </div>
                <div class="modal-body"></div>
                <div class="modal-footer"></div>
            </div>
        `;
        document.body.appendChild(container);
    }
    
    alert(message, title = 'Notice') {
        return new Promise(resolve => {
            const container = document.getElementById('modal-container');
            container.querySelector('.modal-title').textContent = title;
            container.querySelector('.modal-body').innerHTML = `<p>${message}</p>`;
            container.querySelector('.modal-footer').innerHTML = `
                <button class="btn btn-primary" onclick="modalManager.close(); modalManager._resolve()">OK</button>
            `;
            container.classList.add('active');
            this._resolve = resolve;
        });
    }
    
    confirm(message, title = 'Confirm') {
        return new Promise(resolve => {
            const container = document.getElementById('modal-container');
            container.querySelector('.modal-title').textContent = title;
            container.querySelector('.modal-body').innerHTML = `<p>${message}</p>`;
            container.querySelector('.modal-footer').innerHTML = `
                <button class="btn btn-secondary" onclick="modalManager.close(); modalManager._resolve(false)">Cancel</button>
                <button class="btn btn-primary" onclick="modalManager.close(); modalManager._resolve(true)">Confirm</button>
            `;
            container.classList.add('active');
            this._resolve = resolve;
        });
    }
    
    prompt(message, defaultValue = '', title = 'Input Required') {
        return new Promise(resolve => {
            const container = document.getElementById('modal-container');
            container.querySelector('.modal-title').textContent = title;
            container.querySelector('.modal-body').innerHTML = `
                <p>${message}</p>
                <input type="text" class="form-input" id="modal-prompt-input" value="${defaultValue}" style="margin-top: 12px;">
            `;
            container.querySelector('.modal-footer').innerHTML = `
                <button class="btn btn-secondary" onclick="modalManager.close(); modalManager._resolve(null)">Cancel</button>
                <button class="btn btn-primary" onclick="modalManager._resolvePrompt()">OK</button>
            `;
            container.classList.add('active');
            this._resolve = resolve;
            setTimeout(() => {
                const input = document.getElementById('modal-prompt-input');
                input.focus();
                input.select();
                input.addEventListener('keydown', (e) => {
                    if (e.key === 'Enter') {
                        modalManager._resolvePrompt();
                    } else if (e.key === 'Escape') {
                        modalManager.close();
                        modalManager._resolve(null);
                    }
                });
            }, 100);
        });
    }
    
    _resolvePrompt() {
        const input = document.getElementById('modal-prompt-input');
        const value = input ? input.value : null;
        this.close();
        this._resolve(value);
    }
    
    toast(message, type = 'info', duration = 3000) {
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.innerHTML = `
            <span class="toast-icon">${type === 'success' ? '✓' : type === 'error' ? '✕' : 'ℹ'}</span>
            <span class="toast-message">${message}</span>
        `;
        
        let toastContainer = document.getElementById('toast-container');
        if (!toastContainer) {
            toastContainer = document.createElement('div');
            toastContainer.id = 'toast-container';
            document.body.appendChild(toastContainer);
        }
        
        toastContainer.appendChild(toast);
        setTimeout(() => toast.classList.add('show'), 10);
        setTimeout(() => {
            toast.classList.remove('show');
            setTimeout(() => toast.remove(), 300);
        }, duration);
    }
    
    close() {
        document.getElementById('modal-container').classList.remove('active');
    }
}

const modalManager = new ModalManager();
window.showAlert = (msg, title) => modalManager.alert(msg, title);
window.showConfirm = (msg, title) => modalManager.confirm(msg, title);
window.showPrompt = (msg, defaultValue, title) => modalManager.prompt(msg, defaultValue, title);
window.showToast = (msg, type) => modalManager.toast(msg, type);
