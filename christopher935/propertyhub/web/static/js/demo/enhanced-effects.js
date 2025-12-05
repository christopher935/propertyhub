// PropertyHub Enhanced Effects JavaScript
// Functional charts, number counters, and proper animations

document.addEventListener('DOMContentLoaded', function() {
    
    // ============================================
    // INTERSECTION OBSERVER FOR LAZY LOADING
    // ============================================
    
    if ('IntersectionObserver' in window) {
        const lazyObserver = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting) {
                    entry.target.classList.add('loaded');
                    
                    // Trigger number counting animation
                    const counters = entry.target.querySelectorAll('.counter-number');
                    counters.forEach(counter => animateNumber(counter));
                    
                    lazyObserver.unobserve(entry.target);
                }
            });
        }, {
            rootMargin: '50px 0px',
            threshold: 0.1
        });
        
        // Observe all lazy load elements
        document.querySelectorAll('.lazy-load, .lazy-load-enhanced').forEach(element => {
            lazyObserver.observe(element);
        });
    }
    
    // ============================================
    // ANIMATED NUMBER COUNTERS
    // ============================================
    
    function animateNumber(element) {
        const finalValue = element.textContent.trim();
        const numericValue = parseFloat(finalValue.replace(/[^\d.-]/g, ''));
        
        if (isNaN(numericValue)) return;
        
        const duration = 2000; // 2 seconds
        const startTime = performance.now();
        const suffix = finalValue.replace(numericValue.toString(), '');
        
        function updateNumber(currentTime) {
            const elapsed = currentTime - startTime;
            const progress = Math.min(elapsed / duration, 1);
            
            // Ease-out animation
            const easeProgress = 1 - Math.pow(1 - progress, 3);
            const currentValue = Math.round(numericValue * easeProgress);
            
            element.textContent = currentValue + suffix;
            
            if (progress < 1) {
                requestAnimationFrame(updateNumber);
            } else {
                element.textContent = finalValue; // Ensure final value is exact
            }
        }
        
        requestAnimationFrame(updateNumber);
    }
    
    // ============================================
    // FUNCTIONAL CHART CREATION
    // ============================================
    
    function createBarChart(container, data) {
        const barChart = document.createElement('div');
        barChart.className = 'bar-chart';
        
        data.forEach((value, index) => {
            const bar = document.createElement('div');
            bar.className = 'bar';
            bar.setAttribute('data-value', value);
            bar.style.setProperty('--index', index);
            bar.style.height = value + '%';
            barChart.appendChild(bar);
        });
        
        container.innerHTML = '';
        container.appendChild(barChart);
        container.classList.add('chart-functional');
    }
    
    function createDonutChart(container, percentage, label) {
        const donutHTML = `
            <div class="donut-chart">
                <svg viewBox="0 0 42 42">
                    <defs>
                        <linearGradient id="gradient-${Math.random().toString(36).substr(2, 9)}" x1="0%" y1="0%" x2="100%" y2="0%">
                            <stop offset="0%" style="stop-color:var(--navy-primary);stop-opacity:1" />
                            <stop offset="100%" style="stop-color:var(--gold-primary);stop-opacity:1" />
                        </linearGradient>
                    </defs>
                    <circle cx="21" cy="21" r="15.915" class="donut-segment" 
                            stroke="var(--gray-200)" stroke-width="3" 
                            fill="none" stroke-dasharray="100 0"></circle>
                    <circle cx="21" cy="21" r="15.915" class="donut-segment navy" 
                            stroke="url(#gradient-${Math.random().toString(36).substr(2, 9)})" 
                            stroke-width="3" fill="none" 
                            stroke-dasharray="${percentage} ${100-percentage}" 
                            stroke-dashoffset="0"></circle>
                </svg>
                <div class="donut-center">
                    <span class="donut-value counter-number">${percentage}%</span>
                    <span class="donut-label">${label}</span>
                </div>
            </div>
        `;
        
        container.innerHTML = donutHTML;
    }
    
    function createLineChart(container, data) {
        const width = 300;
        const height = 100;
        const maxValue = Math.max(...data);
        const points = data.map((value, index) => {
            const x = (index / (data.length - 1)) * width;
            const y = height - (value / maxValue) * height;
            return `${x},${y}`;
        }).join(' ');
        
        const areaPoints = `0,${height} ${points} ${width},${height}`;
        
        const lineHTML = `
            <svg viewBox="0 0 ${width} ${height}">
                <defs>
                    <linearGradient id="areaGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                        <stop offset="0%" style="stop-color:var(--navy-primary);stop-opacity:0.3" />
                        <stop offset="100%" style="stop-color:var(--navy-primary);stop-opacity:0" />
                    </linearGradient>
                </defs>
                <polygon points="${areaPoints}" class="chart-area" fill="url(#areaGradient)"/>
                <polyline points="${points}" class="chart-line" />
            </svg>
        `;
        
        container.innerHTML = lineHTML;
        container.classList.add('line-chart');
    }
    
    // ============================================
    // REPLACE CHART PLACEHOLDERS WITH FUNCTIONAL CHARTS
    // ============================================
    
    document.querySelectorAll('.chart-placeholder').forEach((placeholder, index) => {
        const parentCard = placeholder.closest('.kpi-card, .chart-card, .stats-card');
        
        if (parentCard) {
            // Determine chart type based on context
            const cardText = parentCard.textContent.toLowerCase();
            
            if (cardText.includes('revenue') || cardText.includes('leads') || cardText.includes('properties')) {
                // Create bar chart with sample data
                createBarChart(placeholder, [85, 92, 78, 95, 88, 91]);
            } else if (cardText.includes('completion') || cardText.includes('rate') || cardText.includes('score')) {
                // Create donut chart
                const percentage = 85 + (index * 3); // Vary the percentage
                createDonutChart(placeholder, percentage, 'Complete');
            } else {
                // Create line chart with trending data
                createLineChart(placeholder, [65, 72, 68, 78, 85, 92, 88, 95, 91]);
            }
        }
    });
    
    // ============================================
    // ENHANCED KPI CARD INTERACTIONS
    // ============================================
    
    document.querySelectorAll('.kpi-card, .stats-card').forEach(card => {
        card.classList.add('kpi-card-enhanced', 'lazy-load-enhanced');
        
        // Add click animation
        card.addEventListener('click', function() {
            this.style.transform = 'scale(0.98)';
            setTimeout(() => {
                this.style.transform = '';
            }, 150);
        });
    });
    
    // ============================================
    // CONTEXTUAL ICON THEMING
    // ============================================
    
    function applyIconTheming() {
        document.querySelectorAll('.icon').forEach(icon => {
            const bgColor = getComputedStyle(icon.closest('.card, .kpi-card, .nav-item') || icon.parentElement).backgroundColor;
            const rgb = bgColor.match(/\d+/g);
            
            if (rgb) {
                const brightness = (parseInt(rgb[0]) * 299 + parseInt(rgb[1]) * 587 + parseInt(rgb[2]) * 114) / 1000;
                
                if (brightness < 128) {
                    icon.classList.add('icon-context-light');
                } else {
                    icon.classList.add('icon-context-dark');
                }
            }
        });
    }
    
    // Apply icon theming after a short delay to ensure styles are loaded
    setTimeout(applyIconTheming, 100);
    
    // ============================================
    // SMOOTH SCROLLING AND STAGGERED ANIMATIONS
    // ============================================
    
    const observerOptions = {
        rootMargin: '0px 0px -100px 0px',
        threshold: 0.1
    };
    
    const staggerObserver = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                const children = entry.target.querySelectorAll('.kpi-card, .chart-card, .property-card, .card');
                children.forEach((child, index) => {
                    setTimeout(() => {
                        child.classList.add('loaded');
                    }, index * 100);
                });
                
                staggerObserver.unobserve(entry.target);
            }
        });
    }, observerOptions);
    
    // Observe sections for staggered animations
    document.querySelectorAll('section, .dashboard-grid, .properties-grid').forEach(section => {
        staggerObserver.observe(section);
    });
    
    // ============================================
    // PERFORMANCE MONITORING
    // ============================================
    
    // Only run animations if user hasn't requested reduced motion
    if (!window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
        console.log('✅ PropertyHub Enhanced Effects Loaded');
    } else {
        // Disable animations for users who prefer reduced motion
        document.documentElement.style.setProperty('--animation-duration', '0.01s');
    }
});

// ============================================
// UTILITY FUNCTIONS
// ============================================

// Debounced resize handler for responsive charts
let resizeTimeout;
window.addEventListener('resize', function() {
    clearTimeout(resizeTimeout);
    resizeTimeout = setTimeout(() => {
        // Recalculate chart dimensions if needed
        document.querySelectorAll('.line-chart svg').forEach(svg => {
            svg.style.width = '100%';
            svg.style.height = '100%';
        });
    }, 150);
});