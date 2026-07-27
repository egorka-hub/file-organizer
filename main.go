package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"maps"
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

	rulesMap := make(map[string]string, len(DefaultRules))
	maps.Copy(rulesMap, DefaultRules)

	return &FileOrganizer{
		sourceDir:  sourceDir,
		rulesMap:   rulesMap,
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

func (fo *FileOrganizer) logSuccess(format string, args ...any) {
	log.Printf("[SUCCESS] "+format, args...)
}

func (fo *FileOrganizer) logError(format string, args ...any) {
	log.Printf("[ERROR] "+format, args...)
}

func (fo *FileOrganizer) Close() error {
	if fo.logFile == nil {
		return nil
	}
	return fo.logFile.Close()
}

func (fo *FileOrganizer) moveFile(sourcePath, targetDir string, info os.FileInfo) error {
	fullPath := filepath.Join(fo.sourceDir, targetDir)

	dirExisted := true
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		dirExisted = false
	}

	if err := os.MkdirAll(fullPath, 0755); err != nil {
		fo.logError("cannot create directory %s: %v", fullPath, err)
		return fmt.Errorf("cannot create directory: %w", err)
	}
	if !dirExisted {
		fo.logSuccess("Created directory: %s", fullPath)
	}

	name := filepath.Base(sourcePath)

	dstPath := filepath.Join(fullPath, name)

	if _, err := os.Stat(dstPath); err == nil {
		ext := filepath.Ext(name)
		base := strings.TrimSuffix(name, ext)
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		name = fmt.Sprintf("%s_%s%s", base, timestamp, ext)
		dstPath = filepath.Join(fullPath, name)
	}

	err := os.Rename(sourcePath, dstPath)
	if err != nil {
		fo.logError("cannot move file %s to %s: %v", sourcePath, dstPath, err)
		return fmt.Errorf("cannot move file: %w", err)
	}
	fo.logSuccess("Moved file: %s -> %s", sourcePath, dstPath)

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

func (stats *FileStats) String() string {
	return fmt.Sprintf("Файлов: %d, Размер: %.2f KB", stats.processedFiles, float64(stats.totalSize)/1024)
}

func (fo *FileOrganizer) generateReport() string {
	var report strings.Builder
	report.WriteString("=== Отчёт о перемещении файлов ===\n\n")
	_, _ = fmt.Fprintf(&report, "Всего обработано файлов: %d\n", fo.processedFiles)
	_, _ = fmt.Fprintf(&report, "Общий размер: %.2f KB\n\n", float64(fo.totalSize)/1024)
	report.WriteString("Статистика по категориям:\n\n")
	for category, stats := range fo.statistics {
		_, _ = fmt.Fprintf(&report, "%s:\n  %s\n\n", category, stats)
	}
	return report.String()
}

func main() {
	fmt.Println("=== File Organizer ===")

	fmt.Print("Введите путь к директории для организации (Enter для текущей директории): ")

	scanner := bufio.NewScanner(os.Stdin)

	var sourcePath string
	if scanner.Scan() {
		sourcePath = strings.TrimSpace(scanner.Text())
	} else if err := scanner.Err(); err != nil {
		fmt.Println("cannot read input:", err)
		os.Exit(1)
	}

	if sourcePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println("cannot get current directory:", err)
			os.Exit(1)
		}
		sourcePath = cwd
	}

	organizer, err := NewFileOrganizer(sourcePath)
	if err != nil {
		fmt.Println("cannot create organizer:", err)
		os.Exit(1)
	}

	defer organizer.Close()

	if err := organizer.Organize(); err != nil {
		fmt.Println("cannot organize files:", err)
		os.Exit(1)
	}

	fmt.Println(organizer.generateReport())

	fmt.Println("Организация завершена! Подробности в файле organizer.log")
}
