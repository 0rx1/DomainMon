package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
)

var (
	dnsClient = &net.Resolver{
		PreferGo: true,
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
	}

	httpClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}
)

func (s *Server) handleWhoisLookup(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "domain parameter is required", http.StatusBadRequest)
		return
	}

	// Check cache first
	if cached, ok := s.cache.Get(fmt.Sprintf("whois:%s", domain)); ok {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		json.NewEncoder(w).Encode(cached)
		return
	}

	whoisInfo := &WhoisInfo{
		DomainName: domain,
	}

	raw, err := whois.Whois(domain)
	if err != nil {
		whoisInfo.Error = fmt.Sprintf("WHOIS lookup failed: %v", err)
	} else {
		result, err := whoisparser.Parse(raw)
		if err != nil {
			whoisInfo.Error = fmt.Sprintf("WHOIS parsing failed: %v", err)
		} else {
			whoisInfo.Registrar = result.Registrar.Name
			whoisInfo.NameServers = result.Domain.NameServers
			whoisInfo.Status = result.Domain.Status
			whoisInfo.DNSSec = result.Domain.DNSSec

			// Parse dates
			if created, err := time.Parse(time.RFC3339, result.Domain.CreatedDate); err == nil {
				whoisInfo.CreatedDate = created
			}
			if expires, err := time.Parse(time.RFC3339, result.Domain.ExpirationDate); err == nil {
				whoisInfo.ExpiryDate = expires
			}
			if updated, err := time.Parse(time.RFC3339, result.Domain.UpdatedDate); err == nil {
				whoisInfo.LastUpdated = updated
			}

			// Parse contacts
			if result.Registrant != nil {
				whoisInfo.Registrant = Contact{
					Name:         result.Registrant.Name,
					Organization: result.Registrant.Organization,
					Street:       result.Registrant.Street,
					City:         result.Registrant.City,
					Province:     result.Registrant.Province,
					PostalCode:   result.Registrant.PostalCode,
					Country:      result.Registrant.Country,
					Phone:        result.Registrant.Phone,
					Email:        result.Registrant.Email,
				}
			}

			if result.Administrative != nil {
				whoisInfo.Administrative = Contact{
					Name:         result.Administrative.Name,
					Organization: result.Administrative.Organization,
					Street:       result.Administrative.Street,
					City:         result.Administrative.City,
					Province:     result.Administrative.Province,
					PostalCode:   result.Administrative.PostalCode,
					Country:      result.Administrative.Country,
					Phone:        result.Administrative.Phone,
					Email:        result.Administrative.Email,
				}
			}

			if result.Technical != nil {
				whoisInfo.Technical = Contact{
					Name:         result.Technical.Name,
					Organization: result.Technical.Organization,
					Street:       result.Technical.Street,
					City:         result.Technical.City,
					Province:     result.Technical.Province,
					PostalCode:   result.Technical.PostalCode,
					Country:      result.Technical.Country,
					Phone:        result.Technical.Phone,
					Email:        result.Technical.Email,
				}
			}
		}
	}

	// Cache the result
	s.cache.Set(fmt.Sprintf("whois:%s", domain), whoisInfo)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(whoisInfo)
}

func (s *Server) handleReverseDNS(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "domain parameter is required", http.StatusBadRequest)
		return
	}

	// Check cache first
	if cached, ok := s.cache.Get(fmt.Sprintf("dns:%s", domain)); ok {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		json.NewEncoder(w).Encode(cached)
		return
	}

	dnsInfo := &DNSInfo{
		Domain: domain,
	}

	// Get all IP addresses (both IPv4 and IPv6)
	ips, err := net.LookupIP(domain)
	if err != nil {
		dnsInfo.Error = fmt.Sprintf("DNS lookup failed: %v", err)
	} else {
		// Process IP addresses
		for _, ip := range ips {
			ipInfo := IPInfo{
				Address: ip.String(),
				Version: 4,
			}
			if ip.To4() == nil {
				ipInfo.Version = 6
			}

			// Get reverse DNS
			names, err := net.LookupAddr(ip.String())
			if err == nil {
				for i := range names {
					names[i] = strings.TrimSuffix(names[i], ".")
				}
				ipInfo.Reverse = names
			}

			dnsInfo.IPAddresses = append(dnsInfo.IPAddresses, ipInfo)
		}

		// Get MX records
		mxs, err := net.LookupMX(domain)
		if err == nil {
			for _, mx := range mxs {
				dnsInfo.MXRecords = append(dnsInfo.MXRecords, MXRecord{
					Host:     strings.TrimSuffix(mx.Host, "."),
					Priority: int(mx.Pref),
				})
			}
		}

		// Get TXT records
		txts, err := net.LookupTXT(domain)
		if err == nil {
			dnsInfo.TXTRecords = txts
		}

		// Get NS records
		nss, err := net.LookupNS(domain)
		if err == nil {
			for _, ns := range nss {
				nsRecord := NSRecord{
					Host: strings.TrimSuffix(ns.Host, "."),
				}
				// Look up nameserver IPs
				if nsIPs, err := net.LookupIP(ns.Host); err == nil {
					for _, ip := range nsIPs {
						nsRecord.IPs = append(nsRecord.IPs, ip.String())
					}
				}
				dnsInfo.NSRecords = append(dnsInfo.NSRecords, nsRecord)
			}
		}

		// Get CNAME records
		cname, err := net.LookupCNAME(domain)
		if err == nil {
			dnsInfo.CNAMEs = append(dnsInfo.CNAMEs, strings.TrimSuffix(cname, "."))
		}
	}

	// Cache the result
	s.cache.Set(fmt.Sprintf("dns:%s", domain), dnsInfo)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(dnsInfo)
}

func (s *Server) performWhoisLookup(domain string) WhoisInfo {
	whoisInfo := WhoisInfo{DomainName: domain}
	raw, err := whois.Whois(domain)
	if err != nil {
		whoisInfo.Error = fmt.Sprintf("WHOIS lookup failed: %v", err)
		return whoisInfo
	}

	result, err := whoisparser.Parse(raw)
	if err != nil {
		whoisInfo.Error = fmt.Sprintf("WHOIS parsing failed: %v", err)
		return whoisInfo
	}

	// Fill in the WHOIS info (same as handleWhoisLookup)
	whoisInfo.Registrar = result.Registrar.Name
	whoisInfo.NameServers = result.Domain.NameServers
	whoisInfo.Status = result.Domain.Status
	whoisInfo.DNSSec = result.Domain.DNSSec
	// ... rest of the WHOIS parsing logic

	return whoisInfo
}

func (s *Server) performDNSLookup(domain string) DNSInfo {
	return DNSInfo{Domain: domain} // Implement full DNS lookup logic as needed
}
