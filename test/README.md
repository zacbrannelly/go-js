# Test262 Testing

This directory contains the setup for running Test262 conformance tests against the go-js JavaScript engine.

## Setup

Please install Node Version Manager (`nvm`) if you have not already.

Before running tests, you need to set up the test environment:

```bash
./setup_tests.sh
```

This will install dependencies of the harness and link the harness to our fork of `eshost`.

## Running the tests

For example:

```bash
./run_tests.sh ../test262/test/**/*.js --save-only-failed
```

For more information on what flags are supported, see the documentation for [test262-harness](https://github.com/tc39/test262-harness).
