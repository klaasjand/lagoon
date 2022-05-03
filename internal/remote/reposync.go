package remote

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type RepoSyncRemote struct {
	id     string
	src    string
	usPath string
	saPath string
}

func NewRepoSyncRemote(id string, src string, usPath string, saPath string) *RepoSyncRemote {
	return &RepoSyncRemote{
		id:     id,
		src:    src,
		usPath: usPath,
		saPath: saPath,
	}
}

func (r RepoSyncRemote) Init() error {
	if _, err := exec.LookPath("reposync"); err != nil {
		return errors.Errorf(fmtErrPreFlight, r.id, err)
	}

	if _, err := exec.LookPath("createrepo"); err != nil {
		return errors.Errorf(fmtErrPreFlight, r.id, err)
	}

	// TODO: Path should be configurable RHEL family is using /etc/yum.repos.d/ use github.com/zcalusic/sysinfo to detect distro
	if err := os.WriteFile(fmt.Sprintf("/etc/yum/repos.d/%s.repo", strings.ToLower(r.id)), []byte(r.src), 0644); err != nil {
		return errors.Errorf(fmtErrPreFlight, r.id, err)
	}

	return nil
}

func (r RepoSyncRemote) Sync() error {
	if repoId, err := getRepoId(r.src); err == nil {
		cmd := exec.Command("reposync", "--delete", fmt.Sprintf("--repoid=%s", repoId), "--norepopath", fmt.Sprintf("--download_path=%s", r.usPath), "--downloadcomps", "--download-metadata")

		log.Debug().Str("repo", r.id).Str("command", cmd.String()).Msg("Executing reposync")

		// TODO: Add repomanage to cleanup old packages?
		return cmd.Run()
	} else {
		return err
	}
}

func (r RepoSyncRemote) Publish(snapshot string) error {
	var cmd *exec.Cmd

	snapPath := filepath.Join(r.saPath, snapshot)
	compsPath := filepath.Join(snapPath, "comps.xml")

	if _, err := os.Stat(compsPath); err == nil {
		log.Debug().Str("repo", r.id).Msg("Groupdata found")

		cmd = exec.Command("createrepo", "--update", "-p", "--workers", "2", "-g", compsPath, snapPath)
	} else {
		log.Debug().Str("repo", r.id).Msg("Groupdata not found")

		cmd = exec.Command("createrepo", "--update", "-p", "--workers", "2", snapPath)
	}

	// TODO: Implement errata support

	return cmd.Run()
}

func getRepoId(src string) (string, error) {
	srcLines := strings.Split(src, "\n")

	if len(srcLines) > 0 {
		r, _ := regexp.Compile(`^\[([A-Za-z0-9-_\.]+)\]$`) // Match text, numbers and - or _ or . between []

		m := r.FindStringSubmatch(srcLines[0])

		if len(m) == 2 {
			return m[1], nil
		}
	}

	return "", errors.New("unable to find repoid")
}
