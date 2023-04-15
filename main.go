/**
* This Golang code used for :
* 1. Show the realtime camera that detect the human face based on `haarcascade_frontalface_default` classification
* 2. Realtime activate the PC speaker, that could speak based on defined `text` on `func Speak(text string)`
*
* Author : Angga Bayu Sejati <anggabs86@gmail.com>
* This project uses some libraries :
* 1. Face detection use go-cv, please see https://gocv.io/ or https://github.com/hybridgroup/gocv
* 2. Speaker use htgo-tts (https://github.com/hegedustibor/htgo-tts)
 */
package main

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	htgotts "github.com/hegedustibor/htgo-tts"
	"gocv.io/x/gocv"
)

//var for sync.WaitGroup
var wg sync.WaitGroup

const (
	//device id `0` mean the source came from default PC webcam
	DeviceID = 0

	//XML file that initialize face
	XmlFile = "haarcascade_frontalface_default.xml"

	//Window Frame title
	WindowFrameTitle = "Face Detection"

	//Text that sign that human face
	TextSignHuman = "Human"

	//Log file detector
	LogFileName = "face_count.log"

	//directory for saving the sound
	AudioDir = "audio"

	//Audio language from gtss
	AudioLanguage = "en"
)

//This func used for open and showing the camera
//The the camera will draw a rectangle around each face on
//the original image, with text identifying as "Human"
func CameraInitialization() {
	//open webcam
	//Read source of webcam
	//you can change into URL like this :
	//webcam, err := gocv.VideoCaptureFile("http://192.168.43.100:4747/video")
	//The URL is address the camera, could be from cctv or Droidcam
	webcam, err := gocv.VideoCaptureDevice(DeviceID)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer webcam.Close()

	//Open the Window Frame
	window := gocv.NewWindow(WindowFrameTitle)
	defer window.Close()

	//prepare image matrix
	img := gocv.NewMat()
	defer img.Close()

	//color for the rect when faces detected
	rectColor := color.RGBA{0, 0, 255, 0}

	//load classifier to recognize faces104 062 9102
	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	if !classifier.Load(XmlFile) {
		fmt.Printf("Error reading cascade file: %v\n", XmlFile)
		return
	}

	fmt.Printf("Camera : %v\n initialized", DeviceID)
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Cannot read device %d\n", DeviceID)
			return
		}

		if img.Empty() {
			continue
		}

		//face detection
		rects := classifier.DetectMultiScale(img)

		//Get face count and then save into file
		faceCountDetected := len(rects)
		go SaveLog(faceCountDetected)

		//draw a rectangle around each face on the original image, with text identifying as "Human"
		for _, r := range rects {
			gocv.Rectangle(&img, r, rectColor, 3)
			size := gocv.GetTextSize(TextSignHuman, gocv.FontHersheyPlain, 1.2, 2)
			pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
			gocv.PutText(&img, TextSignHuman, pt, gocv.FontHersheyPlain, 1.2, rectColor, 2)
		}

		//show the image in the window
		window.IMShow(img)
		if window.WaitKey(1) >= 0 {
			break
		}
	}

	//call sync.WaitGroup.Done() func
	wg.Done()
}

//Save the faceCountDetected to `LogFileName`
//parameter faceCountDetected int
func SaveLog(faceCountDetected int) {
	text := "0"
	if faceCountDetected > 0 {
		text = strconv.Itoa(faceCountDetected)
	}

	f, err := os.OpenFile(LogFileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	time.Sleep(500 * time.Millisecond)
	if _, err := f.WriteString(text); err != nil {
		log.Println(err)
	}
}

//Initialize the speaker
//The speaker will speak `n` detected face count
func InitializedSpeak() {
	for {
		//sleep 1000 miliseconds (1 second)
		time.Sleep(1000 * time.Millisecond)

		content, err := ioutil.ReadFile(LogFileName)
		if err != nil {
			log.Fatal(err)
		}

		contentTxt := string(content)
		if false == strings.EqualFold(contentTxt, "0") {
			Speak(fmt.Sprintf("Human face %s count detected", contentTxt))
		}
	}
}

//Func for activate speaker to speak `text`
//param `text` string
func Speak(text string) {
	speech := htgotts.Speech{Folder: AudioDir, Language: AudioLanguage}
	speech.Speak(text)
}

//func MAIN
func main() {
	//Add 2 sync.WaitGroup
	wg.Add(2)

	//initialize goroutine
	go CameraInitialization()
	go InitializedSpeak()

	//wait the sync.WaitGroup
	wg.Wait()
}
