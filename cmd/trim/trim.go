// Copyright (c) 2013-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/database"
	//	"github.com/btcsuite/btcd/limits"
	"github.com/btcsuite/btclog"
)

const (
	// blockDbNamePrefix is the prefix for the btcd block database.
	blockDbNamePrefix = "blocks"
)

var (
	cfg *config
	log btclog.Logger
)

func trim(db database.DB) {
	db.Update(func(tx database.Tx) error {
		tx.PruneBlocks(uint64(cfg.Trim) * 1024 * 1024)
		return nil
	})
}

// loadBlockDB opens the block database and returns a handle to it.
func loadBlockDB() (database.DB, error) {
	// The database name is based on the database type.
	dbName := blockDbNamePrefix + "_" + defaultDbType
	dbPath := filepath.Join(cfg.DataDir, dbName)

	log.Infof("Loading block database from '%s'", dbPath)
	db, err := database.Open(defaultDbType, dbPath, activeNetParams.Net)
	if err != nil {
		return nil, fmt.Errorf("database does not exist")
	}

	log.Info("Block database loaded")
	return db, nil
}

// realMain is the real main function for the utility.  It is necessary to work
// around the fact that deferred functions do not run when os.Exit() is called.
func realMain() error {
	// Load configuration and parse command line.
	tcfg, _, err := loadConfig()
	if err != nil {
		return err
	}
	cfg = tcfg

	// Setup logging.
	backendLogger := btclog.NewBackend(os.Stdout)
	defer os.Stdout.Sync()
	log = backendLogger.Logger("MAIN")
	database.UseLogger(backendLogger.Logger("BCDB"))
	blockchain.UseLogger(backendLogger.Logger("CHAN"))

	// Load the block database.
	db, err := loadBlockDB()
	if err != nil {
		log.Errorf("Failed to load database: %v", err)
		return err
	}
	defer db.Close()
	
	if err != nil {
		return err
	}

	trim(db)

	return nil
}

func main() {
	// up some limits.
	//	if err := limits.SetLimits(); err != nil {
	//		os.Exit(1)
	//	}

	// Work around defer not working after os.Exit()
	if err := realMain(); err != nil {
		os.Exit(1)
	}
}
