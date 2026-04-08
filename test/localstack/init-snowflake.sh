#!/bin/bash

# LocalStack initialization script for Snowflake emulator
# This script runs when LocalStack starts and sets up the test environment

set -e

echo "🚀 Initializing LocalStack Snowflake environment..."

# Wait for Snowflake to be ready
echo "⏳ Waiting for Snowflake emulator..."
sleep 5

# Create test databases
echo "📦 Creating test databases..."
awslocal snowflake create-database --name TEST_DB || true
awslocal snowflake create-database --name PROD_DB || true
awslocal snowflake create-database --name DEV_DB || true

# Create test warehouses
echo "🏭 Creating test warehouses..."
awslocal snowflake create-warehouse --name ANALYTICS_WH || true
awslocal snowflake create-warehouse --name COMPUTE_WH || true

# Create test schemas
echo "📋 Creating test schemas..."
awslocal snowflake create-schema --database TEST_DB --name PUBLIC || true
awslocal snowflake create-schema --database PROD_DB --name RAW_DATA || true
awslocal snowflake create-schema --database PROD_DB --name ANALYTICS || true
awslocal snowflake create-schema --database DEV_DB --name SANDBOX || true

# Create test tables
echo "📊 Creating test tables..."
awslocal snowflake execute-sql --database TEST_DB --schema PUBLIC --sql "CREATE TABLE IF NOT EXISTS CUSTOMERS (id INT, name VARCHAR, email VARCHAR)" || true
awslocal snowflake execute-sql --database TEST_DB --schema PUBLIC --sql "CREATE TABLE IF NOT EXISTS ORDERS (id INT, customer_id INT, amount DECIMAL)" || true
awslocal snowflake execute-sql --database PROD_DB --schema RAW_DATA --sql "CREATE TABLE IF NOT EXISTS EVENTS (id INT, event_type VARCHAR, timestamp TIMESTAMP)" || true
awslocal snowflake execute-sql --database PROD_DB --schema ANALYTICS --sql "CREATE TABLE IF NOT EXISTS USER_METRICS (user_id INT, metric_name VARCHAR, value DECIMAL)" || true
awslocal snowflake execute-sql --database PROD_DB --schema ANALYTICS --sql "CREATE TABLE IF NOT EXISTS REVENUE_REPORTS (report_id INT, revenue DECIMAL, date DATE)" || true
awslocal snowflake execute-sql --database DEV_DB --schema SANDBOX --sql "CREATE TABLE IF NOT EXISTS TEST_TABLE (id INT, data VARCHAR)" || true

echo "✅ LocalStack Snowflake environment initialized!"
echo ""
echo "📍 Connection details:"
echo "   Account: test.localstack"
echo "   Host: localhost:4566"
echo "   User: test"
echo "   Password: test"
echo ""
