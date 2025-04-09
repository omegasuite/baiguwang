module github.com/btcsuite/btcd/database

replace (
	github.com/btcsuite/btcd/chaincfg/chainhash => f:/gopath/src/github.com/btcsuite/btcd/chaincfg/chainhash
	github.com/btcsuite/btcd/chaincfg => f:/gopath/src/github.com/btcsuite/btcd/chaincfg
	github.com/btcsuite/btclog => f:/gopath/src/github.com/btcsuite/btclog
	github.com/syndtr/goleveldb/leveldb => f:/gopath/src/github.com/syndtr/goleveldb/leveldb
	github.com/syndtr/goleveldb/leveldb/comparer => f:/gopath/src/github.com/syndtr/goleveldb/leveldb/comparer
	github.com/btcsuite/btcd/wire => f:/gopath/src/github.com/btcsuite/btcd/wire
	github.com/syndtr/goleveldb/leveldb/errors => f:/gopath/src/github.com/syndtr/goleveldb/leveldb/errors
	github.com/syndtr/goleveldb/leveldb/opt => f:/gopath/src/github.com/syndtr/goleveldb/leveldb/opt
	github.com/syndtr/goleveldb/leveldb/iterator => f:/gopath/src/github.com/syndtr/goleveldb/leveldb/iterator
	github.com/syndtr/goleveldb/leveldb/util => f:/gopath/src/github.com/syndtr/goleveldb/leveldb/util
	github.com/syndtr/goleveldb/leveldb/filter => f:/gopath/src/github.com/syndtr/goleveldb/leveldb/filter
	github.com/btcsuite/btcd/database => f:/gopath/src/github.com/btcsuite/btcd/database
	github.com/btcsuite/btcd/btcutil => f:/gopath/src/github.com/btcsuite/btcd/btcutil
	github.com/btcsuite/btcd/database/internal/treap => f:/gopath/src/github.com/btcsuite/btcd/database/internal/treap
)

go 1.22.1
