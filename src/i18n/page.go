package i18n

import (
	"fmt"
	"path/filepath"
)

func GetPageTranslation(pageName string, language Language, translationsDirPath string) (*Translation, error) {
	translation, err := FromFile(
		filepath.Join(translationsDirPath, language.String(), pageName+".json"),
	)
	if err != nil {
		return nil, err
	}

	if translation.Language != language {
		return translation, fmt.Errorf(
			"translation language (%s) differs from what was requested (%s)",
			translation.Language,
			language,
		)
	}

	return translation, nil
}
