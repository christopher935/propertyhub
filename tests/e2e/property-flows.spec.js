/**
 * End-to-End Tests for Property Flows
 * Test Framework: Playwright
 * 
 * Run with: npm test
 * or: npx playwright test
 */

const { test, expect } = require('@playwright/test');

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

test.describe('Property Discovery Flows', () => {
  
  test('Homepage loads and displays featured properties', async ({ page }) => {
    await page.goto(BASE_URL);
    
    // Check page loaded
    await expect(page).toHaveTitle(/Landlords of Texas/);
    
    // Check hero section
    await expect(page.locator('h1')).toContainText('Schedule Your Tour');
    
    // Check featured properties section
    const featuredSection = page.locator('[x-data*="homepageProperties"]');
    await expect(featuredSection).toBeVisible();
    
    // Wait for properties to load (max 5 seconds)
    await page.waitForSelector('.property-card', { timeout: 5000 }).catch(() => {
      console.warn('⚠️ No property cards found on homepage');
    });
    
    // Count property cards
    const propertyCards = page.locator('.property-card');
    const count = await propertyCards.count();
    
    // Verify at least 1 property displays
    expect(count).toBeGreaterThan(0);
    console.log(`✅ Homepage displays ${count} properties`);
    
    // Verify property card has required elements
    if (count > 0) {
      const firstCard = propertyCards.first();
      await expect(firstCard.locator('.property-price')).toBeVisible();
      await expect(firstCard.locator('.property-address')).toBeVisible();
      await expect(firstCard.locator('.property-details')).toContainText(/bed/i);
      await expect(firstCard.locator('.property-details')).toContainText(/bath/i);
    }
  });

  test('Properties browse page displays property grid', async ({ page }) => {
    await page.goto(`${BASE_URL}/properties`);
    
    // Check page loaded
    await expect(page.locator('h1')).toContainText(/Find Your Perfect Home|Browse Properties/i);
    
    // Wait for properties to load
    await page.waitForSelector('#propertyGrid', { timeout: 5000 });
    
    // Check for property cards or empty state
    const propertyCards = page.locator('.property-card');
    const emptyState = page.locator('.empty-state');
    
    const hasProperties = await propertyCards.count() > 0;
    const showsEmpty = await emptyState.isVisible();
    
    // Should show either properties or empty state, not both
    expect(hasProperties || showsEmpty).toBeTruthy();
    
    if (hasProperties) {
      console.log(`✅ Properties page shows ${await propertyCards.count()} properties`);
    } else {
      console.log('⚠️ Properties page shows empty state');
    }
  });

  test('Property detail page displays full information', async ({ page }) => {
    // First, get a property ID from the API
    const response = await page.request.get(`${BASE_URL}/api/v1/properties?limit=1`);
    const data = await response.json();
    
    if (!data.data || !data.data.properties || data.data.properties.length === 0) {
      test.skip('No properties available to test');
      return;
    }
    
    const propertyId = data.data.properties[0].id;
    console.log(`Testing property detail for ID: ${propertyId}`);
    
    // Visit property detail page
    await page.goto(`${BASE_URL}/property/${propertyId}`);
    
    // Check NOT showing [ENCRYPTED]
    const h1 = page.locator('h1');
    const h1Text = await h1.textContent();
    expect(h1Text).not.toContain('[ENCRYPTED]');
    console.log(`✅ Property address displays correctly: ${h1Text}`);
    
    // Check required elements
    await expect(page.locator('.property-price')).toBeVisible();
    await expect(page.locator('.property-details')).toContainText(/bed/i);
    await expect(page.locator('.property-details')).toContainText(/bath/i);
    
    // Check bathroom formatting (should be X.X not X.XXXXXX)
    const detailsText = await page.locator('.property-details').textContent();
    const bathroomMatch = detailsText.match(/(\d+\.\d+)\s*bath/i);
    if (bathroomMatch) {
      const bathrooms = bathroomMatch[1];
      // Should have 1 decimal place max
      expect(bathrooms.split('.')[1].length).toBeLessThanOrEqual(1);
      console.log(`✅ Bathrooms formatted correctly: ${bathrooms}`);
    }
    
    // Check Schedule Tour button
    const scheduleButton = page.locator('a[href*="book-showing"]');
    await expect(scheduleButton).toBeVisible();
    await expect(scheduleButton).toContainText(/Schedule|Book/i);
  });

  test('Property search and filters work', async ({ page }) => {
    await page.goto(`${BASE_URL}/properties`);
    
    // Test price filter
    await page.click('[data-filter="price-low"]');
    await page.waitForTimeout(1000); // Wait for filter to apply
    
    // Verify URL updated
    expect(page.url()).toContain('max_price');
    
    // Test bedroom filter
    await page.selectOption('#bedroomFilter', '2');
    await page.waitForTimeout(1000);
    
    // Verify URL updated
    expect(page.url()).toContain('bedrooms');
    
    // Clear filters
    await page.click('button:has-text("Clear All")');
    await page.waitForTimeout(1000);
    
    // Verify filters cleared from URL
    const urlAfterClear = page.url();
    expect(urlAfterClear).not.toContain('max_price');
    expect(urlAfterClear).not.toContain('bedrooms');
  });

});

test.describe('Booking Flow', () => {
  
  test('Booking page loads with proper form elements', async ({ page }) => {
    await page.goto(`${BASE_URL}/book-showing`);
    
    // Check 4-step wizard displays
    await expect(page.locator('text=Pre-Qualification')).toBeVisible();
    await expect(page.locator('text=Date & Time')).toBeVisible();
    await expect(page.locator('text=Contact Info')).toBeVisible();
    await expect(page.locator('text=Confirmation')).toBeVisible();
    
    // Check rental requirements section
    await expect(page.locator('text=Income Requirement')).toBeVisible();
    await expect(page.locator('text=Criminal Background')).toBeVisible();
    
    // Check "Next Step" button
    await expect(page.locator('button:has-text("Next Step")')).toBeVisible();
  });

  test('Booking form validation works', async ({ page }) => {
    await page.goto(`${BASE_URL}/book-showing`);
    
    // Try to proceed without checking requirements
    await page.click('button:has-text("Next Step")');
    
    // Should stay on step 1 (validation should prevent progress)
    await expect(page.locator('text=Pre-Qualification')).toBeVisible();
    
    // Check all requirements
    await page.check('[name="income"]');
    await page.check('[name="felony"]');
    await page.check('[name="eviction"]');
    await page.check('[name="credit"]');
    
    // Now should be able to proceed
    await page.click('button:has-text("Next Step")');
    
    // Should advance to step 2
    await expect(page.locator('.calendar')).toBeVisible({ timeout: 2000 });
  });

  test('Date and time selection works', async ({ page }) => {
    await page.goto(`${BASE_URL}/book-showing`);
    
    // Complete step 1
    await page.check('[name="income"]');
    await page.check('[name="felony"]');
    await page.check('[name="eviction"]');
    await page.check('[name="credit"]');
    await page.click('button:has-text("Next Step")');
    
    // Select a date (click on a future date)
    await page.waitForSelector('.calendar');
    const availableDates = page.locator('.calendar-day:not(.disabled):not(.past)');
    const firstAvailable = availableDates.first();
    await firstAvailable.click();
    
    // Select a time slot
    await page.click('button:has-text("10:00 AM")');
    
    // Proceed to step 3
    await page.click('button:has-text("Next Step")');
    
    // Should show contact form
    await expect(page.locator('input[name="firstName"]')).toBeVisible({ timeout: 2000 });
  });

  test.skip('Complete booking submission (requires CSRF token)', async ({ page }) => {
    // Skip this test until CSRF token issue is resolved
    // This would test the full booking flow from start to confirmation
  });

});

test.describe('API Endpoints', () => {
  
  test('Health check endpoint returns OK', async ({ request }) => {
    const response = await request.get(`${BASE_URL}/health`);
    expect(response.ok()).toBeTruthy();
    
    const data = await response.json();
    expect(data.status).toBe('ok');
    expect(data.database).toBe(true);
    console.log(`✅ Health check passed: ${JSON.stringify(data)}`);
  });

  test('Properties API returns data', async ({ request }) => {
    const response = await request.get(`${BASE_URL}/api/v1/properties?limit=10`);
    expect(response.ok()).toBeTruthy();
    
    const data = await response.json();
    expect(data.success).toBe(true);
    expect(data.data).toHaveProperty('properties');
    expect(Array.isArray(data.data.properties)).toBeTruthy();
    
    const count = data.data.properties.length;
    console.log(`✅ Properties API returned ${count} properties`);
    
    // Verify property structure
    if (count > 0) {
      const property = data.data.properties[0];
      expect(property).toHaveProperty('id');
      expect(property).toHaveProperty('address');
      expect(property).toHaveProperty('city');
      expect(property).toHaveProperty('price');
      expect(property).toHaveProperty('status');
      
      // Address should NOT be [ENCRYPTED] in API response
      expect(property.address).not.toBe('[ENCRYPTED]');
    }
  });

  test('Single property API returns decrypted data', async ({ request }) => {
    // Get property list first
    const listResponse = await request.get(`${BASE_URL}/api/v1/properties?limit=1`);
    const listData = await listResponse.json();
    
    if (listData.data.properties.length === 0) {
      test.skip('No properties to test');
      return;
    }
    
    const propertyId = listData.data.properties[0].id;
    
    // Get single property
    const response = await request.get(`${BASE_URL}/api/v1/properties/${propertyId}`);
    expect(response.ok()).toBeTruthy();
    
    const data = await response.json();
    expect(data.success).toBe(true);
    expect(data.data).toHaveProperty('id');
    expect(data.data).toHaveProperty('address');
    
    // Address should be decrypted
    expect(data.data.address).not.toBe('[ENCRYPTED]');
    expect(data.data.address).not.toBe('');
    console.log(`✅ Property API returned address: ${data.data.address}`);
  });

  test('Booking API requires CSRF token', async ({ request }) => {
    const response = await request.post(`${BASE_URL}/api/v1/bookings`, {
      data: {
        property_id: 1,
        first_name: 'Test',
        last_name: 'User',
        email: 'test@example.com',
        phone: '555-555-5555',
        showing_date: '2025-12-15',
        showing_time: '10:00 AM'
      }
    });
    
    // Should fail without CSRF token
    expect(response.status()).toBe(403);
    
    const data = await response.json();
    expect(data.error).toContain('CSRF');
    console.log('✅ CSRF protection working correctly');
  });

});

test.describe('Admin Flows', () => {
  
  test('Admin login page loads', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin`);
    
    // Check login form displays
    await expect(page.locator('h1, h2')).toContainText(/Admin|Sign In/i);
    await expect(page.locator('input[type="email"]')).toBeVisible();
    await expect(page.locator('input[type="password"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
  });

  test('Admin login shows rate limit message correctly', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin?rate_limited=true&retry_after=60`);
    
    // Should show rate limit message
    await expect(page.locator('text=/Too many login attempts/i')).toBeVisible();
    
    // Should show countdown number (not just "seconds")
    const messageText = await page.locator('.rate-limit-notice, .error-message').textContent();
    expect(messageText).toMatch(/\d+\s*seconds/i);
    console.log('✅ Rate limit message displays correctly');
  });

  test.skip('Admin login with valid credentials (requires test account)', async ({ page }) => {
    // Skip this test until test admin account is created
  });

});

test.describe('Mobile Responsiveness', () => {
  
  test.use({ viewport: { width: 375, height: 667 } }); // iPhone SE

  test('Homepage is mobile responsive', async ({ page }) => {
    await page.goto(BASE_URL);
    
    // Check hamburger menu appears
    const hamburger = page.locator('.hamburger, .mobile-menu-toggle, button[aria-label*="menu"]');
    await expect(hamburger).toBeVisible({ timeout: 2000 });
    
    // Check hero section stacks vertically
    const hero = page.locator('.hero, [class*="hero"]');
    const heroBox = await hero.boundingBox();
    expect(heroBox.width).toBeLessThanOrEqual(375);
    
    console.log('✅ Homepage mobile responsive');
  });

  test('Properties grid stacks on mobile', async ({ page }) => {
    await page.goto(`${BASE_URL}/properties`);
    
    await page.waitForSelector('#propertyGrid', { timeout: 5000 });
    
    const grid = page.locator('#propertyGrid');
    const gridBox = await grid.boundingBox();
    
    // Grid should not cause horizontal scroll
    expect(gridBox.width).toBeLessThanOrEqual(375);
    
    console.log('✅ Properties grid mobile responsive');
  });

  test('Booking form is usable on mobile', async ({ page }) => {
    await page.goto(`${BASE_URL}/book-showing`);
    
    // All form elements should be visible and within viewport
    const form = page.locator('form');
    const formBox = await form.boundingBox();
    expect(formBox.width).toBeLessThanOrEqual(375);
    
    // Buttons should be full width on mobile
    const nextButton = page.locator('button:has-text("Next Step")');
    const buttonBox = await nextButton.boundingBox();
    expect(buttonBox.width).toBeGreaterThan(300); // Nearly full width
    
    console.log('✅ Booking form mobile responsive');
  });

});
