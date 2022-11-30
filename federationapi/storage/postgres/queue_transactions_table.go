// Copyright 2022 The Matrix.org Foundation C.I.C.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package postgres

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"github.com/matrix-org/dendrite/internal"
	"github.com/matrix-org/dendrite/internal/sqlutil"
	"github.com/matrix-org/gomatrixserverlib"
)

const queueTransactionsSchema = `
CREATE TABLE IF NOT EXISTS federationsender_queue_transactions (
    -- The transaction ID that was generated before persisting the event.
	transaction_id TEXT NOT NULL,
    -- The destination server that we will send the event to.
	server_name TEXT NOT NULL,
	-- The JSON NID from the federationsender_transaction_json table.
	json_nid BIGINT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS federationsender_queue_transactions_transaction_json_nid_idx
    ON federationsender_queue_transactions (json_nid, server_name);
CREATE INDEX IF NOT EXISTS federationsender_queue_transactions_json_nid_idx
    ON federationsender_queue_transactions (json_nid);
CREATE INDEX IF NOT EXISTS federationsender_queue_transactions_server_name_idx
    ON federationsender_queue_transactions (server_name);
`

const insertQueueTransactionSQL = "" +
	"INSERT INTO federationsender_queue_transactions (transaction_id, server_name, json_nid)" +
	" VALUES ($1, $2, $3)"

const deleteQueueTransactionsSQL = "" +
	"DELETE FROM federationsender_queue_transactions WHERE server_name = $1 AND json_nid = ANY($2)"

const selectQueueTransactionsSQL = "" +
	"SELECT json_nid FROM federationsender_queue_transactions" +
	" WHERE server_name = $1" +
	" LIMIT $2"

const selectQueueTransactionsCountSQL = "" +
	"SELECT COUNT(*) FROM federationsender_queue_transactions" +
	" WHERE server_name = $1"

type queueTransactionsStatements struct {
	db                               *sql.DB
	insertQueueTransactionStmt       *sql.Stmt
	deleteQueueTransactionsStmt      *sql.Stmt
	selectQueueTransactionsStmt      *sql.Stmt
	selectQueueTransactionsCountStmt *sql.Stmt
}

func NewPostgresQueueTransactionsTable(db *sql.DB) (s *queueTransactionsStatements, err error) {
	s = &queueTransactionsStatements{
		db: db,
	}
	_, err = s.db.Exec(queueTransactionsSchema)
	if err != nil {
		return
	}
	if s.insertQueueTransactionStmt, err = s.db.Prepare(insertQueueTransactionSQL); err != nil {
		return
	}
	if s.deleteQueueTransactionsStmt, err = s.db.Prepare(deleteQueueTransactionsSQL); err != nil {
		return
	}
	if s.selectQueueTransactionsStmt, err = s.db.Prepare(selectQueueTransactionsSQL); err != nil {
		return
	}
	if s.selectQueueTransactionsCountStmt, err = s.db.Prepare(selectQueueTransactionsCountSQL); err != nil {
		return
	}
	return
}

func (s *queueTransactionsStatements) InsertQueueTransaction(
	ctx context.Context,
	txn *sql.Tx,
	transactionID gomatrixserverlib.TransactionID,
	serverName gomatrixserverlib.ServerName,
	nid int64,
) error {
	stmt := sqlutil.TxStmt(txn, s.insertQueueTransactionStmt)
	_, err := stmt.ExecContext(
		ctx,
		transactionID, // the transaction ID that we initially attempted
		serverName,    // destination server name
		nid,           // JSON blob NID
	)
	return err
}

func (s *queueTransactionsStatements) DeleteQueueTransactions(
	ctx context.Context, txn *sql.Tx,
	serverName gomatrixserverlib.ServerName,
	jsonNIDs []int64,
) error {
	stmt := sqlutil.TxStmt(txn, s.deleteQueueTransactionsStmt)
	_, err := stmt.ExecContext(ctx, serverName, pq.Int64Array(jsonNIDs))
	return err
}

func (s *queueTransactionsStatements) SelectQueueTransactions(
	ctx context.Context, txn *sql.Tx,
	serverName gomatrixserverlib.ServerName,
	limit int,
) ([]int64, error) {
	stmt := sqlutil.TxStmt(txn, s.selectQueueTransactionsStmt)
	rows, err := stmt.QueryContext(ctx, serverName, limit)
	if err != nil {
		return nil, err
	}
	defer internal.CloseAndLogIfError(ctx, rows, "queueFromStmt: rows.close() failed")
	var result []int64
	for rows.Next() {
		var nid int64
		if err = rows.Scan(&nid); err != nil {
			return nil, err
		}
		result = append(result, nid)
	}

	return result, rows.Err()
}

func (s *queueTransactionsStatements) SelectQueueTransactionCount(
	ctx context.Context, txn *sql.Tx, serverName gomatrixserverlib.ServerName,
) (int64, error) {
	var count int64
	stmt := sqlutil.TxStmt(txn, s.selectQueueTransactionsCountStmt)
	err := stmt.QueryRowContext(ctx, serverName).Scan(&count)
	if err == sql.ErrNoRows {
		// It's acceptable for there to be no rows referencing a given
		// JSON NID but it's not an error condition. Just return as if
		// there's a zero count.
		return 0, nil
	}
	return count, err
}