package fixture_test

import (
	"github.com/hashicorp/go-memdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMemDBInherit(t *testing.T) {
	type Inherited struct {
		Name string
		Id   string
	}

	type Record struct {
		Inherited
	}

	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"record": {
				Name: "record",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:   "id",
						Unique: true,
						Indexer: &memdb.StringFieldIndex{
							Field: "Name",
						},
					},
				},
			},
		},
	}
	storage, err := memdb.NewMemDB(schema)
	require.NoError(t, err)

	txn := storage.Txn(true)
	err = txn.Insert("record", &Record{
		Inherited: Inherited{
			Name: "test-1",
		},
	})
	assert.NoError(t, err)
	txn.Commit()
}
