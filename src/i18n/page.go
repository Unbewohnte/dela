package i18n

import (
	"path/filepath"
)

func GetPageTranslation(pageName string, language Language, translationsDirPath string) (*Translation, error) {
	translation, err := FromFile(
		filepath.Join(translationsDirPath, language.String(), pageName+".json"),
	)
	if err != nil {
		return nil, err
	}

	return translation, nil
}
