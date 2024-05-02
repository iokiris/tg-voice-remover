package main

import (
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type AudioProcessRequest struct {
	InputPath  string `json:"input_path"`
	OutputPath string `json:"output_path"`
}

func removeVocals(input []byte) ([]byte, error) {
	uid := uuid.New().String()
	inputFileName := filepath.Join("temp", uid+"_input.mp3")
	outputDir := filepath.Join("output", uid)
	if err := os.MkdirAll(filepath.Dir(inputFileName), 0755); err != nil {
		log.Println("error creating input directory:", err)
		return nil, err
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Println("error creating output directory:", err)
		return nil, err
	}
	err := os.WriteFile(inputFileName, input, 0644)
	if err != nil {
		log.Println("error writing file:", err)
		return nil, err
	}
	requestData := AudioProcessRequest{
		InputPath:  inputFileName,
		OutputPath: outputDir,
	}
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", "http://localhost:8000/process-audio/", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("Error creating request:", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Non-OK HTTP status: %s", resp.Status)
		return nil, err
	}

	processedAudioPath := filepath.Join(outputDir, uid+"_input", "accompaniment.mp3")
	if _, err := os.Stat(processedAudioPath); os.IsNotExist(err) {
		log.Println("processed file doesn't exists:", processedAudioPath)
		return nil, err
	}
	processedAudio, err := os.ReadFile(processedAudioPath)
	if err != nil {
		log.Println("processing audio error:", err)
		return nil, err
	}
	// remove files
	cleanup(inputFileName, outputDir)

	return processedAudio, nil
}

func cleanup(inputFileName, outputDir string) {
	if err := os.Remove(inputFileName); err != nil {
		log.Println("Failed to remove input file:", inputFileName, err)
	}
	if err := os.RemoveAll(outputDir); err != nil {
		log.Println("Failed to remove output directory:", outputDir, err)
	}
}
