package install

import (
	"github.com/spf13/cobra"

	kumactl_cmd "github.com/kumahq/kuma/app/kumactl/pkg/cmd"
)

func NewInstallCmd(pctx *kumactl_cmd.RootContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Kuma on Kubernetes",
		Long:  `Install Kuma on Kubernetes.`,
	}
	// sub-commands
	cmd.AddCommand(newInstallControlPlaneCmd(pctx))
	cmd.AddCommand(newInstallMetrics(pctx))
	cmd.AddCommand(newInstallTracing())
	cmd.AddCommand(newInstallDNS())
	cmd.AddCommand(newInstallLogging())
	return cmd
}
