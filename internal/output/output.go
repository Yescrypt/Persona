// Package output writes a wordlist to a file and reports basic statistics.
package output

import (
	"bufio"
	"fmt"
	"os"
)

// Stats summarizes a written file.
type Stats struct {
	Count int
	Bytes int64
}

// Write writes one line per entry to path (newline-terminated).
func Write(path string, lines []string) (Stats, error) {
	f, err := os.Create(path)
	if err != nil {
		return Stats{}, err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	var n int64
	for _, l := range lines {
		m, err := w.WriteString(l + "\n")
		if err != nil {
			return Stats{}, err
		}
		n += int64(m)
	}
	if err := w.Flush(); err != nil {
		return Stats{}, err
	}
	return Stats{Count: len(lines), Bytes: n}, nil
}

// HumanSize renders a byte count as a short human-readable string.
func HumanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGT"[exp])
}
