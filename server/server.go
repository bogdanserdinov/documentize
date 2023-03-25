package server

import (
	"context"
	"documentize/users"
	"errors"
	"github.com/gorilla/mux"
	"github.com/zeebo/errs"
	"golang.org/x/sync/errgroup"
	"html/template"
	"net"
	"net/http"
	"path/filepath"
)

// Config contains configuration for console web server.
type Config struct {
	Address   string `json:"address"`
	StaticDir string `json:"staticDir"`
}

// Server represents console web server.
//
// architecture: Endpoint
type Server struct {
	config Config

	Listener net.Listener
	server   http.Server

	templates struct {
		users UserTemplates
	}
}

func New(config Config, listener net.Listener, users *users.Service) (*Server, error) {
	server := &Server{
		config:   config,
		Listener: listener,
	}

	err := server.initializeTemplates()
	if err != nil {
		return nil, err
	}

	router := mux.NewRouter()

	userRouter := router.PathPrefix("/users").Subrouter()
	userController := NewUsers(users, server.templates.users)
	userRouter.HandleFunc("", userController.List).Methods(http.MethodGet)
	userRouter.HandleFunc("/create", userController.Create).Methods(http.MethodGet, http.MethodPost)
	userRouter.HandleFunc("/export-data", userController.Export).Methods(http.MethodGet)
	userRouter.HandleFunc("/generate/{id}", userController.Generate).Methods(http.MethodGet)

	css := http.FileServer(http.Dir(server.config.StaticDir))
	router.PathPrefix("/css/").Handler(http.StripPrefix("/css/", css))

	server.server = http.Server{
		Handler: router,
	}

	return server, nil
}

func (server *Server) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	var group errgroup.Group
	group.Go(func() error {
		<-ctx.Done()
		return server.server.Shutdown(context.Background())
	})
	group.Go(func() error {
		defer cancel()
		err := server.server.Serve(server.Listener)
		isCancelled := errs.IsFunc(err, func(err error) bool { return errors.Is(err, context.Canceled) })
		if isCancelled || errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		return err
	})

	return group.Wait()
}

// initializeTemplates initializes and caches templates for managers controller.
func (server *Server) initializeTemplates() (err error) {
	server.templates.users.List, err = template.ParseFiles(filepath.Join(server.config.StaticDir, "users", "list.html"))
	if err != nil {
		return err
	}
	server.templates.users.Create, err = template.ParseFiles(filepath.Join(server.config.StaticDir, "users", "create.html"))
	if err != nil {
		return err
	}

	return nil
}

// Close closes server and underlying listener.
func (server *Server) Close() error {
	return server.server.Close()
}
