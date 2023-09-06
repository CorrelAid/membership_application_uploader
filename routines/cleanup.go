package routines

import (
	"log"
	"time"

	"github.com/CorrelAid/membership_application_uploader/models"
	"github.com/hashicorp/go-memdb"
)

func StartCleanupRoutine(db *memdb.MemDB) {
	go cleanupRoutine(db)

	ticker := time.NewTicker(24 * time.Hour) // Run every 24 hours
	defer ticker.Stop()

	for range ticker.C {
		go cleanupRoutine(db)
	}
}
func cleanupRoutine(db *memdb.MemDB) {
	currentTime := time.Now()

	txn := db.Txn(true)
	defer txn.Abort()

	memberTable, err := txn.Get("member", "expiry")
	if err != nil {
		panic(err)
	}

	for obj := memberTable.Next(); obj != nil; obj = memberTable.Next() {
		member := obj.(*models.Member)
		expiryTime, err := time.Parse(time.RFC1123, member.Expiry)
		if err != nil {
			panic(err)
		}

		if expiryTime.Before(currentTime) {
			txn.Delete("member", obj)
			log.Printf("Deleted expired member: email=%s", member.Email)
		}
	}

	txn.Commit()
}
