package i18n

import (
	"encoding/json"
	"io"
	"os"
)

type Language string

func (l *Language) String() string {
	return string(*l)
}

const (
	RU  Language = "RU"
	ENG Language = "ENG"
)

type Translations []*Translation

func (ts *Translations) Add(translation *Translation) {
	*ts = append(*ts, translation)
}

type Translation struct {
	Language Language  `json:"language"`
	Messages []Message `json:"messages"`
}

type Message struct {
	ID          string `json:"id"`
	Translation string `json:"translation"`
}

func FromFile(filePath string) (*Translation, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	contents, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var translation Translation
	err = json.Unmarshal(contents, &translation)
	if err != nil {
		return nil, err
	}

	return &translation, nil
}
