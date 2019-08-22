package format

import (
	"fmt"
	"strings"
	"time"
)

func DurationHumanReadable(d time.Duration) string {
	var parts []string
	if d.Hours() > 24 {
		days := int(d.Hours() / 24)
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if (int(d.Hours()) % 24) >= 1.0 {
		hours := int(d.Hours()) % 24
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if int(d.Minutes())%60 > 0 {
		minutes := int(d.Minutes()) % 60
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if d.Seconds() > 0 {
		seconds := int(d.Seconds()) % 60
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}

	// Take only the two most relevant parts
	upperBound := 2
	if len(parts) < upperBound {
		upperBound = len(parts)
	}

	parts = parts[:upperBound]
	return fmt.Sprintf(strings.Join(parts, ", "))
}
