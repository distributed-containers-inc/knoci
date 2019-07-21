FROM python:3.7.4-alpine3.9

WORKDIR /usr/src/app

COPY server.py ./
RUN chmod 755 server.py
CMD [ "./server.py" ]