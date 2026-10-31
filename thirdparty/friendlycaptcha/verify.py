from http.server import HTTPServer, SimpleHTTPRequestHandler
import json
import requests


class Handler(SimpleHTTPRequestHandler):
    def do_POST(self):
        if self.path == "/verify":
            length = int(self.headers["Content-Length"])
            body = json.loads(self.rfile.read(length))

            res = requests.post("https://global.frcapi.com/api/v2/captcha/siteverify", json={
                "response": body["response"],
                "sitekey": "" # sitekey here
            }, headers={
                "X-API-Key": "" # api key here
            })

            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Access-Control-Allow-Origin", "*")
            self.end_headers()
            self.wfile.write(res.content)
        else:
            self.send_response(404)
            self.end_headers()

    def do_OPTIONS(self):
        self.send_response(200)
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "POST, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type")
        self.end_headers()


print("Serving on http://localhost:8000")
HTTPServer(("localhost", 8000), Handler).serve_forever()
