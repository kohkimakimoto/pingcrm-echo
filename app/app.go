package app

import (
	"context"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kohkimakimoto/inertia-echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type App struct {
	Echo   *echo.Echo
	Config *Config
	DB     *DB
}

func NewApp(c *Config) (*App, error) {
	a := &App{
		Echo:   echo.New(),
		Config: c,
	}

	// echo
	e := a.Echo
	e.HideBanner = true
	e.HidePort = true
	e.Server.Addr = c.Addr

	// logger
	logger := log.New("pingcrm")
	logger.SetHeader(`${time_rfc3339} ${level}`)
	if c.IsDev {
		logger.SetLevel(log.DEBUG)
	}
	e.Logger = logger

	// renderer
	e.Renderer = inertia.NewRendererWithFS(c.ViewsFs, "views/*.html", map[string]interface{}{
		"vite_entry": inertia.ViteEntry(c.GetViteManifest()),
		"is_dev": func() bool {
			return c.IsDev
		},
	})

	e.HTTPErrorHandler = ErrorHandler

	// database
	if c.DataDir == "" {
		return nil, errors.New("DataDir ('-data-dir' flag or 'DATA_DIR' environment variable) is required. ")
	}

	if _, err := os.Stat(c.DataDir); os.IsNotExist(err) {
		err = os.MkdirAll(c.DataDir, os.FileMode(0755))
		if err != nil {
			return nil, err
		}
		logger.Infof("Created date directory: %s", c.DataDir)
	}

	db, err := NewDB(c.DatabasePath())
	if err != nil {
		return nil, err
	}
	if db.IsNotExist() {
		if err := db.Init(c.DatabaseFs); err != nil {
			return nil, err
		}
		logger.Infof("Initialized database: %s", db.Path)
	}
	a.DB = db

	// middleware
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	// inject app instance to echo context.
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(appKey, a)
			return next(c)
		}
	})
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: middleware.DefaultSkipper,
		Format:  `${time_rfc3339} ${remote_ip} ${host} ${method} ${uri} ${status} ${latency} ${latency_human} ${bytes_in} ${bytes_out}` + "\n",
	}))
	pubDir, _ := fs.Sub(c.PublicFs, "public")
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root:       "/",
		Filesystem: http.FS(pubDir),
	}))
	e.Use(SessionMiddleware(c.Secret))
	e.Use(inertia.MiddlewareWithConfig(inertia.MiddlewareConfig{
		Share: func(c echo.Context) (map[string]interface{}, error) {
			sess := r.Session(c)
			var user map[string]interface{}
			if u, _ := sess.User(); u != nil {
				user = map[string]interface{}{
					"id":         u.Id,
					"first_name": u.FirstName,
					"last_name":  u.LastName,
					"email":      u.Email,
					"owner":      u.Owner,
					"account": map[string]interface{}{
						"id":   u.AccountId,
						"name": u.Account.Name,
					},
				}
			}

			flash, err := sess.Flash()
			if err != nil {
				return nil, err
			}

			flashErrs, err := sess.Errors()
			if err != nil {
				return nil, err
			}

			return map[string]interface{}{
				"auth": map[string]interface{}{
					"user": user,
				},
				"flash": map[string]interface{}{
					"success": flash.Success,
					"error":   flash.Error,
				},
				"errors": flashErrs,
			}, nil
		},
	}))
	e.Use(inertia.CSRF())

	auth := AuthRequired()
	guest := RedirectIfAuthenticated()

	// auth
	e.GET("login", AuthCreateHandler, guest)
	e.POST("login", AuthStoreHandler, guest)
	e.DELETE("logout", AuthDestroyHandler)

	// dashboard
	e.GET("/", DashboardIndexHandler, auth)

	// users
	e.GET("/users", UsersIndexHandler, auth)
	e.GET("/users/create", UsersCreateHandler, auth)
	e.POST("/users", UsersStoreHandler, auth)
	e.GET("/users/:id/edit", UsersEditHandler, auth)
	e.PUT("/users/:id", UsersUpdateHandler, auth)
	e.DELETE("/users/:id", UsersDestroyHandler, auth)
	e.PUT("/users/:id/restore", UsersRestoreHandler, auth)

	// organizations
	e.GET("/organizations", OrganizationsIndexHandler, auth)
	e.GET("/organizations/create", OrganizationsCreateHandler, auth)
	e.POST("/organizations", OrganizationsStoreHandler, auth)
	e.GET("/organizations/:id/edit", OrganizationsEditHandler, auth)
	e.PUT("/organizations/:id", OrganizationsUpdateHandler, auth)
	e.DELETE("/organizations/:id", OrganizationsDestroyHandler, auth)
	e.PUT("/organizations/:id/restore", OrganizationsRestoreHandler, auth)

	// contacts
	e.GET("/contacts", ContactsIndexHandler, auth)
	e.GET("/contacts/create", ContactsCreateHandler, auth)
	e.POST("/contacts", ContactsStoreHandler, auth)
	e.GET("/contacts/:id/edit", ContactsEditHandler, auth)
	e.PUT("/contacts/:id", ContactsUpdateHandler, auth)
	e.DELETE("/contacts/:id", ContactsDestroyHandler, auth)
	e.PUT("/contacts/:id/restore", ContactsRestoreHandler, auth)

	// reports
	e.GET("/reports", ReportsIndexHandler, auth)

	// images
	e.GET("/img/:path", ImagesShowHandler)

	return a, nil
}

func (a *App) Start() error {
	e := a.Echo
	logger := e.Logger

	if a.Config.IsDev {
		logger.Debug("Run on dev mode")
	}

	if a.Config.IsDemo {
		logger.Info("Run on demo mode")
	}

	// start http server
	go func() {
		if err := e.Start(e.Server.Addr); err != nil {
			logger.Info(err)
		}
	}()
	logger.Infof("The server listening on %s (pid: %d)", e.Server.Addr, os.Getpid())

	// see https://echo.labstack.com/cookbook/graceful-shutdown
	// Wait for interrupt signal to gracefully shut down the server with a timeout of 'ShutdownTimeoutSec' seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Infof("Received signal: %v", sig)
	timeout := time.Duration(10) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Info("Shutting down the server")
	if err := e.Shutdown(ctx); err != nil {
		logger.Errorf("failed to shut down echo http server: %v", err)
		return err
	}

	// done
	logger.Infof("Successfully shutdown")
	return nil
}

func (a *App) Close() {
	if a.DB != nil {
		if err := a.DB.Close(); err != nil {
			a.Echo.Logger.Error(err)
		}
	}
}

const appKey = "app"

func getApp(c echo.Context) *App {
	return c.Get(appKey).(*App)
}
