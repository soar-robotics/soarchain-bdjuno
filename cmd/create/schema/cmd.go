package schema

import (
	parsecmdtypes "github.com/forbole/juno/v4/cmd/parse/types"
	"github.com/spf13/cobra"
)

// NewCreateSchemaCmd returns the Cobra command allowing to create schema 
// from SQL files inside the database
func NewCreateSchemaCmd(parseConfig *parsecmdtypes.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "database",
		Short: "Create database schema",
	}

	cmd.AddCommand(
		createSchema(parseConfig),
	)

	return cmd
}
