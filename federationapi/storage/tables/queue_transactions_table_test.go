package tables_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/matrix-org/dendrite/federationapi/storage/postgres"
	"github.com/matrix-org/dendrite/federationapi/storage/sqlite3"
	"github.com/matrix-org/dendrite/federationapi/storage/tables"
	"github.com/matrix-org/dendrite/internal/sqlutil"
	"github.com/matrix-org/dendrite/setup/config"
	"github.com/matrix-org/dendrite/test"
	"github.com/matrix-org/gomatrixserverlib"
	"github.com/stretchr/testify/assert"
)

type QueueTransactionsDatabase struct {
	DB     *sql.DB
	Writer sqlutil.Writer
	Table  tables.FederationQueueTransactions
}

func mustCreateQueueTransactionsTable(t *testing.T, dbType test.DBType) (database QueueTransactionsDatabase, close func()) {
	t.Helper()
	connStr, close := test.PrepareDBConnectionString(t, dbType)
	db, err := sqlutil.Open(&config.DatabaseOptions{
		ConnectionString: config.DataSource(connStr),
	}, sqlutil.NewExclusiveWriter())
	assert.NoError(t, err)
	var tab tables.FederationQueueTransactions
	switch dbType {
	case test.DBTypePostgres:
		tab, err = postgres.NewPostgresQueueTransactionsTable(db)
		assert.NoError(t, err)
	case test.DBTypeSQLite:
		tab, err = sqlite3.NewSQLiteQueueTransactionsTable(db)
		assert.NoError(t, err)
	}
	assert.NoError(t, err)

	database = QueueTransactionsDatabase{
		DB:     db,
		Writer: sqlutil.NewDummyWriter(),
		Table:  tab,
	}
	return database, close
}

func TestShoudInsertQueueTransaction(t *testing.T) {
	ctx := context.Background()
	test.WithAllDatabases(t, func(t *testing.T, dbType test.DBType) {
		db, close := mustCreateQueueTransactionsTable(t, dbType)
		defer close()

		transactionID := gomatrixserverlib.TransactionID(fmt.Sprintf("%d", time.Now().UnixNano()))
		serverName := gomatrixserverlib.ServerName("domain")
		nid := int64(1)
		err := db.Table.InsertQueueTransaction(ctx, nil, transactionID, serverName, nid)
		if err != nil {
			t.Fatalf("Failed inserting transaction: %s", err.Error())
		}
	})
}

func TestShouldRetrieveInsertedQueueTransaction(t *testing.T) {
	ctx := context.Background()
	test.WithAllDatabases(t, func(t *testing.T, dbType test.DBType) {
		db, close := mustCreateQueueTransactionsTable(t, dbType)
		defer close()

		transactionID := gomatrixserverlib.TransactionID(fmt.Sprintf("%d", time.Now().UnixNano()))
		serverName := gomatrixserverlib.ServerName("domain")
		nid := int64(1)

		err := db.Table.InsertQueueTransaction(ctx, nil, transactionID, serverName, nid)
		if err != nil {
			t.Fatalf("Failed inserting transaction: %s", err.Error())
		}

		retrievedNids, err := db.Table.SelectQueueTransactions(ctx, nil, serverName, 10)
		if err != nil {
			t.Fatalf("Failed retrieving transaction: %s", err.Error())
		}

		assert.Equal(t, retrievedNids[0], nid)
		assert.Equal(t, len(retrievedNids), 1)
	})
}

func TestShouldDeleteQueueTransaction(t *testing.T) {
	ctx := context.Background()
	test.WithAllDatabases(t, func(t *testing.T, dbType test.DBType) {
		db, close := mustCreateQueueTransactionsTable(t, dbType)
		defer close()

		transactionID := gomatrixserverlib.TransactionID(fmt.Sprintf("%d", time.Now().UnixNano()))
		serverName := gomatrixserverlib.ServerName("domain")
		nid := int64(1)

		err := db.Table.InsertQueueTransaction(ctx, nil, transactionID, serverName, nid)
		if err != nil {
			t.Fatalf("Failed inserting transaction: %s", err.Error())
		}

		_ = db.Writer.Do(db.DB, nil, func(txn *sql.Tx) error {
			err = db.Table.DeleteQueueTransactions(ctx, txn, serverName, []int64{nid})
			return err
		})
		if err != nil {
			t.Fatalf("Failed deleting transaction: %s", err.Error())
		}

		count, err := db.Table.SelectQueueTransactionCount(ctx, nil, serverName)
		if err != nil {
			t.Fatalf("Failed retrieving transaction count: %s", err.Error())
		}
		assert.Equal(t, count, int64(0))
	})
}

func TestShouldDeleteOnlySpecifiedQueueTransaction(t *testing.T) {
	ctx := context.Background()
	test.WithAllDatabases(t, func(t *testing.T, dbType test.DBType) {
		db, close := mustCreateQueueTransactionsTable(t, dbType)
		defer close()

		transactionID := gomatrixserverlib.TransactionID(fmt.Sprintf("%d", time.Now().UnixNano()))
		serverName := gomatrixserverlib.ServerName("domain")
		nid := int64(1)
		transactionID2 := gomatrixserverlib.TransactionID(fmt.Sprintf("%d2", time.Now().UnixNano()))
		serverName2 := gomatrixserverlib.ServerName("domain2")
		nid2 := int64(2)
		transactionID3 := gomatrixserverlib.TransactionID(fmt.Sprintf("%d3", time.Now().UnixNano()))

		err := db.Table.InsertQueueTransaction(ctx, nil, transactionID, serverName, nid)
		if err != nil {
			t.Fatalf("Failed inserting transaction: %s", err.Error())
		}
		err = db.Table.InsertQueueTransaction(ctx, nil, transactionID2, serverName2, nid)
		if err != nil {
			t.Fatalf("Failed inserting transaction: %s", err.Error())
		}
		err = db.Table.InsertQueueTransaction(ctx, nil, transactionID3, serverName, nid2)
		if err != nil {
			t.Fatalf("Failed inserting transaction: %s", err.Error())
		}

		_ = db.Writer.Do(db.DB, nil, func(txn *sql.Tx) error {
			err = db.Table.DeleteQueueTransactions(ctx, txn, serverName, []int64{nid})
			return err
		})
		if err != nil {
			t.Fatalf("Failed deleting transaction: %s", err.Error())
		}

		count, err := db.Table.SelectQueueTransactionCount(ctx, nil, serverName)
		if err != nil {
			t.Fatalf("Failed retrieving transaction count: %s", err.Error())
		}
		assert.Equal(t, count, int64(1))

		count, err = db.Table.SelectQueueTransactionCount(ctx, nil, serverName2)
		if err != nil {
			t.Fatalf("Failed retrieving transaction count: %s", err.Error())
		}
		assert.Equal(t, count, int64(1))
	})
}
