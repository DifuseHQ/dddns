package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/DifuseHQ/dddns/pkg/logger"
	"os"
)

type Config struct {
	DBPath   string `json:"db_path"`
	LogPath  string `json:"log_path"`
	DNSAddr  string `json:"dns_addr"`
	DNSPort  string `json:"dns_port"`
	HTTPAddr string `json:"http_addr"`
	HTTPPort string `json:"http_port"`
	Domain   string `json:"domain"`
}

func InitConfig() Config {
	var cfg Config
	configPath := flag.String("config", "", "Path to JSON config file")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, logger.AsciiArt)
		fmt.Fprintf(os.Stderr, "Configuration options:\n")
		flag.PrintDefaults()
	}

	flag.StringVar(&cfg.DBPath, "db-path", "./data/ddns.db", "Database path")
	flag.StringVar(&cfg.LogPath, "log-path", "./data/dddns.log", "Log file path")
	flag.StringVar(&cfg.HTTPPort, "dns-addr", "::", "DNS server bind address")
	flag.StringVar(&cfg.DNSPort, "dns-port", "5544", "DNS server port")
	flag.StringVar(&cfg.HTTPPort, "http-addr", "::", "HTTP server bind address")
	flag.StringVar(&cfg.HTTPPort, "http-port", "3000", "HTTP server port")
	flag.StringVar(&cfg.Domain, "domain", "dddns.network", "Domain to use for DNS records")

	flag.Parse()

	if *configPath != "" {
		err := loadConfigFromJSON(*configPath, &cfg)
		if err != nil {
			logger.Log.Fatal("Error loading config from JSON", err)
		}
	}

	return cfg
}

func loadConfigFromJSON(filePath string, config *Config) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(config)
}
