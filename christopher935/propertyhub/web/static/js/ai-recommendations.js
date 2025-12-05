const AIRecommendations = {
    sessionId: null,

    init() {
        this.sessionId = localStorage.getItem('property_session_id');
        this.loadRecommendations();
    },

    async loadRecommendations() {
        const container = document.getElementById('aiRecommendations');
        if (!container) return;

        try {
            const url = this.sessionId ? 
                `/api/recommendations?session_id=${this.sessionId}&limit=6` :
                `/api/recommendations?limit=6`;
            
            const response = await fetch(url);
            if (!response.ok) throw new Error('Failed to load recommendations');
            
            const data = await response.json();
            
            if (data.recommendations && data.recommendations.length > 0) {
                this.renderRecommendations(data.recommendations, data.personalized);
            } else {
                container.style.display = 'none';
            }
        } catch (error) {
            console.error('Error loading recommendations:', error);
            container.style.display = 'none';
        }
    },

    renderRecommendations(recommendations, isPersonalized) {
        const container = document.getElementById('aiRecommendations');
        const grid = container.querySelector('.recommendations-grid');
        
        const title = container.querySelector('h2');
        if (title && isPersonalized) {
            title.innerHTML = `
                <svg width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24" style="display:inline-block;vertical-align:middle;margin-right:8px;">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M9.813 15.904 9 18.75l-.813-2.846a4.5 4.5 0 0 0-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 0 0 3.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 0 0 3.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 0 0-3.09 3.09ZM18.259 8.715 18 9.75l-.259-1.035a3.375 3.375 0 0 0-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 0 0 2.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 0 0 2.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 0 0-2.456 2.456ZM16.894 20.567 16.5 21.75l-.394-1.183a2.25 2.25 0 0 0-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 0 0 1.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 0 0 1.423 1.423l1.183.394-1.183.394a2.25 2.25 0 0 0-1.423 1.423Z" />
                </svg>
                Recommended For You
            `;
        }
        
        grid.innerHTML = recommendations.map(rec => {
            const prop = rec.property;
            const bedrooms = prop.bedrooms || '—';
            const bathrooms = prop.bathrooms || '—';
            const sqft = prop.square_feet ? `${prop.square_feet.toLocaleString()} sq ft` : '—';
            
            return `
                <div class="property-card" data-property-id="${prop.id}">
                    ${rec.type === 'personalized' ? 
                        `<div class="ai-badge">
                            <svg width="14" height="14" fill="currentColor" viewBox="0 0 24 24">
                                <path d="M9.813 15.904 9 18.75l-.813-2.846a4.5 4.5 0 0 0-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 0 0 3.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 0 0 3.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 0 0-3.09 3.09Z" />
                            </svg>
                            AI Match
                        </div>` : ''
                    }
                    <button class="save-heart" data-property-id="${prop.id}">
                        <svg class="heart-icon" fill="none" stroke="currentColor" stroke-width="1.5" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M21 8.25c0-2.485-2.099-4.5-4.688-4.5-1.935 0-3.597 1.126-4.312 2.733-.715-1.607-2.377-2.733-4.313-2.733C5.1 3.75 3 5.765 3 8.25c0 7.22 9 12 9 12s9-4.78 9-12z" />
                        </svg>
                    </button>
                    <div class="property-image">
                        ${prop.featured_image ? 
                            `<img src="${prop.featured_image}" alt="${prop.address}">` : 
                            `<div class="property-placeholder">
                                <svg width="48" height="48" fill="currentColor">
                                    <use href="#icon-home-modern"></use>
                                </svg>
                            </div>`
                        }
                    </div>
                    <div class="property-details">
                        <div class="property-price">$${prop.price.toLocaleString()}</div>
                        <div class="property-address">${prop.address}</div>
                        <div class="property-location">${prop.city}, ${prop.state}</div>
                        <div class="property-features">
                            <span>${bedrooms} bed</span>
                            <span>•</span>
                            <span>${bathrooms} bath</span>
                            <span>•</span>
                            <span>${sqft}</span>
                        </div>
                        ${rec.reason ? `<div class="ai-reason">${rec.reason}</div>` : ''}
                        <a href="/property/${prop.id}" class="btn btn-primary btn-sm btn-block">View Details</a>
                    </div>
                </div>
            `;
        }).join('');
        
        container.style.display = 'block';
    }
};

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => AIRecommendations.init());
} else {
    AIRecommendations.init();
}
