package schema

import (
	modulestypes "github.com/forbole/bdjuno/v4/modules/types"

	parsecmdtypes "github.com/forbole/juno/v4/cmd/parse/types"
	"github.com/forbole/juno/v4/types/config"
	"github.com/spf13/cobra"

	"github.com/forbole/bdjuno/v4/database"
)

// createSchema returns the Cobra command allowing to refresh x/bank total supply
func createSchema(parseConfig *parsecmdtypes.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "schema",
		Short: "Create schema in database",
		RunE: func(cmd *cobra.Command, args []string) error {
			parseCtx, err := parsecmdtypes.GetParserContext(config.Cfg, parseConfig)
			if err != nil {
				return err
			}

			sources, err := modulestypes.BuildSources(config.Cfg.Node, parseCtx.EncodingConfig)
			if err != nil {
				return err
			}

			// Get the database
			db := database.Cast(parseCtx.Database)

			// Create schema inside the database
			err = db.CreateSchemaInsideDatabase()
			if err != nil {
				return err
			}

			return nil
		},
	}
}
