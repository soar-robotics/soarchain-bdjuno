package create

import (
	parse "github.com/forbole/juno/v4/cmd/parse/types"
	"github.com/spf13/cobra"

	createschema "github.com/forbole/bdjuno/v4/cmd/create/schema"
)

// NewCreateCmd returns the Cobra command allowing to create static data
func NewCreateCmd(parseCfg *parse.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create static data",
	}

	cmd.AddCommand(
		createschema.NewCreateSchemaCmd(parseCfg),
	)

	return cmd
}
