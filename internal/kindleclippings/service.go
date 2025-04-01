package kindleclippings

import (
	"errors"
	"fmt"

	"github.com/nsr888/kindle-highlights-to-obsidian/internal/model"
	"github.com/nsr888/kindle-highlights-to-obsidian/internal/parser"
	"github.com/nsr888/kindle-highlights-to-obsidian/internal/storage"
)

func Parse(inputFile string) (model.Books, error) {
	var (
		books   = make(model.Books, 0)
		booksMap = make(map[string]int)
	)

	translMap, err := storage.ReadTranslations()
	if err != nil {
		return nil, fmt.Errorf("load translation map: %w", err)
	}

	rawClippings, err := storage.ReadRawClippings(inputFile)
	if err != nil {
		return nil, fmt.Errorf("read clippings: %w", err)
	}

	for _, c := range rawClippings {
		entry, errP := parser.ParseClippingsEntry(c, translMap)
		if errP != nil {
			if errors.Is(errP, parser.ErrInvalidEntry) ||
				errors.Is(errP, parser.ErrEmptyHighlight) {
				continue
			}
			return nil, fmt.Errorf("parse clippings entry: %w", errP)
		}

		key := fmt.Sprintf("%s%s", entry.BookTitle, entry.BookAuthor)

		if _, exists := booksMap[key]; !exists {
			booksMap[key] = len(books)
			books = append(books, model.Book{
				Title:            entry.BookTitle,
				Filename:         entry.Filename,
				Author:           entry.BookAuthor,
				FirstHighlightDt: entry.Date,
				Highlights:       make([]model.Highlight, 0),
			})
		}

		index := booksMap[key]
		bk := books[index]
		bk.LastHighlightDt = entry.Date
		bk.Highlights = append(bk.Highlights, model.Highlight{
			Date: entry.Date,
			Text: entry.HighlightText,
		})
		books[index] = bk
	}

	return books, nil
}
