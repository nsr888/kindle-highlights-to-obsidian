package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nsr888/kindle-highlights-to-obsidian/internal/kindleclippings"
	"github.com/nsr888/kindle-highlights-to-obsidian/internal/output"
	"github.com/nsr888/kindle-highlights-to-obsidian/internal/prompt"
)

func main() {
	inputFile := flag.String("input", "My Clippings.txt", "Path to My Clippings.txt")
	outputDir := flag.String("output", "./highlights", "Output directory")
	help := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		return
	}

	if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
		fmt.Println("Input file does not exist")
		return
	}

	books, err := kindleclippings.Parse(*inputFile)
	if err != nil {
		fmt.Println("Error processing kindle clippings from input file:", err)
		return
	}

	userResponse, err := prompt.Run(books)
	if err != nil {
		fmt.Println("Error prompting user for books:", err)
		return
	}

	requestedBooks := books.FilterByIndex(userResponse)

	existingHighlightsMap, err := output.ReadExistingExport(*outputDir)
	if err != nil {
		fmt.Println("Error processing existing highlights from output dir:", err)
		return
	}

	err = output.WriteBooks(*outputDir, requestedBooks, existingHighlightsMap)
	if err != nil {
		fmt.Println("Error writing books to output directory:", err)
		return
	}
}
