#!/bin/bash
set -e

# Ensure the target directory exists
mkdir -p src/frontend/dist

# Copy the built frontend files
cp -r web/dist/* src/frontend/dist/

echo "Frontend files copied successfully" 