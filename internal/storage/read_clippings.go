package storage

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	separator = "=========="
)

func ReadRawClippings(inputFile string) ([]string, error) {
	fd, err := os.Open(inputFile)
	if err != nil {
		return nil, fmt.Errorf("open input file: %w", err)
	}
	defer fd.Close()

	br := bufio.NewReader(fd)
	r, _, err := br.ReadRune()
	if err != nil {
		return nil, fmt.Errorf("read rune: %w", err)
	}

	if r != '\uFEFF' {
		err = br.UnreadRune() // Not a BOM -- put the rune back
		if err != nil {
			return nil, fmt.Errorf("unread rune: %w", err)
		}
	}

	scanner := bufio.NewScanner(br)
	var content strings.Builder
	for scanner.Scan() {
		content.WriteString(scanner.Text())
		content.WriteString("\n")
	}

	contentStr := content.String()

	entries := strings.Split(contentStr, separator)

	return entries, nil
}
