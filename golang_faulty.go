package main

import (
	"bufio"
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
    defer file.Close()

    var wg sync.WaitGroup
    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        line := scanner.Text()
        if strings.Contains(line, "ERROR") {
            wg.Add(1)
						go func(l string) {
                defer wg.Done()

                // Sending log to API
								// If fatal, the whole process exit
                resp, err := http.Post(apiURL, "plain/text", strings.NewReader(l))
                if err != nil {
                    fmt.Println("Failed to send log:", err)
                } else {
                    defer resp.Body.Close()
                    body, _ := ioutil.ReadAll(resp.Body)
                    fmt.Printf("Response: %s\n", string(body))
                }
					}(line)
        }
    }

    if err := scanner.Err(); err != nil {
        fmt.Println("Error reading file:", err)
    }
    wg.Wait()
}

func main() {
    readAndSendLogs("server.log", "https://jsonplaceholder.typicode.com/posts")
}

