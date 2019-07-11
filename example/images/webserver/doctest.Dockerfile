FROM python:3.7.4-alpine3.9

WORKDIR /usr/src/app

COPY server.py ./
CMD [ "python", "-m", "doctest", "server.py" ]