package workspace

import (
	"fmt"
	"strings"
	"unicode"
)

// sanitizeContainerName ensures the container name is valid for Docker
// Valid characters: [a-zA-Z0-9][a-zA-Z0-9_.-]
func sanitizeContainerName(name string) string {
	// If empty, use default
	if name == "" {
		return "workspace"
	}

	// Convert to lowercase
	result := strings.ToLower(name)

	// Replace invalid characters with underscore
	var cleaned []rune
	for _, r := range result {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.' || r == '-' {
			cleaned = append(cleaned, r)
		} else {
			cleaned = append(cleaned, '_')
		}
	}

	// Ensure first character is alphanumeric
	nameStr := string(cleaned)
	if len(nameStr) > 0 {
		first := rune(nameStr[0])
		if !unicode.IsLetter(first) && !unicode.IsDigit(first) {
			nameStr = "w" + nameStr[1:]
		}
	}

	// Limit length
	if len(nameStr) > 128 {
		nameStr = nameStr[:128]
	}

	return nameStr
}

// GetContainerName returns a valid Docker container name
func GetContainerName(workspaceName string) string {
	return fmt.Sprintf("codepod-%s", sanitizeContainerName(workspaceName))
}
