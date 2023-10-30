package db

import (
	"database/sql"
	"fmt"
	"github.com/DifuseHQ/dddns/internal/db/model"
	"github.com/DifuseHQ/dddns/internal/utils"
	"github.com/DifuseHQ/dddns/pkg/logger"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

var Database *sql.DB

func InitDB(domain string) {
	var err error

	cwd, err := os.Getwd()
	if err != nil {
		logger.Log.Fatal("Error getting current working directory", err)
	}

	dbPath := fmt.Sprintf("%s/data/dddns.db", cwd)

	Database, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.Log.Fatal("Error opening database connection", err)
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

	CREATE TABLE IF NOT EXISTS query_log (
		ip_address VARCHAR(45),
		query_type INT,
		total_queries INT DEFAULT 0,
		successful_queries INT DEFAULT 0,
		failed_queries INT DEFAULT 0,
		last_query_timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_records_domain ON records (domain);
    CREATE INDEX IF NOT EXISTS idx_ip_address ON query_log(ip_address);
	CREATE INDEX IF NOT EXISTS idx_query_type ON query_log(query_type);

	PRAGMA journal_mode=WAL;
    `

	_, err = Database.Exec(createTablesSQL)

	if err != nil {
		logger.Log.Fatal("Error creating records table", err)
	}

	_, err = InsertOrUpdateRecord(Database, &model.Record{
		UUID:       "a10af838-c870-4291-a81f-ee09098d7247",
		Domain:     "loopback." + domain,
		ARecord:    "127.0.0.1",
		AAAARecord: "::1",
	})

	if err != nil {
		logger.Log.Fatal("Error inserting initial record", err)
	}

	logger.Log.Info("Database initialized")
}

func InsertOrUpdateRecord(database *sql.DB, record *model.Record) (bool, error) {
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
		return false, err
	}

	return true, nil
}

func DeleteRecord(database *sql.DB, uuid string) (bool, error) {
	deleteSQL := `DELETE FROM records WHERE uuid = ?;`

	_, err := database.Exec(deleteSQL, uuid)
	if err != nil {
		return false, err
	}

	return true, nil
}

func UpdateQueryLog(database *sql.DB, ipAddress string, queryType int, foundRecord bool) {
	var log model.QueryLog
	err := database.QueryRow(`
		SELECT ip_address, 
		       query_type, 
		       total_queries, 
		       successful_queries, 
		       failed_queries, 
		       last_query_timestamp FROM query_log 
		WHERE ip_address = ? AND query_type = ?`, ipAddress, queryType).Scan(&log.IPAddress, &log.QueryType, &log.TotalQueries, &log.SuccessfulQueries, &log.FailedQueries, &log.LastQueryTimestamp)

	if err == sql.ErrNoRows {
		_, err = database.Exec(`
			INSERT INTO query_log (ip_address, 
								   query_type, 
								   total_queries, 
								   successful_queries, 
								   failed_queries, 
								   last_query_timestamp) 
			VALUES (?, ?, 1, ?, ?, CURRENT_TIMESTAMP)`,
			ipAddress, queryType, utils.BoolToInt(foundRecord), utils.BoolToInt(!foundRecord))
		if err != nil {
			logger.Log.Error("Error inserting new query log", err)
		}
	} else if err == nil {
		log.TotalQueries++
		if foundRecord {
			log.SuccessfulQueries++
		} else {
			log.FailedQueries++
		}
		_, err = database.Exec("UPDATE query_log SET total_queries = ?, successful_queries = ?, failed_queries = ?, last_query_timestamp = CURRENT_TIMESTAMP WHERE ip_address = ? AND query_type = ?",
			log.TotalQueries, log.SuccessfulQueries, log.FailedQueries, ipAddress, queryType)
		if err != nil {
			logger.Log.Error("Error updating query log", err)
		}
	} else {
		logger.Log.Error("Error querying query log", err)
	}
}
