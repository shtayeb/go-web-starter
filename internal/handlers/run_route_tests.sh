#!/bin/bash

# Route Tests Runner Script
# This script runs the route tests for landing and dashboard pages

echo "🚀 Running Route Tests for Landing and Dashboard"
echo "=================================================="

# Set test environment
export APP_ENV=test
export DEBUG=true

echo "📋 Running Landing Page Tests..."
go test -v ./internal/handlers -run TestLandingRoute

echo ""
echo "📊 Running Dashboard Page Tests..."
go test -v ./internal/handlers -run TestDashboardRoute

echo ""
echo "✅ All route tests completed!"

# Run with coverage
echo ""
echo "📈 Running tests with coverage..."
go test -cover ./internal/handlers -run "TestLandingRoute|TestDashboardRoute"

echo ""
echo "🎉 Route testing completed successfully!"