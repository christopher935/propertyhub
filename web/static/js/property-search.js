// Property Search & Filter Enhancement

document.addEventListener('DOMContentLoaded', function() {
    const filterPills = document.querySelectorAll('.filter-pill');
    const activeFiltersContainer = document.getElementById('activeFilters');
    const activeFilters = new Set();
    
    // Filter pill click handler
    filterPills.forEach(pill => {
        pill.addEventListener('click', function() {
            const filter = this.dataset.filter;
            
            // Handle "All Properties" special case
            if (filter === 'all') {
                filterPills.forEach(p => p.classList.remove('active'));
                this.classList.add('active');
                activeFilters.clear();
                updateActiveFiltersDisplay();
                return;
            }
            
            // Remove "All" if any specific filter is selected
            const allPill = document.querySelector('.filter-pill[data-filter="all"]');
            if (allPill) {
                allPill.classList.remove('active');
            }
            
            // Toggle current filter
            if (this.classList.contains('active')) {
                this.classList.remove('active');
                activeFilters.delete(filter);
                
                // If no filters active, activate "All"
                if (activeFilters.size === 0 && allPill) {
                    allPill.classList.add('active');
                }
            } else {
                this.classList.add('active');
                activeFilters.add(filter);
            }
            
            updateActiveFiltersDisplay();
            applyFilters();
        });
    });
    
    function updateActiveFiltersDisplay() {
        if (!activeFiltersContainer) return;
        
        if (activeFilters.size === 0) {
            activeFiltersContainer.style.display = 'none';
            return;
        }
        
        activeFiltersContainer.style.display = 'flex';
        activeFiltersContainer.innerHTML = '';
        
        activeFilters.forEach(filter => {
            const filterTag = document.createElement('span');
            filterTag.className = 'filter-tag';
            
            const filterText = document.querySelector(`.filter-pill[data-filter="${filter}"]`)?.textContent || filter;
            filterTag.innerHTML = `
                ${filterText}
                <button class="filter-remove" data-filter="${filter}" aria-label="Remove filter">×</button>
            `;
            
            // Remove button handler
            const removeBtn = filterTag.querySelector('.filter-remove');
            removeBtn.addEventListener('click', function(e) {
                e.stopPropagation();
                const filterToRemove = this.dataset.filter;
                activeFilters.delete(filterToRemove);
                
                const pill = document.querySelector(`.filter-pill[data-filter="${filterToRemove}"]`);
                if (pill) {
                    pill.classList.remove('active');
                }
                
                // If no filters active, activate "All"
                if (activeFilters.size === 0) {
                    const allPill = document.querySelector('.filter-pill[data-filter="all"]');
                    if (allPill) {
                        allPill.classList.add('active');
                    }
                }
                
                updateActiveFiltersDisplay();
                applyFilters();
            });
            
            activeFiltersContainer.appendChild(filterTag);
        });
    }
    
    function applyFilters() {
        // This would typically trigger a backend API call or client-side filtering
        console.log('Active filters:', Array.from(activeFilters));
        
        // Show loading skeleton
        const loadingSkeleton = document.getElementById('loadingSkeleton');
        const propertyGrid = document.getElementById('propertyGrid');
        
        if (loadingSkeleton && propertyGrid) {
            loadingSkeleton.classList.remove('hidden');
            propertyGrid.style.opacity = '0.5';
            
            // Simulate API call with timeout
            setTimeout(() => {
                loadingSkeleton.classList.add('hidden');
                propertyGrid.style.opacity = '1';
            }, 500);
        }
    }
    
    // Clear all filters button
    const clearFiltersButtons = document.querySelectorAll('.clear-filters-btn');
    clearFiltersButtons.forEach(btn => {
        btn.addEventListener('click', function() {
            activeFilters.clear();
            filterPills.forEach(p => p.classList.remove('active'));
            
            const allPill = document.querySelector('.filter-pill[data-filter="all"]');
            if (allPill) {
                allPill.classList.add('active');
            }
            
            // Reset selects
            document.querySelectorAll('.form-select').forEach(select => {
                select.selectedIndex = 0;
            });
            
            updateActiveFiltersDisplay();
            applyFilters();
        });
    });
    
    // Advanced filter selects
    const advancedFilters = document.querySelectorAll('.filter-group-select');
    advancedFilters.forEach(select => {
        select.addEventListener('change', function() {
            applyFilters();
        });
    });
});
