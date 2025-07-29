# MCS Testing Framework

## Directory Structure

```
test/
├── coverage/           # Test coverage files
│   ├── unit/          # Unit test coverage files
│   ├── integration/   # Integration test coverage files
│   └── e2e/          # End-to-end test coverage files
├── docs/             # Test documentation and planning
├── fixtures/         # Test data and fixtures
├── integration/      # Integration test suites
├── reports/          # Test reports (XML, JSON, HTML)
└── scripts/          # Test automation scripts
```

## Coverage Files

- Unit test coverage files (.out) are stored in `test/coverage/unit/`
- Integration test coverage files are stored in `test/coverage/integration/`
- E2E test coverage files are stored in `test/coverage/e2e/`

## Usage

Use the Makefile targets for clean test management:

```bash
make test-unit          # Run unit tests with coverage
make test-integration   # Run integration tests
make test-e2e          # Run end-to-end tests
make test-coverage     # Generate coverage reports
make test-clean        # Clean test artifacts
```

## Reports

Test reports are generated in multiple formats and stored in `test/reports/`.