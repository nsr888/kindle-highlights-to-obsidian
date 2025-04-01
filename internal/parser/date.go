package parser

import (
	"fmt"
	"strings"
	"time"

	"github.com/goodsign/monday"

	"github.com/nsr888/kindle-highlights-to-obsidian/internal/model"
)

const (
	debugModeEnabled  = false
)

// example metaInfo: - Your Highlight Location 1293-1294 | Added on Sunday, December 1, 2013 7:49:48 PM
func Date(
	metaInfo string,
	transMap map[string]model.Translation,
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
