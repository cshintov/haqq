package v150

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"
)

func (r *RevestingUpgradeHandler) forceDequeueUnbondingAndRedelegation() error {
	r.ctx.Logger().Info("Force unbonding and redelegation.")

	blockTime := r.ctx.BlockHeader().Time
	unbondingPeriod := r.StakingKeeper.UnbondingTime(r.ctx)
	unbondedCoins := sdk.NewCoins(sdk.NewCoin(r.StakingKeeper.BondDenom(r.ctx), sdk.ZeroInt()))
	failedUnbondingAttempts := 0
	processedPairs := make(map[string]bool)

	// Remove all unbonding delegations from the ubd queue.
	ubdq := r.StakingKeeper.DequeueAllMatureUBDQueue(r.ctx, blockTime.Add(unbondingPeriod))
	r.ctx.Logger().Info(fmt.Sprintf("Found %d unbonding delegations to be completed:", len(ubdq)))
	for _, dvPair := range ubdq {
		valAddr, err := sdk.ValAddressFromBech32(dvPair.ValidatorAddress)
		if err != nil {
			panic(err)
		}
		delegatorAddress := sdk.MustAccAddressFromBech32(dvPair.DelegatorAddress)

		if _, done := processedPairs[fmt.Sprintf("%s-%s", valAddr.String(), delegatorAddress.String())]; done {
			continue
		}

		balances, err := r.completeUnbonding(r.ctx, delegatorAddress, valAddr)
		if err != nil {
			r.ctx.Logger().Error(fmt.Sprintf(" - failed unbonding %s -> %s: %s", valAddr.String(), delegatorAddress.String(), err.Error()))
			failedUnbondingAttempts++
			continue
		}

		processedPairs[fmt.Sprintf("%s-%s", valAddr.String(), delegatorAddress.String())] = true

		unbondedCoins = unbondedCoins.Add(balances...)
		r.ctx.Logger().Info(fmt.Sprintf(" - unbonded %s -> %s: %s", valAddr.String(), delegatorAddress.String(), balances.String()))

		r.ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCompleteUnbonding,
				sdk.NewAttribute(sdk.AttributeKeyAmount, balances.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, dvPair.ValidatorAddress),
				sdk.NewAttribute(types.AttributeKeyDelegator, dvPair.DelegatorAddress),
			),
		)
	}
	r.ctx.Logger().Info("Total unbonded tokens: " + unbondedCoins.String())
	r.ctx.Logger().Error(fmt.Sprintf("Failed attempts: %d", failedUnbondingAttempts))

	// Remove all mature redelegations from the red queue.
	matureRedelegations := r.StakingKeeper.DequeueAllMatureRedelegationQueue(r.ctx, blockTime.Add(unbondingPeriod))
	for _, dvvTriplet := range matureRedelegations {
		valSrcAddr, err := sdk.ValAddressFromBech32(dvvTriplet.ValidatorSrcAddress)
		if err != nil {
			panic(err)
		}
		valDstAddr, err := sdk.ValAddressFromBech32(dvvTriplet.ValidatorDstAddress)
		if err != nil {
			panic(err)
		}
		delegatorAddress := sdk.MustAccAddressFromBech32(dvvTriplet.DelegatorAddress)

		balances, err := r.completeRedelegation(
			r.ctx,
			delegatorAddress,
			valSrcAddr,
			valDstAddr,
		)
		if err != nil {
			return errors.Wrap(err, "failed to complete redelegation")
		}

		r.ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCompleteRedelegation,
				sdk.NewAttribute(sdk.AttributeKeyAmount, balances.String()),
				sdk.NewAttribute(types.AttributeKeyDelegator, dvvTriplet.DelegatorAddress),
				sdk.NewAttribute(types.AttributeKeySrcValidator, dvvTriplet.ValidatorSrcAddress),
				sdk.NewAttribute(types.AttributeKeyDstValidator, dvvTriplet.ValidatorDstAddress),
			),
		)
	}

	return nil
}

func (r *RevestingUpgradeHandler) completeUnbonding(
	ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress,
) (sdk.Coins, error) {
	ubd, found := r.StakingKeeper.GetUnbondingDelegation(ctx, delAddr, valAddr)
	if !found {
		return nil, types.ErrNoUnbondingDelegation
	}

	bondDenom := r.StakingKeeper.GetParams(ctx).BondDenom
	balances := sdk.NewCoins()

	delegatorAddress, err := sdk.AccAddressFromBech32(ubd.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	// loop through all the entries and complete unbonding mature entries
	for i := 0; i < len(ubd.Entries); i++ {
		entry := ubd.Entries[i]
		ubd.RemoveEntry(int64(i))
		i--

		// track undelegation only when remaining or truncated shares are non-zero
		if !entry.Balance.IsZero() {
			amt := sdk.NewCoin(bondDenom, entry.Balance)
			if err := r.BankKeeper.UndelegateCoinsFromModuleToAccount(
				ctx, types.NotBondedPoolName, delegatorAddress, sdk.NewCoins(amt),
			); err != nil {
				return nil, err
			}

			balances = balances.Add(amt)
		}
	}

	// set the unbonding delegation or remove it if there are no more entries
	switch {
	case len(ubd.Entries) == 0:
		r.StakingKeeper.RemoveUnbondingDelegation(ctx, ubd)
	default:
		r.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
	}

	return balances, nil
}

func (r *RevestingUpgradeHandler) completeRedelegation(
	ctx sdk.Context, delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress,
) (sdk.Coins, error) {
	red, found := r.StakingKeeper.GetRedelegation(ctx, delAddr, valSrcAddr, valDstAddr)
	if !found {
		return nil, types.ErrNoRedelegation
	}

	bondDenom := r.StakingKeeper.GetParams(ctx).BondDenom
	balances := sdk.NewCoins()

	// loop through all the entries and complete mature redelegation entries
	for i := 0; i < len(red.Entries); i++ {
		entry := red.Entries[i]
		red.RemoveEntry(int64(i))
		i--

		if !entry.InitialBalance.IsZero() {
			balances = balances.Add(sdk.NewCoin(bondDenom, entry.InitialBalance))
		}
	}

	// set the redelegation or remove it if there are no more entries
	switch {
	case len(red.Entries) == 0:
		r.StakingKeeper.RemoveRedelegation(ctx, red)
	default:
		r.StakingKeeper.SetRedelegation(ctx, red)
	}

	return balances, nil
}

func (r *RevestingUpgradeHandler) checkUnbondingPoolBalance() sdk.Coin {
	bondDenom := r.StakingKeeper.GetParams(r.ctx).BondDenom
	poolAcc := r.AccountKeeper.GetModuleAccount(r.ctx, types.NotBondedPoolName)
	balance := r.BankKeeper.GetBalance(r.ctx, poolAcc.GetAddress(), bondDenom)

	return balance
}
