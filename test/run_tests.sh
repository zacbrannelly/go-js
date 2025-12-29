#!/bin/bash

set -e

echo "🏗️  Building go-js binary..."
cd ../
./build.sh
cd test

echo ""
echo "🧪 Running Test262 tests..."
echo ""

# Go to the harness directory.
cd test262-harness

# Use the correct node version.
export NVM_DIR=$HOME/.nvm;
source $NVM_DIR/nvm.sh;

nvm use --silent

# Run the tests.
node bin/run.js \
  --host-type=go-js \
  --host-path=../../bin/go-js \
  --test262-dir ../test262 \
  "$@"
