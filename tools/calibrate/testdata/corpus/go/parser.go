package ledger

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ParseEntries reads account,amount pairs while ignoring comments.
func ParseEntries(source io.Reader) ([]Entry, error) {
	scanner := bufio.NewScanner(source)
	entries := make([]Entry, 0)
	for line := 1; scanner.Scan(); line++ {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		account, amountText, ok := strings.Cut(text, ",")
		if !ok {
			return nil, fmt.Errorf("line %d: expected account,amount", line)
		}
		amount, err := strconv.ParseInt(strings.TrimSpace(amountText), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: parse amount: %w", line, err)
		}
		entries = append(entries, Entry{Account: strings.TrimSpace(account), Amount: amount})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan entries: %w", err)
	}
	return entries, nil
}
