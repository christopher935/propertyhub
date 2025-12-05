// Admin Sidebar Navigation
document.addEventListener('DOMContentLoaded', function() {
    // Handle nav section toggle
    const navSections = document.querySelectorAll('.nav-section');
    
    navSections.forEach(section => {
        const header = section.querySelector('.nav-section-header');
        const items = section.querySelector('.nav-section-items');
        const toggle = section.querySelector('.nav-section-toggle');
        
        if (header && items) {
            header.addEventListener('click', function(e) {
                e.preventDefault();
                
                // Toggle active state
                const isActive = section.classList.contains('active');
                
                if (isActive) {
                    section.classList.remove('active');
                    items.style.maxHeight = '0';
                    if (toggle) toggle.textContent = '▶';
                } else {
                    section.classList.add('active');
                    items.style.maxHeight = items.scrollHeight + 'px';
                    if (toggle) toggle.textContent = '▼';
                }
            });
        }
    });
    
    // Set active section based on current page
    const currentPath = window.location.pathname;
    const navItems = document.querySelectorAll('.nav-item');
    
    navItems.forEach(item => {
        const link = item.querySelector('a');
        if (link && link.getAttribute('href') === currentPath) {
            item.classList.add('active');
            
            // Expand parent section
            const parentSection = item.closest('.nav-section');
            if (parentSection) {
                parentSection.classList.add('active');
                const items = parentSection.querySelector('.nav-section-items');
                const toggle = parentSection.querySelector('.nav-section-toggle');
                if (items) items.style.maxHeight = items.scrollHeight + 'px';
                if (toggle) toggle.textContent = '▼';
            }
        }
    });
});

