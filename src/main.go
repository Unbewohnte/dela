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

package main

import (
	"Unbewohnte/dela/conf"
	"Unbewohnte/dela/logger"
	"Unbewohnte/dela/server"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const Version string = "0.2.0"

var (
	printVersion *bool = flag.Bool("version", false, "Print version information and exit")
)

const ConfName string = "conf.json"

var (
	WDir string
	Conf conf.Conf
)

func init() {
	// Parse flags
	flag.Parse()
	if *printVersion {
		fmt.Printf("dela v%s - a web TODO list\n(c) 2023 Kasyanov Nikolay Alexeyevich (Unbewohnte)\n", Version)
		os.Exit(0)
	}

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
