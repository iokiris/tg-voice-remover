package main

import (
	"bytes"
	"github.com/google/uuid"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

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
	cmd := exec.Command("python", "process_audio.py", inputFileName, outputDir)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Printf("cmdrun failed: %s", stderr.String())
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
