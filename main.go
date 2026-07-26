package main

import (
	"fmt"
	"os"
)

var DefaultRules = map[string]string{
	".jpg":  "Images",
	".jpeg": "Images",
	".png":  "Images",
	".pdf":  "Documents",
	".doc":  "Documents",
	".docx": "Documents",
	".txt":  "Documents",
	".mp3":  "Music",
	".wav":  "Music",
	".mp4":  "Video",
	".avi":  "Video",
	".zip":  "Archives",
	".rar":  "Archives",
}

type FileOrganizer struct {
	sourceDir      string
	rulesMap       map[string]string
	processedFiles int
	logFile        *os.File
}

func NewFileOrganizer(sourceDir string) (*FileOrganizer, error) {
	if sourceDir == "" {
		return nil, fmt.Errorf("sourceDir cannot be empty")
	}

	info, err := os.Stat(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("cannot stat sourceDir: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("sourceDir must be a directory")
	}

	return &FileOrganizer{
		sourceDir: sourceDir,
		rulesMap:  DefaultRules,
	}, nil

}

func main() {
	organizer, err := NewFileOrganizer("./test_files")
	if err != nil {
		fmt.Println("cannot create organizer:", err)
		os.Exit(1)
	}
	fmt.Printf("FileOrganizer создан для директории: %s\n", organizer.sourceDir)
}
