#!/bin/sh

# Script to run all pre-compiled tests without having the go toolkit installed.
# It looks for all files with the .test extension and executes them with the -test.short and -test.v flags.
# If any test fails, it will print "-> FAIL" and exit with a non-zero status. If all tests pass, it will print "-> OK".

TEST_HELPERS_PATH=${TEST_HELPERS:-./build}
TEST_HELPERS=$(realpath "$TEST_HELPERS_PATH")
export TEST_HELPERS
echo "Using test helpers from: $TEST_HELPERS"

ALL_TESTS=$(find . -type f -name "*.test")
for test in $ALL_TESTS; do
  echo "===== Starting $test ====="
  $test -test.short -test.v || command_failed=1
done

if [ ${command_failed:-0} -eq 1 ]
then
  echo "-> FAIL"
  exit 1
fi
echo "-> OK"
