document.addEventListener('DOMContentLoaded', function() {
    const filterPills = document.querySelectorAll('.filter-pill');
    const activeFiltersContainer = document.getElementById('activeFilters');
    const activeFilters = new Set();
    const propertyGrid = document.getElementById('propertyGrid');
    const loadingSkeleton = document.getElementById('loadingSkeleton');
    const emptyState = document.getElementById('emptyState');
    const searchInput = document.getElementById('searchInput');
    
    let currentFilters = {
        bedrooms: '',
        bathrooms: '',
        city: '',
        sortBy: 'created_at',
        sortOrder: 'desc',
        search: ''
    };

    function initFromURL() {
        const params = new URLSearchParams(window.location.search);
        
        params.getAll('filter').forEach(filter => {
            activeFilters.add(filter);
            const pill = document.querySelector(`.filter-pill[data-filter="${filter}"]`);
            if (pill) {
                pill.classList.add('active');
            }
        });

        if (params.has('bedrooms')) currentFilters.bedrooms = params.get('bedrooms');
        if (params.has('bathrooms')) currentFilters.bathrooms = params.get('bathrooms');
        if (params.has('city')) currentFilters.city = params.get('city');
        if (params.has('sort_by')) currentFilters.sortBy = params.get('sort_by');
        if (params.has('search')) {
            currentFilters.search = params.get('search');
            if (searchInput) searchInput.value = currentFilters.search;
        }

        const bedroomSelect = document.getElementById('bedroomFilter');
        const bathroomSelect = document.getElementById('bathroomFilter');
        const locationSelect = document.getElementById('locationFilter');
        const sortSelect = document.getElementById('sortSelect');

        if (bedroomSelect && currentFilters.bedrooms) {
            bedroomSelect.value = currentFilters.bedrooms;
        }
        if (bathroomSelect && currentFilters.bathrooms) {
            bathroomSelect.value = currentFilters.bathrooms;
        }
        if (locationSelect && currentFilters.city) {
            locationSelect.value = currentFilters.city;
        }
        if (sortSelect && currentFilters.sortBy) {
            const sortMap = {
                'created_at': 'newest',
                'price_asc': 'price-low',
                'price_desc': 'price-high',
                'bedrooms': 'bedrooms',
                'square_feet': 'size'
            };
            const sortKey = currentFilters.sortBy + (currentFilters.sortOrder === 'asc' ? '_asc' : '');
            for (const [key, value] of Object.entries(sortMap)) {
                if (key === sortKey || key === currentFilters.sortBy) {
                    sortSelect.value = value;
                    break;
                }
            }
        }

        updateActiveFiltersDisplay();
        
        if (params.toString() && activeFilters.size > 1) {
            applyFilters();
        }
    }

    function updateURL() {
        const params = new URLSearchParams();
        
        activeFilters.forEach(filter => {
            if (filter !== 'all') {
                params.append('filter', filter);
            }
        });
        
        if (currentFilters.bedrooms) params.set('bedrooms', currentFilters.bedrooms);
        if (currentFilters.bathrooms) params.set('bathrooms', currentFilters.bathrooms);
        if (currentFilters.city) params.set('city', currentFilters.city);
        if (currentFilters.sortBy !== 'created_at') params.set('sort_by', currentFilters.sortBy);
        if (currentFilters.search) params.set('search', currentFilters.search);
        
        const newURL = params.toString() ? 
            `${window.location.pathname}?${params}` : 
            window.location.pathname;
        
        window.history.replaceState({}, '', newURL);
    }
    
    filterPills.forEach(pill => {
        pill.addEventListener('click', function() {
            const filter = this.dataset.filter;
            
            if (filter === 'all') {
                filterPills.forEach(p => p.classList.remove('active'));
                this.classList.add('active');
                activeFilters.clear();
                activeFilters.add('all');
                updateActiveFiltersDisplay();
                applyFilters();
                return;
            }
            
            const allPill = document.querySelector('.filter-pill[data-filter="all"]');
            if (allPill) {
                allPill.classList.remove('active');
                activeFilters.delete('all');
            }
            
            if (this.classList.contains('active')) {
                this.classList.remove('active');
                activeFilters.delete(filter);
                
                if (activeFilters.size === 0 && allPill) {
                    allPill.classList.add('active');
                    activeFilters.add('all');
                }
            } else {
                const priceFilters = ['price-low', 'price-mid', 'price-high'];
                if (priceFilters.includes(filter)) {
                    priceFilters.forEach(pf => {
                        activeFilters.delete(pf);
                        const pfPill = document.querySelector(`.filter-pill[data-filter="${pf}"]`);
                        if (pfPill) pfPill.classList.remove('active');
                    });
                }
                
                this.classList.add('active');
                activeFilters.add(filter);
            }
            
            updateActiveFiltersDisplay();
            applyFilters();
        });
    });
    
    function updateActiveFiltersDisplay() {
        if (!activeFiltersContainer) return;
        
        const displayFilters = Array.from(activeFilters).filter(f => f !== 'all');
        
        if (displayFilters.length === 0) {
            activeFiltersContainer.style.display = 'none';
            return;
        }
        
        activeFiltersContainer.style.display = 'flex';
        activeFiltersContainer.innerHTML = '';
        
        displayFilters.forEach(filter => {
            const filterTag = document.createElement('span');
            filterTag.className = 'filter-tag';
            
            const filterText = document.querySelector(`.filter-pill[data-filter="${filter}"]`)?.textContent || filter;
            filterTag.innerHTML = `
                ${filterText}
                <button class="filter-remove" data-filter="${filter}" aria-label="Remove filter">×</button>
            `;
            
            const removeBtn = filterTag.querySelector('.filter-remove');
            removeBtn.addEventListener('click', function(e) {
                e.stopPropagation();
                const filterToRemove = this.dataset.filter;
                activeFilters.delete(filterToRemove);
                
                const pill = document.querySelector(`.filter-pill[data-filter="${filterToRemove}"]`);
                if (pill) {
                    pill.classList.remove('active');
                }
                
                if (activeFilters.size === 0) {
                    const allPill = document.querySelector('.filter-pill[data-filter="all"]');
                    if (allPill) {
                        allPill.classList.add('active');
                        activeFilters.add('all');
                    }
                }
                
                updateActiveFiltersDisplay();
                applyFilters();
            });
            
            activeFiltersContainer.appendChild(filterTag);
        });
    }
    
    async function applyFilters() {
        showLoadingSkeleton();
        
        const params = new URLSearchParams();
        
        if (activeFilters.has('price-low')) {
            params.set('max_price', '3500');
        } else if (activeFilters.has('price-mid')) {
            params.set('min_price', '3500');
            params.set('max_price', '4500');
        } else if (activeFilters.has('price-high')) {
            params.set('min_price', '4500');
        }
        
        if (currentFilters.bedrooms) {
            params.set('bedrooms', currentFilters.bedrooms);
        }
        if (currentFilters.bathrooms) {
            params.set('bathrooms', currentFilters.bathrooms);
        }
        if (currentFilters.city) {
            params.set('city', currentFilters.city);
        }
        if (currentFilters.search) {
            params.set('search', currentFilters.search);
        }
        
        if (currentFilters.sortBy) {
            params.set('sort_by', currentFilters.sortBy);
            if (currentFilters.sortOrder) {
                params.set('sort_order', currentFilters.sortOrder);
            }
        }
        
        params.set('page', '1');
        params.set('limit', '20');
        
        try {
            const response = await fetch(`/api/v1/properties?${params}`);
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data = await response.json();
            
            if (data.success) {
                const properties = data.data || [];
                const pagination = data.pagination || {};
                
                renderProperties(properties);
                updateResultsCount(pagination.total_count || properties.length);
            } else {
                throw new Error(data.message || 'Failed to fetch properties');
            }
        } catch (error) {
            console.error('Filter error:', error);
            showErrorState('Unable to load properties. Please try again.');
        } finally {
            hideLoadingSkeleton();
            updateURL();
        }
    }
    
    function showLoadingSkeleton() {
        if (loadingSkeleton) {
            loadingSkeleton.classList.remove('hidden');
        }
        if (propertyGrid) {
            propertyGrid.style.opacity = '0.3';
            propertyGrid.style.pointerEvents = 'none';
        }
        if (emptyState) {
            emptyState.classList.add('hidden');
        }
    }
    
    function hideLoadingSkeleton() {
        if (loadingSkeleton) {
            loadingSkeleton.classList.add('hidden');
        }
        if (propertyGrid) {
            propertyGrid.style.opacity = '1';
            propertyGrid.style.pointerEvents = 'auto';
        }
    }
    
    function renderProperties(properties) {
        if (!propertyGrid) return;
        
        if (!properties || properties.length === 0) {
            propertyGrid.classList.add('hidden');
            if (emptyState) {
                emptyState.classList.remove('hidden');
            }
            return;
        }
        
        propertyGrid.classList.remove('hidden');
        if (emptyState) {
            emptyState.classList.add('hidden');
        }
        
        propertyGrid.innerHTML = properties.map(property => createPropertyCard(property)).join('');
    }
    
    function createPropertyCard(property) {
        const image = property.featured_image || property.images?.[0] || '/static/images/placeholder-property.jpg';
        const price = property.price ? `$${Math.round(property.price).toLocaleString()}` : 'Contact for Price';
        const bedrooms = property.bedrooms || '—';
        const bathrooms = property.bathrooms || '—';
        const sqft = property.square_feet ? property.square_feet.toLocaleString() : '—';
        const city = property.city || 'Houston';
        const state = property.state || 'TX';
        const address = property.address || 'Address Available Upon Request';
        const neighborhood = property.neighborhood || city;
        const zipCode = property.zip_code || '';
        
        const statusBadgeClass = property.status === 'available' ? 'success' : 
                                 property.status === 'just-listed' ? 'info' : 'secondary';
        const statusText = property.status === 'available' ? 'Available' :
                          property.status === 'just-listed' ? 'New Listing' : 'For Rent';
        
        return `
            <div class="property-card" data-property-id="${property.id}">
                <div class="property-image">
                    <img src="${image}" alt="${address}" loading="lazy" onerror="this.style.display='none'; this.parentElement.querySelector('.property-placeholder').style.display='flex'">
                    <div class="property-placeholder" style="display: none;">
                        <svg width="120" height="120" style="color: var(--gray-400);"><use href="#icon-home"></use></svg>
                    </div>
                    <span class="badge badge-${statusBadgeClass}">${statusText}</span>
                </div>
                <div class="property-details">
                    <h3 class="property-title">
                        <a href="/properties/${property.id}">${address}</a>
                    </h3>
                    <p class="property-location">
                        <svg width="16" height="16"><use href="#icon-map-pin"></use></svg>
                        ${neighborhood}, ${city}, ${state} ${zipCode}
                    </p>
                    <div class="property-price">${price}/month</div>
                    <p class="property-specs">${bedrooms} bed • ${bathrooms} bath • ${sqft} sq ft</p>
                    <a class="btn btn-primary" href="/book-showing?property=${property.id}">Schedule A Tour</a>
                </div>
            </div>
        `;
    }
    
    function showErrorState(message) {
        if (!propertyGrid) return;
        
        propertyGrid.innerHTML = `
            <div style="grid-column: 1 / -1; text-align: center; padding: 3rem;">
                <h3 style="color: var(--error-500); margin-bottom: 1rem;">Error Loading Properties</h3>
                <p style="color: var(--gray-600);">${message}</p>
                <button class="btn btn-primary" onclick="location.reload()" style="margin-top: 1.5rem;">
                    Retry
                </button>
            </div>
        `;
    }
    
    function updateResultsCount(count) {
        const countElements = document.querySelectorAll('.properties-count .count-number');
        countElements.forEach(el => {
            el.textContent = count || 0;
        });
    }
    
    const clearFiltersButtons = document.querySelectorAll('.clear-filters-btn');
    clearFiltersButtons.forEach(btn => {
        btn.addEventListener('click', function() {
            activeFilters.clear();
            activeFilters.add('all');
            filterPills.forEach(p => p.classList.remove('active'));
            
            const allPill = document.querySelector('.filter-pill[data-filter="all"]');
            if (allPill) {
                allPill.classList.add('active');
            }
            
            document.querySelectorAll('.form-select').forEach(select => {
                select.selectedIndex = 0;
            });
            
            currentFilters = {
                bedrooms: '',
                bathrooms: '',
                city: '',
                sortBy: 'created_at',
                sortOrder: 'desc',
                search: ''
            };
            
            if (searchInput) {
                searchInput.value = '';
            }
            
            updateActiveFiltersDisplay();
            applyFilters();
        });
    });
    
    const bedroomFilter = document.getElementById('bedroomFilter');
    const bathroomFilter = document.getElementById('bathroomFilter');
    const locationFilter = document.getElementById('locationFilter');
    const sortSelect = document.getElementById('sortSelect');
    
    if (bedroomFilter) {
        bedroomFilter.addEventListener('change', function() {
            currentFilters.bedrooms = this.value;
            applyFilters();
        });
    }
    
    if (bathroomFilter) {
        bathroomFilter.addEventListener('change', function() {
            currentFilters.bathrooms = this.value;
            applyFilters();
        });
    }
    
    if (locationFilter) {
        locationFilter.addEventListener('change', function() {
            currentFilters.city = this.value;
            applyFilters();
        });
    }
    
    if (sortSelect) {
        sortSelect.addEventListener('change', function() {
            const sortValue = this.value;
            
            switch(sortValue) {
                case 'newest':
                    currentFilters.sortBy = 'created_at';
                    currentFilters.sortOrder = 'desc';
                    break;
                case 'price-low':
                    currentFilters.sortBy = 'price';
                    currentFilters.sortOrder = 'asc';
                    break;
                case 'price-high':
                    currentFilters.sortBy = 'price';
                    currentFilters.sortOrder = 'desc';
                    break;
                case 'bedrooms':
                    currentFilters.sortBy = 'bedrooms';
                    currentFilters.sortOrder = 'desc';
                    break;
                case 'size':
                    currentFilters.sortBy = 'square_feet';
                    currentFilters.sortOrder = 'desc';
                    break;
                default:
                    currentFilters.sortBy = 'created_at';
                    currentFilters.sortOrder = 'desc';
            }
            
            applyFilters();
        });
    }
    
    if (searchInput) {
        let searchTimeout;
        searchInput.addEventListener('input', function() {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(() => {
                currentFilters.search = this.value.trim();
                applyFilters();
            }, 500);
        });
        
        searchInput.addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                clearTimeout(searchTimeout);
                currentFilters.search = this.value.trim();
                applyFilters();
            }
        });
    }
    
    const searchButton = document.querySelector('.properties-search-compact .btn-primary');
    if (searchButton) {
        searchButton.addEventListener('click', function() {
            if (searchInput) {
                currentFilters.search = searchInput.value.trim();
                applyFilters();
            }
        });
    }
    
    initFromURL();
});
