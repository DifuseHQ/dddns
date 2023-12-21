package utils

import (
	"fmt"
	"github.com/DifuseHQ/dddns/pkg/logger"
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

func ParseIPAddressFromHostname(hostname string) string {
	parts := strings.Split(hostname, ".")
	ipParts := parts[:len(parts)-3]
	for i, j := 0, len(ipParts)-1; i < j; i, j = i+1, j-1 {
		ipParts[i], ipParts[j] = ipParts[j], ipParts[i]
	}
	reversedIP := strings.Join(ipParts, ".")
	if strings.Contains(reversedIP, ":") {
		reversedIP = strings.ReplaceAll(reversedIP, ".", ":")
	}

	return reversedIP
}
