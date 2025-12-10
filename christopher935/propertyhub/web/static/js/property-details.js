// Property Detail Page Enhancements

document.addEventListener('DOMContentLoaded', function() {
    // Smooth scroll active state tracking
    const sections = document.querySelectorAll('section[id]');
    const navLinks = document.querySelectorAll('.property-nav .nav-link');
    
    if (sections.length > 0 && navLinks.length > 0) {
        const observerOptions = {
            root: null,
            rootMargin: '-20% 0px -60% 0px',
            threshold: 0
        };
        
        const observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    const id = entry.target.getAttribute('id');
                    navLinks.forEach(link => {
                        link.classList.remove('active');
                        if (link.getAttribute('href') === `#${id}`) {
                            link.classList.add('active');
                        }
                    });
                }
            });
        }, observerOptions);
        
        sections.forEach(section => observer.observe(section));
    }
    
    // Image gallery lightbox (if images are present)
    const galleryImages = document.querySelectorAll('.property-gallery img');
    galleryImages.forEach(img => {
        img.style.cursor = 'pointer';
        img.addEventListener('click', function() {
            // Create simple lightbox overlay
            const overlay = document.createElement('div');
            overlay.style.cssText = `
                position: fixed;
                inset: 0;
                background: rgba(0, 0, 0, 0.9);
                z-index: 1000;
                display: flex;
                align-items: center;
                justify-content: center;
                cursor: pointer;
            `;
            
            const largeImg = document.createElement('img');
            largeImg.src = this.src;
            largeImg.style.cssText = `
                max-width: 90%;
                max-height: 90%;
                object-fit: contain;
                border-radius: 12px;
            `;
            
            overlay.appendChild(largeImg);
            document.body.appendChild(overlay);
            document.body.style.overflow = 'hidden';
            
            overlay.addEventListener('click', function() {
                document.body.removeChild(overlay);
                document.body.style.overflow = '';
            });
        });
    });
});
