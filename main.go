package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)
import _ "embed"

//go:embed index.html
var html string

//go:embed config.json
var config string

type configType struct {
	OutDir              string `json:"outDir"`
	ConvertAudioFormat  string `json:"convertAudioFormat"`
	ConvertAudioQuality string `json:"convertAudioQuality"`
	ConvertVideoFormat  string `json:"convertVideoFormat"`
	NicoVideoUserName   string `json:"nicoVideoUserName"`
	NicoVideoPassword   string `json:"nicoVideoPassword"`
}

type queue struct {
	SoundUrl       string `json:"soundUrl"`
	ForceAudioOnly bool   `json:"forceAudioOnly"`
	CreatedAt      string `json:"createdAt"`
}

type addQueueBody struct {
	SoundUrl       string `json:"soundUrl"`
	ForceAudioOnly bool   `json:"forceAudioOnly"`
}

type getQueuesBody struct {
	Queues []queue `json:"queues"`
}

var operationQueue []queue
var isOperationRunning bool = false
var parsedConfig configType

func main() {
	parseConfig()
	http.HandleFunc("/", htmlHandler)
	http.HandleFunc("/addQueue", addQueueHandler)
	http.HandleFunc("/getQueues", getQueuesHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func htmlHandler(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprintf(w, html)
	if err != nil {
		fmt.Printf("error htmlHandler: %v\n", err)
	}
}

func addQueueHandler(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("error read all: %v\n", err)
	}

	var requestBody addQueueBody
	err = json.Unmarshal(bodyBytes, &requestBody)
	if err != nil {
		fmt.Printf("error json unmarshal: %v\n", err)
	}

	operationQueue = append(operationQueue, queue{
		SoundUrl:       requestBody.SoundUrl,
		ForceAudioOnly: requestBody.ForceAudioOnly,
		CreatedAt:      time.Now().Format("2006-01-02 15:04:05"),
	})

	getQueuesHandler(w, r)

	if !isOperationRunning {
		go tryQueuePopAndDownload()
	}
}

func getQueuesHandler(w http.ResponseWriter, r *http.Request) {
	var responseBody getQueuesBody
	responseBody.Queues = operationQueue

	responseBytes, err := json.Marshal(responseBody)
	if err != nil {
		fmt.Printf("error json marshal: %v\n", err)
	}

	_, err = fmt.Fprintf(w, string(responseBytes))
	if err != nil {
		fmt.Printf("error fmt Fprintf: %v\n", err)
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

func downloadTarget(input queue) {
	todayString := time.Now().Format("2006-01-02")
	outDir := parsedConfig.OutDir + "/" + todayString

	err := os.MkdirAll(outDir, os.ModePerm)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	// yt-dlp default template
	fileNameTemplate := "%(title).50s [%(id)s].%(ext)s"
	fileOutputArg := outDir + "/" + fileNameTemplate

	var args []string

	args = append(args, "-o", fileOutputArg)
	if input.ForceAudioOnly {
		args = append(args, getAudioOptionParameters()...)
	}

	// TODO: 直近はこれで問題なさそうではあるけど暫定対応で、もし他のサイトで音声取ってきたくなったら書き直す必要が出てきてしまう
	//   代替案として、mp3>mp3/mp4というフォーマットを設定に書くことで、mp3の場合にmp4に詰め直すような処理を回避させることは可能かもしれない
	//   とはいえ実際に存在しうる音声ファイルのフォーマットを網羅するのは非現実的でもあるので悩ましい
	if !strings.Contains(input.SoundUrl, "soundcloud.com") && !input.ForceAudioOnly {
		args = append(args, getVideoOptionParameters()...)
	}

	if strings.Contains(input.SoundUrl, "nicovideo.jp") {
		args = append(args, getNicoVideoParameters()...)
	}

	args = append(args, input.SoundUrl)

	cmd := exec.Command("yt-dlp", args...)
	err = cmd.Run()

	if err != nil {
		fmt.Printf("error download: %v\n", err)
		return
	}

	fmt.Printf("finished download target\n")
}

func parseConfig() {
	err := json.Unmarshal([]byte(config), &parsedConfig)
	if err != nil {
		panic(err)
	}
}

func getVideoOptionParameters() []string {
	return []string{"--recode-video", parsedConfig.ConvertVideoFormat}
}

func getAudioOptionParameters() []string {
	return []string{"-x", "--audio-format", parsedConfig.ConvertAudioFormat, "--audio-quality", parsedConfig.ConvertAudioQuality}
}

func getNicoVideoParameters() []string {
	return []string{"--username", parsedConfig.NicoVideoUserName, "--password", parsedConfig.NicoVideoPassword}
}
