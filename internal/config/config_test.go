package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExposureMetadataPathResolvesRelativeToConfigFile(t *testing.T) {
	configPath = ""
	t.Setenv("PAISA_EXPOSURE_METADATA_FILE", "")
	configFile := filepath.Join(t.TempDir(), "paisa.yaml")

	require.NoError(t, LoadConfig([]byte("journal_path: main.ledger\ndb_path: paisa.db\nexposure_metadata_file: exposure.json\n"), configFile))
	require.Equal(t, filepath.Join(filepath.Dir(configFile), "exposure.json"), GetExposureMetadataPath())
}

func TestExposureMetadataEnvironmentOverrideWins(t *testing.T) {
	configPath = ""
	configFile := filepath.Join(t.TempDir(), "paisa.yaml")
	environmentPath := filepath.Join(t.TempDir(), "mounted-exposure.json")

	require.NoError(t, LoadConfig([]byte("journal_path: main.ledger\ndb_path: paisa.db\nexposure_metadata_file: exposure.json\n"), configFile))
	t.Setenv("PAISA_EXPOSURE_METADATA_FILE", environmentPath)

	require.Equal(t, environmentPath, GetExposureMetadataPath())
}
