package output

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nsr888/kindle-highlights-to-obsidian/pkg/hashs"
)

func ReadExistingExport(outputDir string) (map[string]map[string]struct{}, error) {
	files, err := filepath.Glob(outputDir + "/*.md")
	if err != nil {
		return nil, fmt.Errorf("glob failed: %w", err)
	}

	hashMap := make(map[string]map[string]struct{}, len(files))
	for _, file := range files {
		fd, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("open file: %w", err)
		}
		defer fd.Close()

		basename := filepath.Base(file)
		hashMap[basename] = make(map[string]struct{})

		br := bufio.NewReader(fd)
		scanner := bufio.NewScanner(br)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "- ") {
				hash := hashs.FNV64a(strings.TrimSpace(line[2:]))
				hashMap[basename][hash] = struct{}{}
			}
		}
	}

	if len(hashMap) > 0 {
		fmt.Println("Found", len(hashMap), "existing books in output directory", outputDir)
	}

	return hashMap, nil
}
