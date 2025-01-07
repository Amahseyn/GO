package main

import (
	"fmt"

	"gocv.io/x/gocv"
)

func main() {
	// Open the video file
	video, err := gocv.VideoCaptureFile("a09.mp4") // Replace with your video file path
	if err != nil {
		fmt.Println("Error opening video file:", err)
		return
	}
	defer video.Close()

	// Create a window to display the video
	window := gocv.NewWindow("Video Player")
	defer window.Close()

	// Create a Mat to hold frames
	img := gocv.NewMat()
	defer img.Close()

	for {
		if ok := video.Read(&img); !ok {
			fmt.Println("Reached end of video or cannot read frame.")
			break
		}

		if img.Empty() {
			continue
		}

		// Display the frame in the window
		window.IMShow(img)

		// Break the loop when 'q' is pressed
		if window.WaitKey(1) == 'q' {
			break
		}
	}
}
