package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ricoberger/aks-state-exporter/pkg/config"
	"github.com/ricoberger/aks-state-exporter/pkg/exporter"
	"github.com/ricoberger/aks-state-exporter/pkg/logger"
	"github.com/ricoberger/aks-state-exporter/pkg/server"
	"github.com/ricoberger/aks-state-exporter/pkg/version"

	"github.com/alecthomas/kong"
	"github.com/prometheus/client_golang/prometheus"
)

var cli struct {
	Start startCmd `cmd:"" default:"1" help:"Start the exporter."`
}

func main() {
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

type startCmd struct {
	Config   string          `env:"CONFIG" default:"config.yaml" help:"The path to the configuration file for the server."`
	Log      logger.Config   `json:"log" embed:"" prefix:"log." envprefix:"LOG_"`
	Exporter exporter.Config `json:"exporter" kong:"-"`
	Server   server.Config   `json:"server" embed:"" prefix:"server." envprefix:"SERVER_"`
}

func (r *startCmd) Run() error {
	cfg, err := config.Load(r.Config, *r)
	if err != nil {
		return err
	}

	logger := logger.New(cfg.Log)
	logger.Info("Version information", "version", slog.GroupValue(version.Info()...))
	logger.Info("Build information", "build", slog.GroupValue(version.BuildContext()...))

	exporter, err := exporter.New(cfg.Exporter)
	if err != nil {
		slog.Error("Failed to create exporter", slog.String("error", err.Error()))
		return err
	}
	prometheus.MustRegister(exporter)

	server, err := server.New(cfg.Server)
	if err != nil {
		slog.Error("Failed to create server", slog.String("error", err.Error()))
		return err
	}
	go server.Start()
	defer server.Stop()

	// All components should be terminated gracefully. For that we are listen
	// for the SIGINT and SIGTERM signals and try to gracefully shutdown the
	// started components. This ensures that established connections or tasks
	// are not interrupted.
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	logger.Debug("Start listining for SIGINT and SIGTERM signal")
	<-done
	logger.Info("Shutdown started...")

	return nil
}
