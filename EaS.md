# Examples as Specifications (EaS)

EaS (Examples as Specifications) is a testing methodology that combines executable documentation with automated testing. Tests are written as examples that serve both as documentation and specifications.

## Philosophy

- **Examples are Specifications**: Each example demonstrates expected behavior and serves as a living specification
- **Executable Documentation**: Examples can be executed to verify the system works as documented
- **Simplicity**: Test syntax should be intuitive and readable by both humans and machines
- **Self-Documenting**: The test files themselves serve as comprehensive documentation

## File Structure

```
examples/
├── category1/
│   ├── spec.txt       # Test specifications
│   ├── data.json      # Test data (if needed)
│   ├── setup*         # Optional setup script
│   └── teardown*      # Optional teardown script
├── category2/
│   ├── spec.txt
│   └── data.json
└── ...
```

## Test Syntax

### Basic Test Case
```
# Test description
command_to_execute
expected_output
```

### Error Test Case (expects non-zero exit code)
```
# Test description  
!command_that_should_fail
```

### Setup Command (output ignored)
```
# Test description
-setup_command_here
```

### Standalone Commands
```
-setup_command
!error_command
```

## Test Types

### 1. Regular Tests
- **Format**: `# description` → `command` → `expected output`
- **Purpose**: Verify command produces expected output
- **Validation**: 
  - JSON comparison (normalized with `jq -S`)
  - Multi-line string comparison
  - Direct string comparison

### 2. Error Tests  
- **Format**: `# description` → `!command`
- **Purpose**: Verify command exits with non-zero code
- **Validation**: Exit code ≠ 0
- **No expected output line needed**

### 3. Setup Commands
- **Format**: `# description` → `-command` or standalone `-command`
- **Purpose**: Prepare test environment
- **Validation**: None (output and errors ignored)

## Test Execution

### Runner Features
- **Automatic Discovery**: Finds all `examples/*/spec.txt` files
- **Category Filtering**: Run specific test categories
- **Colored Output**: Visual feedback with pass/fail indicators
- **JSON Normalization**: Handles JSON output comparison intelligently
- **Error Reporting**: Shows expected vs actual output on failures

### Execution Context
- **Working Directory**: Each test runs in its category directory
- **PATH Enhancement**: Binary under test is added to PATH
- **Environment**: Clean environment for each test

## Example Categories

### Typical Categories
- **basic_queries**: Core functionality tests
- **error_handling**: Error condition tests  
- **integration**: End-to-end scenario tests
- **edge_cases**: Boundary condition tests
- **performance**: Performance-related tests

## Best Practices

### Test Organization
- Group related tests into logical categories
- Use descriptive test names that explain the scenario
- Keep test data files small and focused
- Use meaningful category names

### Test Writing
- Write tests that demonstrate real usage patterns
- Include both positive and negative test cases
- Test edge cases and error conditions
- Keep expected outputs concise and relevant

### Error Testing
- Use `!` prefix for commands that should fail
- Focus on exit code verification rather than error message parsing
- Test various error conditions systematically

### Data Management
- Keep test data minimal and focused
- Use realistic but simplified data structures
- Include edge cases in test data (empty values, boundary dates, etc.)

## Implementation Details

### Test Runner Components
- **Parser**: Processes spec.txt files and identifies test types
- **Executor**: Runs commands in appropriate context
- **Validator**: Compares outputs using multiple strategies
- **Reporter**: Provides formatted test results

### Comparison Strategies
1. **JSON Comparison**: Normalizes and compares JSON structures
2. **Multi-line Text**: Handles multi-line expected outputs
3. **String Comparison**: Direct string matching for simple cases

### Error Handling
- Setup/teardown failures don't stop test execution
- Missing files are reported but don't cause runner failure
- Malformed test cases are reported and skipped

## Usage Examples

### Running Tests
```bash
# Run all tests
./test_examples.sh

# Run specific category
./test_examples.sh basic_queries

# Run multiple categories  
./test_examples.sh basic_queries error_handling

# Show help
./test_examples.sh --help
```

### Sample spec.txt
```
# Basic functionality test
mycommand --option value input.json
{"result": "success", "count": 42}

# Error condition test
!mycommand --invalid-option

# Setup for subsequent tests
-echo '{"test": "data"}' > temp.json

# Test using setup data
mycommand temp.json
{"processed": true}
```

## Benefits

### For Developers
- **Living Documentation**: Tests document actual system behavior
- **Regression Prevention**: Automated verification of examples
- **Refactoring Safety**: Changes that break examples are immediately detected

### For Users
- **Clear Examples**: Real usage patterns are demonstrated
- **Reliable Documentation**: Examples are guaranteed to work
- **Comprehensive Coverage**: Edge cases and error conditions are documented

### For Teams
- **Shared Understanding**: Examples provide common reference point
- **Quality Assurance**: Systematic testing of documented features
- **Maintenance**: Changes to behavior require updating documentation

## Integration

### CI/CD Pipeline
- Run EaS tests as part of automated testing
- Fail builds when examples don't match reality
- Generate documentation from successful examples

### Development Workflow
- Write examples before implementing features (TDD approach)
- Update examples when changing behavior
- Use examples for manual testing and debugging

## Advanced Features

### Custom Validation
- JSON schema validation for complex outputs
- Custom comparison functions for specific formats
- Performance thresholds for timing-sensitive tests

### Environment Management
- Docker integration for consistent test environments
- Parameterized tests for different configurations
- Resource cleanup and isolation

This specification defines EaS as a powerful methodology that bridges the gap between documentation and testing, ensuring that examples remain accurate and useful throughout the software development lifecycle.