package cli

import (
	"path/filepath"
	"testing"

	viper "github.com/spf13/viper"
)

func TestInitConfigHyphenatedEnvKeys(t *testing.T) {
	t.Setenv("SERVER_URL", "http://example.invalid:9999")

	originalCfgFile := cfgFile
	defer func() { cfgFile = originalCfgFile }()

	cfgFile = filepath.Join(t.TempDir(), "absent.yaml")
	initConfig()

	if got := viper.GetString("server-url"); got != "http://example.invalid:9999" {
		t.Errorf("server-url = %q, want SERVER_URL value", got)
	}
}
