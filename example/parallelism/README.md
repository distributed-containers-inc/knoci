# Parallelism

Knoci allows test images to automatically parallelize tests by implementing two features:
1. Running `docker run --rm (test image) --entrypoint /num_tests` should return an integer (as in a print statement)
2. If #1 has been implemented, running `docker run --rm (test image) 5 10` should run tests 5, 6, 7, 8, 9, 10 exclusively (and no others)

## This image

This image implements a dummy set of tests, numbered 1-50, where the first waits one second before succeeding, the second waits two seconds, and so on.
