#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

# Test counter
TESTS=0
PASSED=0

# Test function for comparing output
run_test() {
    local name="$1"
    local cmd="$2"
    local expected="$3"
    local test_dir="$4"
    
    TESTS=$((TESTS + 1))
    echo -n "  Testing: $name ... "
    
    # Change to test directory and add current dir to PATH for clean jdq command
    if output=$(cd "$test_dir" && PATH="../../:$PATH" eval "$cmd" 2>&1); then
        # Try JSON comparison first
        if output_normalized=$(echo "$output" | jq -S . 2>/dev/null) && 
           expected_normalized=$(echo "$expected" | jq -S . 2>/dev/null); then
            if [ "$output_normalized" = "$expected_normalized" ]; then
                echo -e "${GREEN}PASS${NC}"
                PASSED=$((PASSED + 1))
                return
            fi
        fi
        
        # Multi-line string comparison
        if [ "$(echo "$output" | wc -l)" -gt 1 ] || [ "$(echo "$expected" | wc -l)" -gt 1 ]; then
            if [ "$output" = "$expected" ]; then
                echo -e "${GREEN}PASS${NC}"
                PASSED=$((PASSED + 1))
                return
            fi
        fi
        
        # Direct string comparison
        if [ "$output" = "$expected" ]; then
            echo -e "${GREEN}PASS${NC}"
            PASSED=$((PASSED + 1))
        else
            echo -e "${RED}FAIL${NC}"
            echo "    Expected: '$expected'"
            echo "    Got: '$output'"
        fi
    else
        echo -e "${RED}FAIL${NC} (command failed)"
        echo "    Error: $output"
    fi
}

# Setup command (ignore output and errors)
run_setup() {
    local cmd="$1"
    local test_dir="$2"
    echo -e "  ${BLUE}Setup: $cmd${NC}"
    (cd "$test_dir" && PATH="../../:$PATH" eval "$cmd" > /dev/null 2>&1) || true
}

# Run tests for a specific category
run_category_tests() {
    local category_dir="$1"
    local test_file="$category_dir/spec.txt"
    
    if [[ ! -f "$test_file" ]]; then
        echo -e "${YELLOW}Warning: No spec.txt found in $category_dir${NC}"
        return
    fi
    
    echo -e "${CYAN}Running $(basename "$category_dir") tests...${NC}"
    
    # Run setup script if it exists
    if [[ -x "$category_dir/setup" ]]; then
        echo -e "  ${BLUE}Running setup script...${NC}"
        (cd "$category_dir" && PATH="../../:$PATH" ./setup) || true
    fi
    
    # Read test cases from file
    while IFS= read -r line; do
        # Skip empty lines
        [[ -z "$line" ]] && continue
        
        # Parse test name from comment
        if [[ "$line" =~ ^#[[:space:]]*(.*) ]]; then
            test_name="${BASH_REMATCH[1]}"
            
            # Read command line
            if ! IFS= read -r cmd_line; then
                echo -e "  ${RED}Error: Missing command line after test name: $test_name${NC}"
                continue
            fi
            
            # Check if this is a setup command (starts with -)
            if [[ "$cmd_line" =~ ^-(.+) ]]; then
                # Setup command - execute but don't test output
                setup_cmd="${BASH_REMATCH[1]}"
                run_setup "$setup_cmd" "$category_dir"
            else
                # Regular test command - read expected output
                if ! IFS= read -r expected_output; then
                    echo -e "  ${RED}Error: Missing expected output after command: $cmd_line${NC}"
                    continue
                fi
                
                # Run the test
                run_test "$test_name" "$cmd_line" "$expected_output" "$category_dir"
            fi
        # Handle standalone setup commands
        elif [[ "$line" =~ ^-(.+) ]]; then
            setup_cmd="${BASH_REMATCH[1]}"
            run_setup "$setup_cmd" "$category_dir"
        fi
    done < "$test_file"
    
    # Run teardown script if it exists
    if [[ -x "$category_dir/teardown" ]]; then
        echo -e "  ${BLUE}Running teardown script...${NC}"
        (cd "$category_dir" && PATH="../../:$PATH" ./teardown) || true
    fi
    
    echo ""
}

# Help function
show_help() {
    echo "Usage: $0 [category...]"
    echo ""
    echo "Run examples for jdq. If no category is specified, run all examples."
    echo ""
    echo "Available example categories:"
    for dir in examples/*/; do
        if [[ -d "$dir" && -f "$dir/spec.txt" ]]; then
            category=$(basename "$dir")
            echo "  $category"
        fi
    done
    echo ""
    echo "Examples:"
    echo "  $0                    # Run all tests"
    echo "  $0 basic_queries      # Run only basic query tests"
    echo "  $0 date_handling value_inheritance  # Run specific categories"
}

# Main execution
if [[ "$1" == "--help" || "$1" == "-h" ]]; then
    show_help
    exit 0
fi

echo "Running examples..."
echo ""

# If specific categories provided, run only those
if [[ $# -gt 0 ]]; then
    for category in "$@"; do
        # Support both "basic_queries" and "examples/basic_queries" formats
        if [[ "$category" == examples/* ]]; then
            category_dir="$category"
        else
            category_dir="examples/$category"
        fi
        
        if [[ -d "$category_dir" ]]; then
            run_category_tests "$category_dir"
        else
            echo -e "${RED}Error: Category '$category' not found${NC}"
            exit 1
        fi
    done
else
    # Run all categories
    for category_dir in examples/*/; do
        if [[ -d "$category_dir" ]]; then
            run_category_tests "$category_dir"
        fi
    done
fi

# Summary
echo "======================================="
echo "Tests: $TESTS, Passed: $PASSED, Failed: $((TESTS - PASSED))"
echo "======================================="

if [ $PASSED -eq $TESTS ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
fi