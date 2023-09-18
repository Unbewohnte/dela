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
