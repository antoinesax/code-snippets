import os
import re
import requests


def process_log_and_post(file_path, api_url):
    try:
        file = open(file_path, "r")
        for line in file:
            if "ERROR" in line:
                error_match = re.search(r"\[(.*)\] ERROR (.*)", line)
                if error_match:
                    timestamp, error_message = error_match.groups()
                    print(f"Error at {timestamp}: {error_message}")

                    # Preparing payload
                    data = {"timestamp": timestamp, "message": error_message}

                    # Sending data to API
                    response = requests.post(api_url, json=data)
                    print(f"POST response: {response.text}")
        file.close()
    except FileNotFoundError:
        print("File not found. Please check the file path.")
    except Exception as e:
        print("An error occurred:", e)


# Example usage
log_file = "server.log"
api_url = "https://jsonplaceholder.typicode.com/posts"
process_log_and_post(log_file, api_url)
