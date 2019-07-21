FROM python:3.7.4-stretch

WORKDIR /usr/src/app

COPY server.py ./
CMD [ "python", "-m", "doctest", "server.py" ]