/**
 * PropertyHub Navigation & Header - Interactive Behaviors
 * Implements: Section expansion, user menu dropdown, search, keyboard navigation
 * Methodology: Cosmetic Polish + Behavioral Excellence = Pixel-Perfect Excellence
 */

// ============================================
// NAVIGATION SECTION EXPANSION/COLLAPSE
// ============================================

function initNavigationExpansion() {
  const navSections = document.querySelectorAll('.nav-section');
  
  navSections.forEach(section => {
    const header = section.querySelector('.nav-section-header');
    const toggle = section.querySelector('.nav-section-toggle');
    
    if (!header || !toggle) return; // Skip if no toggle (like Overview)
    
    header.addEventListener('click', (e) => {
      e.preventDefault();
      toggleNavSection(section);
    });
    
    // Keyboard support
    header.addEventListener('keydown', (e) => {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        toggleNavSection(section);
      }
    });
  });
}

function toggleNavSection(section) {
  const isExpanded = section.classList.contains('expanded');
  
  // Optional: Close other sections (accordion behavior)
  // document.querySelectorAll('.nav-section.expanded').forEach(s => {
  //   if (s !== section) s.classList.remove('expanded');
  // });
  
  if (isExpanded) {
    section.classList.remove('expanded');
  } else {
    section.classList.add('expanded');
  }
}

// ============================================
// USER MENU DROPDOWN
// ============================================

function initUserMenuDropdown() {
  const userMenuButton = document.querySelector('.admin-user-menu');
  const userDropdown = document.querySelector('.admin-user-dropdown');
  
  if (!userMenuButton || !userDropdown) return;
  
  // Toggle dropdown on click
  userMenuButton.addEventListener('click', (e) => {
    e.stopPropagation();
    toggleUserDropdown();
  });
  
  // Close dropdown when clicking outside
  document.addEventListener('click', (e) => {
    if (!e.target.closest('.admin-user-menu-container')) {
      closeUserDropdown();
    }
  });
  
  // Keyboard support
  userMenuButton.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      toggleUserDropdown();
    } else if (e.key === 'Escape') {
      closeUserDropdown();
    }
  });
  
  // Close on Escape key
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      closeUserDropdown();
    }
  });
}

function toggleUserDropdown() {
  const userMenuButton = document.querySelector('.admin-user-menu');
  const userDropdown = document.querySelector('.admin-user-dropdown');
  
  const isOpen = userDropdown.classList.contains('show');
  
  if (isOpen) {
    closeUserDropdown();
  } else {
    openUserDropdown();
  }
}

function openUserDropdown() {
  const userMenuButton = document.querySelector('.admin-user-menu');
  const userDropdown = document.querySelector('.admin-user-dropdown');
  
  userMenuButton.classList.add('open');
  userDropdown.classList.add('show');
  userDropdown.style.display = 'block';
  
  // Focus first item
  const firstItem = userDropdown.querySelector('.dropdown-item');
  if (firstItem) {
    setTimeout(() => firstItem.focus(), 100);
  }
}

function closeUserDropdown() {
  const userMenuButton = document.querySelector('.admin-user-menu');
  const userDropdown = document.querySelector('.admin-user-dropdown');
  
  if (!userMenuButton || !userDropdown) return;
  
  userMenuButton.classList.remove('open');
  userDropdown.classList.remove('show');
  
  setTimeout(() => {
    if (!userDropdown.classList.contains('show')) {
      userDropdown.style.display = 'none';
    }
  }, 200);
}

// ============================================
// STICKY HEADER SCROLL BEHAVIOR
// ============================================

function initStickyHeaderBehavior() {
  const header = document.querySelector('.admin-header');
  if (!header) return;
  
  let lastScrollTop = 0;
  
  window.addEventListener('scroll', () => {
    const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
    
    if (scrollTop > 50) {
      header.classList.add('scrolled');
    } else {
      header.classList.remove('scrolled');
    }
    
    lastScrollTop = scrollTop;
  });
}

// ============================================
// SEARCH FUNCTIONALITY (BASIC)
// ============================================

function initSearchFunctionality() {
  const searchInput = document.querySelector('.admin-search input');
  if (!searchInput) return;
  
  let searchTimeout;
  
  searchInput.addEventListener('input', (e) => {
    clearTimeout(searchTimeout);
    
    const query = e.target.value.trim();
    
    if (query.length < 2) return;
    
    // Debounce search
    searchTimeout = setTimeout(() => {
      performSearch(query);
    }, 300);
  });
  
  // Clear search on Escape
  searchInput.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      searchInput.value = '';
      searchInput.blur();
    }
  });
}

function performSearch(query) {
  // TODO: Implement actual search with results dropdown
  // This would call an API endpoint and show results
}

// ============================================
// KEYBOARD NAVIGATION
// ============================================

function initKeyboardNavigation() {
  const navHeaders = document.querySelectorAll('.nav-section-header');
  
  navHeaders.forEach((header, index) => {
    header.addEventListener('keydown', (e) => {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        const nextHeader = navHeaders[index + 1];
        if (nextHeader) nextHeader.focus();
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        const prevHeader = navHeaders[index - 1];
        if (prevHeader) prevHeader.focus();
      }
    });
  });
  
  // Dropdown keyboard navigation
  const dropdownItems = document.querySelectorAll('.dropdown-item');
  dropdownItems.forEach((item, index) => {
    item.addEventListener('keydown', (e) => {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        const nextItem = dropdownItems[index + 1];
        if (nextItem) nextItem.focus();
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        const prevItem = dropdownItems[index - 1];
        if (prevItem) prevItem.focus();
      }
    });
  });
}

// ============================================
// ACTIVE STATE MANAGEMENT
// ============================================

function initActiveStateManagement() {
  const currentPath = window.location.pathname;
  const navLinks = document.querySelectorAll('.nav-section-header[href]');
  
  navLinks.forEach(link => {
    if (link.getAttribute('href') === currentPath) {
      link.classList.add('active');
    } else {
      link.classList.remove('active');
    }
  });
}

// ============================================
// INITIALIZATION
// ============================================

document.addEventListener('DOMContentLoaded', () => {
  
  initNavigationExpansion();
  initUserMenuDropdown();
  initStickyHeaderBehavior();
  initSearchFunctionality();
  initKeyboardNavigation();
  initActiveStateManagement();
  
});
