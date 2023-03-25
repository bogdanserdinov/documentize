package documentize

import (
	"context"
	"documentize/server"
	"documentize/users"
	"errors"
	"github.com/zeebo/errs"
	"golang.org/x/sync/errgroup"
	"net"
)

type DB interface {
	Users() users.DB

	Close() error
	CreateSchema(ctx context.Context) error
}

type Config struct {
	DatabaseURL     string `env:"DATABASE_URL,notEmpty"`
	ServerAddress   string `env:"SERVER_ADDRESS,notEmpty"`
	StaticDir       string `env:"STATIC_DIR,notEmpty"`
	ExportDataPath  string `env:"EXPORT_DATA_PATH,notEmpty"`
	DocTemplatePath string `env:"DOC_TEMPLATE_PATH,notEmpty"`
}

type Documentize struct {
	config *Config

	users *users.Service

	server *server.Server
}

func New(config *Config, db DB) (*Documentize, error) {
	app := &Documentize{
		config: config,
	}

	{
		cfg := users.Config{
			ExportDataPath:  config.ExportDataPath,
			DocTemplatePath: config.DocTemplatePath,
		}

		app.users = users.New(cfg, db.Users())
	}

	{
		listener, err := net.Listen("tcp", config.ServerAddress)
		if err != nil {
			return nil, err
		}

		cfg := server.Config{
			Address:   config.ServerAddress,
			StaticDir: config.StaticDir,
		}

		app.server, err = server.New(
			cfg,
			listener,
			app.users,
		)
	}

	return app, nil
}

func (documentize *Documentize) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return ignoreCancel(documentize.server.Run(ctx))
	})
	return group.Wait()
}

// Close closes all the resources.
func (documentize *Documentize) Close() error {
	var errlist errs.Group
	errlist.Add(documentize.server.Close())
	return errlist.Err()
}

// we ignore cancellation and stopping errors since they are expected.
func ignoreCancel(err error) error {
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}
