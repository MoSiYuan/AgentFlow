# Testing Skill

## Description

Comprehensive testing automation including unit, integration, and E2E tests.

## Capabilities

- Test suite creation and execution
- Coverage reporting
- Parallel test execution
- Test result analysis
- Regression detection

## Usage Examples

```bash
# Run all tests
agentflow create "Full test suite" -d "npm test"

# Test specific module
agentflow create "Test auth module" \
  -d "npm test -- auth.test.ts"

# Parallel testing
agentflow create "Parallel tests" -d "
  - Unit: npm run test:unit
  - Integration: npm run test:integration
  - E2E: npm run test:e2e
"

# Coverage analysis
agentflow create "Coverage check" \
  -d "npm test -- --coverage --watchAll=false"
```

## Test Templates

### Unit Test Template
```typescript
describe('FeatureName', () => {
  it('should handle expected case', () => {
    // Arrange
    const input = { /* test data */ };

    // Act
    const result = featureName(input);

    // Assert
    expect(result).toEqual({ /* expected output */ });
  });

  it('should handle edge case', () => {
    // Test edge cases
  });
});
```

### Integration Test Template
```typescript
describe('FeatureName Integration', () => {
  beforeAll(async () => {
    // Setup: database, server, etc.
  });

  afterAll(async () => {
    // Cleanup
  });

  it('should integrate with other modules', () => {
    // Integration test
  });
});
```

## Coverage Goals

- **Statements**: > 80%
- **Branches**: > 75%
- **Functions**: > 80%
- **Lines**: > 80%

## Parallel Test Execution

AgentFlow can run tests in parallel for faster feedback:

```bash
# Create test tasks
agentflow create "Frontend tests" -d "cd frontend && npm test"
agentflow create "Backend tests" -d "cd backend && npm test"
agentflow create "API tests" -d "cd api && npm test"

# All execute simultaneously
# Results collected and reported together
```
