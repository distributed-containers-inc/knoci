#!/usr/bin/env python3

from http.server import BaseHTTPRequestHandler, HTTPServer


class Server(BaseHTTPRequestHandler):
  def response(self):
    """Return the response data for a given path as (text, status code)
    >>> MockServer('/').response()
    ('Index', 200)
    >>> MockServer('/a').response()
    ('404 Not Found', 404)
    """
    if self.path == '/':
      return "Index", 200
    return "404 Not Found", 404

  def do_GET(self):
    resp, status = self.response()
    resp = resp.encode('utf-8')

    self.send_response(status)
    self.send_header('Content-Type', 'text/plain')
    self.send_header('Content-Length', len(resp))
    self.end_headers()
    self.wfile.write(resp)


class MockServer(Server):
  def __init__(self, path):
    self.path = path


def main():
  s = HTTPServer(('0.0.0.0', 8080), Server)
  try:
    s.serve_forever()
  except KeyboardInterrupt:
    pass


if __name__ == '__main__':
  main()
