/*
  	dela - web TODO list
    Copyright (C) 2023, 2024  Kasyanov Nikolay Alexeyevich (Unbewohnte)

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package server

import (
	"Unbewohnte/dela/db"
	"Unbewohnte/dela/i18n"
	"path/filepath"
)

type PageData struct {
	Translation map[string]string
	Data        interface{}
}

func (s *Server) GetPageData(templateNames []string, language i18n.Language) (*PageData, error) {
	translation := make(map[string]string)

	translationsDirPath := filepath.Join(s.config.BaseContentDir, TranslationsDirName)
	for _, page := range templateNames {
		pageTranslation, err := i18n.GetPageTranslation(page, language, translationsDirPath)
		if err != nil {
			// Try ENG
			pageTranslation, err = i18n.GetPageTranslation(page, i18n.Eng, translationsDirPath)
			if err != nil {
				return nil, err
			}
		}

		// Merge translations
		for _, message := range pageTranslation.Messages {
			translation[message.ID] = message.Translation
		}
	}

	return &PageData{
		Translation: translation,
		Data:        nil,
	}, nil
}

type IndexPageData struct {
	Groups []*db.TodoGroup `json:"groups"`
}

func GetIndexPageData(db *db.DB, login string) (*IndexPageData, error) {
	groups, err := db.GetAllUserTodoGroups(login)
	if err != nil {
		return nil, err
	}

	return &IndexPageData{
		Groups: groups,
	}, nil
}

type CategoryPageData struct {
	Groups         []*db.TodoGroup `json:"groups"`
	CurrentGroupId uint64          `json:"currentGroupId"`
	Todos          []*db.Todo      `json:"todos"`
}

func GetCategoryPageData(db *db.DB, login string, groupId uint64) (*CategoryPageData, error) {
	groups, err := db.GetAllUserTodoGroups(login)
	if err != nil {
		return nil, err
	}

	todos, err := db.GetGroupTodos(groupId)
	if err != nil {
		return nil, err
	}

	return &CategoryPageData{
		Groups:         groups,
		CurrentGroupId: groupId,
		Todos:          todos,
	}, nil
}
