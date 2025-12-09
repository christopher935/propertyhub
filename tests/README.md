# PropertyHub Automated Test Suite

Comprehensive end-to-end testing for PropertyHub using Playwright.

## Quick Start

```bash
# Install dependencies
cd tests
npm install

# Install Playwright browsers
npx playwright install

# Run all tests
npm test

# Run tests with UI
npm run test:ui

# Run specific browser
npm run test:chrome
npm run test:firefox
npm run test:safari
```

## Test Coverage

### Property Flows
- ✅ Homepage featured properties display
- ✅ Properties browse page grid and filters
- ✅ Property detail page with all information
- ✅ Search and filter functionality
- ✅ Property address NOT showing "[ENCRYPTED]"
- ✅ Bathroom formatting (2.5 not 2.500000)

### Booking Flows
- ✅ Booking form loads with all steps
- ✅ Form validation works
- ✅ Date and time selection functional
- ⏭️ Complete booking submission (pending CSRF fix)

### API Endpoints
- ✅ Health check returns OK
- ✅ Properties API returns data
- ✅ Single property API decrypts address
- ✅ Booking API enforces CSRF protection

### Admin Flows
- ✅ Admin login page loads
- ✅ Rate limit message displays correctly
- ⏭️ Admin authentication (pending test account)

### Mobile Responsiveness
- ✅ Homepage responsive layout
- ✅ Properties grid stacks on mobile
- ✅ Booking form usable on mobile

## Test Commands

```bash
# Development
npm test                    # Run all tests
npm run test:headed         # Run with visible browser
npm run test:debug          # Run in debug mode
npm run test:ui             # Run with Playwright UI

# Specific Browsers
npm run test:chrome         # Chrome only
npm run test:firefox        # Firefox only
npm run test:safari         # Safari/WebKit only
npm run test:mobile         # Mobile Chrome + Safari

# CI/CD
npm run test:ci             # Run with GitHub reporter
npm run report              # View last test report
```

## Configuration

**Base URL:** Set in `playwright.config.js` or environment variable
```bash
export BASE_URL=http://209.38.116.238:8080
npm test
```

**Browsers:** Tests run against:
- Chromium (desktop)
- Firefox (desktop)
- WebKit/Safari (desktop)
- Mobile Chrome (Pixel 5)
- Mobile Safari (iPhone 13)

## Test Files

- **`e2e/property-flows.spec.js`** - All property-related tests
- **`playwright.config.js`** - Playwright configuration
- **`package.json`** - Dependencies and scripts

## Writing New Tests

```javascript
const { test, expect } = require('@playwright/test');

test('my new test', async ({ page }) => {
  await page.goto('http://localhost:8080/my-page');
  
  // Check element visible
  await expect(page.locator('h1')).toBeVisible();
  
  // Check text content
  await expect(page.locator('h1')).toContainText('Expected Text');
  
  // Click button
  await page.click('button:has-text("Submit")');
  
  // Wait for navigation
  await page.waitForURL('**/success');
});
```

## Debugging Failed Tests

### View Test Report
```bash
npm run report
# Opens HTML report in browser
```

### Run in Debug Mode
```bash
npm run test:debug
# Pauses at each step, shows browser DevTools
```

### Run Specific Test
```bash
npx playwright test -g "Homepage loads"
# Runs only tests matching "Homepage loads"
```

### View Screenshots
Failed tests automatically capture screenshots:
```
tests/test-results/
  ├── property-flows-chromium/
  │   ├── test-failed-1-actual.png
  │   └── trace.zip
```

### View Traces
```bash
npx playwright show-trace tests/test-results/.../trace.zip
# Shows timeline of test execution with DOM snapshots
```

## CI/CD Integration

### GitHub Actions

**File:** `.github/workflows/tests.yml`
```yaml
name: E2E Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: 18
      - name: Install dependencies
        run: |
          cd tests
          npm ci
          npx playwright install --with-deps
      - name: Run tests
        run: |
          cd tests
          npm run test:ci
        env:
          BASE_URL: ${{ secrets.BASE_URL }}
      - name: Upload test results
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: playwright-report
          path: tests/playwright-report/
```

### Jenkins

```groovy
pipeline {
  agent any
  stages {
    stage('E2E Tests') {
      steps {
        sh '''
          cd tests
          npm ci
          npx playwright install
          npm test
        '''
      }
    }
  }
  post {
    always {
      publishHTML([
        reportDir: 'tests/playwright-report',
        reportFiles: 'index.html',
        reportName: 'Playwright Test Report'
      ])
    }
  }
}
```

## Prerequisites

Before tests will pass, ensure:
1. ✅ P0 Fix #1: Properties displaying on frontend
2. ✅ P0 Fix #2: Property addresses not encrypted
3. ✅ P0 Fix #3: Booking form submission working
4. ✅ P0 Fix #4: Admin login rate limit error fixed

## Known Issues / Skipped Tests

- **Booking submission test:** Skipped until CSRF token issue resolved
- **Admin authentication test:** Skipped until test admin account created
- **Integration tests:** Require FUB/SES/SNS configuration (see INTEGRATION_TESTING_GUIDE.md)

## Troubleshooting

### Tests Timeout
Increase timeout in `playwright.config.js`:
```javascript
timeout: 60 * 1000  // 60 seconds
```

### Browser Not Installed
```bash
npx playwright install
```

### Port Already in Use
Change BASE_URL to different port:
```bash
export BASE_URL=http://localhost:8081
```

### Tests Pass Locally But Fail in CI
- Check environment variables (BASE_URL, API keys)
- Verify database seeded with test data
- Check CI server can access application URL
- Review CI logs for network/permission errors

## Test Data Management

### Using Test Database
```bash
# Before tests
export DB_NAME=propertyhub_test
npm run test

# After tests
dropdb propertyhub_test  # Clean up
```

### Seeding Test Data
```bash
# Create test properties before running tests
psql -d propertyhub_test < tests/fixtures/properties.sql
```

## Performance Benchmarks

**Target Test Duration:**
- Full suite: < 5 minutes
- Single test file: < 2 minutes
- Individual test: < 30 seconds

**Current Performance:** Run `npm test` to see timing.

## Contributing

When adding new features:
1. Write tests FIRST (TDD approach)
2. Ensure tests pass locally
3. Add tests to appropriate spec file
4. Update this README if new test categories added

## Support

- **Documentation:** https://playwright.dev
- **Issues:** File in GitHub repository
- **Integration Testing:** See `INTEGRATION_TESTING_GUIDE.md`
- **Bug Report Template:** `.github/issues/P0-XX-template.md`

---

**Last Updated:** December 9, 2025  
**Test Framework:** Playwright v1.40.0  
**Node Version:** 18.x or higher
