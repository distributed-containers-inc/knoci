FROM debian:stretch

RUN { \
    echo "#!/usr/bin/env bash"; \
    echo "100"; \
} > /num_tests && chmod 755 /num_tests

RUN mkdir /tests
RUN for i in $(seq 1 50); do { \
    echo "#!/usr/bin/env bash"; \
    echo "echo 'Running test ${i}'"; \
    echo "echo 'This test takes ${i} seconds.'"; \
    echo "sleep ${i}"; \
    echo "echo 'Done!'"; \
} > /tests/"test_${i}.sh" && chmod 755 /tests/"test_${i}.sh"; \
done

ADD entrypoint.sh /entrypoint
RUN chmod 755 /entrypoint
ENTRYPOINT [ "/entrypoint" ]