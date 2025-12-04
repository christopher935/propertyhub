const properties = [
  {
    address: "123 Main St",
    status: "leased",
    closeDate: "Nov 18, 2024",
    cdom: 14,
    price: 3200,
    specs: { beds: 4, baths: 3, sqft: 2400 },
    leads: { total: 12, hot: 8, warm: 3, cold: 1 },
    showings: { total: 3, scheduled: 2, completed: 1, feedback: ["Great space", "Price concerns"] },
    applications: { total: 1, pending: 0, approved: 1 },
    market: { avgDaysOnMarket: 21, comparablePrice: { low: 2800, high: 3400 }, demandLevel: "high", areaType: "luxury" },
    engine: { conversionProbability: 95, recommendedAction: "none", performanceLevel: "strong", insights: ["Strong lead quality", "Below market average time", "High HOT lead engagement"] },
    conversion: { daysToLease: 14, leadSource: "Zillow", leadScore: 87, leadName: "Sarah Johnson", viewCount: 12, showingCount: 2, vsPortfolioAvg: -7 }
  },
  {
    address: "456 Oak Ave",
    status: "active",
    cdom: 14,
    price: 3200,
    specs: { beds: 4, baths: 3, sqft: 2400 },
    leads: { total: 12, hot: 8, warm: 3, cold: 1 },
    showings: { total: 3, scheduled: 3, completed: 0, feedback: [] },
    applications: { total: 0, pending: 0, approved: 0 },
    market: { avgDaysOnMarket: 21, comparablePrice: { low: 2800, high: 3400 }, demandLevel: "high", areaType: "luxury" },
    engine: { conversionProbability: 65, recommendedAction: "none", performanceLevel: "strong", insights: ["Strong lead quality", "Below market average time", "High HOT lead engagement"] }
  },
  {
    address: "789 Pine St",
    status: "active",
    cdom: 28,
    price: 2800,
    specs: { beds: 3, baths: 2, sqft: 1800 },
    leads: { total: 4, hot: 1, warm: 2, cold: 1 },
    showings: { total: 1, scheduled: 1, completed: 1, feedback: ["Price is a bit high"] },
    applications: { total: 0, pending: 0, approved: 0 },
    market: { avgDaysOnMarket: 20, comparablePrice: { low: 2400, high: 2600 }, demandLevel: "medium", areaType: "mid-range" },
    engine: { conversionProbability: 20, recommendedAction: "price_adjustment", performanceLevel: "needs_attention", insights: ["Low traffic", "Price concerns from feedback", "Comps suggest lower price"] }
  },
  {
    address: "234 Elm Dr",
    status: "active",
    cdom: 21,
    price: 2200,
    specs: { beds: 2, baths: 2, sqft: 1200 },
    leads: { total: 8, hot: 3, warm: 4, cold: 1 },
    showings: { total: 3, scheduled: 3, completed: 0, feedback: [] },
    applications: { total: 0, pending: 0, approved: 0 },
    market: { avgDaysOnMarket: 25, comparablePrice: { low: 2100, high: 2300 }, demandLevel: "medium", areaType: "mid-range" },
    engine: { conversionProbability: 45, recommendedAction: "none", performanceLevel: "steady", insights: ["Steady traffic", "Good lead quality for price point"] }
  }
];
