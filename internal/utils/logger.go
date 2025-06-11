// internal/utils/logger.go
package utils

import (
	"fmt"
	"log"
)

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

// Colored logging functions
func LogInfo(message string, args ...interface{}) {
	log.Printf(ColorBlue+"[INFO]"+ColorReset+" "+message, args...)
}

func LogSuccess(message string, args ...interface{}) {
	log.Printf(ColorGreen+"[SUCCESS]"+ColorReset+" "+message, args...)
}

func LogWarning(message string, args ...interface{}) {
	log.Printf(ColorYellow+"[WARNING]"+ColorReset+" "+message, args...)
}

func LogError(message string, args ...interface{}) {
	log.Printf(ColorRed+"[ERROR]"+ColorReset+" "+message, args...)
}

func LogRequest(method, path, userID string) {
	log.Printf(ColorCyan+"[REQUEST]"+ColorReset+" %s %s | User: %s", method, path, userID)
}

func LogAuth(message string, args ...interface{}) {
	log.Printf(ColorPurple+"[AUTH]"+ColorReset+" "+message, args...)
}

func LogDB(message string, args ...interface{}) {
	log.Printf(ColorWhite+"[DB]"+ColorReset+" "+message, args...)
}

func PrintStartupBanner() {
	fmt.Printf("%s%s", ColorBold, ColorGreen)
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Println("║    🚀 User Management System 🚀     ║")
	fmt.Println("║           Starting Server           ║")
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Printf("%s", ColorReset)
}
