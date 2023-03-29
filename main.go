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
}

type addQueueBody struct {
	SoundUrl       string `json:"soundUrl"`
	ForceAudioOnly bool   `json:"forceAudioOnly"`
}

var operationQueue []addQueueBody
var isOperationRunning bool = false
var parsedConfig configType

func main() {
	parseConfig()
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

	operationQueue = append(operationQueue, requestBody)

	if !isOperationRunning {
		go tryQueuePopAndDownload()
	}

	// TODO: Queueの状態を返せるといいかも？
	//  と思ったけどQueueを返すエンドポイントくらいあっていいのでは
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
	today := time.Now()
	todayString := today.Format("2006-01-02")
	outDir := parsedConfig.OutDir + "/" + todayString

	err := os.Mkdir(outDir, os.ModePerm)
	if err != nil {
		return
	}

	// yt-dlp default template
	fileNameTemplate := "%(title).50s [%(id)s].%(ext)s"
	fileOutputArg := outDir + "/" + fileNameTemplate

	var args []string

	args = append(args, "-o", fileOutputArg)
	if input.ForceAudioOnly {
		args = append(args, "-x")
		args = append(
			args,
			"--audio-format",
			parsedConfig.ConvertAudioFormat,
		)
		args = append(
			args,
			"--audio-quality",
			parsedConfig.ConvertAudioQuality,
		)
	}

	// TODO: 直近はこれで問題なさそうではあるけど暫定対応で、もし他のサイトで音声取ってきたくなったら書き直す必要が出てきてしまう
	//   代替案として、mp3>mp3/mp4というフォーマットを設定に書くことで、mp3の場合にmp4に詰め直すような処理を回避させることは可能かもしれない
	//   とはいえ実際に存在しうる音声ファイルのフォーマットを網羅するのは非現実的でもあるので悩ましい
	if !strings.Contains(input.SoundUrl, "soundcloud.com") {
		args = append(
			args,
			"--recode-video",
			parsedConfig.ConvertVideoFormat,
		)
	}
	args = append(args, input.SoundUrl)

	cmd := exec.Command("yt-dlp", args...)
	err = cmd.Run()

	if err != nil {
		fmt.Printf("error download: %v\n", err)
	}

	fmt.Printf("finished download target\n")
}

func parseConfig() {
	err := json.Unmarshal([]byte(config), &parsedConfig)
	if err != nil {
		panic(err)
	}
}
