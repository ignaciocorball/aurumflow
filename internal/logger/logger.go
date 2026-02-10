package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/term"
)

const (
	ansiReset  = "\033[0m"
	ansiDim    = "\033[2m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiRed    = "\033[31m"
)

var (
	jsonOutput bool
	useColor   bool
	logFile    *os.File
	fileMu     sync.Mutex
)

func init() {
	jsonOutput = os.Getenv("AURUMFLOW_LOG_JSON") == "1" || os.Getenv("AURUMFLOW_LOG_JSON") == "true"
	// Color: only when not JSON, and (TTY or AURUMFLOW_LOG_COLOR=1), and not AURUMFLOW_LOG_COLOR=0
	if jsonOutput {
		useColor = false
	} else if v := os.Getenv("AURUMFLOW_LOG_COLOR"); v == "0" || v == "false" {
		useColor = false
	} else if v == "1" || v == "true" {
		useColor = true
	} else {
		useColor = term.IsTerminal(int(os.Stdout.Fd()))
	}
}

type logEntry struct {
	Time  string `json:"time"`
	Level string `json:"level"`
	Msg   string `json:"msg"`
}

func writeToFile(level, msg string) {
	if logFile == nil {
		return
	}
	fileMu.Lock()
	defer fileMu.Unlock()
	ts := time.Now().UTC().Format(time.RFC3339)
	_, _ = fmt.Fprintf(logFile, "%s [%s] %s\n", ts, level, msg)
}

func consoleInfo(msg string) {
	if jsonOutput {
		b, _ := json.Marshal(logEntry{Time: time.Now().UTC().Format(time.RFC3339), Level: "info", Msg: msg})
		log.Println(string(b))
	} else if useColor {
		log.Printf("%s[AurumFlow] [INFO] %s%s", ansiDim, msg, ansiReset)
	} else {
		log.Printf("[AurumFlow] [INFO] %s", msg)
	}
}

func consoleWarn(msg string) {
	if jsonOutput {
		b, _ := json.Marshal(logEntry{Time: time.Now().UTC().Format(time.RFC3339), Level: "warn", Msg: msg})
		log.Println(string(b))
	} else if useColor {
		log.Printf("%s[AurumFlow] [WARN] %s%s", ansiYellow, msg, ansiReset)
	} else {
		log.Printf("[AurumFlow] [WARN] %s", msg)
	}
}

func consoleError(msg string) {
	if jsonOutput {
		b, _ := json.Marshal(logEntry{Time: time.Now().UTC().Format(time.RFC3339), Level: "error", Msg: msg})
		log.Println(string(b))
	} else if useColor {
		log.Printf("%s[AurumFlow] [ERROR] %s%s", ansiRed, msg, ansiReset)
	} else {
		log.Printf("[AurumFlow] [ERROR] %s", msg)
	}
}

func consoleSuccess(msg string) {
	if jsonOutput {
		b, _ := json.Marshal(logEntry{Time: time.Now().UTC().Format(time.RFC3339), Level: "info", Msg: msg})
		log.Println(string(b))
	} else if useColor {
		log.Printf("%s[AurumFlow] [INFO] %s%s", ansiGreen, msg, ansiReset)
	} else {
		log.Printf("[AurumFlow] [INFO] %s", msg)
	}
}

// InitFile creates logDir if needed and opens a new log file for this run.
// Filename: logDir/aurumflow_YYYY-MM-DD_THH-MM-SS.log (UTC).
// If logDir is empty, file logging is disabled.
func InitFile(logDir string) error {
	if logDir == "" {
		return nil
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}
	name := fmt.Sprintf("aurumflow_%s.log", time.Now().UTC().Format("2006-01-02_T15-04-05"))
	path := filepath.Join(logDir, name)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	fileMu.Lock()
	logFile = f
	fileMu.Unlock()
	return nil
}

// CloseFile closes the log file if open (call from main defer).
func CloseFile() {
	fileMu.Lock()
	defer fileMu.Unlock()
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
}

// Info logs an info message (dim on console when color enabled).
func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	writeToFile("INFO", msg)
	consoleInfo(msg)
}

// Success logs a success message (green on console when color enabled). Written as INFO in file.
func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	writeToFile("INFO", msg)
	consoleSuccess(msg)
}

// Warn logs a warning message (yellow on console when color enabled).
func Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	writeToFile("WARN", msg)
	consoleWarn(msg)
}

// Error logs an error message (red on console when color enabled).
func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	writeToFile("ERROR", msg)
	consoleError(msg)
}
