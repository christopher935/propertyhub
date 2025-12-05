const SavedProperties = {
    sessionId: null,
    savedPropertyIds: new Set(),

    init() {
        this.sessionId = this.getOrCreateSession();
        this.loadSavedProperties();
        this.attachEventListeners();
    },

    getOrCreateSession() {
        let sessionId = localStorage.getItem('property_session_id');
        if (!sessionId) {
            sessionId = 'sess_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
            localStorage.setItem('property_session_id', sessionId);
        }
        return sessionId;
    },

    async loadSavedProperties() {
        try {
            const response = await fetch(`/api/properties/saved?session_id=${this.sessionId}`);
            if (response.ok) {
                const data = await response.json();
                data.saved_properties.forEach(sp => {
                    this.savedPropertyIds.add(sp.property_id);
                });
                this.updateAllHeartIcons();
                this.updateSavedCount();
            }
        } catch (error) {
            console.error('Failed to load saved properties:', error);
        }
    },

    async saveProperty(propertyId) {
        try {
            const response = await fetch('/api/properties/save', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    property_id: parseInt(propertyId),
                    session_id: this.sessionId
                })
            });

            if (response.ok) {
                this.savedPropertyIds.add(parseInt(propertyId));
                this.updateHeartIcon(propertyId, true);
                this.updateSavedCount();
                this.showNotification('Property saved!', 'success');
                return true;
            } else if (response.status === 409) {
                this.showNotification('Already saved', 'info');
            } else {
                throw new Error('Failed to save');
            }
        } catch (error) {
            console.error('Error saving property:', error);
            this.showNotification('Failed to save property', 'error');
            return false;
        }
    },

    async unsaveProperty(propertyId) {
        try {
            const response = await fetch(`/api/properties/save/${propertyId}?session_id=${this.sessionId}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                this.savedPropertyIds.delete(parseInt(propertyId));
                this.updateHeartIcon(propertyId, false);
                this.updateSavedCount();
                this.showNotification('Property removed', 'info');
                return true;
            }
        } catch (error) {
            console.error('Error unsaving property:', error);
            this.showNotification('Failed to remove property', 'error');
            return false;
        }
    },

    toggleSave(propertyId) {
        const id = parseInt(propertyId);
        if (this.savedPropertyIds.has(id)) {
            return this.unsaveProperty(id);
        } else {
            return this.saveProperty(id);
        }
    },

    isSaved(propertyId) {
        return this.savedPropertyIds.has(parseInt(propertyId));
    },

    updateHeartIcon(propertyId, isSaved) {
        const hearts = document.querySelectorAll(`[data-property-id="${propertyId}"] .save-heart`);
        hearts.forEach(heart => {
            if (isSaved) {
                heart.classList.add('saved');
                heart.innerHTML = '<svg class="heart-icon" fill="currentColor" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M21 8.25c0-2.485-2.099-4.5-4.688-4.5-1.935 0-3.597 1.126-4.312 2.733-.715-1.607-2.377-2.733-4.313-2.733C5.1 3.75 3 5.765 3 8.25c0 7.22 9 12 9 12s9-4.78 9-12z" /></svg><span class="save-text">Saved</span>';
            } else {
                heart.classList.remove('saved');
                heart.innerHTML = '<svg class="heart-icon" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" d="M21 8.25c0-2.485-2.099-4.5-4.688-4.5-1.935 0-3.597 1.126-4.312 2.733-.715-1.607-2.377-2.733-4.313-2.733C5.1 3.75 3 5.765 3 8.25c0 7.22 9 12 9 12s9-4.78 9-12z" /></svg><span class="save-text">Save</span>';
            }
        });
    },

    updateAllHeartIcons() {
        document.querySelectorAll('[data-property-id]').forEach(element => {
            const propertyId = parseInt(element.dataset.propertyId);
            this.updateHeartIcon(propertyId, this.isSaved(propertyId));
        });
    },

    updateSavedCount() {
        const count = this.savedPropertyIds.size;
        const badges = document.querySelectorAll('.saved-count-badge');
        badges.forEach(badge => {
            badge.textContent = count;
            badge.style.display = count > 0 ? 'inline-block' : 'none';
        });
    },

    attachEventListeners() {
        document.addEventListener('click', (e) => {
            const saveBtn = e.target.closest('.save-heart');
            if (saveBtn) {
                e.preventDefault();
                const card = saveBtn.closest('[data-property-id]');
                if (card) {
                    const propertyId = card.dataset.propertyId;
                    this.toggleSave(propertyId);
                }
            }
        });
    },

    showNotification(message, type = 'info') {
        const toast = document.createElement('div');
        toast.className = `toast toast-${type}`;
        toast.innerHTML = `
            <div class="toast-message">${message}</div>
            <button onclick="this.remove()" style="background:none;border:none;cursor:pointer;color:var(--gray-500);font-size:20px;line-height:1;">&times;</button>
        `;
        document.body.appendChild(toast);
        setTimeout(() => {
            toast.style.opacity = '0';
            toast.style.transform = 'translateY(20px)';
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    }
};

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => SavedProperties.init());
} else {
    SavedProperties.init();
}
