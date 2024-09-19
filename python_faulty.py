import re
import requests
import logging


def process_log_and_post(file_path, api_url):
    pattern = re.compile(r'\[(.*)\] ERROR (.*)')
    try:
        with open(file_path, mode='r') as file, requests.session() as session:
            for line in file:
                error_match = re.search(pattern, line)

                if error_match is None:
                    continue

                timestamp, error_message = error_match.groups()
                logging.info(f"Error at {timestamp}: {error_message}")

                # Preparing payload
                data = {
                    'timestamp': timestamp,
                    'message': error_message
                }

                # Sending data to API
                response = session.post(api_url, json=data)
                status = response.status_code
                if 200 <= status < 300:
                    logging.info(f"POST response: {response.text}")
                else:
                    logging.error(f"An error occured: '{response.reason}' ({status})")

    except FileNotFoundError as e:
        logging.error("File not found. Please check the file path.", e)
    except Exception as e:
        logging.error("An error occurred:", e)

# Example usage
log_file = 'server.log'
api_url = 'https://jsonplaceholder.typicode.com/posts'
process_log_and_post(log_file, api_url)
