package main

import (
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

func SoarchainAddressesParser(cdc codec.Marshaler, cosmosMsg sdk.Msg) ([]string, error)
