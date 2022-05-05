package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
)
import _ "embed"

//go:embed index.html
var html string

type addQueueBody struct {
	SoundUrl     string `json:"soundUrl"`
	IsForceSound bool   `json:"isForceSound"`
}

var operationQueue []addQueueBody
var isOperationRunning bool = false

func main() {
	http.HandleFunc("/", htmlHandler)
	http.HandleFunc("/addQueue", addQueueHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func htmlHandler(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprintf(w, html)
	if err != nil {
		fmt.Printf("error htmlHandler: %v", err)
	}
}

func addQueueHandler(w http.ResponseWriter, r *http.Request) {

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("error read all: %v", err)
	}

	var requestBody addQueueBody
	err = json.Unmarshal(bodyBytes, &requestBody)
	if err != nil {
		fmt.Printf("error json unmarshal: %v", err)
	}

	operationQueue = append(operationQueue, requestBody)

	if !isOperationRunning {
		tryQueuePopAndDownload()
	}
}

func tryQueuePopAndDownload() {
	if len(operationQueue) == 0 {
		isOperationRunning = false
		return
	}

	isOperationRunning = true
	poppedQueue := operationQueue[0]
	operationQueue = operationQueue[1:]
	downloadTarget(poppedQueue)
	tryQueuePopAndDownload()
}

func downloadTarget(input addQueueBody) {
	cmd := exec.Command("yt-dlp", input.SoundUrl)
	err := cmd.Run()

	if err != nil {
		fmt.Printf("error download: %v", err)
	}

	fmt.Printf("finished download target")
}
