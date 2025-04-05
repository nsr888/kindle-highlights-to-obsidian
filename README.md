# Kindle highlights to Obsidian Markdown notes, written in Go.

## Installing

### Using releases

Check the [releases page](https://github.com/nsr888/kindle-highlights-to-obsidian/releases) for direct downloads of the binary. After you download it, you can add it to your `$PATH`.

### Using `go install`

Make sure that `$GOPATH/bin` is in your `$PATH`, because that's where this gets installed:

```
go install github.com/nsr888/kindle-highlights-to-obsidian@latest
```

## Usage

Run with the following command:

```
./kindle-highlights-to-obsidian -input [<path to My Clippings.txt>] -output [<output directory>]
```

Select which books to process from the interactive menu:

* Enter 0 to process all books
* Enter a single number to process one specific book
* Enter multiple numbers separated by spaces to process several books

## Tested device

- Amazon Kindle Paperwhite 5th Generation (EY21)
- Fedora Linux

## Links

- Inspired by https://github.com/dannberg/kindle-clippings-to-obsidian
