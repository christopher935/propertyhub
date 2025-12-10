/**
 * PropertyHub Charts - Advanced Analytics Visualization
 * Chart.js integration for enterprise real estate analytics
 * Supports all PropertyHub dashboard charts and Friday reports
 */

class PropertyHubCharts {
    constructor() {
        this.charts = new Map();
        this.colors = {
            primary: '#2563eb',
            secondary: '#10b981',
            accent: '#f59e0b',
            danger: '#ef4444',
            warning: '#f97316',
            info: '#06b6d4',
            success: '#22c55e',
            muted: '#6b7280'
        };
        this.init();
    }

    init() {
        this.setupChartDefaults();
        this.loadCharts();
        this.setupRealTimeUpdates();
    }

    setupChartDefaults() {
        Chart.defaults.font.family = "'Inter', -apple-system, BlinkMacSystemFont, sans-serif";
        Chart.defaults.font.size = 12;
        Chart.defaults.color = '#374151';
        Chart.defaults.plugins.legend.position = 'bottom';
        Chart.defaults.plugins.legend.labels.usePointStyle = true;
        Chart.defaults.plugins.legend.labels.padding = 20;
        Chart.defaults.responsive = true;
        Chart.defaults.maintainAspectRatio = false;
    }

    async loadCharts() {
        // Load all dashboard charts
        await this.loadPropertyPerformanceChart();
        await this.loadBookingTrendsChart();
        await this.loadRevenueChart();
        await this.loadMarketDataChart();
        await this.loadLeadSourceChart();
        await this.loadOccupancyChart();
        await this.loadMaintenanceChart();
        await this.loadFridayReportCharts();
    }

    async loadPropertyPerformanceChart() {
        const canvas = document.getElementById('property-performance-chart');
        if (!canvas) return;

        try {
            const response = await fetch('/api/analytics/property-performance', {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
            });
            const data = await response.json();

            const chart = new Chart(canvas.getContext('2d'), {
                type: 'line',
                data: {
                    labels: data.labels,
                    datasets: [{
                        label: 'Average Rent',
                        data: data.avgRent,
                        borderColor: this.colors.primary,
                        backgroundColor: this.colors.primary + '20',
                        fill: true,
                        tension: 0.4
                    }, {
                        label: 'Occupancy Rate %',
                        data: data.occupancyRate,
                        borderColor: this.colors.secondary,
                        backgroundColor: this.colors.secondary + '20',
                        fill: true,
                        tension: 0.4,
                        yAxisID: 'y1'
                    }]
                },
                options: {
                    scales: {
                        y: {
                            type: 'linear',
                            display: true,
                            position: 'left',
                            title: { display: true, text: 'Average Rent ($)' }
                        },
                        y1: {
                            type: 'linear',
                            display: true,
                            position: 'right',
                            title: { display: true, text: 'Occupancy Rate (%)' },
                            grid: { drawOnChartArea: false }
                        }
                    },
                    plugins: {
                        title: { display: true, text: 'Property Performance Trends' }
                    }
                }
            });

            this.charts.set('property-performance', chart);
        } catch (error) {
            console.error('Error loading property performance chart:', error);
        }
    }

    async loadBookingTrendsChart() {
        const canvas = document.getElementById('booking-trends-chart');
        if (!canvas) return;

        try {
            const response = await fetch('/api/analytics/booking-trends', {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
            });
            const data = await response.json();

            const chart = new Chart(canvas.getContext('2d'), {
                type: 'bar',
                data: {
                    labels: data.labels,
                    datasets: [{
                        label: 'New Bookings',
                        data: data.newBookings,
                        backgroundColor: this.colors.success,
                        borderRadius: 4
                    }, {
                        label: 'Cancelled Bookings',
                        data: data.cancelled,
                        backgroundColor: this.colors.danger,
                        borderRadius: 4
                    }, {
                        label: 'Completed Bookings',
                        data: data.completed,
                        backgroundColor: this.colors.primary,
                        borderRadius: 4
                    }]
                },
                options: {
                    plugins: {
                        title: { display: true, text: 'Booking Trends Analysis' }
                    },
                    scales: {
                        y: { beginAtZero: true, title: { display: true, text: 'Number of Bookings' } }
                    }
                }
            });

            this.charts.set('booking-trends', chart);
        } catch (error) {
            console.error('Error loading booking trends chart:', error);
        }
    }

    async loadRevenueChart() {
        const canvas = document.getElementById('revenue-chart');
        if (!canvas) return;

        try {
            const response = await fetch('/api/analytics/revenue', {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
            });
            const data = await response.json();

            const chart = new Chart(canvas.getContext('2d'), {
                type: 'line',
                data: {
                    labels: data.labels,
                    datasets: [{
                        label: 'Total Revenue',
                        data: data.totalRevenue,
                        borderColor: this.colors.success,
                        backgroundColor: this.colors.success + '20',
                        fill: true,
                        tension: 0.4
                    }, {
                        label: 'Net Profit',
                        data: data.netProfit,
                        borderColor: this.colors.primary,
                        backgroundColor: this.colors.primary + '20',
                        fill: true,
                        tension: 0.4
                    }]
                },
                options: {
                    plugins: {
                        title: { display: true, text: 'Revenue & Profit Analysis' },
                        tooltip: {
                            callbacks: {
                                label: function(context) {
                                    return context.dataset.label + ': $' + 
                                           context.parsed.y.toLocaleString();
                                }
                            }
                        }
                    },
                    scales: {
                        y: { 
                            beginAtZero: true,
                            title: { display: true, text: 'Amount ($)' },
                            ticks: {
                                callback: function(value) {
                                    return '$' + value.toLocaleString();
                                }
                            }
                        }
                    }
                }
            });

            this.charts.set('revenue', chart);
        } catch (error) {
            console.error('Error loading revenue chart:', error);
        }
    }

    async loadMarketDataChart() {
        const canvas = document.getElementById('market-data-chart');
        if (!canvas) return;

        try {
            const response = await fetch('/api/analytics/market-data', {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
            });
            const data = await response.json();

            const chart = new Chart(canvas.getContext('2d'), {
                type: 'scatter',
                data: {
                    datasets: [{
                        label: 'Market Properties',
                        data: data.marketProperties,
                        backgroundColor: this.colors.info + '60',
                        borderColor: this.colors.info
                    }, {
                        label: 'Your Properties',
                        data: data.yourProperties,
                        backgroundColor: this.colors.accent + '60',
                        borderColor: this.colors.accent
                    }]
                },
                options: {
                    plugins: {
                        title: { display: true, text: 'Market Positioning Analysis' }
                    },
                    scales: {
                        x: { 
                            title: { display: true, text: 'Square Footage' },
                            beginAtZero: true
                        },
                        y: { 
                            title: { display: true, text: 'Rent per Month ($)' },
                            beginAtZero: true,
                            ticks: {
                                callback: function(value) {
                                    return '$' + value.toLocaleString();
                                }
                            }
                        }
                    }
                }
            });

            this.charts.set('market-data', chart);
        } catch (error) {
            console.error('Error loading market data chart:', error);
        }
    }

    async loadLeadSourceChart() {
        const canvas = document.getElementById('lead-source-chart');
        if (!canvas) return;

        try {
            const response = await fetch('/api/analytics/lead-sources', {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
            });
            const data = await response.json();

            const chart = new Chart(canvas.getContext('2d'), {
                type: 'doughnut',
                data: {
                    labels: data.labels,
                    datasets: [{
                        data: data.values,
                        backgroundColor: [
                            this.colors.primary,
                            this.colors.secondary,
                            this.colors.accent,
                            this.colors.info,
                            this.colors.warning,
                            this.colors.success
                        ],
                        borderWidth: 2,
                        borderColor: '#ffffff'
                    }]
                },
                options: {
                    plugins: {
                        title: { display: true, text: 'Lead Sources Distribution' },
                        tooltip: {
                            callbacks: {
                                label: function(context) {
                                    const label = context.label || '';
                                    const value = context.parsed;
                                    const total = context.dataset.data.reduce((a, b) => a + b, 0);
                                    const percentage = ((value / total) * 100).toFixed(1);
                                    return `${label}: ${value} (${percentage}%)`;
                                }
                            }
                        }
                    }
                }
            });

            this.charts.set('lead-source', chart);
        } catch (error) {
            console.error('Error loading lead source chart:', error);
        }
    }

    async loadOccupancyChart() {
        const canvas = document.getElementById('occupancy-chart');
        if (!canvas) return;

        try {
            const response = await fetch('/api/analytics/occupancy', {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
            });
            const data = await response.json();

            const chart = new Chart(canvas.getContext('2d'), {
                type: 'bar',
                data: {
                    labels: data.labels,
                    datasets: [{
                        label: 'Occupied',
                        data: data.occupied,
                        backgroundColor: this.colors.success,
                        stack: 'occupancy'
                    }, {
                        label: 'Vacant',
                        data: data.vacant,
                        backgroundColor: this.colors.warning,
                        stack: 'occupancy'
                    }, {
                        label: 'Maintenance',
                        data: data.maintenance,
                        backgroundColor: this.colors.danger,
                        stack: 'occupancy'
                    }]
                },
                options: {
                    plugins: {
                        title: { display: true, text: 'Property Occupancy Status' }
                    },
                    scales: {
                        y: { 
                            stacked: true,
                            beginAtZero: true,
                            title: { display: true, text: 'Number of Units' }
                        },
                        x: { stacked: true }
                    }
                }
            });

            this.charts.set('occupancy', chart);
        } catch (error) {
            console.error('Error loading occupancy chart:', error);
        }
    }

    async loadMaintenanceChart() {
        const canvas = document.getElementById('maintenance-chart');
        if (!canvas) return;

        try {
            const response = await fetch('/api/analytics/maintenance', {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
            });
            const data = await response.json();

            const chart = new Chart(canvas.getContext('2d'), {
                type: 'line',
                data: {
                    labels: data.labels,
                    datasets: [{
                        label: 'Maintenance Requests',
                        data: data.requests,
                        borderColor: this.colors.warning,
                        backgroundColor: this.colors.warning + '20',
                        tension: 0.4
                    }, {
                        label: 'Completed',
                        data: data.completed,
                        borderColor: this.colors.success,
                        backgroundColor: this.colors.success + '20',
                        tension: 0.4
                    }]
                },
                options: {
                    plugins: {
                        title: { display: true, text: 'Maintenance Request Trends' }
                    },
                    scales: {
                        y: { 
                            beginAtZero: true,
                            title: { display: true, text: 'Number of Requests' }
                        }
                    }
                }
            });

            this.charts.set('maintenance', chart);
        } catch (error) {
            console.error('Error loading maintenance chart:', error);
        }
    }

    async loadFridayReportCharts() {
        // Weekly Performance Chart
        await this.loadWeeklyPerformanceChart();
        
        // Portfolio Summary Chart
        await this.loadPortfolioSummaryChart();
        
        // Action Items Chart
        await this.loadActionItemsChart();
    }

    async loadWeeklyPerformanceChart() {
        const canvas = document.getElementById('weekly-performance-chart');
        if (!canvas) return;

        try {
            const response = await fetch('/api/analytics/weekly-performance', {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
            });
            const data = await response.json();

            const chart = new Chart(canvas.getContext('2d'), {
                type: 'radar',
                data: {
                    labels: ['Occupancy', 'Revenue', 'Maintenance', 'Lead Conversion', 'Tenant Satisfaction', 'Market Performance'],
                    datasets: [{
                        label: 'This Week',
                        data: data.thisWeek,
                        borderColor: this.colors.primary,
                        backgroundColor: this.colors.primary + '20',
                        pointBackgroundColor: this.colors.primary
                    }, {
                        label: 'Last Week',
                        data: data.lastWeek,
                        borderColor: this.colors.muted,
                        backgroundColor: this.colors.muted + '20',
                        pointBackgroundColor: this.colors.muted
                    }]
                },
                options: {
                    plugins: {
                        title: { display: true, text: 'Weekly Performance Radar' }
                    },
                    scales: {
                        r: {
                            beginAtZero: true,
                            max: 100,
                            ticks: { stepSize: 20 }
                        }
                    }
                }
            });

            this.charts.set('weekly-performance', chart);
        } catch (error) {
            console.error('Error loading weekly performance chart:', error);
        }
    }

    async loadPortfolioSummaryChart() {
        const canvas = document.getElementById('portfolio-summary-chart');
        if (!canvas) return;

        try {
            const response = await fetch('/api/analytics/portfolio-summary', {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
            });
            const data = await response.json();

            const chart = new Chart(canvas.getContext('2d'), {
                type: 'polarArea',
                data: {
                    labels: data.labels,
                    datasets: [{
                        data: data.values,
                        backgroundColor: [
                            this.colors.primary + '80',
                            this.colors.secondary + '80',
                            this.colors.accent + '80',
                            this.colors.info + '80',
                            this.colors.success + '80'
                        ],
                        borderColor: [
                            this.colors.primary,
                            this.colors.secondary,
                            this.colors.accent,
                            this.colors.info,
                            this.colors.success
                        ],
                        borderWidth: 2
                    }]
                },
                options: {
                    plugins: {
                        title: { display: true, text: 'Portfolio Distribution' }
                    }
                }
            });

            this.charts.set('portfolio-summary', chart);
        } catch (error) {
            console.error('Error loading portfolio summary chart:', error);
        }
    }

    async loadActionItemsChart() {
        const canvas = document.getElementById('action-items-chart');
        if (!canvas) return;

        try {
            const response = await fetch('/api/analytics/action-items', {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
            });
            const data = await response.json();

            const chart = new Chart(canvas.getContext('2d'), {
                type: 'bar',
                data: {
                    labels: data.labels,
                    datasets: [{
                        label: 'High Priority',
                        data: data.high,
                        backgroundColor: this.colors.danger
                    }, {
                        label: 'Medium Priority',
                        data: data.medium,
                        backgroundColor: this.colors.warning
                    }, {
                        label: 'Low Priority',
                        data: data.low,
                        backgroundColor: this.colors.info
                    }]
                },
                options: {
                    plugins: {
                        title: { display: true, text: 'Action Items by Priority' }
                    },
                    indexAxis: 'y',
                    scales: {
                        x: { 
                            beginAtZero: true,
                            title: { display: true, text: 'Number of Items' }
                        }
                    }
                }
            });

            this.charts.set('action-items', chart);
        } catch (error) {
            console.error('Error loading action items chart:', error);
        }
    }

    setupRealTimeUpdates() {
        // Update charts every 5 minutes
        setInterval(() => {
            this.updateAllCharts();
        }, 300000);

        // Update critical charts every minute
        setInterval(() => {
            this.updateCriticalCharts();
        }, 60000);
    }

    async updateAllCharts() {
        for (const [name, chart] of this.charts) {
            try {
                await this.updateChart(name, chart);
            } catch (error) {
                console.error(`Error updating ${name} chart:`, error);
            }
        }
    }

    async updateCriticalCharts() {
        const criticalCharts = ['property-performance', 'booking-trends', 'revenue'];
        for (const chartName of criticalCharts) {
            const chart = this.charts.get(chartName);
            if (chart) {
                try {
                    await this.updateChart(chartName, chart);
                } catch (error) {
                    console.error(`Error updating critical chart ${chartName}:`, error);
                }
            }
        }
    }

    async updateChart(name, chart) {
        const endpoint = `/api/analytics/${name.replace('-', '-')}`;
        const response = await fetch(endpoint, {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        });
        
        if (response.ok) {
            const data = await response.json();
            this.updateChartData(chart, data);
        }
    }

    updateChartData(chart, newData) {
        chart.data.labels = newData.labels || chart.data.labels;
        
        chart.data.datasets.forEach((dataset, index) => {
            if (newData.datasets && newData.datasets[index]) {
                dataset.data = newData.datasets[index].data;
            }
        });
        
        chart.update('none');
    }

    destroyChart(name) {
        const chart = this.charts.get(name);
        if (chart) {
            chart.destroy();
            this.charts.delete(name);
        }
    }

    destroyAllCharts() {
        for (const [name, chart] of this.charts) {
            chart.destroy();
        }
        this.charts.clear();
    }

    // Export chart as image
    exportChart(chartName, filename) {
        const chart = this.charts.get(chartName);
        if (chart) {
            const link = document.createElement('a');
            link.download = filename || `${chartName}-chart.png`;
            link.href = chart.toBase64Image();
            link.click();
        }
    }

    // Print all charts for Friday reports
    printFridayReports() {
        const printWindow = window.open('', '_blank');
        printWindow.document.write(`
            <html>
                <head>
                    <title>PropertyHub Friday Report - Charts</title>
                    <style>
                        body { font-family: Arial, sans-serif; }
                        .chart-container { page-break-inside: avoid; margin: 20px 0; }
                        .chart-title { font-size: 18px; font-weight: bold; margin-bottom: 10px; }
                        img { max-width: 100%; height: auto; }
                        @media print { .chart-container { page-break-after: always; } }
                    </style>
                </head>
                <body>
                    <h1>PropertyHub Analytics - Friday Report</h1>
                    <p>Generated on: ${new Date().toLocaleDateString()}</p>
        `);

        for (const [name, chart] of this.charts) {
            printWindow.document.write(`
                <div class="chart-container">
                    <div class="chart-title">${name.replace('-', ' ').replace(/\b\w/g, l => l.toUpperCase())} Chart</div>
                    <img src="${chart.toBase64Image()}" alt="${name} Chart">
                </div>
            `);
        }

        printWindow.document.write('</body></html>');
        printWindow.document.close();
        printWindow.print();
    }
}

// Initialize PropertyHub Charts when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.PropertyHubCharts = new PropertyHubCharts();
});

// Export for module usage
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PropertyHubCharts;
}
