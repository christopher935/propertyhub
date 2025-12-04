/**
 * PropertyHub Alpine.js Components
 * Reusable Alpine.js components for property displays across the site
 * SSR + Alpine.js hybrid approach
 */

/**
 * Component 1: Property Pagination
 * Handles client-side pagination for property lists
 * Usage: x-data="propertyPagination(properties, 9)"
 */
document.addEventListener('alpine:init', () => {
    Alpine.data('propertyPagination', (properties = [], perPage = 9) => ({
        allProperties: properties,
        perPage: perPage,
        currentPage: 1,
        
        get totalPages() {
            return Math.ceil(this.allProperties.length / this.perPage);
        },
        
        get paginatedProperties() {
            const start = (this.currentPage - 1) * this.perPage;
            const end = start + this.perPage;
            return this.allProperties.slice(start, end);
        },
        
        get startIndex() {
            return (this.currentPage - 1) * this.perPage + 1;
        },
        
        get endIndex() {
            const end = this.currentPage * this.perPage;
            return end > this.allProperties.length ? this.allProperties.length : end;
        },
        
        get pageNumbers() {
            const pages = [];
            for (let i = 1; i <= this.totalPages; i++) {
                pages.push(i);
            }
            return pages;
        },
        
        nextPage() {
            if (this.currentPage < this.totalPages) {
                this.currentPage++;
            }
        },
        
        prevPage() {
            if (this.currentPage > 1) {
                this.currentPage--;
            }
        },
        
        goToPage(page) {
            if (page >= 1 && page <= this.totalPages) {
                this.currentPage = page;
            }
        },
        
        isCurrentPage(page) {
            return this.currentPage === page;
        }
    }));

    /**
     * Component 2: Property Filter
     * Handles filtering and sorting for property lists
     * Usage: x-data="propertyFilter(properties)"
     */
    Alpine.data('propertyFilter', (properties = []) => ({
        allProperties: properties,
        filters: {
            city: '',
            minPrice: 0,
            maxPrice: 999999,
            minBeds: 0,
            maxBeds: 99,
            minBaths: 0,
            maxBaths: 99,
            minSqft: 0,
            maxSqft: 999999,
            status: ''
        },
        sortBy: 'created_at',
        sortOrder: 'desc',
        
        get filteredProperties() {
            let filtered = [...this.allProperties];
            
            // Filter by city
            if (this.filters.city) {
                const cityLower = this.filters.city.toLowerCase();
                filtered = filtered.filter(p => 
                    p.City && p.City.toLowerCase().includes(cityLower)
                );
            }
            
            // Filter by price range
            if (this.filters.minPrice > 0) {
                filtered = filtered.filter(p => p.Price >= this.filters.minPrice);
            }
            if (this.filters.maxPrice < 999999) {
                filtered = filtered.filter(p => p.Price <= this.filters.maxPrice);
            }
            
            // Filter by bedrooms
            if (this.filters.minBeds > 0) {
                filtered = filtered.filter(p => p.Bedrooms >= this.filters.minBeds);
            }
            if (this.filters.maxBeds < 99) {
                filtered = filtered.filter(p => p.Bedrooms <= this.filters.maxBeds);
            }
            
            // Filter by bathrooms
            if (this.filters.minBaths > 0) {
                filtered = filtered.filter(p => p.Bathrooms >= this.filters.minBaths);
            }
            if (this.filters.maxBaths < 99) {
                filtered = filtered.filter(p => p.Bathrooms <= this.filters.maxBaths);
            }
            
            // Filter by square footage
            if (this.filters.minSqft > 0) {
                filtered = filtered.filter(p => p.SquareFeet >= this.filters.minSqft);
            }
            if (this.filters.maxSqft < 999999) {
                filtered = filtered.filter(p => p.SquareFeet <= this.filters.maxSqft);
            }
            
            // Filter by status
            if (this.filters.status) {
                filtered = filtered.filter(p => p.Status === this.filters.status);
            }
            
            // Sort
            filtered.sort((a, b) => {
                let aVal = a[this.sortBy];
                let bVal = b[this.sortBy];
                
                // Handle null/undefined values
                if (aVal === null || aVal === undefined) aVal = 0;
                if (bVal === null || bVal === undefined) bVal = 0;
                
                if (this.sortOrder === 'asc') {
                    return aVal > bVal ? 1 : -1;
                } else {
                    return aVal < bVal ? 1 : -1;
                }
            });
            
            return filtered;
        },
        
        get filterCount() {
            let count = 0;
            if (this.filters.city) count++;
            if (this.filters.minPrice > 0) count++;
            if (this.filters.maxPrice < 999999) count++;
            if (this.filters.minBeds > 0) count++;
            if (this.filters.maxBeds < 99) count++;
            if (this.filters.minBaths > 0) count++;
            if (this.filters.maxBaths < 99) count++;
            if (this.filters.minSqft > 0) count++;
            if (this.filters.maxSqft < 999999) count++;
            if (this.filters.status) count++;
            return count;
        },
        
        resetFilters() {
            this.filters = {
                city: '',
                minPrice: 0,
                maxPrice: 999999,
                minBeds: 0,
                maxBeds: 99,
                minBaths: 0,
                maxBaths: 99,
                minSqft: 0,
                maxSqft: 999999,
                status: ''
            };
        },
        
        setSortBy(field) {
            if (this.sortBy === field) {
                // Toggle order if clicking same field
                this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
            } else {
                this.sortBy = field;
                this.sortOrder = 'asc';
            }
        }
    }));

    /**
     * Component 3: Property Search
     * Handles real-time search for properties
     * Usage: x-data="propertySearch(properties)"
     */
    Alpine.data('propertySearch', (properties = []) => ({
        allProperties: properties,
        searchQuery: '',
        searchFields: ['Address', 'City', 'Description', 'Neighborhood'],
        
        get searchedProperties() {
            if (!this.searchQuery || this.searchQuery.length < 2) {
                return this.allProperties;
            }
            
            const query = this.searchQuery.toLowerCase();
            return this.allProperties.filter(property => {
                return this.searchFields.some(field => {
                    const value = property[field];
                    return value && value.toString().toLowerCase().includes(query);
                });
            });
        },
        
        get searchResultCount() {
            return this.searchedProperties.length;
        },
        
        clearSearch() {
            this.searchQuery = '';
        }
    }));

    /**
     * Component 4: Combined Property Manager
     * Combines pagination, filtering, and search
     * Usage: x-data="propertyManager(properties, { perPage: 9 })"
     */
    Alpine.data('propertyManager', (properties = [], options = {}) => ({
        // Merge all component functionality
        ...Alpine.raw(Alpine.data('propertyPagination')(properties, options.perPage || 9)),
        ...Alpine.raw(Alpine.data('propertyFilter')(properties)),
        ...Alpine.raw(Alpine.data('propertySearch')(properties)),
        
        // Override to chain filtering -> searching -> pagination
        get displayedProperties() {
            // First filter
            let result = this.filteredProperties;
            
            // Then search
            if (this.searchQuery && this.searchQuery.length >= 2) {
                const query = this.searchQuery.toLowerCase();
                result = result.filter(property => {
                    return this.searchFields.some(field => {
                        const value = property[field];
                        return value && value.toString().toLowerCase().includes(query);
                    });
                });
            }
            
            // Update allProperties for pagination
            this.allProperties = result;
            
            // Then paginate
            const start = (this.currentPage - 1) * this.perPage;
            const end = start + this.perPage;
            return result.slice(start, end);
        },
        
        get totalResults() {
            return this.filteredProperties.length;
        }
    }));
});
