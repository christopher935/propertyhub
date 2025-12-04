/**
 * Property List Page - Interactive Behaviors
 * Behavioral Excellence Layer
 */

(function() {
  'use strict';

  // ========== PROPERTY CARD CLICKS ==========
  
  function initPropertyCards() {
    const cards = document.querySelectorAll('.property-card');
    
    cards.forEach(card => {
      card.addEventListener('click', function(e) {
        // Don't trigger if clicking action buttons
        if (e.target.closest('.property-action-btn')) return;
        
        const address = this.querySelector('.property-address')?.textContent;
        console.log('Property card clicked:', address);
        
        // TODO: Navigate to property detail page or open modal
        // window.location.href = `/admin/property/${propertyId}`;
      });
    });
  }

  // ========== FILTER FUNCTIONALITY ==========
  
  function initFilters() {
    const filterSelects = document.querySelectorAll('.filter-select');
    const filterInput = document.querySelector('.filter-input');
    const clearBtn = document.querySelector('.filter-btn');
    const propertiesGrid = document.querySelector('.properties-grid');
    
    // Mark active filters
    filterSelects.forEach(select => {
      select.addEventListener('change', function() {
        if (this.value && this.value !== this.options[0].value) {
          this.classList.add('active');
        } else {
          this.classList.remove('active');
        }
        applyFilters();
      });
    });
    
    // Search input
    if (filterInput) {
      let searchTimeout;
      filterInput.addEventListener('input', function() {
        clearTimeout(searchTimeout);
        
        if (this.value) {
          this.classList.add('has-value');
        } else {
          this.classList.remove('has-value');
        }
        
        // Debounce search
        searchTimeout = setTimeout(() => {
          applyFilters();
        }, 300);
      });
    }
    
    // Clear filters
    if (clearBtn) {
      clearBtn.addEventListener('click', function() {
        filterSelects.forEach(select => {
          select.selectedIndex = 0;
          select.classList.remove('active');
        });
        
        if (filterInput) {
          filterInput.value = '';
          filterInput.classList.remove('has-value');
        }
        
        applyFilters();
      });
    }
    
    function applyFilters() {
      if (!propertiesGrid) return;
      
      // Show loading state
      propertiesGrid.classList.add('loading');
      
      // Simulate filter delay
      setTimeout(() => {
        propertiesGrid.classList.remove('loading');
        console.log('Filters applied');
        // TODO: Actual filtering logic or API call
      }, 300);
    }
  }

  // ========== TABLE SORTING ==========
  
  function initTableSorting() {
    const tables = document.querySelectorAll('.performance-table, .showing-schedule, .pre-listing-table');
    
    tables.forEach(table => {
      const headers = table.querySelectorAll('th[data-sortable]');
      
      headers.forEach(header => {
        header.style.cursor = 'pointer';
        
        header.addEventListener('click', function() {
          const column = this.dataset.sortable;
          const currentSort = this.dataset.sort || 'none';
          
          // Remove sorted class from all headers
          headers.forEach(h => h.classList.remove('sorted'));
          
          // Toggle sort direction
          let newSort;
          if (currentSort === 'none' || currentSort === 'desc') {
            newSort = 'asc';
          } else {
            newSort = 'desc';
          }
          
          this.dataset.sort = newSort;
          this.classList.add('sorted');
          
          // Update sort indicator
          updateSortIndicator(this, newSort);
          
          // Perform sort
          sortTable(table, column, newSort);
        });
      });
    });
  }
  
  function updateSortIndicator(header, direction) {
    let indicator = header.querySelector('.sort-indicator');
    
    if (!indicator) {
      indicator = document.createElement('span');
      indicator.className = 'sort-indicator';
      header.appendChild(indicator);
    }
    
    indicator.innerHTML = direction === 'asc' ? '↑' : '↓';
  }
  
  function sortTable(table, column, direction) {
    const tbody = table.querySelector('tbody');
    const rows = Array.from(tbody.querySelectorAll('tr'));
    
    rows.sort((a, b) => {
      const aValue = a.querySelector(`[data-column="${column}"]`)?.textContent || '';
      const bValue = b.querySelector(`[data-column="${column}"]`)?.textContent || '';
      
      // Try numeric comparison first
      const aNum = parseFloat(aValue.replace(/[^0-9.-]/g, ''));
      const bNum = parseFloat(bValue.replace(/[^0-9.-]/g, ''));
      
      if (!isNaN(aNum) && !isNaN(bNum)) {
        return direction === 'asc' ? aNum - bNum : bNum - aNum;
      }
      
      // Fall back to string comparison
      return direction === 'asc' 
        ? aValue.localeCompare(bValue)
        : bValue.localeCompare(aValue);
    });
    
    // Re-append rows in sorted order
    rows.forEach(row => tbody.appendChild(row));
  }

  // ========== AI SUGGESTION ACTIONS ==========
  
  function initAISuggestions() {
    const aiButtons = document.querySelectorAll('.ai-action-btn');
    
    aiButtons.forEach(button => {
      button.addEventListener('click', function() {
        const action = this.textContent.trim();
        
        // Show loading state
        this.classList.add('loading');
        this.disabled = true;
        
        // Simulate action
        setTimeout(() => {
          this.classList.remove('loading');
          this.classList.add('success');
          this.disabled = false;
          
          // Show toast
          showToast(`${action} completed!`, 'success');
          
          // Reset after 2 seconds
          setTimeout(() => {
            this.classList.remove('success');
          }, 2000);
        }, 1000);
      });
    });
  }

  // ========== TOAST NOTIFICATIONS ==========
  
  function showToast(message, type = 'info') {
    let container = document.querySelector('.toast-container');
    
    if (!container) {
      container = document.createElement('div');
      container.className = 'toast-container';
      document.body.appendChild(container);
    }
    
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    
    const icon = type === 'success' ? '✓' : type === 'error' ? '✕' : 'ℹ';
    
    toast.innerHTML = `
      <div class="toast-icon">${icon}</div>
      <div class="toast-message">${message}</div>
      <button class="toast-close">×</button>
    `;
    
    container.appendChild(toast);
    
    // Close button
    toast.querySelector('.toast-close').addEventListener('click', () => {
      toast.remove();
    });
    
    // Auto-dismiss after 3 seconds
    setTimeout(() => {
      toast.style.animation = 'slideOutRight 300ms ease-out';
      setTimeout(() => toast.remove(), 300);
    }, 3000);
  }

  // ========== INITIALIZE ALL ==========
  
  document.addEventListener('DOMContentLoaded', function() {
    initPropertyCards();
    initFilters();
    initTableSorting();
    initAISuggestions();
    
    console.log('Property List interactions initialized');
  });

})();
