package logger

import "log"

// ANSI color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

func Info(message string, args ...interface{}) {
	log.Printf(ColorBlue+"[INFO]"+ColorReset+" "+message, args...)
}

func Success(message string, args ...interface{}) {
	log.Printf(ColorGreen+"[SUCCESS]"+ColorReset+" "+message, args...)
}

func Warning(message string, args ...interface{}) {
	log.Printf(ColorYellow+"[WARNING]"+ColorReset+" "+message, args...)
}

func Error(message string, args ...interface{}) {
	log.Printf(ColorRed+"[ERROR]"+ColorReset+" "+message, args...)
}

func Request(method, path, userID string) {
	log.Printf(ColorCyan+"[REQUEST]"+ColorReset+" %s %s | User: %s", method, path, userID)
}

func Auth(message string, args ...interface{}) {
	log.Printf(ColorPurple+"[AUTH]"+ColorReset+" "+message, args...)
}

func DB(message string, args ...interface{}) {
	log.Printf(ColorWhite+"[DB]"+ColorReset+" "+message, args...)
}