#!/usr/bin/env python3
"""Seed Gorse via the Go wrapper API's load-data endpoint."""

import sys
import time
import urllib.request
import urllib.error
import json
import os

BACKEND_URL = os.environ.get("BACKEND_URL", "http://localhost:3001")
ENDPOINT = f"{BACKEND_URL}/api/demo/load-data"

def wait_for_backend(max_tries=10, delay=3):
    health_url = f"{BACKEND_URL}/health"
    for i in range(max_tries):
        try:
            with urllib.request.urlopen(health_url, timeout=5) as resp:
                data = json.loads(resp.read())
                print(f"Backend healthy: {data}")
                return True
        except Exception as e:
            print(f"Waiting for backend ({i+1}/{max_tries})... {e}")
            time.sleep(delay)
    return False

def seed():
    print(f"Seeding data via {ENDPOINT}")
    req = urllib.request.Request(
        ENDPOINT,
        method="POST",
        headers={"Content-Type": "application/json"},
        data=b"{}"
    )
    try:
        with urllib.request.urlopen(req, timeout=60) as resp:
            result = json.loads(resp.read())
            print("Seed result:")
            print(json.dumps(result, indent=2))
            return True
    except urllib.error.HTTPError as e:
        body = e.read().decode()
        print(f"HTTP Error {e.code}: {body}")
        return False
    except Exception as e:
        print(f"Error: {e}")
        return False

if __name__ == "__main__":
    if not wait_for_backend():
        print("Backend not available. Start the backend first.")
        sys.exit(1)
    success = seed()
    sys.exit(0 if success else 1)
