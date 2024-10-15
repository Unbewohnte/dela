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
	"Unbewohnte/dela/conf"
	"Unbewohnte/dela/db"
	"Unbewohnte/dela/logger"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	PagesDirName   string = "pages"
	StaticDirName  string = "static"
	ScriptsDirName string = "scripts"
)

type Server struct {
	config conf.Conf
	db     *db.DB
	http   http.Server
}

// Creates a new server instance with provided config
func New(config conf.Conf) (*Server, error) {
	var server Server = Server{}
	server.config = config

	// check if required directories are present
	_, err := os.Stat(filepath.Join(config.BaseContentDir, PagesDirName))
	if err != nil {
		logger.Error("[Server] A directory with HTML pages is not available: %s", err)
		return nil, err
	}

	_, err = os.Stat(filepath.Join(config.BaseContentDir, ScriptsDirName))
	if err != nil {
		logger.Error("[Server] A directory with scripts is not available: %s", err)
		return nil, err
	}

	_, err = os.Stat(filepath.Join(config.BaseContentDir, StaticDirName))
	if err != nil {
		logger.Error("[Server] A directory with static content is not available: %s", err)
		return nil, err
	}

	// get database working
	serverDB, err := db.FromFile(filepath.Join(config.BaseContentDir, config.ProdDBName))
	if err != nil {
		// Create one then
		serverDB, err = db.Create(filepath.Join(config.BaseContentDir, config.ProdDBName))
		if err != nil {
			logger.Error("Failed to create a new database: %s", err)
			return nil, err
		}
	}
	server.db = serverDB
	logger.Info("Opened a database successfully")

	// start constructing an http server configuration
	server.http = http.Server{
		Addr: fmt.Sprintf(":%d", server.config.Port),
	}

	// configure paths' callbacks
	mux := http.NewServeMux()
	mux.Handle(
		"/static/",
		http.StripPrefix("/static/", http.FileServer(
			http.Dir(filepath.Join(server.config.BaseContentDir, StaticDirName))),
		),
	)

	mux.Handle(
		"/scripts/",
		http.StripPrefix("/scripts/", http.FileServer(
			http.Dir(filepath.Join(server.config.BaseContentDir, ScriptsDirName))),
		),
	)

	// handle page requests
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		switch req.URL.Path {
		case "/":
			requestedPage, err := getPage(
				filepath.Join(server.config.BaseContentDir, PagesDirName), "base.html", "index.html",
			)
			if err != nil {
				http.Redirect(w, req, "/about", http.StatusTemporaryRedirect)
				logger.Error("[Server][/] Failed to get a page: %s", err)
				return
			}

			requestedPage.ExecuteTemplate(w, "index.html", nil)
		default:
			requestedPage, err := getPage(
				filepath.Join(server.config.BaseContentDir, PagesDirName),
				"base.html",
				req.URL.Path[1:]+".html",
			)
			if err == nil {
				requestedPage.ExecuteTemplate(w, req.URL.Path[1:]+".html", nil)
			} else {
				http.Error(w, "Page processing error", http.StatusInternalServerError)
			}
		}
	})
	mux.HandleFunc("/api/user/get", server.EndpointUserGet)              // Non specific
	mux.HandleFunc("/api/user/delete", server.EndpointUserDelete)        // Non specific
	mux.HandleFunc("/api/user/update", server.EndpointUserUpdate)        // Non specific
	mux.HandleFunc("/api/user/create", server.EndpointUserCreate)        // Non specific
	mux.HandleFunc("/api/todo/get", server.EndpointUserTodosGet)         // Non specific
	mux.HandleFunc("/api/todo/delete/", server.EndpointTodoDelete)       // Specific
	mux.HandleFunc("/api/todo/update/", server.EndpointTodoUpdate)       // Specific
	mux.HandleFunc("/api/group/create", server.EndpointTodoGroupCreate)  // Non specific
	mux.HandleFunc("/api/group/get/", server.EndpointTodoGroupGet)       // Specific
	mux.HandleFunc("/api/group/update/", server.EndpointTodoGroupUpdate) // Specific
	mux.HandleFunc("/api/group/delete/", server.EndpointTodoGroupDelete) // Specific

	server.http.Handler = mux
	logger.Info("[Server] Created an HTTP server instance")

	return &server, nil
}

// Launches server instance
func (s *Server) Start() error {
	if s.config.CertFilePath != "" && s.config.KeyFilePath != "" {
		logger.Info("[Server] Using TLS")
		logger.Info("[Server] HTTP server is going live on port %d!", s.config.Port)

		err := s.http.ListenAndServeTLS(s.config.CertFilePath, s.config.KeyFilePath)
		if err != nil && err != http.ErrServerClosed {
			logger.Error("[Server] Fatal server error: %s", err)
			return err
		}
	} else {
		logger.Info("[Server] Not using TLS")
		logger.Info("[Server] HTTP server is going live on port %d!", s.config.Port)

		err := s.http.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error("[Server] Fatal server error: %s", err)
			return err
		}
	}

	return nil
}

// Stops the server immediately
func (s *Server) Stop() {
	ctx, cfunc := context.WithDeadline(context.Background(), time.Now().Add(time.Second*10))
	s.http.Shutdown(ctx)
	cfunc()
}
