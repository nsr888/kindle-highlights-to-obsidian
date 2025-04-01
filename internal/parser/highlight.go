package parser

import (
	"fmt"
	"errors"
	"strings"
	"time"

	"github.com/nsr888/kindle-highlights-to-obsidian/internal/model"
)

const (
	safeFilenameChars = " abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯабвгдеёжзийклмнопрстуфхцчшщъыьэюя"
)

var (
	ErrInvalidEntry   = errors.New("invalid entry")
	ErrEmptyHighlight = errors.New("empty highlight")
)

type HighlightData struct {
	BookTitle     string
	Filename      string
	BookAuthor    string
	Date          time.Time
	HighlightText string
}

func ParseClippingsEntry(
	entry string,
	transMap map[string]model.Translation,
) (HighlightData, error) {
	lines := strings.Split(strings.TrimSpace(entry), "\n")
	if len(lines) < 3 {
		return HighlightData{}, ErrInvalidEntry
	}
	bookInfo := lines[0]
	metaInfo := lines[1]

	bookTitle := getBookTitle(bookInfo)
	bookAuthor := getBookAuthor(bookInfo)

	noteDate, err := Date(metaInfo, transMap)
	if err != nil {
		return HighlightData{}, err
	}

	highlightText := strings.Join(lines[3:], "\n")

	if strings.TrimSpace(highlightText) == "" {
		return HighlightData{}, ErrEmptyHighlight
	}

	return HighlightData{
		BookTitle:     bookTitle,
		Filename:      getValidFilename(bookTitle, bookAuthor),
		BookAuthor:    bookAuthor,
		Date:          noteDate,
		HighlightText: highlightText,
	}, nil
}

func getBookTitle(s string) string {
	parts := strings.Split(s, "(")
	if len(parts) < 2 {
		return s
	}
	return strings.TrimSpace(parts[0])
}

func getBookAuthor(s string) string {
	parts := strings.Split(s, "(")
	if len(parts) < 2 {
		return s
	}
	part := strings.TrimSpace(parts[1])

	return strings.TrimRight(part, ")")
}

func getValidFilename(bookTitle, bookAuthor string) string {
	extention := "md"
	bookAuthor = strings.TrimSpace(bookAuthor)
	bookTitle = strings.TrimSpace(bookTitle)
	bookTitleAuthor := fmt.Sprintf("%s - %s", bookTitle, bookAuthor)
	// clean up the string
	var builder strings.Builder
	for _, c := range bookTitleAuthor {
		if strings.ContainsRune(safeFilenameChars, c) {
			builder.WriteRune(c)
		}
	}
	bookTitleAuthor = builder.String()

	return fmt.Sprintf("%s.%s", bookTitleAuthor, extention)
}
