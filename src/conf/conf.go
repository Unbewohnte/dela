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

package conf

import (
	"encoding/json"
	"io"
	"os"
)

type Conf struct {
	Port           uint16 `json:"port"`
	CertFilePath   string `json:"cert_file_path"`
	KeyFilePath    string `json:"key_file_path"`
	BaseContentDir string `json:"base_content_dir"`
	ProdDBName     string `json:"production_db_name"`
}

// Creates a default server configuration
func Default() Conf {
	return Conf{
		Port:           8080,
		CertFilePath:   "",
		KeyFilePath:    "",
		BaseContentDir: ".",
		ProdDBName:     "dela.db",
	}
}

// Tries to retrieve configuration from given json file
func FromFile(path string) (Conf, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return Default(), err
	}
	defer configFile.Close()

	confBytes, err := io.ReadAll(configFile)
	if err != nil {
		return Default(), err
	}

	var config Conf
	err = json.Unmarshal(confBytes, &config)
	if err != nil {
		return Default(), err
	}

	return config, nil
}

// Create empty configuration file
func Create(path string, conf Conf) (Conf, error) {
	configFile, err := os.Create(path)
	if err != nil {
		return Default(), err
	}
	defer configFile.Close()

	configJsonBytes, err := json.MarshalIndent(conf, "", " ")
	if err != nil {
		return conf, err
	}

	_, err = configFile.Write(configJsonBytes)
	if err != nil {
		return conf, nil
	}

	return conf, nil
}
