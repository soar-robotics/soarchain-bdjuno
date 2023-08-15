package gov

import (
	"fmt"
	"strconv"
	"time"

	modulestypes "github.com/forbole/bdjuno/v4/modules/types"

	"github.com/forbole/bdjuno/v4/database"
	"github.com/forbole/bdjuno/v4/modules/distribution"
	"github.com/forbole/bdjuno/v4/modules/gov"
	"github.com/forbole/bdjuno/v4/modules/mint"
	"github.com/forbole/bdjuno/v4/modules/slashing"
	"github.com/forbole/bdjuno/v4/modules/staking"
	parsecmdtypes "github.com/forbole/juno/v4/cmd/parse/types"
	"github.com/forbole/juno/v4/types/config"
	"github.com/spf13/cobra"
)

// proposalVotesCmd returns the Cobra command allowing to fix all things related to a proposal
func proposalVotesCmd(parseConfig *parsecmdtypes.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "proposal-votes [id]",
		Short: "Get the votes of a proposal given its id",
		RunE: func(cmd *cobra.Command, args []string) error {
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

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

			// Build expected modules of gov modules for handleParamChangeProposal
			distrModule := distribution.NewModule(sources.DistrSource, parseCtx.EncodingConfig.Codec, db)
			mintModule := mint.NewModule(sources.MintSource, parseCtx.EncodingConfig.Codec, db)
			slashingModule := slashing.NewModule(sources.SlashingSource, parseCtx.EncodingConfig.Codec, db)
			stakingModule := staking.NewModule(sources.StakingSource, parseCtx.EncodingConfig.Codec, db)

			// Build the gov module
			govModule := gov.NewModule(sources.GovSource, nil, distrModule, mintModule, slashingModule, stakingModule, parseCtx.EncodingConfig.Codec, db)

			err = refreshProposalVotes(parseCtx, proposalID, govModule)
			if err != nil {
				return err
			}

			// Update the proposal to the latest status
			height, err := parseCtx.Node.LatestHeight()
			if err != nil {
				return fmt.Errorf("error while getting chain latest block height: %s", err)
			}

			err = govModule.UpdateProposal(height, time.Now(), proposalID)
			if err != nil {
				return err
			}

			return nil
		},
	}
}
