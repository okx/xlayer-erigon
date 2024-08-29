package stages

import (
	"context"
	"fmt"
	"github.com/gateway-fm/cdk-erigon-lib/kv"
	"github.com/ledgerwatch/erigon/core/state"
	db2 "github.com/ledgerwatch/erigon/smt/pkg/db"
	smtNs "github.com/ledgerwatch/erigon/smt/pkg/smt"
	"github.com/ledgerwatch/erigon/zk/hermez_db"
	"github.com/ledgerwatch/log/v3"
)

type stageDb struct {
	ctx context.Context
	db  kv.RwDB

	tx          kv.RwTx
	hermezDb    *hermez_db.HermezDb
	eridb       *db2.EriDb
	stateReader *state.PlainStateReader
	smt         *smtNs.SMT

	smtGenerateInMemory bool
}

func newStageDb(ctx context.Context, db kv.RwDB, smtGenerateInMemory bool) (sdb *stageDb, err error) {
	var tx kv.RwTx
	if tx, err = db.BeginRw(ctx); err != nil {
		return nil, err
	}

	sdb = &stageDb{
		ctx:                 ctx,
		db:                  db,
		smtGenerateInMemory: smtGenerateInMemory,
	}
	sdb.SetTx(ctx, tx)
	return sdb, nil
}

func (sdb *stageDb) SetTx(ctx context.Context, tx kv.RwTx) {
	sdb.tx = tx
	sdb.hermezDb = hermez_db.NewHermezDb(tx)
	sdb.eridb = db2.NewEriDb(tx)
	sdb.stateReader = state.NewPlainStateReader(tx)

	if sdb.smtGenerateInMemory {
		sdb.eridb.OpenBatch(ctx.Done())
	}

	sdb.smt = smtNs.NewSMT(sdb.eridb, false)
}

func (sdb *stageDb) CommitAndStart(ctx context.Context) (err error) {
	if sdb.smtGenerateInMemory {
		log.Info(fmt.Sprintf("Commit with smt generating in MEMORY"))
		if err = sdb.eridb.CommitBatch(); err != nil {
			return err
		}
	} else {
		log.Info(fmt.Sprintf("Commit with smt generating in DATABASE"))
	}

	if err = sdb.tx.Commit(); err != nil {
		return err
	}

	tx, err := sdb.db.BeginRw(sdb.ctx)
	if err != nil {
		return err
	}

	sdb.SetTx(ctx, tx)
	return nil
}
