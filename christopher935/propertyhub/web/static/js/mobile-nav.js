document.addEventListener('DOMContentLoaded', function() {
    const navToggle = document.querySelector('.nav-toggle');
    const navMenu = document.querySelector('.nav-menu');
    
    if (!navToggle || !navMenu) return;
    
    let backdrop = document.querySelector('.nav-backdrop');
    if (!backdrop) {
        backdrop = document.createElement('div');
        backdrop.className = 'nav-backdrop';
        document.body.appendChild(backdrop);
    }
    
    function openMenu() {
        navMenu.classList.add('open');
        backdrop.classList.add('open');
        navToggle.setAttribute('aria-expanded', 'true');
        document.body.style.overflow = 'hidden';
    }
    
    function closeMenu() {
        navMenu.classList.remove('open');
        backdrop.classList.remove('open');
        navToggle.setAttribute('aria-expanded', 'false');
        document.body.style.overflow = '';
    }
    
    function toggleMenu() {
        if (navMenu.classList.contains('open')) {
            closeMenu();
        } else {
            openMenu();
        }
    }
    
    navToggle.addEventListener('click', toggleMenu);
    
    backdrop.addEventListener('click', closeMenu);
    
    const navLinks = navMenu.querySelectorAll('.nav-link');
    navLinks.forEach(link => {
        link.addEventListener('click', () => {
            if (window.innerWidth <= 768) {
                closeMenu();
            }
        });
    });
    
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape' && navMenu.classList.contains('open')) {
            closeMenu();
        }
    });
    
    window.addEventListener('resize', () => {
        if (window.innerWidth > 768 && navMenu.classList.contains('open')) {
            closeMenu();
        }
    });
});
