package dns

import (
	"database/sql"
	"github.com/DifuseHQ/dddns/pkg/logger"
	"strings"
	"sync"
	"time"

	"github.com/DifuseHQ/dddns/internal/db"
	"github.com/DifuseHQ/dddns/internal/db/model"
	"github.com/miekg/dns"
)

func InitDNSServer(database *sql.DB, address string) {
	dns.HandleFunc(".", func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		for _, q := range m.Question {
			clientIP := strings.Split(w.RemoteAddr().String(), ":")[0]

			switch q.Qtype {
			case dns.TypeA:
				handleAQuery(database, m, q, clientIP)
			case dns.TypeAAAA:
				handleAAAAQuery(database, m, q, clientIP)
			}
		}
		w.WriteMsg(m)
	})

	err := dns.ListenAndServe(address, "udp", nil)
	if err != nil {
		logger.Log.Fatal("Failed to set up DNS server", err)
	}
}

type CachedRecord struct {
	Record    *model.Record
	Timestamp time.Time
}

var (
	cache      = make(map[string]CachedRecord)
	cacheMutex = &sync.RWMutex{}
)

func handleAQuery(database *sql.DB, m *dns.Msg, q dns.Question, clientIP string) {
	domain := strings.TrimSuffix(q.Name, ".")
	queryType := model.QueryTypeA

	cacheMutex.RLock()
	cached, found := cache[domain]
	cacheMutex.RUnlock()

	if found && time.Since(cached.Timestamp) < time.Minute {
		if cached.Record.ARecord != "" {
			rr, _ := dns.NewRR(q.Name + " IN A " + cached.Record.ARecord)
			m.Answer = append(m.Answer, rr)
			db.UpdateQueryLog(database, clientIP, queryType, true)
			return
		}
	}

	record := getRecordFromDB(database, domain)

	if record != nil && record.ARecord != "" {
		rr, _ := dns.NewRR(q.Name + " IN A " + record.ARecord)
		m.Answer = append(m.Answer, rr)
		db.UpdateQueryLog(database, clientIP, queryType, true)
		cacheMutex.Lock()
		cache[domain] = CachedRecord{Record: record, Timestamp: time.Now()}
		cacheMutex.Unlock()
	} else {
		m.SetRcode(m, dns.RcodeNameError)
		db.UpdateQueryLog(database, clientIP, queryType, false)
	}
}

func handleAAAAQuery(database *sql.DB, m *dns.Msg, q dns.Question, clientIP string) {
	domain := strings.TrimSuffix(q.Name, ".")
	queryType := model.QueryTypeAAAA

	cacheMutex.RLock()
	cached, found := cache[domain]
	cacheMutex.RUnlock()

	if found && time.Since(cached.Timestamp) < time.Minute {
		if cached.Record.AAAARecord != "" {
			rr, _ := dns.NewRR(q.Name + " IN AAAA " + cached.Record.AAAARecord)
			m.Answer = append(m.Answer, rr)
			db.UpdateQueryLog(database, clientIP, queryType, true)
			return
		}
	}

	record := getRecordFromDB(database, domain)

	if record != nil && record.AAAARecord != "" {
		rr, _ := dns.NewRR(q.Name + " IN AAAA " + record.AAAARecord)
		m.Answer = append(m.Answer, rr)
		db.UpdateQueryLog(database, clientIP, queryType, true)
		cacheMutex.Lock()
		cache[domain] = CachedRecord{Record: record, Timestamp: time.Now()}
		cacheMutex.Unlock()
	} else {
		m.SetRcode(m, dns.RcodeNameError)
		db.UpdateQueryLog(database, clientIP, queryType, false)
	}
}

func getRecordFromDB(database *sql.DB, domain string) *model.Record {
	record := &model.Record{}
	query := `SELECT uuid, domain, a_record, aaaa_record, created_at, last_update_at FROM records WHERE domain = ?`

	err := database.QueryRow(query, domain).Scan(&record.UUID, &record.Domain, &record.ARecord, &record.AAAARecord, &record.CreatedAt, &record.LastUpdateAt)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Debug("No record found for domain", domain)
			return nil
		}

		logger.Log.Error("Error querying database", err)
		return nil
	}

	logger.Log.Debug("Found record", record.ARecord)
	return record
}
