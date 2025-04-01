package model

import (
	"time"
)

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

type Books []Book

func (b *Books) FilterByIndex(indexes []int) Books {
	if b == nil {
		return Books{}
	}

	if len(indexes) == 0 {
		return Books{}
	}

	if len(indexes) == 1 && indexes[0] < 0 {
		return *b
	}

	var res Books

	for _, idx := range indexes {
		if idx < 0 || idx >= len(*b) {
			continue
		}

		res = append(res, (*b)[idx])
	}
	return res
}
