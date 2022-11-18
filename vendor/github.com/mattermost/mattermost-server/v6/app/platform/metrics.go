// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package platform

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"runtime"
	"sync"
	"text/template"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mattermost/mattermost-server/v6/einterfaces"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/shared/mlog"
	"github.com/pkg/errors"
)

const TimeToWaitForConnectionsToCloseOnServerShutdown = time.Second

type platformMetrics struct {
	server *http.Server
	router *mux.Router
	lock   sync.Mutex
	logger *mlog.Logger

	metricsImpl einterfaces.MetricsInterface

	cfgFn func() *model.Config
}

// resetMetrics resets the metrics server. Clears the metrics if the metrics are disabled by the config.
func (ps *PlatformService) resetMetrics(metricsImpl einterfaces.MetricsInterface, cfgFn func() *model.Config) error {
	if !*cfgFn().MetricsSettings.Enable {
		if ps.metrics != nil {
			return ps.metrics.stopMetricsServer()
		}
		return nil
	}

	if ps.metrics != nil {
		if err := ps.metrics.stopMetricsServer(); err != nil {
			return err
		}
	}

	ps.metrics = &platformMetrics{
		cfgFn:       cfgFn,
		metricsImpl: metricsImpl,
		logger:      ps.logger,
	}

	if err := ps.metrics.initMetricsRouter(); err != nil {
		return err
	}

	if metricsImpl != nil {
		metricsImpl.Register()
	}

	return ps.metrics.startMetricsServer()
}

func (pm *platformMetrics) stopMetricsServer() error {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	if pm.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), TimeToWaitForConnectionsToCloseOnServerShutdown)
		defer cancel()

		if err := pm.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("could not shutdown metrics server: %v", err)
		}

		pm.logger.Info("Metrics and profiling server is stopped")
	}

	return nil
}

func (pm *platformMetrics) startMetricsServer() error {
	var notify chan struct{}
	pm.lock.Lock()
	defer func() {
		if notify != nil {
			<-notify
		}
		pm.lock.Unlock()
	}()

	l, err := net.Listen("tcp", *pm.cfgFn().MetricsSettings.ListenAddress)
	if err != nil {
		return err
	}

	notify = make(chan struct{})
	pm.server = &http.Server{
		Handler:      handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(pm.router),
		ReadTimeout:  time.Duration(*pm.cfgFn().ServiceSettings.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(*pm.cfgFn().ServiceSettings.WriteTimeout) * time.Second,
	}

	go func() {
		close(notify)
		if err := pm.server.Serve(l); err != nil && err != http.ErrServerClosed {
			pm.logger.Critical(err.Error())
		}
	}()

	pm.logger.Info("Metrics and profiling server is started", mlog.String("address", l.Addr().String()))
	return nil
}

func (pm *platformMetrics) initMetricsRouter() error {
	pm.router = mux.NewRouter()
	runtime.SetBlockProfileRate(*pm.cfgFn().MetricsSettings.BlockProfileRate)

	metricsPage := `
			<html>
				<body>{{if .}}
					<div><a href="/metrics">Metrics</a></div>{{end}}
					<div><a href="/debug/pprof/">Profiling Root</a></div>
					<div><a href="/debug/pprof/cmdline">Profiling Command Line</a></div>
					<div><a href="/debug/pprof/symbol">Profiling Symbols</a></div>
					<div><a href="/debug/pprof/goroutine">Profiling Goroutines</a></div>
					<div><a href="/debug/pprof/heap">Profiling Heap</a></div>
					<div><a href="/debug/pprof/threadcreate">Profiling Threads</a></div>
					<div><a href="/debug/pprof/block">Profiling Blocking</a></div>
					<div><a href="/debug/pprof/trace">Profiling Execution Trace</a></div>
					<div><a href="/debug/pprof/profile">Profiling CPU</a></div>
				</body>
			</html>
		`
	metricsPageTmpl, err := template.New("page").Parse(metricsPage)
	if err != nil {
		return errors.Wrap(err, "failed to create template")
	}

	rootHandler := func(w http.ResponseWriter, r *http.Request) {
		metricsPageTmpl.Execute(w, pm.metricsImpl != nil)
	}

	pm.router.HandleFunc("/", rootHandler)
	pm.router.StrictSlash(true)

	pm.router.Handle("/debug", http.RedirectHandler("/", http.StatusMovedPermanently))
	pm.router.HandleFunc("/debug/pprof/", pprof.Index)
	pm.router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	pm.router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	pm.router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	pm.router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// Manually add support for paths linked to by index page at /debug/pprof/
	pm.router.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	pm.router.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	pm.router.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	pm.router.Handle("/debug/pprof/block", pprof.Handler("block"))

	return nil
}

func (ps *PlatformService) HandleMetrics(route string, h http.Handler) {
	if ps.metrics != nil {
		ps.metrics.router.Handle(route, h)
	}
}

func (ps *PlatformService) RestartMetrics() error {
	return ps.resetMetrics(ps.serviceConfig.Metrics, ps.serviceConfig.ConfigStore.Get)
}

func (ps *PlatformService) Metrics() einterfaces.MetricsInterface {
	if ps.metrics == nil {
		return nil
	}
	return ps.metrics.metricsImpl
}
