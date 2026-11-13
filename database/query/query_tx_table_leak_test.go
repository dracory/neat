package query_test

import (
	"testing"

	"github.com/dracory/neat/contracts/database/orm"
)

// Regression test: inside Transaction() the tx object is a single mutable
// *Query. resetStatement() must clear q.table and q.model along with the
// rest of the per-statement builder state; otherwise a statement that sets
// the table explicitly (via .Model()/.Table()) leaks it into the next
// statement, where inferTableFromValue() skips inference because
// q.table != "".
//
// Before the fix, after tx.Model(&edgeLeakRow{}).Delete(), a subsequent
// tx.Create(&nodeLeakRow{...}) ran "INSERT INTO leak_edges (...)" with node
// columns and failed with "table leak_edges has no column named type".
func TestTransactionTableDoesNotLeakIntoInferredCreate(t *testing.T) {
	w := openSQLiteQuery(t)

	execSQL(t, w, `CREATE TABLE leak_nodes (id TEXT PRIMARY KEY, type TEXT)`)
	execSQL(t, w, `CREATE TABLE leak_edges (id TEXT PRIMARY KEY, source TEXT)`)
	execSQL(t, w, `INSERT INTO leak_edges (id, source) VALUES ('e1', 'n1')`)

	err := w.Q.Transaction(func(tx orm.Query) error {
		// Explicitly model-scoped statement sets q.table = "leak_edges".
		if _, err := tx.Model(&edgeLeakRow{}).Where("id = ?", "e1").Delete(); err != nil {
			return err
		}
		// Inference must resolve "leak_nodes" from TableName(); a leaked
		// q.table makes this INSERT target leak_edges instead.
		return tx.Create(&nodeLeakRow{ID: "n1", Type: "x"})
	})
	if err != nil {
		t.Fatalf("transaction failed (inferred Create hit the wrong table): %v", err)
	}

	var got []nodeLeakRow
	if err := w.Q.Table("leak_nodes").Find(&got); err != nil {
		t.Fatalf("find: %v", err)
	}
	if len(got) != 1 || got[0].Type != "x" {
		t.Fatalf("expected 1 node row, got %+v", got)
	}
}

type nodeLeakRow struct {
	ID   string `db:"id"`
	Type string `db:"type"`
}

func (nodeLeakRow) TableName() string { return "leak_nodes" }

type edgeLeakRow struct {
	ID     string `db:"id"`
	Source string `db:"source"`
}

func (edgeLeakRow) TableName() string { return "leak_edges" }

// Regression test: model-derived state must not leak into the next
// statement either. A statement scoped via .Model() picks up the model's
// soft-delete filter (from q.model) and its GlobalScopes(); a following
// .Table() statement must not inherit them. SQLite treats unknown quoted
// identifiers as string literals, so leaked filters don't error — they
// silently return zero rows; the row count assertion below catches that.
func TestTransactionModelScopeDoesNotLeakIntoNextStatement(t *testing.T) {
	w := openSQLiteQuery(t)

	execSQL(t, w, `CREATE TABLE scoped_docs (id TEXT PRIMARY KEY, tenant_id TEXT, deleted_at TEXT)`)
	execSQL(t, w, `INSERT INTO scoped_docs (id, tenant_id, deleted_at) VALUES ('d1', 't1', NULL), ('d2', 't2', '2024-01-01')`)
	execSQL(t, w, `CREATE TABLE plain_docs (id TEXT PRIMARY KEY, name TEXT)`)
	execSQL(t, w, `INSERT INTO plain_docs (id, name) VALUES ('p1', 'x'), ('p2', 'y')`)

	var plain []plainLeakRow
	err := w.Q.Transaction(func(tx orm.Query) error {
		var docs []scopedLeakRow
		// Applies the global scope (tenant_id = 't1') and the soft-delete
		// filter (deleted_at IS NULL): exactly one row matches.
		if err := tx.Model(&scopedLeakRow{}).Find(&docs); err != nil {
			return err
		}
		if len(docs) != 1 || docs[0].ID != "d1" {
			t.Fatalf("scoped statement: expected [d1], got %+v", docs)
		}
		return tx.Table("plain_docs").Find(&plain)
	})
	if err != nil {
		t.Fatalf("transaction failed: %v", err)
	}
	// Leaked soft-delete/global-scope filters would make this 0.
	if len(plain) != 2 {
		t.Fatalf("expected both plain_docs rows, got %+v — model scope/soft-delete state leaked", plain)
	}
}

type scopedLeakRow struct {
	ID        string  `db:"id"`
	TenantID  string  `db:"tenant_id"`
	DeletedAt *string `db:"deleted_at"`
}

func (scopedLeakRow) TableName() string           { return "scoped_docs" }
func (scopedLeakRow) SoftDeletedAtColumn() string { return "deleted_at" }
func (scopedLeakRow) GlobalScopes() []func(orm.Query) orm.Query {
	return []func(orm.Query) orm.Query{
		func(q orm.Query) orm.Query { return q.Where("tenant_id = ?", "t1") },
	}
}

type plainLeakRow struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}
