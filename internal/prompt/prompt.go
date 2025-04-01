package prompt

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"

	"github.com/nsr888/kindle-highlights-to-obsidian/internal/model"
)

func generateBooksPrompt(books model.Books) string {
	var sb strings.Builder

	for i, book := range books {
		sb.WriteString(fmt.Sprintf("[%d] %s - %s\n", i+1, book.Title, book.Author))
	}

	return sb.String()
}

func Run(
	books model.Books,
) ([]int, error) {
	booksCount := len(books)
	booksPrompt := generateBooksPrompt(books)

	fmt.Println(booksPrompt)

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
			if val <= 0 || val > booksCount {
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
		return nil, fmt.Errorf("prompt failed: %w", err)
	}

	fmt.Println("Selected value:", result)

	strs := strings.Split(result, " ")

	res := make([]int, 0)

	for _, str := range strs {
		bookIndex, err := strconv.Atoi(str)
		if err != nil {
			return nil, errors.New("invalid number")
		}
		res = append(res, bookIndex-1)
	}

	return res, nil
}
