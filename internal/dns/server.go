package dns

import (
	"database/sql"
	"fmt"
	"github.com/DifuseHQ/dddns/internal/db"
	"github.com/DifuseHQ/dddns/internal/db/model"
	"github.com/DifuseHQ/dddns/internal/utils"
	"github.com/DifuseHQ/dddns/pkg/logger"
	"github.com/miekg/dns"
	"net"
	"strings"
	"time"
)

type DNSStatistics struct {
	TotalQueries      int64
	SuccessfulQueries int64
	FailedQueries     int64
	SOAQueries        int64
	NSQueries         int64
	AQueries          int64
	AAAAQueries       int64
}

type DNSServer struct {
	addr       string
	protocol   string
	nameserver string
	domain     string
	mailbox    string
	StartTime  int64
	authority  bool
	Stats      DNSStatistics
	TunnelA    string
	TunnelAAAA string
}

func (s *DNSServer) InitDNSServer(dnsAddr string, dnsPort string, nameServer string, domain string, mailbox string, authority bool, tunnelA string, tunnelAAAA string) {
	srv := &dns.Server{
		Addr:    dnsAddr + ":" + dnsPort,
		Net:     "udp",
		Handler: s,
	}

	if !strings.HasSuffix(nameServer, ".") {
		nameServer = nameServer + "."
	}

	if !strings.HasSuffix(mailbox, ".") {
		mailbox = mailbox + "."
	}

	s.nameserver = nameServer
	s.mailbox = mailbox
	s.domain = domain
	s.authority = authority
	s.StartTime = time.Now().Unix()
	s.TunnelA = tunnelA
	s.TunnelAAAA = tunnelAAAA

	if err := srv.ListenAndServe(); err != nil {
		logger.Log.Fatal("Failed to start DNS server ", err.Error())
	}
}

func (s *DNSServer) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	qname := strings.ToLower(dns.Name(r.Question[0].Name).String())
	qtype := r.Question[0].Qtype

	logger.Log.Debug("Received DNS query: ", qname, " Type: ", qtype)

	var answers []dns.RR
	var responseCode int

	s.Stats.TotalQueries++

	if utils.DomainEndsWith(qname, fmt.Sprintf(".backname.%s.", s.domain)) || utils.DomainEndsWith(qname, fmt.Sprintf(".backname.%s", s.domain)) {
		subdomainOnly := strings.TrimSuffix(qname, ".backname.difusedns.com.")
		if qtype == dns.TypeA {
			ipAddress := utils.ParseIPv4Subdomain(subdomainOnly)
			if ipAddress != "" {
				answers = append(answers, aRecord(qname, 60, ipAddress))
				responseCode = dns.RcodeSuccess
				s.Stats.AQueries++
			} else {
				responseCode = dns.RcodeSuccess
			}
		} else if qtype == dns.TypeAAAA {
			ipAddress := utils.ParseIPv6Subdomain(subdomainOnly)
			if ipAddress != "" {
				answers = append(answers, aaaaRecord(qname, 60, ipAddress))
				responseCode = dns.RcodeSuccess
				s.Stats.AAAAQueries++
			} else {
				responseCode = dns.RcodeSuccess
			}
		} else {
			responseCode = dns.RcodeNameError
		}
	} else if utils.StringContains(qname, "tunnel.difusedns.com") {
		if qtype == dns.TypeA {
			answers = append(answers, aRecord(qname, 60, s.TunnelA))
			responseCode = dns.RcodeSuccess
			s.Stats.AQueries++
		} else if qtype == dns.TypeAAAA {
			answers = append(answers, aaaaRecord(qname, 60, s.TunnelAAAA))
			responseCode = dns.RcodeSuccess
			s.Stats.AAAAQueries++
		} else {
			responseCode = dns.RcodeNameError
		}
	} else {
		record := getRecordFromDB(db.Database, qname)
		logger.Log.Debug("Queried DB for record: ", qname)

		if record != nil {
			logger.Log.Debug("Record found for ", qname)
			if qtype == dns.TypeA && record.ARecord != "" {
				answers = append(answers, aRecord(qname, 60, record.ARecord))
				responseCode = dns.RcodeSuccess
				s.Stats.AQueries++
				logger.Log.Debug("A record found for ", qname)
			} else if qtype == dns.TypeAAAA && record.AAAARecord != "" {
				answers = append(answers, aaaaRecord(qname, 60, record.AAAARecord))
				responseCode = dns.RcodeSuccess
				s.Stats.AAAAQueries++
				logger.Log.Debug("AAAA record found for ", qname)
			} else {
				responseCode = dns.RcodeSuccess
				logger.Log.Debug("No matching record type found for ", qname)
			}
		} else {
			responseCode = dns.RcodeSuccess
			logger.Log.Debug("No record found for ", qname)
		}
	}

	if qtype == dns.TypeSOA {
		soaR := soaRecord(s, qname)
		if soaR == nil {
			responseCode = dns.RcodeNameError
			logger.Log.Debug("No SOA record found for ", qname)
		} else {
			answers = append(answers, soaR)
			responseCode = dns.RcodeSuccess
			logger.Log.Debug("SOA record appended for ", qname)
		}
		s.Stats.SOAQueries++
	} else if qtype == dns.TypeNS {
		nsR := nsRecord(s, qname)
		if nsR == nil {
			responseCode = dns.RcodeNameError
			logger.Log.Debug("No NS record found for ", qname)
		} else {
			answers = append(answers, nsR)
			responseCode = dns.RcodeSuccess
			logger.Log.Debug("NS record appended for ", qname)
		}
		s.Stats.NSQueries++
	}

	if responseCode == dns.RcodeSuccess {
		s.Stats.SuccessfulQueries++
	} else {
		s.Stats.FailedQueries++
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = s.authority
	m.Answer = answers
	m.Rcode = responseCode

	w.WriteMsg(m)

	logger.Log.Debug("Response sent for ", qname, " with Rcode: ", responseCode)
}

func aRecord(name string, ttl uint32, ipAddress string) *dns.A {
	return &dns.A{
		Hdr: dns.RR_Header{Name: name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: ttl},
		A:   net.ParseIP(ipAddress).To4(),
	}
}

func aaaaRecord(name string, ttl uint32, ipAddress string) *dns.AAAA {
	return &dns.AAAA{
		Hdr:  dns.RR_Header{Name: name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: ttl},
		AAAA: net.ParseIP(ipAddress),
	}
}

func soaRecord(s *DNSServer, qname string) *dns.SOA {
	serial := utils.GenerateSerial()
	refresh := 3600
	retry := 600
	expire := 1209600
	minTTL := 300

	if !strings.HasSuffix(qname, ".") {
		qname = qname + "."
	}

	if utils.StringContains(qname, s.domain) {
		return &dns.SOA{
			Hdr:     dns.RR_Header{Name: qname, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 60},
			Ns:      s.nameserver,
			Mbox:    s.mailbox,
			Serial:  serial,
			Refresh: uint32(refresh),
			Retry:   uint32(retry),
			Expire:  uint32(expire),
			Minttl:  uint32(minTTL),
		}
	} else {
		return nil
	}
}

func nsRecord(s *DNSServer, qname string) *dns.NS {
	if utils.StringContains(qname, s.domain) {
		return &dns.NS{
			Hdr: dns.RR_Header{Name: qname, Rrtype: dns.TypeNS, Class: dns.ClassINET, Ttl: 60},
			Ns:  s.nameserver,
		}
	} else {
		return nil
	}
}

func getRecordFromDB(database *sql.DB, domain string) *model.Record {
	record := &model.Record{}

	if strings.HasSuffix(domain, ".") {
		domain = domain[:len(domain)-1]
	}

	query := `SELECT domain, a_record, aaaa_record FROM records WHERE domain = ?`

	err := database.QueryRow(query, domain).Scan(&record.Domain, &record.ARecord, &record.AAAARecord)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Debug("No record found for domain", domain)
			return nil
		}

		logger.Log.Error("Error querying database", err)
		return nil
	}

	return record
}
