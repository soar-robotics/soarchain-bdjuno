package gov

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	modulestypes "github.com/forbole/bdjuno/v4/modules/types"

	parsecmdtypes "github.com/forbole/juno/v4/cmd/parse/types"
	"github.com/forbole/juno/v4/types/config"
	"github.com/spf13/cobra"

	"github.com/forbole/bdjuno/v4/types"

	"github.com/forbole/bdjuno/v4/database"
	"github.com/forbole/bdjuno/v4/modules/distribution"
	"github.com/forbole/bdjuno/v4/modules/gov"
	"github.com/forbole/bdjuno/v4/modules/mint"
	"github.com/forbole/bdjuno/v4/modules/slashing"
	"github.com/forbole/bdjuno/v4/modules/staking"
)

const (
	flagAmount           = "amount"
	flagDenom            = "denom"
	flagDepositID        = "depositid"
	flagDepositor        = "depositor"
	flagDepositTimestamp = "depositTimestamp"
	flagBlockHeight      = "height"
)

// depositCmd returns the Cobra command allowing to fix all things related to a proposal
func proposalDepositsCmd(parseConfig *parsecmdtypes.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposal-deposit [id]",
		Short: "Update given proposal deposit info",
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

			distrModule := distribution.NewModule(sources.DistrSource, parseCtx.EncodingConfig.Codec, db)
			mintModule := mint.NewModule(sources.MintSource, parseCtx.EncodingConfig.Codec, db)
			slashingModule := slashing.NewModule(sources.SlashingSource, parseCtx.EncodingConfig.Codec, db)
			stakingModule := staking.NewModule(sources.StakingSource, parseCtx.EncodingConfig.Codec, db)

			// Build the gov module
			govModule := gov.NewModule(sources.GovSource, nil, distrModule, mintModule, slashingModule, stakingModule, parseCtx.EncodingConfig.Codec, db)

			// Get the flag values
			depositID, _ := cmd.Flags().GetInt64(flagDepositID)
			depositor, _ := cmd.Flags().GetString(flagDepositor)
			am, _ := cmd.Flags().GetInt64(flagAmount)
			timestampDeposit, _ := cmd.Flags().GetString(flagDepositTimestamp)
			blockheight, _ := cmd.Flags().GetInt64(flagBlockHeight)

			depositProposalID := uint64(depositID)
			depositAmount := sdk.NewCoins(sdk.NewCoin(flagDenom, sdk.NewInt(am)))

			depositTimestamp, err := time.Parse(time.RFC3339, timestampDeposit)
			if err != nil {
				return fmt.Errorf("error while parsing timestamp: %s", err)
			}

			depositHeight := int64(blockheight)
			err = govModule.SaveDepositsInDB([]types.Deposit{types.NewDeposit(depositProposalID, depositor, depositAmount, depositTimestamp, depositHeight)})
			if err != nil {
				return fmt.Errorf("error while saving deposits in db: %s", err)
			}

			return nil
		},
	}

	cmd.Flags().Int64(flagAmount, 0, "Amount")
	cmd.Flags().String(flagDenom, "", "Denom")
	cmd.Flags().Int64(flagBlockHeight, 0, "Block Height")
	cmd.Flags().String(flagDepositTimestamp, "", "Timestamp")
	cmd.Flags().Int64(flagDepositID, 0, "Deposit ID")
	cmd.Flags().String(flagDepositor, "", "Depositor Address")
	return cmd
}
