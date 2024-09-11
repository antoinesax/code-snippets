package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sync"
)

func readAndSendLogs(filePath, apiURL string) {
	file, err := os.Open(filePath)
	if err != nil {
		// As in the Python version, we exit with status 1 for "file not
		// found" and 2 for a general error.
		if os.IsNotExist(err) { // Check if the error is a "file not found" error
			fmt.Fprintln(os.Stderr, "File not found. Please check the file path.")
			os.Exit(1)
		} else {
			fmt.Fprintln(os.Stderr, "An error occurred:", err)
			os.Exit(2)
		}
	}
	// This is an unhandled error, but I wouldn't change it because there's not
	// much we can do if Close() fails.
	defer file.Close()

	// Here's a major improvement from the Python version: we're using
	// goroutines and a WaitGroup (it's always nice to see WaitGroups)
	// to submit requests to the API in parallel.  This should produce a
	// nice performance improvement at minimal cost!
	var wg sync.WaitGroup
	// Using a Scanner is a good idea here since it offers a better interface
	// than Reader in this use case.
	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`\[(.*)] ERROR (.*)`)

	for scanner.Scan() {
		line := scanner.Text()
		// Originally we just looked for the text ERROR and sent the entire
		// line; this was a change of functionality from the Python version.
		// I've changed it to use a regular expression to perform a more strict
		// check and parse fields into a timestamp and message.

		match := re.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		wg.Add(1)
		go func() {
			// Sending log to API
			defer wg.Done()

			fmt.Printf("Error at %s: %s\n", match[1], match[2])

			// I ran the program and it failed due to an error from the backend:
			// `SyntaxError: Unexpected token W in JSON at position 1`
			// Looking back at the code, the issue is obvious: we're sending
			// the raw string rather than JSON.
			data := map[string]string{"timestamp": match[1], "message": match[2]}
			jsonData, err := json.Marshal(data)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Failed to encode JSON:", err)
				return
			}

			resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				// As with the Python version, we now send errors to STDERR.
				// Unfortunately, we can't easily exit with an error status
				// without some refactoring - we'd need a mechanism (e.g.
				// a channel) to communicate the failure in the goroutine
				// back to the main program.
				fmt.Fprintln(os.Stderr, "Failed to send log:", err)
				return
			}
			// As well as checking for errors from the Post() call, we now
			// check the status code of the response, as in the Python version.
			if resp.StatusCode != http.StatusCreated {
				fmt.Fprintln(os.Stderr, "Unexpected error from API:", resp.StatusCode)
				return
			}

			// As above, this is an unhandled error that I'm comfortable with
			defer resp.Body.Close()
			// ioutil.ReadAll is deprecated, so I've replaced it with the version in `io`.
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("Response: %s\n", string(body))

			// I noticed while working on the code that the wg.Done()
			// call here will only happen on success, leading to the
			// main goroutine blocking if any errors happen.  I've fixed
			// this using `defer`.
			//wg.Done()
		}()

	}

	if err := scanner.Err(); err != nil {
		// Here we print to STDERR for consistency with the Python version, and
		// we _can_ exit with an error status.  As in the Python version, we
		// exit with status 2 for a general error and status 1 (detected at the
		// start of this function) for "file not found".
		fmt.Fprintln(os.Stderr, "Error reading file:", err)
		os.Exit(2)
	}
	wg.Wait()
}

func main() {
	readAndSendLogs("server.log", "https://jsonplaceholder.typicode.com/posts")
}

// Overall I find this code straightforward and clear, and after the `bufio` fix
// I believe it will work.
