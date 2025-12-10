/**
 * Properties Loader - Fetches and displays properties from API
 * Usage: Include this script in properties-grid.html or any property listing page
 */

(function() {
    'use strict';

    // Configuration
    const API_BASE_URL = '/api/v1';
    const DEFAULT_LIMIT = 20;

    /**
     * Fetch properties from API
     */
    async function fetchProperties(options = {}) {
        const params = new URLSearchParams({
            page: options.page || 1,
            limit: options.limit || DEFAULT_LIMIT,
            listing_type: options.listingType || '',
            city: options.city || '',
            property_type: options.propertyType || '',
            min_price: options.minPrice || '',
            max_price: options.maxPrice || '',
            bedrooms: options.bedrooms || '',
            bathrooms: options.bathrooms || '',
            sort_by: options.sortBy || 'created_at',
            sort_order: options.sortOrder || 'desc',
            search: options.search || ''
        });

        // Remove empty params
        for (const [key, value] of [...params.entries()]) {
            if (!value) params.delete(key);
        }

        try {
            const response = await fetch(`${API_BASE_URL}/properties?${params}`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();
            return data.data || data; // Handle both wrapped and unwrapped responses
        } catch (error) {
            console.error('Error fetching properties:', error);
            throw error;
        }
    }

    /**
     * Fetch single property by ID
     */
    async function fetchPropertyById(id) {
        try {
            const response = await fetch(`${API_BASE_URL}/properties/${id}`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();
            return data.data?.property || data.property;
        } catch (error) {
            console.error('Error fetching property:', error);
            throw error;
        }
    }

    /**
     * Fetch property statistics
     */
    async function fetchPropertyStats() {
        try {
            const response = await fetch(`${API_BASE_URL}/properties/stats`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();
            return data.data?.stats || data.stats;
        } catch (error) {
            console.error('Error fetching stats:', error);
            throw error;
        }
    }

    /**
     * Format price for display
     */
    function formatPrice(price) {
        if (!price || price === 0) return 'Contact for Price';
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: 'USD',
            minimumFractionDigits: 0,
            maximumFractionDigits: 0
        }).format(price);
    }

    /**
     * Create property card HTML
     */
    function createPropertyCard(property) {
        const safe = Sanitizer.sanitizeObject(property);
        const image = safe.featured_image || safe.images?.[0] || '/static/images/placeholder-property.jpg';
        const price = formatPrice(property.price);
        const bedrooms = property.bedrooms || '—';
        const bathrooms = property.bathrooms || '—';
        const sqft = property.square_feet ? property.square_feet.toLocaleString() : '—';
        const address = safe.city ? `${safe.city}, ${safe.state || 'TX'}` : 'Houston, TX';

        return `
            <div class="property-card" data-property-id="${safe.id}">
                <div class="property-image">
                    <img src="${image}" alt="Property" loading="lazy">
                    <span class="property-badge">${safe.listing_type || 'For Sale'}</span>
                </div>
                <div class="property-details">
                    <div class="property-price">${price}</div>
                    <div class="property-address">${address}</div>
                    <div class="property-specs">
                        <span><i class="icon-bed"></i> ${bedrooms} bd</span>
                        <span><i class="icon-bath"></i> ${bathrooms} ba</span>
                        <span><i class="icon-sqft"></i> ${sqft} sqft</span>
                    </div>
                    <div class="property-actions">
                        <a href="/property/${safe.id}" class="btn btn-primary">View Details</a>
                        <a href="/book-showing?property_id=${safe.id}" class="btn btn-secondary">Book Showing</a>
                    </div>
                </div>
            </div>
        `;
    }

    /**
     * Render properties to container
     */
    function renderProperties(properties, containerId = 'properties-grid') {
        const container = document.getElementById(containerId);
        if (!container) {
            console.error(`Container #${containerId} not found`);
            return;
        }

        if (!properties || properties.length === 0) {
            container.innerHTML = `
                <div class="no-properties">
                    <h3>No properties found</h3>
                    <p>Try adjusting your search filters or check back later.</p>
                </div>
            `;
            return;
        }

        container.innerHTML = properties.map(createPropertyCard).join('');
    }

    /**
     * Render pagination
     */
    function renderPagination(pagination, containerId = 'pagination') {
        const container = document.getElementById(containerId);
        if (!container || !pagination) return;

        const { page, total_pages, has_prev, has_next } = pagination;
        
        let html = '<div class="pagination">';
        
        // Previous button
        if (has_prev) {
            html += `<button class="page-btn" data-page="${page - 1}">Previous</button>`;
        }
        
        // Page numbers
        for (let i = Math.max(1, page - 2); i <= Math.min(total_pages, page + 2); i++) {
            const active = i === page ? 'active' : '';
            html += `<button class="page-btn ${active}" data-page="${i}">${i}</button>`;
        }
        
        // Next button
        if (has_next) {
            html += `<button class="page-btn" data-page="${page + 1}">Next</button>`;
        }
        
        html += '</div>';
        container.innerHTML = html;

        // Add click handlers
        container.querySelectorAll('.page-btn').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const newPage = parseInt(e.target.dataset.page);
                loadProperties({ page: newPage });
            });
        });
    }

    /**
     * Load and display properties
     */
    async function loadProperties(options = {}) {
        try {
            // Show loading state
            const container = document.getElementById('properties-grid');
            if (container) {
                container.innerHTML = '<div class="loading">Loading properties...</div>';
            }

            // Fetch properties
            const response = await fetchProperties(options);
            
            // Render properties
            renderProperties(response.properties);
            
            // Render pagination
            if (response.pagination) {
                renderPagination(response.pagination);
            }

            // Update property count
            const countEl = document.getElementById('property-count');
            if (countEl && response.pagination) {
                countEl.textContent = `${response.pagination.total} properties found`;
            }

        } catch (error) {
            console.error('Error loading properties:', error);
            const container = document.getElementById('properties-grid');
            if (container) {
                container.innerHTML = `
                    <div class="error">
                        <h3>Error loading properties</h3>
                        <p>Please try again later or contact support.</p>
                    </div>
                `;
            }
        }
    }

    /**
     * Initialize property filters
     */
    function initializeFilters() {
        const filterForm = document.getElementById('property-filters');
        if (!filterForm) return;

        filterForm.addEventListener('submit', (e) => {
            e.preventDefault();
            
            const formData = new FormData(filterForm);
            const options = {
                listingType: formData.get('listing_type'),
                city: formData.get('city'),
                propertyType: formData.get('property_type'),
                minPrice: formData.get('min_price'),
                maxPrice: formData.get('max_price'),
                bedrooms: formData.get('bedrooms'),
                bathrooms: formData.get('bathrooms'),
                search: formData.get('search')
            };

            loadProperties(options);
        });
    }

    /**
     * Initialize on page load
     */
    function init() {
        // Load properties on page load
        if (document.getElementById('properties-grid')) {
            loadProperties();
        }

        // Initialize filters
        initializeFilters();

        // Load stats if stats container exists
        if (document.getElementById('property-stats')) {
            fetchPropertyStats().then(stats => {
                const statsEl = document.getElementById('property-stats');
                if (statsEl && stats) {
                    statsEl.innerHTML = `
                        <div class="stat">
                            <div class="stat-value">${stats.total_active || 0}</div>
                            <div class="stat-label">Active Properties</div>
                        </div>
                        <div class="stat">
                            <div class="stat-value">${stats.for_sale || 0}</div>
                            <div class="stat-label">For Sale</div>
                        </div>
                        <div class="stat">
                            <div class="stat-value">${stats.for_rent || 0}</div>
                            <div class="stat-label">For Rent</div>
                        </div>
                        <div class="stat">
                            <div class="stat-value">${formatPrice(stats.average_price)}</div>
                            <div class="stat-label">Avg Price</div>
                        </div>
                    `;
                }
            });
        }
    }

    // Auto-initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

    // Expose API to window for manual use
    window.PropertiesAPI = {
        fetchProperties,
        fetchPropertyById,
        fetchPropertyStats,
        loadProperties,
        formatPrice
    };

})();

