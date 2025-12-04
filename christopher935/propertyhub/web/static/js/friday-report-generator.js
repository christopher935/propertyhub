document.addEventListener('DOMContentLoaded', () => {
    const generateReportBtn = document.getElementById('generate-report-btn');
    const copyReportBtn = document.getElementById('copy-report-btn');
    const reportPreview = document.getElementById('report-preview');

    const generateCommentary = (property) => {
        const { performanceLevel } = property.engine;

        if (property.status === 'leased') {
            const { conversion } = property;
            return `Leased in just ${conversion.daysToLease} days to ${conversion.leadName} after ${conversion.showingCount} showings. Fast conversion - ${Math.abs(conversion.vsPortfolioAvg)} days faster than our portfolio average. Strong initial traction with ${conversion.viewCount} views from multiple qualified leads helped accelerate the process. Market timing was optimal with high demand for ${property.specs.beds}-bed properties in this area.`;
        }

        switch (performanceLevel) {
            case 'strong':
                return `Outperforming with strong lead quality - ${property.leads.hot} highly qualified leads in just ${property.cdom} days, well below market average of ${property.market.avgDaysOnMarket} days. ${property.showings.scheduled} showings scheduled this week with serious prospects. Based on lead engagement patterns and showing feedback, expecting application within 7 days. Pricing is competitive at $${property.price.toLocaleString()} for this ${property.specs.beds}-bed luxury unit.`;
            case 'needs_attention':
                return `Traffic is running 40% below similar ${property.specs.beds}-bed properties in the area. Only ${property.leads.total} views in ${property.cdom} days suggests pricing or presentation issue. Comparable properties leased at $${property.market.comparablePrice.low.toLocaleString()}-$${property.market.comparablePrice.high.toLocaleString()} vs our $${property.price.toLocaleString()} asking. One showing last week but feedback indicated price concerns. Consider adjustment to $${(property.market.comparablePrice.low + property.market.comparablePrice.high) / 2} to accelerate conversion and align with market comps.`;
            case 'steady':
                return `Steady progress with ${property.leads.total} leads and ${property.showings.scheduled} showings scheduled in ${property.cdom} days. This is on track with the market average of ${property.market.avgDaysOnMarket} days. Lead quality is good for this price point, with ${property.leads.hot} serious prospects engaged. Expecting applications to come in within the next 7-10 days.`;
            default:
                return '';
        }
    };

    const generateReport = () => {
        let reportHTML = '';

        // Section 1: Closings This Week
        const closings = properties.filter(p => p.status === 'leased');
        if (closings.length > 0) {
            reportHTML += '<h3>CLOSINGS THIS WEEK</h3>';
            closings.forEach(property => {
                reportHTML += `
                    <div class="report-item">
                        <p class="property-address">${property.address} - Closes ${property.closeDate}</p>
                        <p class="commentary">${generateCommentary(property)}</p>
                    </div>
                `;
            });
        }

        // Section 2: Active Listings
        const activeListings = properties.filter(p => p.status === 'active');
        if (activeListings.length > 0) {
            reportHTML += '<h3>ACTIVE LISTINGS</h3>';
            activeListings.forEach(property => {
                reportHTML += `
                    <div class="report-item">
                        <p class="property-address">${property.address}</p>
                        <p class="metrics">CDOM: ${property.cdom} days | Leads: ${property.leads.total} | Showings: ${property.showings.total} | Applications: ${property.applications.total}</p>
                        <p class="commentary">${generateCommentary(property)}</p>
                    </div>
                `;
            });
        }

        reportPreview.innerHTML = reportHTML;
    };

    const copyReportToClipboard = () => {
        const reportText = reportPreview.innerText;
        navigator.clipboard.writeText(reportText).then(() => {
            alert('Report copied to clipboard!');
        }, () => {
            alert('Failed to copy report.');
        });
    };

    generateReportBtn.addEventListener('click', generateReport);
    copyReportBtn.addEventListener('click', copyReportToClipboard);

    // Initial report generation
    generateReport();
});
