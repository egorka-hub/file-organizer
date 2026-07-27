package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	statistics     map[string]*FileStats
	totalSize      int64
}

type FileStats struct {
	processedFiles int
	totalSize      int64
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
		sourceDir:  sourceDir,
		rulesMap:   DefaultRules,
		statistics: make(map[string]*FileStats),
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

func (fo *FileOrganizer) moveFile(sourcePath, targetDir string, info os.FileInfo) error {
	fullPath := filepath.Join(fo.sourceDir, targetDir)

	err := os.MkdirAll(fullPath, 0755)
	if err != nil {
		fo.logError(fmt.Sprintf("cannot create directory %s: %v", fullPath, err))
		return fmt.Errorf("cannot create directory: %w", err)
	}
	fo.logSuccess(fmt.Sprintf("Created directory: %s", fullPath))

	name := filepath.Base(sourcePath)

	dstPath := filepath.Join(fullPath, name)

	if _, err := os.Stat(dstPath); err == nil {
		ext := filepath.Ext(name)
		base := strings.TrimSuffix(name, ext)
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		name = fmt.Sprintf("%s_%s%s", base, timestamp, ext)
		dstPath = filepath.Join(fullPath, name)
	}

	err = os.Rename(sourcePath, dstPath)
	if err != nil {
		fo.logError(fmt.Sprintf("cannot move file %s to %s: %v", sourcePath, dstPath, err))
		return fmt.Errorf("cannot move file: %w", err)
	}
	fo.logSuccess(fmt.Sprintf("Moved file: %s -> %s", sourcePath, dstPath))

	fo.totalSize += info.Size()

	if _, ok := fo.statistics[targetDir]; !ok {
		fo.statistics[targetDir] = &FileStats{}
	}

	fo.statistics[targetDir].processedFiles++
	fo.statistics[targetDir].totalSize += info.Size()

	return nil
}

func (fo *FileOrganizer) Organize() error {
	if err := fo.initLog(); err != nil {
		return fmt.Errorf("cannot init log: %w", err)
	}
	err := filepath.WalkDir(fo.sourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Dir(path) != fo.sourceDir {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		category, ok := fo.rulesMap[ext]
		if !ok {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		if err := fo.moveFile(path, category, info); err != nil {
			return err
		}

		fo.processedFiles++

		return nil

	})

	return err
}

func (fs *FileStats) String() string {
	return fmt.Sprintf("Файлов: %d, Размер: %.2f KB", fs.processedFiles, float64(fs.totalSize)/1024)
}

func (fo *FileOrganizer) generateReport() string {
	var report strings.Builder
	report.WriteString("=== Отчёт о перемещении файлов ===\n\n")
	report.WriteString(fmt.Sprintf("Всего обработано файлов: %d\n", fo.processedFiles))
	report.WriteString(fmt.Sprintf("Общий размер: %.2f KB\n\n", float64(fo.totalSize)/1024))
	report.WriteString("Статистика по категориям:\n\n")
	for category, stats := range fo.statistics {
		report.WriteString(fmt.Sprintf("%s:\n  %s\n\n", category, stats))
	}
	return report.String()
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
