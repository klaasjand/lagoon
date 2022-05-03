package remote

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type RsyncRemote struct {
	id       string
	src      string
	dest     string
	excludes []string
}

func NewRsyncRemote(id string, src string, dest string, excludes []string) *RsyncRemote {
	return &RsyncRemote{
		id:       id,
		src:      src,
		dest:     dest,
		excludes: excludes,
	}
}

func (r RsyncRemote) Init() error {
	if _, err := exec.LookPath("rsync"); err != nil {
		return errors.Errorf(fmtErrPreFlight, r.id, err)
	}

	if !isRsyncUrl(r.src) {
		return errors.New("incorrect rsync url")
	}

	return nil
}

func (r RsyncRemote) Sync() error {
	// TODO: Factor out I/O related code to add unittests
	if isRsyncUrl(r.src) {
		// NOTE: Somehow pattern --exclude={'file1.txt','dir1/*','dir2'} or --exclude={file1.txt,dir1/*,dir2} does not work, using separate excludes for now
		args := []string{"-avSHP", "--delete"}
		for _, e := range r.excludes {
			args = append(append(args, "--exclude"), e)
		}
		args = append(append(args, r.src), r.dest)

		cmd := exec.Command("rsync", args...)

		log.Debug().Str("repo", r.id).Str("command", cmd.String()).Msg("Executing rsync")

		if err := cmd.Run(); err != nil {
			log.Error().Stack().Err(err).Str("repo", r.id).Msg("")

			return err
		}

		return nil
	} else {
		return errors.New("incorrect rsync url")
	}
}

func (r RsyncRemote) Publish(snapshot string) error {
	return nil
}

func isRsyncUrl(src string) bool {
	return strings.HasPrefix(src, "rsync://")
}
