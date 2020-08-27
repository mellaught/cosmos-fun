package v0_11

import (
	v010staking "github.com/mellaught/cosmos-fun/x/staking/legacy/v0_10"
	v011staking "github.com/mellaught/cosmos-fun/x/staking/legacy/v0_11"
	v010token "github.com/mellaught/cosmos-fun/x/token/legacy/v0_10"
	v011token "github.com/mellaught/cosmos-fun/x/token/legacy/v0_11"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/genutil"
)

// Migrate migrates exported state from v0.10.x to a v0.11.0 genesis state
func Migrate(appState genutil.AppMap) genutil.AppMap {
	v010Codec := codec.New()
	codec.RegisterCrypto(v010Codec)

	v011Codec := codec.New()
	codec.RegisterCrypto(v011Codec)

	// migrate staking state
	if appState[v010staking.ModuleName] != nil {
		var stakingGenState v010staking.GenesisState
		v010Codec.MustUnmarshalJSON(appState[v010staking.ModuleName], &stakingGenState)

		delete(appState, v010staking.ModuleName)
		appState[v011staking.ModuleName] = v011Codec.MustMarshalJSON(v011staking.Migrate(stakingGenState))
	}

	// migrate token state
	if appState[v010token.ModuleName] != nil {
		var tokenGenState v010token.GenesisState
		v010Codec.MustUnmarshalJSON(appState[v010token.ModuleName], &tokenGenState)

		delete(appState, v010token.ModuleName)
		appState[v011token.ModuleName] = v011Codec.MustMarshalJSON(v011token.Migrate(tokenGenState))
	}

	return appState
}
