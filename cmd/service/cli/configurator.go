package cli

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

// FromCommand sets values to the global configuration by obtaining values from the command flags.
func FromCommand(cmd *cobra.Command) {
	if verbose := os.Getenv("debug"); verbose == "true" {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("Debug level enabled")
		//config.Get().Verbose = true
	}
}
