package main

import (
	"Unbewohnte/dela/conf"
	"Unbewohnte/dela/logger"
	"Unbewohnte/dela/server"
	"os"
	"path/filepath"
)

const ConfName string = "conf.json"

var (
	WDir string
	Conf conf.Conf
)

func init() {
	// Initialize logging
	logger.SetOutput(os.Stdout)

	// Work out the working directory
	exePath, err := os.Executable()
	if err != nil {
		logger.Error("[Init] Failed to retrieve executable's path: %s", err)
		os.Exit(1)
	}
	WDir = filepath.Dir(exePath)
	logger.Info("[Init] Working in \"%s\"", WDir)

	// Open configuration, create if does not exist
	Conf, err = conf.FromFile(filepath.Join(WDir, ConfName))
	if err != nil {
		_, err = conf.Create(filepath.Join(WDir, ConfName), conf.Default())
		if err != nil {
			logger.Error("[Init] Failed to create a new configuration file: %s", err)
			os.Exit(1)
		}
		logger.Info("[Init] Created a new configuration file")
		os.Exit(0)
	}
	logger.Info("[Init] Opened existing configuration file")
	if Conf.BaseContentDir == "." {
		Conf.BaseContentDir = WDir
	}

	logger.Info("[Init] Successful initializaion!")
}

func main() {
	server, err := server.New(Conf)
	if err != nil {
		logger.Error("[Main] Failed to initialize a new server with conf (%+v): %s", Conf, err)
		return
	}
	logger.Info("[Main] Successfully initialized a new server instance with conf (%+v)", Conf)

	err = server.Start()
	if err != nil {
		logger.Error("[Main] Fatal server failure: %s. Exiting...", err)
		return
	}
}
