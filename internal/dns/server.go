package dns

import (
	"database/sql"
	"github.com/DifuseHQ/dddns/internal/db"
	"github.com/DifuseHQ/dddns/pkg/logger"
	"github.com/phuslu/fastdns"
	"log"
	"net/netip"
	"os"
	"sync"
	"time"

	"github.com/DifuseHQ/dddns/internal/db/model"
)

type DNSHandler struct {
	Debug bool
}

func InitDNSServer(dnsAddr string, dnsPort string, domain string) {
	dnsBind := dnsAddr + ":" + dnsPort
	server := &fastdns.Server{
		Handler: &DNSHandler{
			Debug: os.Getenv("DEBUG") != "",
		},
		Stats: &fastdns.CoreStats{
			Prefix: "coredns_",
			Family: "1",
			Proto:  "udp",
			Server: "dns://" + dnsBind,
			Zone:   domain + ".",
		},
	}

	err := server.ListenAndServe(dnsBind)

	if err != nil {
		logger.Log.Fatal("Failed to start DNS server ", err)
	}
}

func (h *DNSHandler) ServeDNS(rw fastdns.ResponseWriter, req *fastdns.Message) {
	if h.Debug {
		log.Printf("%s: CLASS %s TYPE %s\n", string(req.Domain), req.Question.Class, req.Question.Type)
	}

	domain := string(req.Domain)

	if domain != "" {
		switch req.Question.Type {
		case fastdns.TypeA:
			h.handleA(domain, rw, req)
		case fastdns.TypeAAAA:
			h.handleAAAA(domain, rw, req)
		default:
			fastdns.Error(rw, req, fastdns.RcodeNXDomain)
		}
	} else {
		fastdns.Error(rw, req, fastdns.RcodeNXDomain)
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

func (h *DNSHandler) handleA(domain string, rw fastdns.ResponseWriter, req *fastdns.Message) {
	cacheMutex.RLock()
	cached, found := cache[domain]
	cacheMutex.RUnlock()

	if found && time.Since(cached.Timestamp) < time.Minute {
		if cached.Record != nil && cached.Record.ARecord != "" {
			ip := netip.MustParseAddr(cached.Record.ARecord)
			fastdns.HOST(rw, req, 60, []netip.Addr{ip})
			return
		}
	}

	record := getRecordFromDB(db.Database, domain)
	if record != nil && record.ARecord != "" {
		ip := netip.MustParseAddr(record.ARecord)
		fastdns.HOST(rw, req, 60, []netip.Addr{ip})

		cacheMutex.Lock()
		cache[domain] = CachedRecord{Record: record, Timestamp: time.Now()}
		cacheMutex.Unlock()
	} else {
		fastdns.Error(rw, req, fastdns.RcodeNXDomain)
	}
}

func (h *DNSHandler) handleAAAA(domain string, rw fastdns.ResponseWriter, req *fastdns.Message) {
	cacheMutex.RLock()
	cached, found := cache[domain]
	cacheMutex.RUnlock()

	if found && time.Since(cached.Timestamp) < time.Minute {
		if cached.Record != nil && cached.Record.AAAARecord != "" {
			ip := netip.MustParseAddr(cached.Record.AAAARecord)
			fastdns.HOST(rw, req, 60, []netip.Addr{ip})
			return
		}
	}

	record := getRecordFromDB(db.Database, domain)
	if record != nil && record.AAAARecord != "" {
		ip := netip.MustParseAddr(record.AAAARecord)
		fastdns.HOST(rw, req, 60, []netip.Addr{ip})

		cacheMutex.Lock()
		cache[domain] = CachedRecord{Record: record, Timestamp: time.Now()}
		cacheMutex.Unlock()
	} else {
		fastdns.Error(rw, req, fastdns.RcodeNXDomain)
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
