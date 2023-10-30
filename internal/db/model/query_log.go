package model

import (
	"time"
)

const (
	QueryTypeA    = 1
	QueryTypeAAAA = 2
)

type QueryLog struct {
	IPAddress          string    `db:"ip_address"`
	QueryType          int       `db:"query_type"`
	TotalQueries       int       `db:"total_queries"`
	SuccessfulQueries  int       `db:"successful_queries"`
	FailedQueries      int       `db:"failed_queries"`
	LastQueryTimestamp time.Time `db:"last_query_timestamp"`
}
