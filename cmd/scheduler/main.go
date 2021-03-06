package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/figment-networks/indexing-engine/health"
	"github.com/figment-networks/indexing-engine/health/database/postgreshealth"
	"github.com/figment-networks/indexing-engine/metrics"
	"github.com/figment-networks/indexing-engine/metrics/prometheusmetrics"
	"go.uber.org/zap"

	"github.com/figment-networks/indexer-scheduler/cmd/scheduler/config"
	"github.com/figment-networks/indexer-scheduler/conn/tray"
	"github.com/figment-networks/indexer-scheduler/core"
	"github.com/figment-networks/indexer-scheduler/destination"
	"github.com/figment-networks/indexer-scheduler/http/auth"
	"github.com/figment-networks/indexer-scheduler/persistence"
	"github.com/figment-networks/indexer-scheduler/persistence/postgresstore"
	"github.com/figment-networks/indexer-scheduler/process"
	"github.com/figment-networks/indexer-scheduler/ui"

	"github.com/figment-networks/indexer-scheduler/runner/lastdata"
	runnerPersistence "github.com/figment-networks/indexer-scheduler/runner/lastdata/persistence"
	runnerDatabase "github.com/figment-networks/indexer-scheduler/runner/lastdata/persistence/postgresstore"
	runnerHTTP "github.com/figment-networks/indexer-scheduler/runner/lastdata/transport/http"
	runnerWS "github.com/figment-networks/indexer-scheduler/runner/lastdata/transport/ws"

	"github.com/figment-networks/indexer-scheduler/runner/syncrange"
	runnerSyncrangePersistence "github.com/figment-networks/indexer-scheduler/runner/syncrange/persistence"
	runnerSyncrangeDatabase "github.com/figment-networks/indexer-scheduler/runner/syncrange/persistence/postgresstore"
	runnerSyncrangeHTTP "github.com/figment-networks/indexer-scheduler/runner/syncrange/transport/http"
	runnerSyncrangeWS "github.com/figment-networks/indexer-scheduler/runner/syncrange/transport/ws"

	"github.com/figment-networks/indexer-scheduler/structures"

	_ "github.com/lib/pq"
)

type flags struct {
	configPath  string
	showVersion bool
}

var configFlags = flags{}

func init() {
	flag.BoolVar(&configFlags.showVersion, "v", false, "Show application version")
	flag.StringVar(&configFlags.configPath, "config", "", "Path to config")
	flag.Parse()
}

func main() {

	ctx := context.Background()

	// Initialize configuration
	cfg, err := initConfig(configFlags.configPath)
	if err != nil {
		log.Fatal(fmt.Errorf("error initializing config [ERR: %+v]", err))
	}

	if cfg.RollbarServerRoot == "" {
		cfg.RollbarServerRoot = "github.com/figment-networks/indexer-scheduler"
	}

	rcfg := &config.RollbarConfig{
		AppEnv:             cfg.AppEnv,
		RollbarAccessToken: cfg.RollbarAccessToken,
		RollbarServerRoot:  cfg.RollbarServerRoot,
		Version:            config.GitSHA,
	}

	if cfg.AppEnv == "development" || cfg.AppEnv == "local" {
		config.InitLogger("console", "debug", []string{"stderr"}, rcfg)
	} else {
		config.InitLogger("json", "info", []string{"stderr"}, rcfg)
	}
	config.Info(config.IdentityString())

	logger := config.GetLogger()
	defer logger.Sync()

	// setup metrics
	prom := prometheusmetrics.New()
	err = metrics.AddEngine(prom)
	if err != nil {
		logger.Fatal("Error running prometheus ", zap.Error(err))
		return

	}
	err = metrics.Hotload(prom.Name())
	if err != nil {
		logger.Fatal("Error running prometheus ", zap.Error(err))
		return
	}

	// connect to database
	logger.Info("[DB] Connecting to database...")
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Error connecting to database", zap.Error(err))
		return
	}

	if err := db.PingContext(ctx); err != nil {
		logger.Fatal("Error connecting to database (ping)", zap.Error(err))
		return
	}
	logger.Info("[DB] Ping successfull...")
	defer db.Close()

	mux := http.NewServeMux()

	attachDynamic(ctx, mux)

	dbMonitor := postgreshealth.NewPostgresMonitorWithMetrics(db, logger)
	monitor := &health.Monitor{}
	monitor.AddProber(ctx, dbMonitor)
	go monitor.RunChecks(ctx, cfg.HealthCheckInterval)
	monitor.AttachHttp(mux)

	attachProfiling(mux)

	logger.Info("[Scheduler] Adding scheduler...")

	d := postgresstore.NewDriver(db)
	sch := process.NewScheduler(logger, d)

	cStore := &persistence.CoreStorage{Driver: d}

	creds := auth.AuthCredentials{
		User:     cfg.AuthUser,
		Password: cfg.AuthPassword,
	}

	c := core.NewCore(cStore, sch, creds, logger)
	c.RegisterHandles(mux)
	scheme := destination.NewScheme(logger, creds)
	scheme.RegisterHandles(mux)

	connTray := tray.NewConnTray(logger)
	cont := destination.NewContainer(logger)

	if err := c.InitialLoad(ctx); err != nil {
		logger.Error("[Scheduler] Error during initial load of scheduler", zap.Error(err))
		logger.Sync()
	}
	if cfg.SchedulesConfig != "" {
		logger.Info("[Scheduler] Loading schedule initial config")

		files, err := ioutil.ReadDir(cfg.SchedulesConfig)
		if err != nil {
			logger.Fatal("Error reading scheduling config dir", zap.Error(err))
			return
		}

		rcs := []structures.RunConfig{}
		for _, fileInfo := range files {
			if fileInfo.IsDir() {
				continue
			}

			file, err := os.Open(path.Join(cfg.SchedulesConfig, fileInfo.Name()))
			if err != nil {
				logger.Fatal("Error reading config file", zap.Error(err), zap.String("filepath", path.Join(cfg.SchedulesConfig, fileInfo.Name())))
				return
			}

			rcp := []structures.RunConfigParams{}
			dec := json.NewDecoder(file)
			err = dec.Decode(&rcp)
			file.Close()
			if err != nil {
				logger.Fatal("Error reading config file (decode)", zap.Error(err), zap.String("filepath", path.Join(cfg.SchedulesConfig, fileInfo.Name())))
				return
			}

			for _, rConf := range rcp {
				duration, err := time.ParseDuration(rConf.Interval)
				if err != nil {
					logger.Fatal("Error parsing duration ", zap.Error(err), zap.String("filepath", path.Join(cfg.SchedulesConfig, fileInfo.Name())))
					return
				}
				rcs = append(rcs, structures.RunConfig{
					Network:  rConf.Network,
					ChainID:  rConf.ChainID,
					Kind:     rConf.Kind,
					TaskID:   rConf.TaskID,
					Version:  "0.0.1",
					Duration: duration,
				})
			}
		}

		if err := c.AddSchedules(ctx, rcs); err != nil {
			logger.Fatal("Error adding schedules", zap.Error(err))
			return
		}
	}

	if cfg.DestinationsValue != "" {
		logger.Info("[Scheduler] Loading destinations initial config from env var")

		trgts := []structures.TargetConfig{}
		dec := json.NewDecoder(strings.NewReader(cfg.DestinationsValue))
		err = dec.Decode(&trgts)
		if err != nil {
			logger.Fatal("Error reading config from env (decode)", zap.Error(err), zap.String("cfg", cfg.DestinationsValue))
			return
		}

		for _, trgt := range trgts {
			err := cont.Add(ctx, trgt, connTray, scheme)
			if err != nil {
				logger.Error("Error adding destination", zap.Error(err))
				return
			}
		}

	} else if cfg.DestinationsConfig != "" {
		logger.Info("[Scheduler] Loading destinations initial config from path")

		files, err := ioutil.ReadDir(cfg.DestinationsConfig)
		if err != nil {
			logger.Fatal("Error reading scheduling config dir", zap.Error(err))
			return
		}

		trgts := []structures.TargetConfig{}
		for _, fileInfo := range files {
			if fileInfo.IsDir() {
				continue
			}

			file, err := os.Open(path.Join(cfg.DestinationsConfig, fileInfo.Name()))
			if err != nil {
				logger.Fatal("Error reading config file", zap.Error(err), zap.String("filepath", path.Join(cfg.DestinationsConfig, fileInfo.Name())))
				return
			}

			dec := json.NewDecoder(file)
			err = dec.Decode(&trgts)
			file.Close()
			if err != nil {
				logger.Fatal("Error reading config file (decode)", zap.Error(err), zap.String("filepath", path.Join(cfg.DestinationsConfig, fileInfo.Name())))
				return
			}

			for _, trgt := range trgts {
				err := cont.Add(ctx, trgt, connTray, scheme)
				if err != nil {
					logger.Error("Error adding destination", zap.Error(err))
					return
				}
			}
		}
	}

	logger.Info("[Scheduler] Running Load")
	go reloadScheduler(ctx, logger, c)

	mux.Handle("/metrics", metrics.Handler())

	pStore := runnerPersistence.NewLastDataStorageTransport(runnerDatabase.NewDriver(db))

	lh := lastdata.NewClient(logger, pStore, creds, scheme)
	rHTTP := runnerHTTP.NewLastDataHTTPTransport(logger)
	lh.AddTransport(runnerHTTP.ConnectionTypeHTTP, rHTTP)
	rWS := runnerWS.NewLastDataWSTransport(logger, connTray)
	lh.AddTransport(runnerWS.ConnectionTypeWS, rWS)
	lh.RegisterHandles(mux)

	pSRStore := runnerSyncrangePersistence.NewLastDataStorageTransport(runnerSyncrangeDatabase.NewDriver(db))
	sr := syncrange.NewClient(logger, pSRStore, creds, scheme)
	rsSRHTTP := runnerSyncrangeHTTP.NewSyncrangeHTTPTransport(logger)
	sr.AddTransport(runnerSyncrangeHTTP.ConnectionTypeHTTP, rsSRHTTP)

	rsWS := runnerSyncrangeWS.NewSyncRangeWSTransport(logger, connTray)
	sr.AddTransport(runnerWS.ConnectionTypeWS, rsWS)
	sr.RegisterHandles(mux)

	c.LoadRunner(lastdata.RunnerName, lh)
	c.LoadRunner(syncrange.RunnerName, sr)

	uInterface := ui.NewUI()
	uInterface.RegisterHandles(mux)

	s := &http.Server{
		Addr:    cfg.Address,
		Handler: mux,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	osSig := make(chan os.Signal)
	exit := make(chan string, 2)
	signal.Notify(osSig, syscall.SIGTERM)
	signal.Notify(osSig, syscall.SIGINT)

	go runHTTP(s, cfg.Address, logger, exit)

RunLoop:
	for {
		select {
		case <-osSig:
			s.Shutdown(ctx)
			break RunLoop
		case <-exit:
			break RunLoop
		}
	}
}

func initConfig(path string) (config.Config, error) {
	cfg := &config.Config{}

	if path != "" {
		if err := config.FromFile(path, cfg); err != nil {
			return *cfg, err
		}
	}

	if cfg.DatabaseURL != "" {
		return *cfg, nil
	}

	if err := config.FromEnv(cfg); err != nil {
		return *cfg, err
	}

	return *cfg, nil
}

func runHTTP(s *http.Server, address string, logger *zap.Logger, exit chan<- string) {
	defer logger.Sync()

	logger.Info(fmt.Sprintf("[HTTP] Listening on %s", address))

	if err := s.ListenAndServe(); err != nil {
		logger.Error("[HTTP] failed to listen", zap.Error(err))
	}
	exit <- "http"
}

func reloadScheduler(ctx context.Context, logger *zap.Logger, c *core.Core) {
	tckr := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-tckr.C:
			if err := c.LoadScheduler(ctx); err != nil {
				logger.Error("[Scheduler] Error during loading of scheduler", zap.Error(err))
				logger.Sync()
			}
		}
	}
}
