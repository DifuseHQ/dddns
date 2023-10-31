package handler

import (
	"github.com/DifuseHQ/dddns/internal/dns"
	"github.com/gofiber/fiber/v2"
	"html/template"
	"time"
)

type DNSStatsPageData struct {
	Stats             dns.DNSStatistics // Assuming dns.DNSStatistics is your stats struct
	HumanReadableTime string
}

func GetDNSStatistics(dns *dns.DNSServer) fiber.Handler {
	return func(c *fiber.Ctx) error {
		stats := dns.Stats
		startTime := time.Unix(dns.StartTime, 0).UTC().Format("2006-01-02 15:04:05 UTC")

		pageData := DNSStatsPageData{
			Stats:             stats,
			HumanReadableTime: startTime,
		}

		htmlContent := `
			<!DOCTYPE html>
			<html>
			<head>
				<title>DDDNS - Statistics</title>
				<style>
					table {
						width: 100%;
						border-collapse: collapse;
					}
					th, td {
						border: 1px solid black;
						padding: 8px;
						text-align: left;
					}
					th {
						background-color: #f2f2f2;
					}
				</style>
			</head>
			<body>
				<h1>DDDNS - DNS Server Statistics</h1>
				<table>
					<tr>
						<th>Statistic</th>
						<th>Value</th>
					</tr>
					<tr>
						<td>Server Start Time</td>
						<td>{{.HumanReadableTime}}</td>
					</tr>
					<tr>
						<td>Total Queries</td>
						<td>{{.Stats.TotalQueries}}</td>
					</tr>
					<tr>
						<td>Successful Queries</td>
						<td>{{.Stats.SuccessfulQueries}}</td>
					</tr>
					<tr>
						<td>Failed Queries</td>
						<td>{{.Stats.FailedQueries}}</td>
					</tr>
					<tr>
						<td>A Record Queries</td>
						<td>{{.Stats.AQueries}}</td>
					</tr>
					<tr>
						<td>AAAA Record Queries</td>
						<td>{{.Stats.AAAAQueries}}</td>
					</tr>
					<tr>
						<td>SOA Record Queries</td>
						<td>{{.Stats.SOAQueries}}</td>
					</tr>
					<tr>
						<td>NS Record Queries</td>
						<td>{{.Stats.NSQueries}}</td>
					</tr>
				</table>
			</body>
			</html>
		`

		tpl, err := template.New("stats").Parse(htmlContent)

		if err != nil {
			return c.SendString("Error generating the template")
		}

		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return tpl.Execute(c.Response().BodyWriter(), pageData)
	}
}
