# End-to-End Testing with Playwright

This directory contains Playwright-based end-to-end tests for the Task Management System frontend.

## Prerequisites

- Node.js 18+ 
- Task Management System built (`make build`)

## Setup

Install Playwright and dependencies:

```bash
make test-e2e-setup
```

This will:
1. Install npm dependencies
2. Install Playwright browsers (Chromium, Firefox, WebKit)

## Running Tests

### Run all tests (headless):
```bash
make test-e2e
```

### Run tests in UI mode (interactive):
```bash
make test-e2e-ui
```

### Run tests in headed mode (see browser):
```bash
cd tests/e2e && npm run test:headed
```

### Run specific test file:
```bash
cd tests/e2e && npx playwright test specs/auth.spec.ts
```

### Debug a test:
```bash
cd tests/e2e && npm run test:debug
```

## Test Structure

```
tests/e2e/
├── package.json           # Dependencies and scripts
├── playwright.config.ts   # Playwright configuration
├── specs/                 # Test specifications
│   ├── auth.spec.ts      # Authentication tests
│   ├── navigation.spec.ts # Navigation and menu tests
│   └── projects.spec.ts   # Project management tests
├── test-results/         # Test execution results (gitignored)
└── playwright-report/    # HTML test reports (gitignored)
```

## Test Categories

### Authentication (`auth.spec.ts`)
- Login page display
- User login with credentials
- User registration
- Input validation
- Error handling

### Navigation (`navigation.spec.ts`)
- Menu open/close
- Screen navigation
- Logout functionality

### Projects (`projects.spec.ts`)
- Project selector
- Project switching
- Board display

## Configuration

The tests are configured to:
- Run on Chromium, Firefox, and WebKit
- Automatically start the Task server on port 8080
- Use a test database (`task.test.db`)
- Take screenshots on failure
- Record traces on retry

## Viewing Reports

After running tests, view the HTML report:

```bash
cd tests/e2e && npm run report
```

## Writing New Tests

Create new test files in `specs/` directory:

```typescript
import { test, expect } from '@playwright/test';

test.describe('Feature Name', () => {
  test.beforeEach(async ({ page }) => {
    // Setup code
  });

  test('should do something', async ({ page }) => {
    // Test code
    await expect(page.locator('#element')).toBeVisible();
  });
});
```

## CI/CD Integration

Tests can be run in CI environments. Set environment variables:

```bash
CI=true npm test
```

This will:
- Enable retries (2 attempts)
- Run tests sequentially
- Fail fast on errors
