package inits

import (
	"time"

	"github.com/CorrelAid/membership_application_uploader/models"
	"github.com/CorrelAid/membership_application_uploader/routines"
	"github.com/hashicorp/go-memdb"
)

var DB *memdb.MemDB

func DBInit() {

	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"member": {
				Name: "member",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:         "id",
						Unique:       true,
						Indexer:      &memdb.StringFieldIndex{Field: "Email"},
						AllowMissing: false,
					},
					"time": {
						Name:         "time",
						Unique:       false,
						Indexer:      &memdb.StringFieldIndex{Field: "Time"},
						AllowMissing: false,
					},
					"expiry": {
						Name:         "expiry",
						Unique:       false,
						Indexer:      &memdb.StringFieldIndex{Field: "Expiry"},
						AllowMissing: false,
					},
				},
			},
		},
	}

	db, err := memdb.NewMemDB(schema)
	if err != nil {
		panic(err)
	}
	member := &models.Member{
		Email:  "test@example.com",
		Time:   "2022-12-31",
		Expiry: time.Now().Add(24 * 14 * time.Hour).Format(time.RFC1123),
	}

	txn := db.Txn(true)
	defer txn.Abort()

	if err := txn.Insert("member", member); err != nil {
		panic(err)
	}
	txn.Commit()
	go routines.StartCleanupRoutine(db)
	DB = db
}
