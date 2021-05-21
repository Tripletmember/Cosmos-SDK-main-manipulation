module github.com/iov-one/cosmos-sdk-crud

go 1.15

require (
	github.com/cosmos/cosmos-sdk v0.42.4
	github.com/gogo/protobuf v1.3.3
	github.com/lucasjones/reggen v0.0.0-20200904144131-37ba4fa293bb
	github.com/pkg/errors v0.9.1
	github.com/tendermint/tendermint v0.34.10
	github.com/tendermint/tm-db v0.6.4
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/iov-one/cosmos-sdk-crud => ./
