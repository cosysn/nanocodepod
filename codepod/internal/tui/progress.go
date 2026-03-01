package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("75"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
)

// PrintSuccess prints a success message
func PrintSuccess(msg string) {
	fmt.Println(successStyle.Render("✓ ") + msg)
}

// PrintError prints an error message
func PrintError(msg string) {
	fmt.Println(errorStyle.Render("✗ ") + msg)
}

// PrintWarning prints a warning message
func PrintWarning(msg string) {
	fmt.Println(warningStyle.Render("⚠ ") + msg)
}

// PrintInfo prints an info message
func PrintInfo(msg string) {
	fmt.Println(infoStyle.Render("ℹ ") + msg)
}

// PrintDim prints a dim message
func PrintDim(msg string) {
	fmt.Println(dimStyle.Render(msg))
}

// ProgressBar renders a simple progress bar
func ProgressBar(label string, current, total int) string {
	width := 40
	percent := float64(current) / float64(total)
	filled := int(float64(width) * percent)

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	return fmt.Sprintf("%s [%s] %d%%", label, bar, int(percent*100))
}

// Spinner returns a spinner frame
func Spinner(frame int) string {
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	return spinners[frame%len(spinners)]
}

// LoadingAnimation runs a loading animation
func LoadingAnimation(label string, done chan bool) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			fmt.Printf("\r%s %s... Done!\n", successStyle.Render("✓"), label)
			return
		case <-ticker.C:
			fmt.Printf("\r%s %s...", frames[i%len(frames)], label)
			i++
		}
	}
}

// Confirm asks for confirmation
func Confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	var input string
	fmt.Scanln(&input)
	return input == "y" || input == "Y"
}
