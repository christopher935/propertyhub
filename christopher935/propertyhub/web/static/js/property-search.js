/**
 * PropertyHub Advanced Property Search
 * Comprehensive property search with filters, maps, saved searches
 * HAR MLS integration with real-time market data and analytics
 */

class PropertyHubSearch {
    constructor() {
        this.searchResults = [];
        this.currentFilters = {};
        this.savedSearches = [];
        this.searchHistory = [];
        this.favorites = [];
        this.mapInstance = null;
        this.mapMarkers = [];
        this.searchTimeout = null;
        this.currentPage = 1;
        this.resultsPerPage = 20;
        this.totalResults = 0;
        
        // Search configuration
        this.config = {
            enableAutoComplete: true,
            enableMapSearch: true,
            enableSavedSearches: true,
            enablePriceAlerts: true,
            minSearchDelay: 300, // ms
            maxSearchRadius: 50, // miles
            defaultSortBy: 'price_asc'
        };

        // Filter definitions
        this.filterDefinitions = {
            location: { type: 'text', label: 'Location' },
            priceMin: { type: 'number', label: 'Min Price', prefix: '$' },
            priceMax: { type: 'number', label: 'Max Price', prefix: '$' },
            bedrooms: { type: 'number', label: 'Bedrooms' },
            bathrooms: { type: 'number', label: 'Bathrooms' },
            squareFootageMin: { type: 'number', label: 'Min Sq Ft' },
            squareFootageMax: { type: 'number', label: 'Max Sq Ft' },
            propertyType: { type: 'select', label: 'Property Type', options: ['house', 'condo', 'townhouse', 'apartment'] },
            yearBuilt: { type: 'range', label: 'Year Built', min: 1900, max: new Date().getFullYear() },
            lotSize: { type: 'number', label: 'Lot Size (acres)' },
            garage: { type: 'checkbox', label: 'Garage' },
            pool: { type: 'checkbox', label: 'Pool' },
            fireplace: { type: 'checkbox', label: 'Fireplace' },
            waterfront: { type: 'checkbox', label: 'Waterfront' },
            newConstruction: { type: 'checkbox', label: 'New Construction' },
            foreclosure: { type: 'checkbox', label: 'Foreclosure' },
            investmentProperty: { type: 'checkbox', label: 'Investment Property' }
        };

        this.init();
    }

    async init() {
        console.log('🏠 PropertyHub Search initializing...');
        
        await this.loadSearchData();
        this.setupSearchInterface();
        this.setupEventListeners();
        this.loadSavedSearches();
        this.loadSearchHistory();
        
        // Initialize map if enabled
        if (this.config.enableMapSearch) {
            await this.initializeMap();
        }
        
        // Load initial results if we have default filters
        if (this.hasDefaultFilters()) {
            this.performSearch();
        }
        
        console.log('✅ PropertyHub Search ready');
    }

    // Search Core Functions
    async performSearch(resetPage = true) {
        if (resetPage) {
            this.currentPage = 1;
        }

        // Show loading state
        this.showSearchLoading();
        
        try {
            // Build search parameters
            const searchParams = this.buildSearchParams();
            
            // Track search analytics
            this.trackSearchAnalytics(searchParams);
            
            // Add to search history
            this.addToSearchHistory(searchParams);
            
            // Perform the actual search
            const response = await fetch('/api/properties/search', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                },
                body: JSON.stringify({
                    ...searchParams,
                    page: this.currentPage,
                    limit: this.resultsPerPage
                })
            });

            if (!response.ok) {
                throw new Error(`Search failed: ${response.status}`);
            }

            const data = await response.json();
            
            // Update search results
            this.searchResults = data.properties;
            this.totalResults = data.total;
            
            // Update UI
            this.renderSearchResults(data);
            this.updateResultsCount(data.total);
            this.updateMapMarkers(data.properties);
            this.updatePagination();
            
            // Enable saved search functionality
            this.enableSaveSearch();
            
        } catch (error) {
            console.error('Search error:', error);
            this.showSearchError(error.message);
        } finally {
            this.hideSearchLoading();
        }
    }

    buildSearchParams() {
        const params = {
            filters: { ...this.currentFilters },
            sortBy: this.getSortBy(),
            includeMarketData: true,
            includePhotos: true,
            includeNeighborhoodData: true
        };

        // Add geographic bounds if map search is active
        if (this.mapInstance && this.isMapSearchActive()) {
            const bounds = this.mapInstance.getBounds();
            params.geoBounds = {
                north: bounds.getNorthEast().lat(),
                south: bounds.getSouthWest().lat(),
                east: bounds.getNorthEast().lng(),
                west: bounds.getSouthWest().lng()
            };
        }

        return params;
    }

    // Filter Management
    updateFilter(filterName, value) {
        if (value === null || value === undefined || value === '') {
            delete this.currentFilters[filterName];
        } else {
            this.currentFilters[filterName] = value;
        }

        // Update filter display
        this.updateFilterDisplay();
        
        // Debounced search
        this.debouncedSearch();
    }

    clearFilter(filterName) {
        delete this.currentFilters[filterName];
        this.updateFilterElement(filterName, '');
        this.updateFilterDisplay();
        this.debouncedSearch();
    }

    clearAllFilters() {
        this.currentFilters = {};
        this.clearAllFilterElements();
        this.updateFilterDisplay();
        this.performSearch();
    }

    updateFilterDisplay() {
        const filterContainer = document.getElementById('active-filters');
        if (!filterContainer) return;

        const activeFilters = Object.entries(this.currentFilters);
        
        if (activeFilters.length === 0) {
            filterContainer.innerHTML = '<p class="no-filters">No filters active</p>';
            return;
        }

        filterContainer.innerHTML = activeFilters.map(([key, value]) => {
            const definition = this.filterDefinitions[key];
            const displayValue = this.formatFilterValue(key, value, definition);
            
            return `
                <div class="active-filter">
                    <span class="filter-label">${definition?.label || key}:</span>
                    <span class="filter-value">${displayValue}</span>
                    <button class="filter-remove" onclick="PropertyHubSearch.clearFilter('${key}')">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
            `;
        }).join('');
    }

    formatFilterValue(key, value, definition) {
        if (definition?.prefix) {
            return `${definition.prefix}${value.toLocaleString()}`;
        }
        
        if (typeof value === 'boolean') {
            return value ? 'Yes' : 'No';
        }
        
        if (Array.isArray(value)) {
            return value.join(', ');
        }
        
        return value.toString();
    }

    // Search Interface Setup
    setupSearchInterface() {
        this.renderSearchFilters();
        this.renderSearchResults();
        this.renderMapContainer();
        this.setupAutoComplete();
    }

    renderSearchFilters() {
        const filterContainer = document.getElementById('search-filters');
        if (!filterContainer) return;

        filterContainer.innerHTML = `
            <div class="search-filters-container">
                <div class="primary-filters">
                    <div class="filter-group">
                        <label>Location</label>
                        <input type="text" 
                               id="location-search" 
                               class="form-control search-input" 
                               placeholder="City, ZIP, Address, or MLS#"
                               data-filter="location">
                    </div>
                    
                    <div class="filter-group">
                        <label>Price Range</label>
                        <div class="price-inputs">
                            <input type="number" 
                                   id="price-min" 
                                   class="form-control" 
                                   placeholder="Min Price"
                                   data-filter="priceMin">
                            <span class="to-separator">to</span>
                            <input type="number" 
                                   id="price-max" 
                                   class="form-control" 
                                   placeholder="Max Price"
                                   data-filter="priceMax">
                        </div>
                    </div>
                    
                    <div class="filter-group">
                        <label>Bedrooms</label>
                        <select id="bedrooms" class="form-control" data-filter="bedrooms">
                            <option value="">Any</option>
                            <option value="1">1+</option>
                            <option value="2">2+</option>
                            <option value="3">3+</option>
                            <option value="4">4+</option>
                            <option value="5">5+</option>
                        </select>
                    </div>
                    
                    <div class="filter-group">
                        <label>Bathrooms</label>
                        <select id="bathrooms" class="form-control" data-filter="bathrooms">
                            <option value="">Any</option>
                            <option value="1">1+</option>
                            <option value="2">2+</option>
                            <option value="3">3+</option>
                            <option value="4">4+</option>
                        </select>
                    </div>
                    
                    <div class="filter-group">
                        <label>Property Type</label>
                        <select id="property-type" class="form-control" data-filter="propertyType">
                            <option value="">All Types</option>
                            <option value="house">House</option>
                            <option value="condo">Condo</option>
                            <option value="townhouse">Townhouse</option>
                            <option value="apartment">Apartment</option>
                        </select>
                    </div>
                    
                    <div class="filter-actions">
                        <button class="btn btn-primary search-btn" onclick="PropertyHubSearch.performSearch()">
                            <i class="fas fa-search"></i> Search
                        </button>
                        <button class="btn btn-secondary" onclick="PropertyHubSearch.toggleAdvancedFilters()">
                            <i class="fas fa-sliders-h"></i> More Filters
                        </button>
                    </div>
                </div>
                
                <div class="advanced-filters" id="advanced-filters" style="display: none;">
                    <div class="filter-row">
                        <div class="filter-group">
                            <label>Square Footage</label>
                            <div class="range-inputs">
                                <input type="number" class="form-control" placeholder="Min Sq Ft" data-filter="squareFootageMin">
                                <span class="to-separator">to</span>
                                <input type="number" class="form-control" placeholder="Max Sq Ft" data-filter="squareFootageMax">
                            </div>
                        </div>
                        
                        <div class="filter-group">
                            <label>Year Built</label>
                            <div class="range-inputs">
                                <input type="number" class="form-control" placeholder="Min Year" data-filter="yearBuiltMin">
                                <span class="to-separator">to</span>
                                <input type="number" class="form-control" placeholder="Max Year" data-filter="yearBuiltMax">
                            </div>
                        </div>
                        
                        <div class="filter-group">
                            <label>Lot Size (acres)</label>
                            <input type="number" class="form-control" placeholder="Minimum" data-filter="lotSize" step="0.1">
                        </div>
                    </div>
                    
                    <div class="filter-row">
                        <div class="checkbox-filters">
                            <label class="checkbox-label">
                                <input type="checkbox" data-filter="garage"> Garage
                            </label>
                            <label class="checkbox-label">
                                <input type="checkbox" data-filter="pool"> Pool
                            </label>
                            <label class="checkbox-label">
                                <input type="checkbox" data-filter="fireplace"> Fireplace
                            </label>
                            <label class="checkbox-label">
                                <input type="checkbox" data-filter="waterfront"> Waterfront
                            </label>
                            <label class="checkbox-label">
                                <input type="checkbox" data-filter="newConstruction"> New Construction
                            </label>
                            <label class="checkbox-label">
                                <input type="checkbox" data-filter="foreclosure"> Foreclosure
                            </label>
                        </div>
                    </div>
                </div>
                
                <div class="active-filters" id="active-filters"></div>
            </div>
        `;
    }

    renderSearchResults(data = null) {
        const resultsContainer = document.getElementById('search-results');
        if (!resultsContainer) return;

        if (!data || !data.properties || data.properties.length === 0) {
            resultsContainer.innerHTML = `
                <div class="no-results">
                    <i class="fas fa-home fa-3x"></i>
                    <h3>No Properties Found</h3>
                    <p>Try adjusting your search filters or expanding your search area.</p>
                    <button class="btn btn-primary" onclick="PropertyHubSearch.clearAllFilters()">
                        Clear All Filters
                    </button>
                </div>
            `;
            return;
        }

        const resultsHTML = `
            <div class="search-results-header">
                <div class="results-info">
                    <h3>${data.total.toLocaleString()} Properties Found</h3>
                    <p>Showing ${((this.currentPage - 1) * this.resultsPerPage) + 1}-${Math.min(this.currentPage * this.resultsPerPage, data.total)} of ${data.total.toLocaleString()}</p>
                </div>
                
                <div class="results-controls">
                    <div class="sort-controls">
                        <label>Sort by:</label>
                        <select id="sort-select" class="form-control">
                            <option value="price_asc">Price: Low to High</option>
                            <option value="price_desc">Price: High to Low</option>
                            <option value="date_desc">Newest First</option>
                            <option value="date_asc">Oldest First</option>
                            <option value="sqft_desc">Largest First</option>
                            <option value="sqft_asc">Smallest First</option>
                        </select>
                    </div>
                    
                    <div class="view-controls">
                        <button class="view-btn ${this.viewMode === 'grid' ? 'active' : ''}" 
                                onclick="PropertyHubSearch.setViewMode('grid')">
                            <i class="fas fa-th-large"></i>
                        </button>
                        <button class="view-btn ${this.viewMode === 'list' ? 'active' : ''}" 
                                onclick="PropertyHubSearch.setViewMode('list')">
                            <i class="fas fa-list"></i>
                        </button>
                        <button class="view-btn ${this.viewMode === 'map' ? 'active' : ''}" 
                                onclick="PropertyHubSearch.setViewMode('map')">
                            <i class="fas fa-map"></i>
                        </button>
                    </div>
                    
                    <button class="btn btn-secondary save-search-btn" onclick="PropertyHubSearch.showSaveSearchModal()">
                        <i class="fas fa-save"></i> Save Search
                    </button>
                </div>
            </div>
            
            <div class="property-results ${this.viewMode || 'grid'}">
                ${data.properties.map(property => this.renderPropertyCard(property)).join('')}
            </div>
        `;

        resultsContainer.innerHTML = resultsHTML;
    }

    renderPropertyCard(property) {
        const isFavorite = this.favorites.includes(property.id);
        const mainPhoto = property.photos && property.photos.length > 0 ? property.photos[0].url : '/static/images/no-image.png';
        
        return `
            <div class="property-card" data-property-id="${property.id}">
                <div class="property-image-container">
                    <img src="${mainPhoto}" alt="${property.address}" class="property-image" loading="lazy">
                    <button class="favorite-btn ${isFavorite ? 'active' : ''}" 
                            onclick="PropertyHubSearch.toggleFavorite('${property.id}')">
                        <i class="fas fa-heart"></i>
                    </button>
                    <div class="property-badges">
                        ${property.status === 'new' ? '<span class="badge new">New</span>' : ''}
                        ${property.openHouse ? '<span class="badge open-house">Open House</span>' : ''}
                        ${property.priceReduced ? '<span class="badge price-reduced">Price Reduced</span>' : ''}
                    </div>
                    <div class="photo-count">
                        <i class="fas fa-camera"></i> ${property.photos?.length || 0}
                    </div>
                </div>
                
                <div class="property-content">
                    <div class="property-price">
                        $${property.price.toLocaleString()}
                        ${property.monthlyPayment ? `<span class="monthly-payment">$${property.monthlyPayment}/mo</span>` : ''}
                    </div>
                    
                    <div class="property-details">
                        <div class="property-specs">
                            <span class="spec">
                                <i class="fas fa-bed"></i> ${property.bedrooms} beds
                            </span>
                            <span class="spec">
                                <i class="fas fa-bath"></i> ${property.bathrooms} baths
                            </span>
                            <span class="spec">
                                <i class="fas fa-ruler-combined"></i> ${property.squareFootage?.toLocaleString()} sqft
                            </span>
                            ${property.lotSize ? `
                                <span class="spec">
                                    <i class="fas fa-map"></i> ${property.lotSize} acres
                                </span>
                            ` : ''}
                        </div>
                    </div>
                    
                    <div class="property-address">
                        <i class="fas fa-map-marker-alt"></i>
                        ${property.address}, ${property.city}, ${property.state} ${property.zipCode}
                    </div>
                    
                    <div class="property-highlights">
                        ${property.highlights?.slice(0, 3).map(highlight => `
                            <span class="highlight">${highlight}</span>
                        `).join('') || ''}
                    </div>
                    
                    <div class="property-listing-info">
                        <span class="listing-agent">${property.listingAgent?.name || 'MLS Listing'}</span>
                        <span class="days-on-market">${property.daysOnMarket} days on market</span>
                    </div>
                    
                    <div class="property-actions">
                        <button class="btn btn-primary" onclick="PropertyHubSearch.viewProperty('${property.id}')">
                            View Details
                        </button>
                        <button class="btn btn-secondary" onclick="PropertyHubSearch.scheduleShowing('${property.id}')">
                            Schedule Showing
                        </button>
                        <button class="btn btn-outline" onclick="PropertyHubSearch.getMoreInfo('${property.id}')">
                            Get Info
                        </button>
                    </div>
                </div>
            </div>
        `;
    }

    // Map Integration
    async initializeMap() {
        if (typeof google === 'undefined') {
            await this.loadGoogleMapsAPI();
        }

        const mapContainer = document.getElementById('search-map');
        if (!mapContainer) return;

        // Default to Houston, TX
        const defaultCenter = { lat: 29.7604, lng: -95.3698 };
        
        this.mapInstance = new google.maps.Map(mapContainer, {
            zoom: 10,
            center: defaultCenter,
            styles: this.getMapStyles()
        });

        // Setup map event listeners
        this.setupMapEventListeners();
    }

    async loadGoogleMapsAPI() {
        return new Promise((resolve, reject) => {
            if (typeof google !== 'undefined') {
                resolve();
                return;
            }

            const script = document.createElement('script');
            script.src = `https://maps.googleapis.com/maps/api/js?key=${window.GOOGLE_MAPS_API_KEY}&libraries=places`;
            script.async = true;
            script.defer = true;
            script.onload = resolve;
            script.onerror = reject;
            document.head.appendChild(script);
        });
    }

    updateMapMarkers(properties) {
        if (!this.mapInstance) return;

        // Clear existing markers
        this.mapMarkers.forEach(marker => marker.setMap(null));
        this.mapMarkers = [];

        // Add new markers
        properties.forEach(property => {
            if (property.latitude && property.longitude) {
                const marker = new google.maps.Marker({
                    position: { lat: property.latitude, lng: property.longitude },
                    map: this.mapInstance,
                    title: property.address,
                    icon: this.getPropertyMarkerIcon(property)
                });

                // Add info window
                const infoWindow = new google.maps.InfoWindow({
                    content: this.createMarkerInfoWindow(property)
                });

                marker.addListener('click', () => {
                    // Close other info windows
                    this.mapMarkers.forEach(m => {
                        if (m.infoWindow) m.infoWindow.close();
                    });
                    
                    infoWindow.open(this.mapInstance, marker);
                });

                marker.infoWindow = infoWindow;
                this.mapMarkers.push(marker);
            }
        });

        // Adjust map bounds to fit all markers
        if (this.mapMarkers.length > 0) {
            const bounds = new google.maps.LatLngBounds();
            this.mapMarkers.forEach(marker => {
                bounds.extend(marker.getPosition());
            });
            this.mapInstance.fitBounds(bounds);
        }
    }

    createMarkerInfoWindow(property) {
        const mainPhoto = property.photos && property.photos.length > 0 ? property.photos[0].url : '/static/images/no-image.png';
        
        return `
            <div class="map-info-window">
                <img src="${mainPhoto}" alt="${property.address}" class="info-image">
                <div class="info-content">
                    <div class="info-price">$${property.price.toLocaleString()}</div>
                    <div class="info-specs">
                        ${property.bedrooms} beds • ${property.bathrooms} baths • ${property.squareFootage?.toLocaleString()} sqft
                    </div>
                    <div class="info-address">${property.address}</div>
                    <button class="btn btn-primary btn-sm" onclick="PropertyHubSearch.viewProperty('${property.id}')">
                        View Details
                    </button>
                </div>
            </div>
        `;
    }

    // Saved Searches
    async saveSearch(searchName, alertSettings = {}) {
        const searchData = {
            name: searchName,
            filters: { ...this.currentFilters },
            sortBy: this.getSortBy(),
            alertSettings: alertSettings,
            createdAt: new Date().toISOString()
        };

        try {
            const response = await fetch('/api/properties/saved-searches', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                },
                body: JSON.stringify(searchData)
            });

            if (!response.ok) {
                throw new Error('Failed to save search');
            }

            const savedSearch = await response.json();
            this.savedSearches.push(savedSearch);
            this.renderSavedSearches();
            
            this.showNotification('Search saved successfully!', 'success');
            this.closeSaveSearchModal();
            
        } catch (error) {
            console.error('Error saving search:', error);
            this.showNotification('Failed to save search', 'error');
        }
    }

    async loadSavedSearches() {
        try {
            const response = await fetch('/api/properties/saved-searches', {
                headers: {
                    'Authorization': `Bearer ${window.PropertyHubAuth?.token}`
                }
            });

            if (response.ok) {
                const data = await response.json();
                this.savedSearches = data.searches || [];
                this.renderSavedSearches();
            }
        } catch (error) {
            console.error('Error loading saved searches:', error);
        }
    }

    // Auto-Complete
    setupAutoComplete() {
        const locationInput = document.getElementById('location-search');
        if (!locationInput || typeof google === 'undefined') return;

        const autocomplete = new google.maps.places.Autocomplete(locationInput, {
            types: ['address'],
            componentRestrictions: { country: 'us' }
        });

        autocomplete.addListener('place_changed', () => {
            const place = autocomplete.getPlace();
            if (place.geometry) {
                this.updateFilter('location', place.formatted_address);
                
                // Update map center if available
                if (this.mapInstance) {
                    this.mapInstance.setCenter(place.geometry.location);
                    this.mapInstance.setZoom(13);
                }
            }
        });
    }

    // Event Listeners
    setupEventListeners() {
        // Filter input events
        document.addEventListener('input', (e) => {
            if (e.target.matches('[data-filter]')) {
                const filterName = e.target.dataset.filter;
                let value = e.target.value;
                
                if (e.target.type === 'checkbox') {
                    value = e.target.checked;
                } else if (e.target.type === 'number') {
                    value = value ? parseInt(value) : null;
                }
                
                this.updateFilter(filterName, value);
            }
        });

        // Sort change
        document.addEventListener('change', (e) => {
            if (e.target.id === 'sort-select') {
                this.performSearch();
            }
        });

        // Search button
        document.addEventListener('click', (e) => {
            if (e.target.matches('.search-btn') || e.target.closest('.search-btn')) {
                e.preventDefault();
                this.performSearch();
            }
        });

        // Property card clicks
        document.addEventListener('click', (e) => {
            if (e.target.matches('.property-card') || e.target.closest('.property-card')) {
                if (!e.target.matches('button') && !e.target.closest('button')) {
                    const propertyId = e.target.closest('.property-card').dataset.propertyId;
                    this.viewProperty(propertyId);
                }
            }
        });

        // Keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            if (e.ctrlKey || e.metaKey) {
                switch (e.key) {
                    case 'f':
                        e.preventDefault();
                        document.getElementById('location-search')?.focus();
                        break;
                    case 's':
                        e.preventDefault();
                        this.showSaveSearchModal();
                        break;
                }
            }
        });
    }

    // Utility Methods
    debouncedSearch() {
        clearTimeout(this.searchTimeout);
        this.searchTimeout = setTimeout(() => {
            this.performSearch();
        }, this.config.minSearchDelay);
    }

    getSortBy() {
        const sortSelect = document.getElementById('sort-select');
        return sortSelect ? sortSelect.value : this.config.defaultSortBy;
    }

    trackSearchAnalytics(searchParams) {
        if (window.PropertyHubAnalytics) {
            window.PropertyHubAnalytics.trackPropertySearch(searchParams, this.totalResults);
        }
    }

    addToSearchHistory(searchParams) {
        const historyEntry = {
            filters: { ...searchParams.filters },
            timestamp: new Date().toISOString(),
            resultCount: this.totalResults
        };
        
        this.searchHistory.unshift(historyEntry);
        
        // Keep only last 50 searches
        if (this.searchHistory.length > 50) {
            this.searchHistory = this.searchHistory.slice(0, 50);
        }
        
        localStorage.setItem('propertySearchHistory', JSON.stringify(this.searchHistory));
    }

    loadSearchHistory() {
        const saved = localStorage.getItem('propertySearchHistory');
        if (saved) {
            this.searchHistory = JSON.parse(saved);
        }
    }

    showSearchLoading() {
        const resultsContainer = document.getElementById('search-results');
        if (resultsContainer) {
            resultsContainer.innerHTML = `
                <div class="search-loading">
                    <div class="spinner"></div>
                    <p>Searching properties...</p>
                </div>
            `;
        }
    }

    hideSearchLoading() {
        // Loading will be replaced by results
    }

    showSearchError(message) {
        const resultsContainer = document.getElementById('search-results');
        if (resultsContainer) {
            resultsContainer.innerHTML = `
                <div class="search-error">
                    <i class="fas fa-exclamation-triangle fa-2x"></i>
                    <h3>Search Error</h3>
                    <p>${message}</p>
                    <button class="btn btn-primary" onclick="PropertyHubSearch.performSearch()">
                        Try Again
                    </button>
                </div>
            `;
        }
    }

    showNotification(message, type = 'info') {
        if (window.PropertyHubDashboard) {
            window.PropertyHubDashboard.displayNotification({
                type: type,
                title: type === 'success' ? 'Success' : 'Notification',
                message: message
            });
        }
    }

    // Public API Methods
    viewProperty(propertyId) {
        // Track property view
        if (window.PropertyHubAnalytics) {
            window.PropertyHubAnalytics.trackPropertyView(propertyId);
        }
        
        window.location.href = `/properties/${propertyId}`;
    }

    scheduleShowing(propertyId) {
        window.location.href = `/properties/${propertyId}/schedule-showing`;
    }

    getMoreInfo(propertyId) {
        window.location.href = `/properties/${propertyId}/contact`;
    }

    toggleFavorite(propertyId) {
        const index = this.favorites.indexOf(propertyId);
        
        if (index > -1) {
            this.favorites.splice(index, 1);
        } else {
            this.favorites.push(propertyId);
        }
        
        // Update UI
        const favoriteBtn = document.querySelector(`[data-property-id="${propertyId}"] .favorite-btn`);
        if (favoriteBtn) {
            favoriteBtn.classList.toggle('active');
        }
        
        // Save to localStorage
        localStorage.setItem('propertyFavorites', JSON.stringify(this.favorites));
        
        // Track analytics
        if (window.PropertyHubAnalytics) {
            const action = index > -1 ? 'removed_favorite' : 'added_favorite';
            window.PropertyHubAnalytics.trackEvent('property_favorite', {
                action: action,
                propertyId: propertyId
            });
        }
    }

    setViewMode(mode) {
        this.viewMode = mode;
        
        // Update view controls
        document.querySelectorAll('.view-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        
        document.querySelector(`[onclick="PropertyHubSearch.setViewMode('${mode}')"]`)?.classList.add('active');
        
        // Update results display
        const resultsContainer = document.querySelector('.property-results');
        if (resultsContainer) {
            resultsContainer.className = `property-results ${mode}`;
        }
        
        // Show/hide map
        const mapContainer = document.getElementById('search-map');
        if (mapContainer) {
            mapContainer.style.display = mode === 'map' ? 'block' : 'none';
        }
    }

    toggleAdvancedFilters() {
        const advancedFilters = document.getElementById('advanced-filters');
        const isVisible = advancedFilters.style.display !== 'none';
        
        advancedFilters.style.display = isVisible ? 'none' : 'block';
        
        // Update button text
        const button = event.target.closest('button');
        if (button) {
            button.innerHTML = isVisible ? 
                '<i class="fas fa-sliders-h"></i> More Filters' : 
                '<i class="fas fa-sliders-h"></i> Fewer Filters';
        }
    }

    showSaveSearchModal() {
        // Implementation for save search modal
        console.log('Showing save search modal');
    }

    closeSaveSearchModal() {
        // Implementation for closing save search modal
        console.log('Closing save search modal');
    }
}

// Initialize PropertyHub Search when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    if (document.querySelector('#search-filters') || document.querySelector('#search-results')) {
        window.PropertyHubSearch = new PropertyHubSearch();
    }
});

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PropertyHubSearch;
}
