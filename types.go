package main

import (
	"io"
	"net/http"
	"time"
)

type Domain struct {
	Name      string       `json:"name"`
	TLD       string       `json:"tld"`
	CreatedAt time.Time    `json:"created_at"`
	Health    DomainHealth `json:"health"`
}

type DomainStats struct {
	TotalDomains   int            `json:"total_domains"`
	DomainsPerTLD  map[string]int `json:"domains_per_tld"`
	LastUpdateTime time.Time      `json:"last_update_time"`
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

type WhoisInfo struct {
	DomainName     string    `json:"domain_name"`
	Registrar      string    `json:"registrar,omitempty"`
	CreatedDate    time.Time `json:"created_date,omitempty"`
	ExpiryDate     time.Time `json:"expiry_date,omitempty"`
	LastUpdated    time.Time `json:"last_updated,omitempty"`
	NameServers    []string  `json:"nameservers,omitempty"`
	Status         []string  `json:"status,omitempty"`
	DNSSec         bool      `json:"dnssec,omitempty"`
	Registrant     Contact   `json:"registrant,omitempty"`
	Administrative Contact   `json:"administrative,omitempty"`
	Technical      Contact   `json:"technical,omitempty"`
	Error          string    `json:"error,omitempty"`
}

type Contact struct {
	Name         string `json:"name,omitempty"`
	Organization string `json:"organization,omitempty"`
	Street       string `json:"street,omitempty"`
	City         string `json:"city,omitempty"`
	Province     string `json:"province,omitempty"`
	PostalCode   string `json:"postal_code,omitempty"`
	Country      string `json:"country,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Email        string `json:"email,omitempty"`
}

type DNSInfo struct {
	Domain      string     `json:"domain"`
	IPAddresses []IPInfo   `json:"ip_addresses,omitempty"`
	Hostnames   []string   `json:"hostnames,omitempty"`
	MXRecords   []MXRecord `json:"mx_records,omitempty"`
	TXTRecords  []string   `json:"txt_records,omitempty"`
	NSRecords   []NSRecord `json:"ns_records,omitempty"`
	CNAMEs      []string   `json:"cnames,omitempty"`
	Error       string     `json:"error,omitempty"`
}

type IPInfo struct {
	Address  string   `json:"address"`
	Version  int      `json:"version"` // 4 or 6
	Reverse  []string `json:"reverse,omitempty"`
	ASN      string   `json:"asn,omitempty"`
	Location string   `json:"location,omitempty"`
	ISP      string   `json:"isp,omitempty"`
}

type MXRecord struct {
	Host     string `json:"host"`
	Priority int    `json:"priority"`
}

type NSRecord struct {
	Host string   `json:"host"`
	IPs  []string `json:"ips,omitempty"`
}

type DomainHealth struct {
	IsOnline     bool          `json:"is_online"`
	Protocol     string        `json:"protocol"` // http/https
	StatusCode   int           `json:"status_code,omitempty"`
	ResponseTime time.Duration `json:"response_time,omitempty"`
	Error        string        `json:"error,omitempty"`
}
