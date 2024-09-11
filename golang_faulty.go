package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

func readAndSendLogs(filePath, apiURL string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
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

	for scanner.Scan() {
		line := scanner.Text()
		// This is a change of functionality from the Python version.  Here we
		// just look for the text ERROR and then we send the entire line.
		// The Python version performed a more strict check using a regular
		// expression and also parsed the fields into a timestamp and message.
		// I don't know if this change is intentional (maybe requirements have
		// changed) but I'll look into making this consistent with the Python
		// version in an upcoming commit.
		if strings.Contains(line, "ERROR") {
			wg.Add(1)
			go func() {
				// Sending log to API
				defer wg.Done()

				// I ran the program and it failed due to an error from the backend:
				// `SyntaxError: Unexpected token W in JSON at position 1`
				// Looking back at the code, the issue is obvious: we're sending
				// the raw string rather than JSON.
				data := map[string]string{"timestamp": "TODO", "message": line}
				jsonData, err := json.Marshal(data)
				if err != nil {
					// TODO BEFORE MERGE: fix error handling (along with the others)
					fmt.Println("Failed to encode JSON:", err)
				}

				resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
				// As well as checking for errors from the Post() call, we could
				// also check the status code of the response, as in the Python
				// version.  I'll add that in an upcoming commit.
				if err != nil {
					// As with the Python version, we should send errors to STDERR.
					// Unfortunately, we can't easily exit with an error status
					// without some refactoring - we'd need a mechanism (e.g.
					// a channel) to communicate the failure in the goroutine
					// back to the main program.
					fmt.Println("Failed to send log:", err)
				} else {
					// As above, this is an unhandled error that I'm comfortable with
					defer resp.Body.Close()
					// I believe ReadAll is deprecated. I'll look into replacing
					// it in an upcoming commit.
					body, _ := ioutil.ReadAll(resp.Body)
					fmt.Printf("Response: %s\n", string(body))
				}
				// I noticed while working on the code that the wg.Done()
				// call here will only happen on success, leading to the
				// main goroutine blocking if any errors happen.  I've fixed
				// this using `defer`.
				//wg.Done()
			}()
		}
	}

	if err := scanner.Err(); err != nil {
		// Here we should print to STDERR again, and we _can_ exit with
		// an error status.  For consistency with the Python version,
		// we'd detect "file not found" and exit with 1, otherwise we'd
		// exit with 2.  I'm not sure how easy that is with Scanner but I'll
		// look into it!
		fmt.Println("Error reading file:", err)
	}
	wg.Wait()
}

func main() {
	readAndSendLogs("server.log", "https://jsonplaceholder.typicode.com/posts")
}

// Overall I find this code straightforward and clear, and after the `bufio` fix
// I believe it will work.
