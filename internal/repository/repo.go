package repository

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/google/uuid"
	"github.com/klaasjand/lagoon/internal/remote"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
)

const fmtSnapshotLayout = "20060102"

// Matches a date in YYYYMMDD format from 19000101 through 20991231
const fmtSnapshotPattern = `(19|20)\d\d(0[1-9]|1[012])(0[1-9]|[12][0-9]|3[01])`

const maxSyncRetries = 5

type RepoMetrics struct {
	SyncTotal    prometheus.Counter
	SyncDuration prometheus.Gauge
}

type Repo struct {
	config    RepoConfig
	waitGroup *sync.WaitGroup
	isRunning bool
	metrics   *RepoMetrics
	usPath    string
	saPath    string
	pubPath   string
	remote    remote.Remote
}

func NewRepo(cfg RepoConfig, wg *sync.WaitGroup) (*Repo, error) {
	syncTotal := promauto.NewCounter(prometheus.CounterOpts{
		Name:        "lagoon_sync_total",
		Help:        "The total number of repo syncs",
		ConstLabels: prometheus.Labels{"repo": cfg.Id, "name": cfg.Name},
	})

	syncDuration := promauto.NewGauge(prometheus.GaugeOpts{
		Name:        "lagoon_sync_duration_seconds",
		Help:        "The sync duration",
		ConstLabels: prometheus.Labels{"repo": cfg.Id, "name": cfg.Name},
	})

	metrics := &RepoMetrics{SyncTotal: syncTotal, SyncDuration: syncDuration}

	switch cfg.Type {
	case "dummy":
		return &Repo{
			config:    cfg,
			waitGroup: wg,
			metrics:   metrics,
			usPath:    getUpstreamPath(cfg.Id, cfg.Dest),
			saPath:    getStagingPath(cfg.Id, cfg.Dest),
			pubPath:   getPublicPath(cfg.Id, cfg.Dest),
			remote:    remote.NewDummyRemote(cfg.Id, getUpstreamPath(cfg.Id, cfg.Dest)),
		}, nil
	case "rsync":
		return &Repo{
			config:    cfg,
			waitGroup: wg,
			metrics:   metrics,
			usPath:    getUpstreamPath(cfg.Id, cfg.Dest),
			saPath:    getStagingPath(cfg.Id, cfg.Dest),
			pubPath:   getPublicPath(cfg.Id, cfg.Dest),
			remote:    remote.NewRsyncRemote(cfg.Id, cfg.Src, getUpstreamPath(cfg.Id, cfg.Dest), cfg.Exclude),
		}, nil
	case "reposync":
		return &Repo{
			config:    cfg,
			waitGroup: wg,
			metrics:   metrics,
			usPath:    getUpstreamPath(cfg.Id, cfg.Dest),
			saPath:    getStagingPath(cfg.Id, cfg.Dest),
			pubPath:   getPublicPath(cfg.Id, cfg.Dest),
			remote:    remote.NewRepoSyncRemote(cfg.Id, cfg.Src, getUpstreamPath(cfg.Id, cfg.Dest), getStagingPath(cfg.Id, cfg.Dest)),
		}, nil
	default:
		return nil, fmt.Errorf("unknown repo type '%s'", cfg.Type)
	}
}

func (m Repo) GetCronSpec() string {
	return m.config.Cron
}

func getUpstreamPath(id string, dest string) string {
	return fmt.Sprintf("%s/upstream/%s/", dest, id)
}

func getStagingPath(id string, dest string) string {
	return fmt.Sprintf("%s/staging/%s/", dest, id)
}

func getPublicPath(id string, dest string) string {
	return fmt.Sprintf("%s/public/%s/", dest, id)
}

func (m Repo) PreReqs() error {
	if err := os.MkdirAll(m.usPath, 0755); err != nil {
		return errors.Errorf("destination %s", err)
	}

	if err := os.MkdirAll(m.saPath, 0755); err != nil {
		return errors.Errorf("destination %s", err)
	}

	if err := os.MkdirAll(m.pubPath, 0755); err != nil {
		return errors.Errorf("destination %s", err)
	}

	if err := m.remote.Init(); err != nil {
		return err
	}

	log.Debug().Msg("All prerequisite checks and actions succeeded")

	return nil
}

func (m *Repo) Sync() {
	if !m.isRunning {
		m.waitGroup.Add(1)
		m.isRunning = true

		jobId := uuid.New()
		syncLog := log.With().Str("repo", m.config.Id).Str("jobid", jobId.String()).Logger()

		syncLog.Info().Msg("Starting sync")

		startTime := time.Now()

		// TODO: Add health check before starting sync action
		syncBackOff := backoff.NewExponentialBackOff()
		syncBackOff.InitialInterval = 30 * time.Second
		syncBackOff.MaxInterval = 5 * time.Minute
		syncBackOff.Multiplier = 1.7

		notify := func(err error, t time.Duration) {
			syncLog.Warn().Msgf("Error while synchronizing, retrying in %v", t)
		}

		err := backoff.RetryNotify(m.remote.Sync, backoff.WithMaxRetries(syncBackOff, maxSyncRetries), notify)
		if err != nil {
			syncLog.Error().Stack().Err(err).Msg("Stopped retrying sync")
		} else {
			syncLog.Debug().Msg("Successful sync")

			if snapshot, err := m.createSnapshot(); err == nil {
				if err := m.publishSnapshot(snapshot); err != nil {
					syncLog.Error().Stack().Err(err).Msg("")
				}
			} else {
				syncLog.Error().Stack().Err(err).Msg("")
			}

			if err := m.cleanupSnapshots(); err != nil {
				syncLog.Error().Stack().Err(err).Msg("")
			}

			endTime := time.Now()
			m.metrics.SyncDuration.Set(endTime.Sub(startTime).Seconds())

			m.metrics.SyncTotal.Inc()
		}

		syncLog.Info().Msg("Exiting sync")

		m.isRunning = false
		m.waitGroup.Done()
	} else {
		log.Info().Str("repo", m.config.Id).Msg("Not starting sync; sync already in progress")
	}
}

func (m Repo) createSnapshot() (string, error) {
	// TODO: Add fs check to preflight checks in order to check if fs supports hardlinks
	snapshot := time.Now().Format(fmtSnapshotLayout)
	snapPath := filepath.Join(m.saPath, snapshot)

	if _, err := os.Stat(snapPath); os.IsNotExist(err) {
		cmd := exec.Command("cp", "-al", m.usPath, snapPath)
		if err := cmd.Run(); err != nil {
			return "", err
		}

		log.Info().Str("repo", m.config.Id).Str("snapshot", snapPath).Msg("Created snapshot")
	} else {
		return "", errors.Errorf("snapshot %v already exists", snapPath)
	}

	return snapshot, nil
}

func (m Repo) publishSnapshot(snapshot string) error {
	var err error

	if err = m.remote.Publish(snapshot); err == nil {
		snapPath := filepath.Join(m.saPath, snapshot)

		if _, err = os.Stat(snapPath); err == nil {
			if err = os.Symlink(snapPath, filepath.Join(m.pubPath, snapshot)); err == nil {
				log.Info().Str("repo", m.config.Id).Str("snapshot", snapPath).Msg("Published snapshot")
				// New snapshot is published, now publish it as latest
				if err = m.unPublishSnapshot("latest"); err == nil {
					log.Info().Str("repo", m.config.Id).Msg("Publishing latest snapshot")

					return os.Symlink(snapPath, filepath.Join(m.pubPath, "latest"))
				}
			}
		}
	}

	return err
}

func (m Repo) unPublishSnapshot(snapshot string) error {
	snapPath := filepath.Join(m.pubPath, snapshot)

	if _, err := os.Stat(snapPath); err == nil {
		log.Info().Str("repo", m.config.Id).Str("snapshot", snapPath).Msg("Removing published snapshot")

		return os.Remove(snapPath)
	}

	return nil
}

func (m Repo) cleanupSnapshots() error {
	var err error
	var fileInfo []fs.FileInfo

	if fileInfo, err = ioutil.ReadDir(getStagingPath(m.config.Id, m.config.Dest)); err == nil {
		// ReadDir already sorts by name, so no need to do it afterwards
		allSnapshots := []string{}

		r, _ := regexp.Compile(fmtSnapshotPattern)

		for _, f := range fileInfo {
			if f.IsDir() {
				if r.MatchString(f.Name()) {
					allSnapshots = append(allSnapshots, f.Name())
				}
			}
		}

		if len(allSnapshots) > m.config.Snapshots {
			remSnapshots := allSnapshots[:len(allSnapshots)-m.config.Snapshots]

			errs := false
			for _, s := range remSnapshots {
				if err = m.unPublishSnapshot(s); err == nil {
					snapPath := filepath.Join(m.saPath, s)
					log.Info().Str("repo", m.config.Id).Str("snapshot", snapPath).Msg("Removing staged snapshot")

					if err = os.RemoveAll(snapPath); err != nil {
						errs = true
					}
				} else {
					break
				}
			}

			if errs {
				err = errors.New("problems encountered while cleaning up snapshots")
			}
		} else {
			log.Info().Str("repo", m.config.Id).Msg("No snapshots to remove")
		}
	}

	return err
}
