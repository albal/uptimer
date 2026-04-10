#!/bin/bash
set -e

# Uptimer - Debian 13 Setup Script
# Run this as root on your Debian 13 host.

echo "Installing dependencies..."
apt-get update
apt-get install -y nginx postgresql postgresql-contrib curl git build-essential

# Configure PostgreSQL
echo "Configuring PostgreSQL..."
sudo -u postgres psql -c "CREATE USER uptimer WITH PASSWORD 'uptimer_password';" || true
sudo -u postgres psql -c "CREATE DATABASE uptimer OWNER uptimer;" || true

# Pre-migration: Enable uuid-ossp
sudo -u postgres psql -d uptimer -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"

# Create application directory
mkdir -p /opt/uptimer/static
chown -R $USER:$USER /opt/uptimer

echo "Setup complete! Please configure your .env file in /opt/uptimer/.env"
echo "and then deploy your application using the GitHub Actions workflow."
