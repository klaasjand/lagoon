package remote

import (
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type DummyRemote struct {
	id   string
	dest string
}

func NewDummyRemote(id string, dest string) *DummyRemote {
	return &DummyRemote{
		id:   id,
		dest: dest,
	}
}

func (r DummyRemote) Init() error {
	return nil
}

func (r DummyRemote) Sync() error {
	rand.Seed(time.Now().UnixNano())
	sleeptime := rand.Intn(30)

	log.Debug().Str("repo", r.id).Int("sleep", sleeptime).Msg("Dummy sync sleeping")

	time.Sleep(time.Duration(sleeptime) * time.Second)

	randErr := func() bool {
		rand.Seed(time.Now().UnixNano())

		return rand.Float32() < 0.5
	}

	if randErr() {
		err := errors.New("dummy sync error")

		log.Error().Stack().Err(err).Str("repo", r.id).Msg("")

		return err
	} else {
		dummyFile, err := os.Create(filepath.Join(r.dest, uuid.New().String()))

		if err != nil {
			return err
		}

		return dummyFile.Close()
	}
}

func (r DummyRemote) Publish(snapshot string) error {
	return nil
}
