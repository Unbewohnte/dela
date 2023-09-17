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
	serverDB, err := db.FromFile(config.ProdDBPath)
	if err != nil {
		// Create one then
		serverDB, err = db.Create(config.ProdDBPath)
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
		// Ignore favicon request. It's not an .ico
		if req.URL.Path == "favicon.ico" {
			return
		}

		switch req.URL.Path {
		case "/":
			requestedPage, err := getPage(
				filepath.Join(server.config.BaseContentDir, PagesDirName), "base.html", "index.html",
			)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
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
				// Redirect to the index
				index, err := getPage(
					filepath.Join(server.config.BaseContentDir, PagesDirName),
					"base.html",
					req.URL.Path[1:]+".html",
				)
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
				}

				index.ExecuteTemplate(w, "index.html", nil)
			}
		}
	})
	mux.HandleFunc("/api/user", server.UserEndpoint)
	mux.HandleFunc("/api/todo", server.TodoEndpoint)
	// mux.HandleFunc("/api/group", server.TodoGroupEndpoint)

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
