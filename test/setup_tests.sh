#!/bin/bash

set -e

echo "🔧 Setting up Test262 test environment..."
echo ""

# Check if Node Version Manager is installed.
export NVM_DIR=$HOME/.nvm;
if [ ! -d $NVM_DIR ]; then
  echo "❌ Node Version Manager is not installed."
  echo "   Please install it from https://github.com/nvm-sh/nvm"
  exit 1
fi

# Allow use of Node Version Manager.
source $NVM_DIR/nvm.sh;

# Go to the eshost directory.
echo "📦 Setting up eshost..."
cd eshost

# Use the correct node version.
nvm use --silent

# Install the dependencies.
echo "   Installing dependencies..."
npm install --silent

# Link the eshost binary.
echo "   Linking eshost binary..."
npm link --silent

# Go to the harness directory.
echo ""
echo "📦 Setting up test262-harness..."
cd ../test262-harness

# Use the correct node version.
nvm use --silent

# Install the harness dependencies.
echo "   Installing dependencies..."
npm install --silent

# Link the harness to the eshost package.
echo "   Linking to eshost..."
npm link eshost --silent

echo ""
echo "✅ Setup complete!"
