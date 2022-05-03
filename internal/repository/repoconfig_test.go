package repository

import (
	"testing"

	. "github.com/go-playground/assert/v2"
	"github.com/go-playground/validator/v10"
)

func TestValidateId(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("repo_id", ValidateId)

	var tests = []struct {
		input string
		valid bool
	}{
		{"", false},
		{"dummy*1", false},
		{"*dummy1", false},
		{"dummy1*", false},
		{"Xdummy1", false},
		{"dummy1X", false},
		{"1dummy", true},
		{"dummy1", true},
		{"dummy-1", true},
		{"dummy_1", true},
	}
	for i, test := range tests {
		errs := validate.Var(test.input, "repo_id")

		if test.valid {
			if !IsEqual(errs, nil) {
				t.Errorf("Test: %d with valid input should not result in error: %s", i, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Errorf("Test: %d with invalid input should result in error", i)
			}
		}
	}
}

func TestValidatePathAbs(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("repo_path", ValidatePathAbs)

	var tests = []struct {
		input string
		valid bool
	}{
		{"", false},
		{"../../share", false},
		{"/var/lib/lagoon", true},
	}
	for i, test := range tests {
		errs := validate.Var(test.input, "repo_path")

		if test.valid {
			if !IsEqual(errs, nil) {
				t.Errorf("Test: %d with valid input should not result in error: %s", i, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Errorf("Test: %d with invalid input should result in error", i)
			}
		}
	}
}

func TestValidateCron(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("repo_cron", ValidateCron)

	var tests = []struct {
		input string
		valid bool
	}{
		{"", false},
		{"incorrect", false},
		{"*/10 * * * * *", true},
	}
	for i, test := range tests {
		errs := validate.Var(test.input, "repo_cron")

		if test.valid {
			if !IsEqual(errs, nil) {
				t.Errorf("Test: %d with valid input should not result in error: %s", i, errs)
			}
		} else {
			if IsEqual(errs, nil) {
				t.Errorf("Test: %d with invalid input should result in error", i)
			}
		}
	}
}
