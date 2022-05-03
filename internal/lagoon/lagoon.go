package lagoon

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/klaasjand/lagoon-dev/internal/config"
	"github.com/klaasjand/lagoon-dev/internal/repository"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

var (
	repos []*repository.Repo
)

func Run() {
	initLogger()

	log.Info().Msg("Starting Lagoon")

	if err := config.LoadConfig(); err != nil {
		log.Fatal().Stack().Err(err).Msg("Unable to load configuration")
	}

	var wg sync.WaitGroup

	names := make(map[string]bool)

	for _, rc := range config.RepoConfigs {
		if _, value := names[rc.Id]; !value {
			names[rc.Id] = true

			if m, err := repository.NewRepo(rc, &wg); err == nil {
				repos = append(repos, m)
			} else {
				log.Warn().Err(err).Msg("Cannot add repo")
			}
		} else {
			log.Fatal().Msgf("Repository id must be unique, found duplicate entry for: %s", rc.Id)
		}
	}

	log.Info().Msg("Running preflight checks")
	for _, m := range repos {
		if err := m.PreReqs(); err != nil {
			log.Fatal().Stack().Err(err).Msg("")
		}
	}

	c := cron.New(cron.WithParser(cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)))

	for _, m := range repos {
		c.AddFunc(m.GetCronSpec(), m.Sync)
	}

	c.Start()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Debug().Msg(sig.String())
		done <- true
	}()

	// Start http server and handle metrics scraping
	httpServer := &http.Server{Addr: ":9000"}
	http.Handle("/metrics", promhttp.Handler())

	// go httpServer.ListenAndServe()
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			log.Error().Stack().Err(err).Msg("")
		}
	}()

	log.Info().Msg("Lagoon running")
	<-done

	c.Stop() // Stop cron scheduler

	log.Info().Msg("Waiting for running sync jobs to exit gracefully")
	wg.Wait()

	// Wait for ListenAndServe goroutine to close.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error().Stack().Err(err).Msg("")
	}

	log.Info().Msg("Exiting")
}

func initLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	debug := flag.Bool("d", false, "sets log level to debug")
	human := flag.Bool("h", false, "sets log output to human readable")

	flag.Parse()

	// Use -d to enable debug logging
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Use -h to enable human readable logging
	if *human {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02T15:04:05.999Z07:00"}).With().Caller().Logger()
	}
}
