package db

import (
	"database/sql"
	"fmt"
	"github.com/DifuseHQ/dddns/internal/db/model"
	"github.com/DifuseHQ/dddns/pkg/logger"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

var Database *sql.DB

func InitDB(domain string) {
	var err error

	cwd, err := os.Getwd()
	if err != nil {
		logger.Log.Fatal("Error getting current working directory ", err.Error())
	}

	dbPath := fmt.Sprintf("%s/data/dddns.db", cwd)

	Database, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.Log.Fatal("Error opening database connection ", err.Error())
	}

	createTablesSQL := `
	CREATE TABLE IF NOT EXISTS records (
		"uuid" TEXT NOT NULL PRIMARY KEY,
		"domain" TEXT,
		"a_record" TEXT,
		"aaaa_record" TEXT,
		"created_at" DATETIME DEFAULT CURRENT_TIMESTAMP,
		"last_update_at" DATETIME DEFAULT CURRENT_TIMESTAMP
    );

	CREATE INDEX IF NOT EXISTS idx_records_domain ON records (domain);

	PRAGMA journal_mode=WAL;
    `

	_, err = Database.Exec(createTablesSQL)

	if err != nil {
		logger.Log.Fatal("Error creating records table", err.Error())
	}

	loopbackDomain := "loopback." + domain

	logger.Log.Debug(fmt.Sprintf("Inserting loopback record %s", loopbackDomain))

	_, err = InsertOrUpdateRecord(Database, &model.Record{
		UUID:       "a10af838-c870-4291-a81f-ee09098d7247",
		Domain:     "loopback." + domain,
		ARecord:    "127.0.0.1",
		AAAARecord: "::1",
	}, domain)

	if err != nil {
		logger.Log.Fatal("Error inserting initial record", err.Error())
	}

	logger.Log.Info("Database initialized")
}

func InsertOrUpdateRecord(database *sql.DB, record *model.Record, domain string) (bool, error) {
	if record.Domain != domain && record.Domain != "loopback."+domain {
		err := fmt.Errorf("record domain %s doesn't match domain %s", record.Domain, domain)
		return false, err
	}

	upsertSQL := `
	INSERT INTO records (uuid, domain, a_record, aaaa_record, last_update_at)
	VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(uuid) DO UPDATE SET 
		a_record = excluded.a_record, 
		domain = excluded.domain,
		aaaa_record = excluded.aaaa_record, 
		last_update_at = CURRENT_TIMESTAMP;
	`

	_, err := database.Exec(upsertSQL, record.UUID, record.Domain, record.ARecord, record.AAAARecord)
	if err != nil {
		logger.Log.Error("Error inserting or updating record ", err.Error())
		return false, err
	}

	logger.Log.Debug("Record inserted or updated ", record)

	return true, nil
}

func DeleteRecord(database *sql.DB, uuid string) (bool, error) {
	deleteSQL := `DELETE FROM records WHERE uuid = ?;`

	_, err := database.Exec(deleteSQL, uuid)
	if err != nil {
		logger.Log.Error("Error deleting record ", err.Error())
		return false, err
	}

	logger.Log.Debug("Record deleted for ", uuid)

	return true, nil
}
