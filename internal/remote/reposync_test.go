package remote

import (
	"testing"

	. "github.com/go-playground/assert/v2"
)

func TestGetRepoId(t *testing.T) {
	var tests = []struct {
		input string
		valid bool
	}{
		{"", false},
		{"[[repoid]]", false},
		{"repoid]", false},
		{"[repoid", false},
		{"repoid", false},
		{`# Mirror Comment
		[Dummy-1.x]
		name=Dummy Mirror - 1`, false},

		{"[repoid]", true},
		{"[RepoId]", true},
		{"[repo-id]", true},
		{"[repo_id]", true},
		{"[repo_id_1]", true},
		{`[Dummy-1.x]
		name=Dummy Mirror - 1`, true},
	}
	for i, test := range tests {
		_, err := getRepoId(test.input)

		if test.valid {
			if !IsEqual(err, nil) {
				t.Errorf("Test: %d with valid input should not result in error: %s", i, err)
			}
		} else {
			if IsEqual(err, nil) {
				t.Errorf("Test: %d with invalid input should result in error", i)
			}
		}
	}
}
