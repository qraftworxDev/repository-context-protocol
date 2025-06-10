#!/bin/bash
set -e

echo "Running integration tests..."

# Ensure binaries exist
if [ ! -f "bin/repocontext" ]; then
    echo "Error: repocontext binary not found. Run 'make build' first."
    exit 1
fi

# Create test directory
TEST_DIR="testdata/integration-test"
rm -rf "$TEST_DIR"
mkdir -p "$TEST_DIR"

# Create a simple Go file for testing
cat > "$TEST_DIR/main.go" << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println(greet("World"))
}

func greet(name string) string {
    return fmt.Sprintf("Hello, %s!", name)
}
EOF

cd "$TEST_DIR"

# Test init command
echo "Testing init command..."
../../bin/repocontext init

# Verify .repocontext directory was created
if [ ! -d ".repocontext" ]; then
    echo "Error: .repocontext directory not created"
    exit 1
fi

echo "Integration tests passed!" 