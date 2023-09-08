package gov

import (
	"fmt"
	"strconv"
	"time"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	juno "github.com/forbole/juno/v4/types"

	abci "github.com/tendermint/tendermint/abci/types"
	tmctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/rs/zerolog/log"
)

// HandleBlock implements modules.BlockModule
func (m *Module) HandleBlock(
	b *tmctypes.ResultBlock, blockResults *tmctypes.ResultBlockResults, txs []*juno.Tx, vals *tmctypes.ResultValidators,
) error {

	err := m.updateProposalsStatus(b.Block.Height, b.Block.Time, vals, blockResults.EndBlockEvents)
	if err != nil {
		log.Error().Str("module", "gov").Int64("height", b.Block.Height).
			Err(err).Msg("error while updating proposals")
	}

	return nil
}

// updateProposalsStatus updates the status of proposals if they have been included in the EndBlockEvents
func (m *Module) updateProposalsStatus(height int64, blockTime time.Time, blockVals *tmctypes.ResultValidators, events []abci.Event) error {
	if len(events) == 0 {
		return nil
	}

	var ids []uint64
	eventsList := juno.FindEventsByType(events, govtypes.EventTypeActiveProposal)
	if len(eventsList) == 0 {
		return nil
	}

	for _, event := range eventsList {
		// find proposal ID
		proposalID, err := juno.FindAttributeByKey(event, govtypes.AttributeKeyProposalID)
		if err != nil {
			return fmt.Errorf("error while getting proposal ID from block events: %s", err)
		}

		// parse proposal ID from []byte to unit64
		id, err := strconv.ParseUint(string(proposalID.Value), 10, 64)
		if err != nil {
			return fmt.Errorf("error while parsing proposal id: %s", err)
		}

		// add proposal ID to ids array
		ids = append(ids, id)
	}

	// update status for proposals IDs stored in ids array
	for _, id := range ids {
		err := m.UpdateProposalStatus(height, blockTime, id)
		if err != nil {
			return fmt.Errorf("error while updating proposal %d status: %s", id, err)
		}
	}

	return nil
}
