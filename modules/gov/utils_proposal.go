package gov

import (
	"fmt"
	"strings"
	"time"

	minttypes "github.com/osmosis-labs/osmosis/v16/x/mint/types"
	"github.com/rs/zerolog/log"

	proposaltypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"google.golang.org/grpc/codes"

	"github.com/forbole/bdjuno/v4/types"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func (m *Module) UpdateProposalStatus(height int64, blockTime time.Time, id uint64) error {
	// Get the proposal
	proposal, err := m.source.Proposal(height, id)
	if err != nil {
		// Check if proposal has reached the voting end time
		passedVotingPeriod := blockTime.After(proposal.VotingEndTime)

		if strings.Contains(err.Error(), codes.NotFound.String()) && passedVotingPeriod {
			// Handle case when a proposal is deleted from the chain (did not pass deposit period)
			return m.updateDeletedProposalStatus(id)
		}

		return fmt.Errorf("error while getting proposal: %s", err)
	}

	err = m.updateProposalStatus(proposal)
	if err != nil {
		return fmt.Errorf("error while updating proposal status: %s", err)
	}

	err = m.handlePassedProposal(proposal, height)
	if err != nil {
		return fmt.Errorf("error while handling passed proposals: %s", err)
	}

	return nil
}

// updateProposalStatus updates given proposal status
func (m *Module) updateProposalStatus(proposal govtypes.Proposal) error {
	return m.db.UpdateProposal(
		types.NewProposalUpdate(
			proposal.ProposalId,
			proposal.Status.String(),
			proposal.VotingStartTime,
			proposal.VotingEndTime,
		),
	)
}

func (m *Module) UpdateProposalsStakingPoolSnapshot() error {
	log.Debug().Str("module", "gov").Msg("refreshing proposal staking pool snapshots")
	blockTime, err := m.db.GetLastBlockTimestamp()
	if err != nil {
		return err
	}

	ids, err := m.db.GetOpenProposalsIds(blockTime)
	if err != nil {
		log.Error().Err(err).Str("module", "gov").Msg("error while getting open proposals ids")
	}

	// Get the latest block height from db
	height, err := m.db.GetLastBlockHeight()
	if err != nil {
		return fmt.Errorf("error while getting latest block height from db: %s", err)
	}

	for _, proposalID := range ids {
		err = m.UpdateProposalStakingPoolSnapshot(height, proposalID)
		if err != nil {
			return fmt.Errorf("error while updating proposal %d staking pool snapshots: %s", proposalID, err)
		}
	}

	return nil
}

// UpdateProposalStakingPoolSnapshot updates the staking pool snapshot associated with the gov
// proposal having the provided id
func (m *Module) UpdateProposalStakingPoolSnapshot(height int64, proposalID uint64) error {
	pool, err := m.stakingModule.GetStakingPoolSnapshot(height)
	if err != nil {
		return fmt.Errorf("error while getting staking pool: %s", err)
	}

	return m.db.SaveProposalStakingPoolSnapshot(
		types.NewProposalStakingPoolSnapshot(proposalID, pool),
	)
}

// updateDeletedProposalStatus updates the proposal having the given id by setting its status
// to the one that represents a deleted proposal
func (m *Module) updateDeletedProposalStatus(id uint64) error {
	stored, err := m.db.GetProposal(id)
	if err != nil {
		return err
	}

	return m.db.UpdateProposal(
		types.NewProposalUpdate(
			stored.ProposalID,
			types.ProposalStatusInvalid,
			stored.VotingStartTime,
			stored.VotingEndTime,
		),
	)
}

// handleParamChangeProposal updates params to the corresponding modules if a ParamChangeProposal has passed
func (m *Module) handleParamChangeProposal(height int64, paramChangeProposal *proposaltypes.ParameterChangeProposal) (err error) {
	for _, change := range paramChangeProposal.Changes {
		// Update the params for corresponding modules
		switch change.Subspace {
		case distrtypes.ModuleName:
			err = m.distrModule.UpdateParams(height)
			if err != nil {
				return fmt.Errorf("error while updating ParamChangeProposal %s params : %s", distrtypes.ModuleName, err)
			}
		case govtypes.ModuleName:
			err = m.UpdateParams(height)
			if err != nil {
				return fmt.Errorf("error while updating ParamChangeProposal %s params : %s", govtypes.ModuleName, err)
			}
		case minttypes.ModuleName:
			err = m.mintModule.UpdateParams(height)
			if err != nil {
				return fmt.Errorf("error while updating ParamChangeProposal %s params : %s", minttypes.ModuleName, err)
			}

			// Update the inflation
			err = m.mintModule.UpdateInflation()
			if err != nil {
				return fmt.Errorf("error while updating inflation with ParamChangeProposal: %s", err)
			}
		case slashingtypes.ModuleName:
			err = m.slashingModule.UpdateParams(height)
			if err != nil {
				return fmt.Errorf("error while updating ParamChangeProposal %s params : %s", slashingtypes.ModuleName, err)
			}
		case stakingtypes.ModuleName:
			err = m.stakingModule.UpdateParams(height)
			if err != nil {
				return fmt.Errorf("error while updating ParamChangeProposal %s params : %s", stakingtypes.ModuleName, err)
			}
		}
	}
	return nil
}

// UpdateAllActiveProposalsTallyResults updates the tally for active proposals
func (m *Module) UpdateAllActiveProposalsTallyResults() error {
	log.Debug().Str("module", "gov").Msg("refreshing proposal tally results")
	blockTime, err := m.db.GetLastBlockTimestamp()
	if err != nil {
		return err
	}

	ids, err := m.db.GetOpenProposalsIds(blockTime)
	if err != nil {
		log.Error().Err(err).Str("module", "gov").Msg("error while getting open proposals ids")
	}

	height, err := m.db.GetLastBlockHeight()
	if err != nil {
		return err
	}

	for _, proposalID := range ids {
		err = m.UpdateProposalTallyResult(proposalID, height)
		if err != nil {
			return fmt.Errorf("error while updating proposal %d tally result : %s", proposalID, err)
		}
	}

	return nil
}

// UpdateProposalTallyResult updates the tally result associated with the given proposal ID
func (m *Module) UpdateProposalTallyResult(proposalID uint64, height int64) error {
	result, err := m.source.TallyResult(height, proposalID)
	if err != nil {
		return fmt.Errorf("error while getting tally result: %s", err)
	}

	return m.db.SaveTallyResults([]types.TallyResult{
		types.NewTallyResult(
			proposalID,
			result.Yes.String(),
			result.Abstain.String(),
			result.No.String(),
			result.NoWithVeto.String(),
			height,
		),
	})
}

func (m *Module) handlePassedProposal(proposal govtypes.Proposal, height int64) error {
	if proposal.Status != govtypes.StatusPassed {
		// If proposal status is not passed, do nothing
		return nil
	}

	// Unpack proposal
	var content govtypes.Content
	err := m.db.Cdc.UnpackAny(proposal.Content, &content)
	if err != nil {
		return fmt.Errorf("error while handling ParamChangeProposal: %s", err)
	}

	switch p := content.(type) {
	case *proposaltypes.ParameterChangeProposal:
		// Update params while ParameterChangeProposal passed
		err = m.handleParamChangeProposal(height, p)
		if err != nil {
			return fmt.Errorf("error while updating params from ParamChangeProposal: %s", err)
		}

	case *upgradetypes.SoftwareUpgradeProposal:
		// Store software upgrade plan while SoftwareUpgradeProposal passed
		err = m.db.SaveSoftwareUpgradePlan(proposal.ProposalId, p.Plan, height)
		if err != nil {
			return fmt.Errorf("error while storing software upgrade plan: %s", err)
		}

	case *upgradetypes.CancelSoftwareUpgradeProposal:
		// Delete software upgrade plan while CancelSoftwareUpgradeProposal passed
		err = m.db.DeleteSoftwareUpgradePlan(proposal.ProposalId)
		if err != nil {
			return fmt.Errorf("error while deleting software upgrade plan: %s", err)
		}
	}
	return nil
}
