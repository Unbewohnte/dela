package server

import (
	"html/template"
	"path/filepath"
)

// Constructs a pageName template via inserting basePageName in pagesDir
func getPage(pagesDir string, basePageName string, pageName string) (*template.Template, error) {
	page, err := template.ParseFiles(
		filepath.Join(pagesDir, basePageName),
		filepath.Join(pagesDir, pageName),
	)
	if err != nil {
		return nil, err
	}

	return page, nil
}
