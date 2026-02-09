#!/bin/bash
# Quick Start Script for tsf

set -e

echo "==================================="
echo "tsf - Quick Start"
echo "==================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go 1.23 or later."
    exit 1
fi

# Build the binary
echo "1. Building sf binary..."
make build
echo "✓ Build complete"
echo ""

# Setup database directory
DB_DIR="$HOME/.config/sf"
DB_PATH="$DB_DIR/sf.db"
mkdir -p "$DB_DIR"

# Initialize database
echo "2. Initializing database at $DB_PATH..."
./sf initdb -f "$DB_PATH"
echo "✓ Database initialized"
echo ""

# Display credentials
echo "==================================="
echo "Default Credentials:"
echo "-----------------------------------"
echo "Username: admin"
echo "Password: admin123"
echo "==================================="
echo ""

# Set environment variables
export SF_URL=http://localhost:8080
export SF_USERNAME=admin
export SF_PASSWORD=admin123

echo "3. Environment variables set:"
echo "   SF_URL=$SF_URL"
echo "   SF_USERNAME=$SF_USERNAME"
echo "   SF_PASSWORD=***"
echo ""

# Instructions
echo "==================================="
echo "Next Steps:"
echo "==================================="
echo ""
echo "To start the server:"
echo "  ./sf server -f $DB_PATH -port 8080"
echo ""
echo "In another terminal, test the CLI:"
echo "  export SF_URL=http://localhost:8080"
echo "  export SF_USERNAME=admin"
echo "  export SF_PASSWORD=admin123"
echo ""
echo "  ./sf project list"
echo "  ./sf task create -title 'My First Task'"
echo "  ./sf task list"
echo ""
echo "Open web browser:"
echo "  http://localhost:8080"
echo ""
echo "To start an orchestrator:"
echo "  ./sf orchestrator -url http://localhost:8080 -username orchestrator -password <password>"
echo ""
echo "To start a worker:"
echo "  ./sf worker -url http://localhost:8080 -username worker1 -password <password>"
echo ""
echo "For more information, see:"
echo "  - README.md"
echo "  - USER_GUIDE.md"
echo "  - CONTRIBUTING.md"
echo ""
echo "==================================="
echo "Setup Complete!"
echo "==================================="
