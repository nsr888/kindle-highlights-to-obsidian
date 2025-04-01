package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nsr888/kindle-highlights-to-obsidian/internal/model"
)

func ReadTranslations() (map[string]model.Translation, error) {
	files, err := filepath.Glob("languages/*.json")
	if err != nil {
		return nil, fmt.Errorf("glob failed: %w", err)
	}

	translationMap := make(map[string]model.Translation, len(files))
	for _, file := range files {
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

		b := filepath.Base(file)
		lang := strings.TrimSuffix(b, filepath.Ext(b))
		translationMap[lang] = translation
	}

	return translationMap, nil
}

func loadTranslationFromJSON(bytes []byte) (model.Translation, error) {
	var translation model.Translation
	err := json.Unmarshal(bytes, &translation)
	if err != nil {
		return model.Translation{}, err
	}
	return translation, nil
}
