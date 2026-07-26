package main

import (
	"fmt"
	"log"
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

func (fo *FileOrganizer) initLog() error {
	file, err := os.OpenFile("./organizer.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("cannot open log file: %w", err)
	}

	fo.logFile = file
	log.SetOutput(file)

	return nil
}

func (fo *FileOrganizer) logSuccess(message string) {
	log.Printf("[SUCCESS] %s", message)
}

func (fo *FileOrganizer) logError(message string) {
	log.Printf("[ERROR] %s", message)
}

func (fo *FileOrganizer) Close() error {
	if fo.logFile == nil {
		return nil
	}
	return fo.logFile.Close()
}

func main() {
	organizer, err := NewFileOrganizer("./test_files")
	if err != nil {
		fmt.Println("cannot create organizer:", err)
		os.Exit(1)
	}

	if err := organizer.initLog(); err != nil {
		fmt.Println("cannot init log:", err)
		os.Exit(1)
	}
	defer organizer.Close()

	fmt.Printf("FileOrganizer создан для директории: %s\n", organizer.sourceDir)
}
