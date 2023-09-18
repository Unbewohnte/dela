/*
  	dela - web TODO list
    Copyright (C) 2023  Kasyanov Nikolay Alexeyevich (Unbewohnte)

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
