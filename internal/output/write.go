package output

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"text/template"

	"github.com/nsr888/kindle-highlights-to-obsidian/pkg/hashs"
	"github.com/nsr888/kindle-highlights-to-obsidian/internal/model"
)

const (
	fileMode = 0644
)

func WriteBooks(
	outputDir string,
	books []model.Book,
	existingClippings map[string]map[string]struct{},
) error {
	err := os.MkdirAll(outputDir, fs.ModePerm)
	if err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	for _, values := range books {
		if _, exists := existingClippings[values.Filename]; exists {
			err := appendClippings(outputDir, values, existingClippings[values.Filename])
			if err != nil {
				return fmt.Errorf("append clippings: %w", err)
			}

			continue
		}

		err := WriteBook(outputDir, values)
		if err != nil {
			return fmt.Errorf("write all clippings: %w", err)
		}
	}

	return nil
}

func appendClippings(
	outputDir string,
	book model.Book,
	existingClippings map[string]struct{},
) error {
	filePath := filepath.Join(outputDir, book.Filename)
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, fileMode)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	fmt.Println("Found", len(book.Highlights), "highlights in", book.Filename)

	cnt := 0
	for _, highlight := range book.Highlights {
		hash := hashs.FNV64a(highlight.Text)
		if _, exists := existingClippings[hash]; exists {
			cnt++
			continue
		}

		_, err = f.WriteString("- " + highlight.Text + "\n")
		if err != nil {
			return fmt.Errorf("write string: %w", err)
		}
		fmt.Println("Appended highlight with hash", hash, "to", book.Filename)
	}

	if cnt > 0 {
		fmt.Println("Skipped", cnt, "highlights in", book.Filename)
	}

	return nil
}

func WriteBook(
	outputDir string,
	book model.Book,
) error {
	tmplDir := "./templates"
	tmplPath := filepath.Join(tmplDir, "obsidian.tmpl")
	baseFile := filepath.Base(tmplPath)

	tmpl, err := template.New(baseFile).ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	filePath := filepath.Join(outputDir, book.Filename)
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	err = tmpl.Execute(f, book)
	if err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	return nil
}
