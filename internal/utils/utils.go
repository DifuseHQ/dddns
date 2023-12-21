package utils

import (
	"fmt"
	"github.com/DifuseHQ/dddns/pkg/logger"
	"net"
	"strings"
	"time"
)

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func DomainEndsWith(domain, suffix string) bool {
	return strings.HasSuffix(strings.ToLower(domain), strings.ToLower(suffix))
}

func GenerateSerial() uint32 {
	const layout = "20060102"
	dateStr := time.Now().Format(layout)
	revision := 1
	serialStr := fmt.Sprintf("%s%02d", dateStr, revision)

	var serial uint32
	_, err := fmt.Sscanf(serialStr, "%d", &serial)
	if err != nil {
		logger.Log.Fatal("Error generating serial for SOA record ", err.Error())
	}

	return serial
}

func StringContains(s string, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func ParseIPv6Subdomain(subdomain string) string {
	subdomainParts := strings.Split(subdomain, ".")
	var possibleIPv6 string
	if strings.Contains(subdomainParts[len(subdomainParts)-1], "-") {
		possibleIPv6 = strings.ReplaceAll(subdomainParts[len(subdomainParts)-1], "-", ":")
	} else {
		if len(subdomainParts) < 8 {
			return ""
		}
		possibleIPv6 = strings.Join(subdomainParts[len(subdomainParts)-8:], ":")
	}
	address := net.ParseIP(possibleIPv6)

	return address.String()
}

func ParseIPv4Subdomain(subdomain string) string {
	subdomainParts := strings.Split(subdomain, ".")

	var possibleIPv4 string

	if strings.Contains(subdomainParts[len(subdomainParts)-1], "-") {
		possibleIPv4 = strings.ReplaceAll(subdomainParts[len(subdomainParts)-1], "-", ".")
	} else {
		if len(subdomainParts) < 4 {
			return ""
		}
		possibleIPv4 = strings.Join(subdomainParts[len(subdomainParts)-4:], ".")
	}

	address := net.ParseIP(possibleIPv4)

	if address.To4() == nil {
		return ""
	}

	return address.String()
}
