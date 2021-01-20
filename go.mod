module github.com/iov-one/cosmos-sdk-crud

go 1.15

require (
	github.com/cosmos/cosmos-sdk v0.40.0
	github.com/gogo/protobuf v1.3.2
	github.com/pkg/errors v0.9.1
	github.com/tendermint/tendermint v0.34.2
	github.com/tendermint/tm-db v0.6.3
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4

replace github.com/iov-one/cosmos-sdk-crud => ./
