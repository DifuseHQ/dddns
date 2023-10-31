# DDDNS - Difuse DNS Service

DDDNS is an easy way to run an authoritative DNS server for your domain. It comes with a rudimentary authentication mechanism that can be adapted to your needs.

## Introduction

The reason this was built was the fact that BIND is a pain to configure and maintain especially if you want to let users manage their own subdomains. DDDNS is a simple solution to this problem. It's written in Go and compiles to a single binary. It's also very fast and can handle thousands of requests per second. It uses sqlite as the database as it's very easy to setup and maintain. This was designed to be used with [Difuse](https://difuse.io) and all users of Difuse devices get a free difusedns.com subdomain.

## Features

* DNS server with support for A, AAAA, SOA, and NS records.
* HTTP API for managing DNS records.
* Customizable logging and database configuration.
* Easy to set up and configure.

## Requirements

* Go 1.11 or higher

## Building

git clone https://github.com/DifuseHQ/dddns.git
cd dddns
make build

## Configuration

The program can be configured using a JSON configuration file or command-line flags. The available configuration options include:

* `db_path`: Database path (default: ./data/ddns.db)
* `log_path`: Log file path (default: ./data/dddns.log)
* `dns_addr`: DNS server bind address (default: ::)
* `dns_port`: DNS server port (default: 5544)
* `http_addr`: HTTP server bind address (default: ::)
* `http_port`: HTTP server port (default: 3000)
* `domain`: Domain for DNS records (default: difusedns.com)
* `name_server_domain`: Domain for name server records (default: ns1.difuse.io)
* `mail_box`: Mailbox for SOA records (default: admin.difusedns.com)
* `authoritative`: Whether the server is authoritative for the domain (default: true)
* `log_level`: Log level (0-1) (default: 0)

### Using Configuration File

Create a JSON file (config.json, for example) with the necessary configurations:

```json
{
    "db_path": "./data/ddns.db",
    "log_path": "./data/dddns.log",
    "dns_addr": "::",
    "dns_port": 5544,
    "http_addr": "::",
    "http_port": 3000,
    "domain": "difusedns.com",
    "name_server_domain": "ns1.difuse.io",
    "mail_box": "admin.difusedns.com",
    "authoritative": true,
    "log_level": 0
}
```

### Using Command-Line Flags

Alternatively, you can specify configuration using command-line flags. Refer to the `--help` flag for more information.

## Running DDDNS

To start the DNS and HTTP servers, run:

```bash
./dddns -config config.json
```

Or, using command-line flags:

```bash
./dddns --dns-addr "::" --dns-port "5544" --http-addr "::" --http-port "3000"
``` 

## API Endpoints

The service provides several HTTP endpoints for DNS record management and querying server statistics:

* `GET /`: Retrieve DNS server statistics.
* `GET /checks/is-domain-available/:domain`: Check if a domain is available.
* `GET /checks/is-domain-taken-by-someone/:domain`: Check if a domain is taken by someone else.
* `POST /manage-record/create-or-update`: Create or update a DNS record.
* `DELETE /manage-record/delete`: Delete a DNS record.

License

DDDNS is licensed under the MIT License. See [LICENSE.md](LICENSE.md) for more information.