package get

import (
	"context"

	"github.com/pkg/errors"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"

	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/pkg/output"
	"github.com/kumahq/kuma/app/kumactl/pkg/output/printers"
	rest_types "github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

func newGetMeshCmd(pctx *getContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mesh NAME",
		Short: "Show a single Mesh resource",
		Long:  `Show a single Mesh resource.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rs, err := pctx.CurrentResourceStore()
			if err != nil {
				return err
			}
			name := args[0]
			mesh := &core_mesh.MeshResource{}
			if err := rs.Get(context.Background(), mesh, store.GetByKey(name, model.NoMesh)); err != nil {
				if store.IsResourceNotFound(err) {
					return errors.Errorf("No resources found in %s mesh", name)
				}
				return errors.Wrapf(err, "failed to get mesh %s", name)
			}
			meshes := &core_mesh.MeshResourceList{
				Items: []*core_mesh.MeshResource{mesh},
			}
			switch format := output.Format(pctx.args.outputFormat); format {
			case output.TableFormat:
				return printMeshes(pctx.Now(), meshes, cmd.OutOrStdout())
			default:
				printer, err := printers.NewGenericPrinter(format)
				if err != nil {
					return err
				}
				return printer.Print(rest_types.From.Resource(mesh), cmd.OutOrStdout())
			}
		},
	}
	return cmd
}
