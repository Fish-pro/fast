package version

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/fast-io/fast/pkg/version"
)

// NewCmdVersion prints out the release version info for this command binary.
// It is used as a subcommand of a parent command.
func NewCmdVersion(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Print the version information",
		Long:    "Print the version information",
		Example: fmt.Sprintf("    %s version", name),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(os.Stdout, "%s version: %s\n", name, version.Get())
		},
	}
	return cmd
}
