#!/usr/bin/env bash
set -e

cd /mnt/c/Users/Admin/Desktop/devops/frontend

echo "=== Installing Frontend Dependencies ==="
npm install

echo "=== Building Production Frontend Bundle ==="
npm run build

echo "FRONTEND_BUILD_SUCCESS"
