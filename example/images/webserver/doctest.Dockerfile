FROM python:3.7.4-alpine3.9

WORKDIR /usr/src/app

RUN echo -e '#!/bin/sh\necho 2' > /num_tests && chmod 755 /num_tests

COPY server.py ./
CMD [ "python", "-m", "doctest", "server.py" ]