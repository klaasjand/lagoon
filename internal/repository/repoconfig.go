package repository

import (
	"path/filepath"
	"regexp"

	"github.com/go-playground/validator/v10"

	"github.com/robfig/cron/v3"
)

type RepoConfig struct {
	Id        string   `yaml:"id" validate:"repo_id"`
	Name      string   `yaml:"name"`
	Type      string   `yaml:"type" validate:"oneof=dummy reposync rsync"`
	Src       string   `yaml:"src"`
	Dest      string   `yaml:"dest" validate:"repo_path"`
	Cron      string   `yaml:"cron" validate:"repo_cron"`
	Exclude   []string `yaml:"exclude"`
	Snapshots int      `yaml:"snapshots" validate:"min=1,max=1024"`
}

func ValidateId(fl validator.FieldLevel) bool {
	r, _ := regexp.Compile(`([a-z0-9_-]+)`)

	matches := r.FindAllString(fl.Field().String(), -1)
	if len(matches) == 1 && len(matches[0]) == len(fl.Field().String()) {
		return true
	} else {
		return false
	}
}

func ValidatePathAbs(fl validator.FieldLevel) bool {
	return filepath.IsAbs(fl.Field().String())
}

func ValidateCron(fl validator.FieldLevel) bool {
	p := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	if _, err := p.Parse(fl.Field().String()); err != nil {
		return false
	}

	return true
}
