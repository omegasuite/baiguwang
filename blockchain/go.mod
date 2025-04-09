module github.com/btcsuite/btcd/blockchain

replace (
	github.com/btcsuite/btcd/btcec => ../btcec
	github.com/btcsuite/btcd/btcutil => ../btcutil
	github.com/btcsuite/btcd/chaincfg => ../chaincfg
	github.com/btcsuite/btcd/chaincfg/chainhash => ../chainhash
	github.com/btcsuite/btcd/database => ../database
	github.com/btcsuite/btcd/txscript => ../txscript
	github.com/btcsuite/btcd/wire => ../wire
	github.com/btcsuite/btclog => ../../btclog
	github.com/btcsuite/btcd/btcutil/gcs => ../btcutil/gcs
	github.com/btcsuite/btcd/btcutil/gcs/builder => ../btcutil/gcs/builder
	github.com/btcsuite/btcd/database/ffldb => ../database/ffldb
	github.com/stretchr/testify/require => ../testify/require
)

go 1.22.1
