package nameshake

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	ModuleName = "nameshake"
	RouterKey  = "nameshake"
	ModuleCdc  = codec.New()
)

func NewAppModule(k Keeper) AppModule {
	return AppModule{}
}

type AppModuleBasic struct{}

func (a AppModuleBasic) Name() string {
	panic("not implemented")
}

func (a AppModuleBasic) RegisterCodec(*codec.Codec) {
	panic("not implemented")
}

func (a AppModuleBasic) DefaultGenesis() json.RawMessage {
	panic("not implemented")
}

func (a AppModuleBasic) ValidateGenesis(json.RawMessage) error {
	panic("not implemented")
}

type AppModule struct{}

func (a AppModule) Name() string {
	panic("not implemented")
}

func (a AppModule) RegisterCodec(*codec.Codec) {
	panic("not implemented")
}

func (a AppModule) DefaultGenesis() json.RawMessage {
	panic("not implemented")
}

func (a AppModule) ValidateGenesis(json.RawMessage) error {
	panic("not implemented")
}

func (a AppModule) InitGenesis(types.Context, json.RawMessage) []abci.ValidatorUpdate {
	panic("not implemented")
}

func (a AppModule) ExportGenesis(types.Context) json.RawMessage {
	panic("not implemented")
}

func (a AppModule) RegisterInvariants(types.InvariantRouter) {
	panic("not implemented")
}

func (a AppModule) Route() string {
	panic("not implemented")
}

func (a AppModule) NewHandler() types.Handler {
	panic("not implemented")
}

func (a AppModule) QuerierRoute() string {
	panic("not implemented")
}

func (a AppModule) NewQuerierHandler() types.Querier {
	panic("not implemented")
}

func (a AppModule) BeginBlock(types.Context, abci.RequestBeginBlock) types.Tags {
	panic("not implemented")
}

func (a AppModule) EndBlock(types.Context, abci.RequestEndBlock) ([]abci.ValidatorUpdate, types.Tags) {
	panic("not implemented")
}

// type check AppModule and AppModule basic
var (
	_ sdk.AppModule      = AppModule{}
	_ sdk.AppModuleBasic = AppModuleBasic{}
)
