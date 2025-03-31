package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"hash/fnv"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/goodsign/monday"
	"github.com/manifoldco/promptui"
)

const (
	separator         = "=========="
	fileMode          = 0644
	alphabet          = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	safeFilenameChars = " abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯабвгдеёжзийклмнопрстуфхцчшщъыьэюя"
	debugModeEnabled  = false
)

var ErrProcessing = errors.New("error processing clippings")

func hash64Base62(data []byte) string {
	h := fnv.New64a()
	h.Write(data)
	num := h.Sum64()

	if num == 0 {
		return string(alphabet[0])
	}

	var result string
	base := uint64(len(alphabet))
	for num > 0 {
		result = string(alphabet[num%base]) + result
		num /= base
	}
	return result
}

type Translation struct {
	AddedOn string `json:"added_on"`
}

func (t Translation) Validate() error {
	if t.AddedOn == "" {
		return errors.New("required: AddedOn")
	}

	return nil
}

func loadTranslationMapFromFiles() (map[string]Translation, error) {
	files, err := filepath.Glob("languages/*.json")
	if err != nil {
		return nil, fmt.Errorf("glob failed: %w", err)
	}

	translationMap := make(map[string]Translation, len(files))
	for _, file := range files {
		baseName := filepath.Base(file)
		baseNameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("read translation file: %w", err)
		}

		translation, err := loadTranslationFromJSON(content)
		if err != nil {
			return nil, fmt.Errorf("load translation from json: %w", err)
		}

		err = translation.Validate()
		if err != nil {
			return nil, fmt.Errorf("validate: %w", err)
		}

		translationMap[baseNameWithoutExt] = translation
	}

	return translationMap, nil
}

func loadTranslationFromJSON(bytes []byte) (Translation, error) {
	var translation Translation
	err := json.Unmarshal(bytes, &translation)
	if err != nil {
		return Translation{}, err
	}
	return translation, nil
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

// example metaInfo: - Your Highlight Location 1293-1294 | Added on Sunday, December 1, 2013 7:49:48 PM
func parseNoteDate(
	metaInfo string,
	transMap map[string]Translation,
) (time.Time, error) {
	dateParts := strings.Split(metaInfo, " | ")
	if len(dateParts) < 2 {
		return time.Time{}, fmt.Errorf("invalid metaInfo: %s", metaInfo)
	}

	datePart := strings.TrimSpace(dateParts[len(dateParts)-1])
	var dateStr string

	for _, v := range transMap {
		if strings.Contains(datePart, v.AddedOn) {
			dateParts = strings.Split(datePart, v.AddedOn)
			if len(dateParts) < 2 {
				return time.Time{}, fmt.Errorf("invalid date string: %s", dateParts)
			}
			dateStr = strings.TrimSpace(dateParts[1])
			break
		}
	}

	dateStr = strings.ReplaceAll(dateStr, " г. ", " ")
	dateStr = strings.ReplaceAll(dateStr, " в ", " ")

	if dateStr == "" {
		return time.Time{}, fmt.Errorf("datePart %s not contains date string from translation map", datePart)
	}

	var (
		dt            time.Time
		dateTemplates = []string{
			"Monday, January 2, 2006 3:04:05 PM",
			"Monday, January 2, 2006 15:04:05",
			"Monday, 2 January 2006 15:04:05",
			"monday, 2 january 2006 15:04:05",
		}
	)
	var err error
	var errorStr string
	for _, tmpl := range dateTemplates {
		dt, err = time.Parse(tmpl, dateStr)
		if err != nil {
			errorStr += fmt.Sprintf("parse date string: %s with template: %s, error: %v\n", dateStr, tmpl, err)
			continue
		} else {
			break
		}
	}
	for _, tmpl := range dateTemplates {
		dt, err = monday.Parse(tmpl, dateStr, monday.LocaleRuRU)
		if err != nil {
			errorStr += fmt.Sprintf("monday parse date string: %s with template: %s, error: %v\n", dateStr, tmpl, err)
			continue
		} else {
			break
		}
	}
	if dt.IsZero() && debugModeEnabled {
		fmt.Println(errorStr)
	}

	return dt, nil
}

type Book struct {
	Title            string
	Filename         string
	Author           string
	FirstHighlightDt time.Time
	LastHighlightDt  time.Time
	Highlights       []Highlight
}

type Highlight struct {
	Date time.Time
	Text string
}

func processKindleClippings(
	inputFile string,
	outputDir string,
	existingClippings map[string]map[string]struct{},
) error {
	transMap, err := loadTranslationMapFromFiles()
	if err != nil {
		return fmt.Errorf("load translation map: %w", err)
	}

	fd, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("open input file: %w", err)
	}
	defer fd.Close()
	br := bufio.NewReader(fd)
	r, _, err := br.ReadRune()
	if err != nil {
		return fmt.Errorf("read rune: %w", err)
	}
	if r != '\uFEFF' {
		err = br.UnreadRune() // Not a BOM -- put the rune back
		if err != nil {
			return fmt.Errorf("unread rune: %w", err)
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

	books := make([]Book, 0)
	bookHashMap := make(map[string]int)

	var bookPrompt string

	idx := 0
	for _, entry := range entries {
		lines := strings.Split(strings.TrimSpace(entry), "\n")
		if len(lines) < 3 {
			continue
		}

		bookInfo := lines[0]
		bookTitle := getBookTitle(bookInfo)
		bookAuthor := getBookAuthor(bookInfo)

		metaInfo := lines[1]
		noteDate, errParse := parseNoteDate(metaInfo, transMap)
		if errParse != nil {
			fmt.Printf("parse metaInfo: %s, error: %v\n", metaInfo, errParse)
			continue
		}

		highlightText := strings.Join(lines[3:], "\n")

		if strings.TrimSpace(highlightText) == "" {
			continue
		}

		bookHash := hash64Base62(fmt.Appendf(nil, "%s%s", bookTitle, bookAuthor))
		var pos int
		pos, exists := bookHashMap[bookHash]
		if !exists {
			bookHashMap[bookHash] = idx
			pos = idx
			idx++
			bookPrompt += fmt.Sprintf("[%d] %s - %s\n", pos+1, bookTitle, bookAuthor)
			books = append(books, Book{
				Title:            bookTitle,
				Filename:         getValidFilename(bookTitle, bookAuthor),
				Author:           bookAuthor,
				FirstHighlightDt: noteDate,
				Highlights:       make([]Highlight, 0),
			})
		}

		book := books[pos]
		book.LastHighlightDt = noteDate
		book.Highlights = append(book.Highlights, Highlight{
			Date: noteDate,
			Text: highlightText,
		})
		books[pos] = book
	}

	fmt.Println(bookPrompt)

	validate := func(input string) error {
		strs := strings.Split(input, " ")

		if len(strs) == 1 && strs[0] == "0" {
			return nil
		}

		for _, str := range strs {
			val, errAtoi := strconv.Atoi(str)
			if errAtoi != nil {
				return fmt.Errorf("parse int: %w", errAtoi)
			}
			if val <= 0 || val > len(books) {
				return errors.New("value out of range")
			}

		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    "Input one or more numbers, separated by a space:",
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}

	fmt.Println("Selected value:", result)

	handlePromptResult := func(result string) ([]Book, error) {
		strs := strings.Split(result, " ")

		if len(strs) == 1 && strs[0] == "0" {
			return books, nil
		}

		var bookIndex int
		res := make([]Book, 0)

		for _, str := range strs {
			bookIndex, err = strconv.Atoi(str)
			if err != nil {
				return nil, errors.New("invalid number")
			}
			res = append(res, books[bookIndex-1])
		}

		return res, nil
	}

	books, err = handlePromptResult(result)
	if err != nil {
		return fmt.Errorf("handle prompt result: %w", err)
	}

	err = os.MkdirAll(outputDir, fs.ModePerm)
	if err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	tmplDir := "./templates"
	tmplPath := filepath.Join(tmplDir, "obsidian.tmpl")
	baseFile := filepath.Base(tmplPath)
	tmpl, err := template.New(baseFile).ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	for _, values := range books {
		if _, exists := existingClippings[values.Filename]; exists {
			err = appendClippings(outputDir, values, existingClippings[values.Filename])
			if err != nil {
				return fmt.Errorf("append clippings: %w", err)
			}

			continue
		}

		err = writeAllClippings(tmpl, outputDir, values)
		if err != nil {
			return fmt.Errorf("write all clippings: %w", err)
		}
	}

	return nil
}

func appendClippings(
	outputDir string,
	book Book,
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
		hash := hash64Base62([]byte(highlight.Text))
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

func writeAllClippings(
	tmpl *template.Template,
	outputDir string,
	book Book,
) error {
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

func processExistingClippings(outputDir string) (map[string]map[string]struct{}, error) {
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
				hash := hash64Base62([]byte(strings.TrimSpace(line[2:])))
				hashMap[basename][hash] = struct{}{}
			}
		}
	}

	if len(hashMap) > 0 {
		fmt.Println("Found", len(hashMap), "existing books in output directory", outputDir)
	}

	return hashMap, nil
}

func main() {
	inputFile := flag.String("input", "My Clippings.txt", "Path to My Clippings.txt")
	outputDir := flag.String("output", "./highlights", "Output directory")
	flag.Parse()

	if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
		fmt.Println("Input file does not exist")
		return
	}

	hashMap, err := processExistingClippings(*outputDir)
	if err != nil {
		fmt.Println("Error processing existing clippings:", err)
		return
	}

	if err := processKindleClippings(*inputFile, *outputDir, hashMap); err != nil {
		fmt.Println("Error processing clippings:", err)
		return
	}
}
