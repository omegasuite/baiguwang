module github.com/btcsuite/trim

replace (
	github.com/btcsuite/btcd/blockchain => ../../blockchain
	github.com/btcsuite/btcd/btcec => ../../btcec
	github.com/btcsuite/btcd/btcutil => ../../btcutil
	github.com/btcsuite/btcd/btcutil/gcs => ../../btcutil/gcs
	github.com/btcsuite/btcd/btcutil/gcs/builder => ../../btcutil/gcs/builder
	github.com/btcsuite/btcd/chaincfg => ../../chaincfg
	github.com/btcsuite/btcd/chaincfg/chainhash => ../../chaincfg/chainhash
	github.com/btcsuite/btcd/database => ../../database
	github.com/btcsuite/btcd/database/ffldb => ../../database/ffldb
	github.com/btcsuite/btcd/txscript => ../../txscript
	github.com/btcsuite/btcd/wire => ../../wire
	github.com/btcsuite/btclog => ../../../btclog

	github.com/golang/snappy => ../../../snappy-go
	github.com/stretchr/testify/require => ../../testify/require
	github.com/syndtr/goleveldb/leveldb => ../../../btcsuite/goleveldb
	github.com/syndtr/goleveldb/leveldb/comparer => ../../../btcsuite/goleveldb/comparer
	github.com/syndtr/goleveldb/leveldb/errors => ../../../btcsuite/goleveldb/errors
	github.com/syndtr/goleveldb/leveldb/filter => ../../../btcsuite/goleveldb/filter
	github.com/syndtr/goleveldb/leveldb/iterator => ../../../btcsuite/goleveldb/iterator
	github.com/syndtr/goleveldb/leveldb/opt => ../../../btcsuite/goleveldb/opt
	github.com/syndtr/goleveldb/leveldb/util => ../../../btcsuite/goleveldb/util
)

require (
	github.com/btcsuite/btcd/blockchain v1.0.0
	github.com/btcsuite/btcd/btcutil v1.1.5
	github.com/btcsuite/btcd/chaincfg v0.0.0-00010101000000-000000000000
	github.com/btcsuite/btcd/chaincfg/chainhash v1.1.0
	github.com/btcsuite/btcd/database v0.0.0-00010101000000-000000000000
	github.com/btcsuite/btcd/database/ffldb v0.0.0-00010101000000-000000000000
	github.com/btcsuite/btcd/wire v0.0.0-00010101000000-000000000000
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f
	github.com/jessevdk/go-flags v1.4.0
)

require (
	github.com/btcsuite/btcd/btcec/v2 v2.1.3 // indirect
	github.com/btcsuite/btcd/txscript v0.0.0-00010101000000-000000000000 // indirect
	github.com/decred/dcrd/crypto/blake256 v1.0.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/sys v0.0.0-20200814200057-3d37ad5750ed // indirect
)

// The retract statements below fixes an accidental push of the tags of a btcd
// fork.
retract (
	v0.18.1
	v0.18.0
	v0.17.1
	v0.17.0
	v0.16.5
	v0.16.4
	v0.16.3
	v0.16.2
	v0.16.1
	v0.16.0

	v0.15.2
	v0.15.1
	v0.15.0

	v0.14.7
	v0.14.6
	v0.14.6
	v0.14.5
	v0.14.4
	v0.14.3
	v0.14.2
	v0.14.1

	v0.14.0
	v0.13.0-beta2
	v0.13.0-beta
)

go 1.22.1
