#!/bin/bash
# Test script to run tests without tokman interception
export PATH=/opt/homebrew/bin:$PATH
export SKIP_ENV_VALIDATION=1
go test -v -run TestSecurity ./tests/integration/...
