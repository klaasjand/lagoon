package config

import (
	"os"
	"testing"
)

func TestLoadConfigNoFile(t *testing.T) {
	if err := LoadConfig(); err == nil {
		t.Errorf("Non existing config file should result in error: %v", err)
	}
}

func TestLoadConfigEmptyFile(t *testing.T) {
	defer removeConfigFile()

	if err := writeConfigFile(""); err == nil {
		if err := LoadConfig(); err == nil {
			t.Errorf("Empty config file should result in error: %v", err)
		}
	} else {
		t.Errorf("Cannot write config file %v", err)
	}
}

func TestLoadConfigIncorrectFile(t *testing.T) {
	defer removeConfigFile()

	if err := writeConfigFile("incorrect"); err == nil {
		if err := LoadConfig(); err == nil {
			t.Errorf("Config file with incorrect content should result in error: %v", err)
		}
	} else {
		t.Errorf("Cannot write config file %v", err)
	}
}

func TestLoadConfigMissingAttr(t *testing.T) {
	defer removeConfigFile()

	config := `
---
repositories:
  - id: dummy1
    name: Dummy Mirror - 1
    type: dummy
    src: None
    dest: /var/lib/lagoon
    cron: "0 0 12-14 ? * *"
    #snapshots: 52
`

	if err := writeConfigFile(config); err == nil {
		if err := LoadConfig(); err == nil {
			t.Errorf("Config file with missing attribute should result in error; %v", err)
		}
	} else {
		t.Errorf("Cannot write config file %v", err)
	}
}

func TestLoadConfigInvalidChars(t *testing.T) {
	defer removeConfigFile()

	config := `
---
repositories:
  - id: dummy1
	name: Dummy Mirror - 1
    type: dummy
    src: None
    dest: /var/lib/lagoon
    cron: "0 0 12-14 ? * *"
    snapshots: 52
`

	if err := writeConfigFile(config); err == nil {
		if err := LoadConfig(); err == nil {
			t.Errorf("Config file with invalid content should result in error; %v", err)
		}
	} else {
		t.Errorf("Cannot write config file %v", err)
	}
}

func TestLoadConfigCorrectFile(t *testing.T) {
	defer removeConfigFile()

	config := `
---
repositories:
  - id: dummy1
    name: Dummy Mirror - 1
    type: dummy
    src: None
    dest: /var/lib/lagoon
    cron: "0 0 12-14 ? * *"
    snapshots: 52
`

	if err := writeConfigFile(config); err == nil {
		if err := LoadConfig(); err != nil || len(RepoConfigs) != 1 {
			t.Errorf("Config file with correct content should not result in error; %v", err)
		}
	} else {
		t.Errorf("Cannot write config file %v", err)
	}
}

func writeConfigFile(cfg string) error {
	content := []byte(cfg)

	return os.WriteFile("lagoon.yml", content, 0644)
}

func removeConfigFile() {
	os.Remove("lagoon.yml")
}
