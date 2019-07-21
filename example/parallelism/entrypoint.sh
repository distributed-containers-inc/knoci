#!/usr/bin/env bash

# knoci first creates a pod calling /num_tests
# this image returns 50 (see Dockerfile)
# then knoci will run individual tests with given (start..end) ranges where
# 1 <= start <= end <= (output of /num_tests)
# the test script is responsible for only running the specified tests
# (knoci might run this as in "docker run --rm parallelism:latest 1 20")
START=$1
END=$2

for i in $(seq ${START} ${END}); do
    /tests/test_${i}.sh
done