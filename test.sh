#!/bin/sh

cd "${0%/*}"

go test -v $(go list ./...) | \
    sed ''/PASS/s//$(printf "\033[32mPASS\033[0m")/'' | \
    sed ''/FAIL/s//$(printf "\033[31mFAIL\033[0m")/'' | \
    grep -v 'no test files'
