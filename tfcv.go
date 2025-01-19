package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	videoPath := "a09.mp4"
	videoFile, err := os.Open(videoPath)
	if err != nil {
		fmt.Println("Error opening video file:", err)
		return
	}
	defer videoFile.Close()

	// Get file size for logging
	fileInfo, _ := videoFile.Stat()
	fmt.Printf("Video size: %d bytes\n", fileInfo.Size())

	// Create a request to send the video data
	req, err := http.NewRequest("POST", "http://127.0.0.1:5000/process_video_stream", videoFile)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", "video/mp4")

	// Send the video data
	client := &http.Client{Timeout: time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Server response status:", resp.Status)

	// Read the processed video stream from the server
	var receivedData bytes.Buffer
	_, err = io.Copy(&receivedData, resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

	fmt.Printf("Received processed video: %d bytes\n", receivedData.Len())

	// Save the received processed video to a file
	outputFile, err := os.Create("processed_video.mp4")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	_, err = outputFile.Write(receivedData.Bytes())
	if err != nil {
		fmt.Println("Error saving processed video:", err)
		return
	}

	fmt.Println("Processed video saved successfully.")
}
