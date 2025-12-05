// Professional Chart Animations - PropertyHub
// Minimal JavaScript for enhanced effects with CSS3 focus

document.addEventListener('DOMContentLoaded', function() {
    
    // Animated Counter Function
    function animateCounter(element, target, duration = 2000) {
        const start = 0;
        const startTime = performance.now();
        
        function update(currentTime) {
            const elapsed = currentTime - startTime;
            const progress = Math.min(elapsed / duration, 1);
            
            // Easing function for smooth animation
            const ease = progress < 0.5 
                ? 4 * progress * progress * progress 
                : (progress - 1) * (2 * progress - 2) * (2 * progress - 2) + 1;
            
            const current = Math.floor(start + (target - start) * ease);
            element.textContent = formatNumber(current);
            
            if (progress < 1) {
                requestAnimationFrame(update);
            } else {
                element.textContent = formatNumber(target);
                element.classList.add('counter-complete');
            }
        }
        
        requestAnimationFrame(update);
    }
    
    // Number Formatting
    function formatNumber(num) {
        if (num >= 1000000) {
            return (num / 1000000).toFixed(1) + 'M';
        } else if (num >= 1000) {
            return (num / 1000).toFixed(1) + 'K';
        }
        return num.toString();
    }
    
    // Initialize Animated Counters
    function initCounters() {
        const counters = document.querySelectorAll('.animated-counter');
        
        const observer = new IntersectionObserver((entries) => {
            entries.forEach(entry => {
                if (entry.isIntersecting && !entry.target.classList.contains('animated')) {
                    entry.target.classList.add('animated');
                    entry.target.classList.add('counter-animate');
                    
                    const text = entry.target.textContent;
                    const number = parseInt(text.replace(/[^\d]/g, ''));
                    
                    if (!isNaN(number) && number > 0) {
                        animateCounter(entry.target, number);
                    }
                }
            });
        }, { threshold: 0.5 });
        
        counters.forEach(counter => observer.observe(counter));
    }
    
    // Generate SVG Charts
    function generateDonutChart(container, data) {
        const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
        svg.setAttribute('class', 'donut-svg');
        svg.setAttribute('viewBox', '0 0 200 200');
        
        const total = data.reduce((sum, item) => sum + item.value, 0);
        let startAngle = 0;
        const colors = ['#C2A875', '#2B4B73', '#10B981', '#F59E0B', '#8B5CF6'];
        
        // Create gradient definitions
        const defs = document.createElementNS('http://www.w3.org/2000/svg', 'defs');
        data.forEach((item, index) => {
            const gradient = document.createElementNS('http://www.w3.org/2000/svg', 'linearGradient');
            gradient.setAttribute('id', `gradient${index}`);
            gradient.setAttribute('x1', '0%');
            gradient.setAttribute('y1', '0%');
            gradient.setAttribute('x2', '100%');
            gradient.setAttribute('y2', '100%');
            
            const stop1 = document.createElementNS('http://www.w3.org/2000/svg', 'stop');
            stop1.setAttribute('offset', '0%');
            stop1.setAttribute('stop-color', colors[index % colors.length]);
            
            const stop2 = document.createElementNS('http://www.w3.org/2000/svg', 'stop');
            stop2.setAttribute('offset', '100%');
            stop2.setAttribute('stop-color', adjustBrightness(colors[index % colors.length], -20));
            
            gradient.appendChild(stop1);
            gradient.appendChild(stop2);
            defs.appendChild(gradient);
        });
        svg.appendChild(defs);
        
        data.forEach((item, index) => {
            const angle = (item.value / total) * 360;
            const path = createDonutPath(100, 100, 70, 40, startAngle, startAngle + angle);
            
            const pathElement = document.createElementNS('http://www.w3.org/2000/svg', 'path');
            pathElement.setAttribute('d', path);
            pathElement.setAttribute('fill', `url(#gradient${index})`);
            pathElement.setAttribute('class', 'donut-segment');
            pathElement.style.animationDelay = `${index * 0.2}s`;
            
            svg.appendChild(pathElement);
            startAngle += angle;
        });
        
        container.appendChild(svg);
    }
    
    // Helper function to create donut path
    function createDonutPath(x, y, outerRadius, innerRadius, startAngle, endAngle) {
        const start = polarToCartesian(x, y, outerRadius, endAngle);
        const end = polarToCartesian(x, y, outerRadius, startAngle);
        const largeArcFlag = endAngle - startAngle <= 180 ? "0" : "1";
        
        const outerArc = [
            "M", start.x, start.y, 
            "A", outerRadius, outerRadius, 0, largeArcFlag, 0, end.x, end.y
        ].join(" ");
        
        const innerStart = polarToCartesian(x, y, innerRadius, endAngle);
        const innerEnd = polarToCartesian(x, y, innerRadius, startAngle);
        
        const innerArc = [
            "L", innerEnd.x, innerEnd.y,
            "A", innerRadius, innerRadius, 0, largeArcFlag, 0, innerStart.x, innerStart.y
        ].join(" ");
        
        return [outerArc, innerArc, "Z"].join(" ");
    }
    
    function polarToCartesian(centerX, centerY, radius, angleInDegrees) {
        const angleInRadians = (angleInDegrees - 90) * Math.PI / 180.0;
        return {
            x: centerX + (radius * Math.cos(angleInRadians)),
            y: centerY + (radius * Math.sin(angleInRadians))
        };
    }
    
    function adjustBrightness(hex, percent) {
        const num = parseInt(hex.replace("#", ""), 16);
        const amt = Math.round(2.55 * percent);
        const R = (num >> 16) + amt;
        const G = (num >> 8 & 0x00FF) + amt;
        const B = (num & 0x0000FF) + amt;
        return "#" + (0x1000000 + (R < 255 ? R < 1 ? 0 : R : 255) * 0x10000 +
            (G < 255 ? G < 1 ? 0 : G : 255) * 0x100 +
            (B < 255 ? B < 1 ? 0 : B : 255)).toString(16).slice(1);
    }
    
    // Initialize Charts
    function initCharts() {
        // Behavioral Intelligence Chart
        const behavioralChart = document.querySelector('.behavioral-chart');
        if (behavioralChart) {
            const barsHTML = `
                <div class="chart-bars">
                    <div class="chart-bar" data-value="65%"></div>
                    <div class="chart-bar" data-value="82%"></div>
                    <div class="chart-bar" data-value="73%"></div>
                    <div class="chart-bar" data-value="87%"></div>
                    <div class="chart-bar" data-value="91%"></div>
                    <div class="chart-bar" data-value="78%"></div>
                </div>
                <div class="chart-labels">
                    <span>Jan</span><span>Feb</span><span>Mar</span><span>Apr</span><span>May</span><span>Jun</span>
                </div>
            `;
            behavioralChart.innerHTML = barsHTML;
        }
        
        // Market Segments Donut Chart
        const donutContainer = document.querySelector('.donut-chart');
        if (donutContainer) {
            const marketData = [
                { label: 'Sugar Land', value: 32 },
                { label: 'Memorial', value: 24 },
                { label: 'Heights', value: 18 },
                { label: 'Katy', value: 15 },
                { label: 'River Oaks', value: 11 }
            ];
            generateDonutChart(donutContainer, marketData);
            
            // Add center text
            const centerDiv = document.createElement('div');
            centerDiv.className = 'donut-center';
            centerDiv.innerHTML = `
                <div class="donut-value animated-counter">247</div>
                <div class="donut-label">Total Leads</div>
            `;
            donutContainer.appendChild(centerDiv);
        }
        
        // Conversion Funnel Chart
        const funnelChart = document.querySelector('.funnel-chart');
        if (funnelChart) {
            const funnelHTML = `
                <div class="funnel-stage">
                    <div class="funnel-bar">
                        <span>Website Visitors</span>
                        <span class="funnel-percentage">100%</span>
                    </div>
                </div>
                <div class="funnel-stage">
                    <div class="funnel-bar">
                        <span>Generated Leads</span>
                        <span class="funnel-percentage">11.2%</span>
                    </div>
                </div>
                <div class="funnel-stage">
                    <div class="funnel-bar">
                        <span>Tour Bookings</span>
                        <span class="funnel-percentage">34.6%</span>
                    </div>
                </div>
                <div class="funnel-stage">
                    <div class="funnel-bar">
                        <span>Leases Signed</span>
                        <span class="funnel-percentage">22.5%</span>
                    </div>
                </div>
            `;
            funnelChart.innerHTML = funnelHTML;
        }
    }
    
    // Add Hover Effects to KPI Cards
    function initKPIEffects() {
        const kpiCards = document.querySelectorAll('.executive-kpi, .kpi-card, .stat-card');
        
        kpiCards.forEach(card => {
            card.addEventListener('mouseenter', function() {
                this.style.transform = 'translateY(-2px)';
                this.style.boxShadow = '0 8px 25px rgba(0, 0, 0, 0.15)';
            });
            
            card.addEventListener('mouseleave', function() {
                this.style.transform = 'translateY(0)';
                this.style.boxShadow = '0 2px 4px rgba(0, 0, 0, 0.05)';
            });
        });
    }
    
    // Initialize all animations
    initCounters();
    initCharts();
    initKPIEffects();
    
    // Add subtle parallax effect to charts
    window.addEventListener('scroll', function() {
        const scrolled = window.pageYOffset;
        const charts = document.querySelectorAll('.chart-container');
        
        charts.forEach((chart, index) => {
            const speed = 0.5 + (index * 0.1);
            chart.style.transform = `translateY(${scrolled * speed * 0.1}px)`;
        });
    });
});