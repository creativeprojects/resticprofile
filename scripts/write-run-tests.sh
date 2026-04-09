#!/bin/sh
OUTPUT=./scripts/run-tests.sh

echo "#!/bin/sh" >$OUTPUT
chmod +x $OUTPUT
echo "export TEST_HELPER=$TEST_HELPER" >>$OUTPUT
find . -type f -name "*.test" | xargs -I % echo "echo \"=== % ===\" && % -test.short -test.v || command_failed=1" >>$OUTPUT
echo "if [ \${command_failed:-0} -eq 1 ]" >>$OUTPUT
echo "then" >>$OUTPUT
echo "  echo \"-> FAIL\"" >>$OUTPUT
echo "  exit 1" >>$OUTPUT
echo "fi" >>$OUTPUT
echo "echo \"-> OK\"" >>$OUTPUT
