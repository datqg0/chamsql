#!/bin/zsh

# Set proper ownership for node_modules and pnpm store
sudo chown -R node:node /workspace/templateUI/client/node_modules 
sudo chown -R node:node /workspace/templateUI/client/.pnpm_store

# Install dependencies
pnpm install

echo "✅ DevContainer setup completed!"
