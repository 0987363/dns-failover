package cloudflare

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/0987363/dns-failover/models"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"
)

const (
	IPV4 = "IPV4"
	IPV6 = "IPV6"

	IPTypeA    = "A"
	IPTypeAAAA = "AAAA"
)

const (
	baseUrl = "https://api.cloudflare.com/client/v4"
)

type CloudflareProvider struct {
	Key string
}

// DNSRecordResponse struct.
type DNSRecordResponse struct {
	Records []DNSRecord `json:"result"`
	Success bool        `json:"success"`
}

// DNSRecordUpdateResponse struct.
type DNSRecordUpdateResponse struct {
	Record  DNSRecord `json:"result"`
	Success bool      `json:"success"`
}

// DNSRecord for Cloudflare API.
type DNSRecord struct {
	ID      string `json:"id"`
	IP      string `json:"content"`
	Name    string `json:"name"`
	Proxied bool   `json:"proxied"`
	Type    string `json:"type"`
	ZoneID  string `json:"zone_id"`
	TTL     int32  `json:"ttl"`
}

// ZoneResponse is a wrapper for Zones.
type ZoneResponse struct {
	Zones   []Zone `json:"result"`
	Success bool   `json:"success"`
}

// Zone object with id and name.
type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (provider *CloudflareProvider) UpdateDNS(dr *models.DomainRecord) error {
	eTLD, err := publicsuffix.EffectiveTLDPlusOne(dr.Name)
	if err != nil {
		return err
	}
	zoneID, err := provider.getZone(eTLD)
	if err != nil {
		return err
	}

	records, err := provider.getDNSRecords(zoneID, dr.IPType)
	if err != nil {
		return err
	}

	for _, rec := range records {
		if rec.Name != dr.Name {
			log.Debug("Skipping record:", rec.Name)
			continue
		}

		log.Debug("Found record:", rec.Name)
		if rec.IP == dr.IP {
			log.Info("Skipping update, Same IP ", rec.IP)
			return nil
		}

		log.Infof("IP mismatch: Current(%+v) vs Cloudflare(%+v)\n", dr.IP, rec.IP)
		rec.ZoneID = zoneID
		rec.IP = dr.IP
		rec.TTL = 60
		if err := provider.updateRecord(rec); err != nil {
			return err
		}

		return nil
	}

	record := &DNSRecord{
		Type:    getRecordType(dr.IPType),
		IP:      dr.IP,
		TTL:     60,
		Name:    dr.Name,
		Proxied: dr.Proxied,
		ZoneID:  zoneID,
	}

	log.Debugf("Record %+v not found, will create it.\n", record)
	if err := provider.createRecord(record); err != nil {
		return err
	}
	log.Infof("Record [%+v] created with IP address\n", dr)

	return nil
}

// Create a new request with auth in place and optional proxy.
func (provider *CloudflareProvider) newRequest(method, url string, body io.Reader) (*http.Request, *http.Client) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, _ := http.NewRequest(method, baseUrl+url, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", provider.Key))

	return req, client
}

// Find the correct zone via domain name.
func (provider *CloudflareProvider) getZone(domain string) (string, error) {
	var z ZoneResponse

	req, client := provider.newRequest("GET", fmt.Sprintf("/zones?name=%s", domain), nil)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body, _ := io.ReadAll(resp.Body)
	log.Debug("Response body: ", string(body))
	if err = json.Unmarshal(body, &z); err != nil {
		return "", err
	}
	log.Debugf("Get zone response: %+v\n", z)
	if !z.Success {
		return "", models.Error("Get zone unsuccess: ", z)
	}

	for _, zone := range z.Zones {
		if zone.Name == domain {
			return zone.ID, nil
		}
	}
	return "", errors.New("Not found " + domain)
}

// Get all DNS A records for a zone.
func (provider *CloudflareProvider) getDNSRecords(zoneID, ipType string) ([]DNSRecord, error) {
	var r DNSRecordResponse
	recordType := getRecordType(ipType)

	log.Debug("Querying records with type: ", recordType)
	req, client := provider.newRequest("GET", fmt.Sprintf("/zones/"+zoneID+"/dns_records?type=%s&page=1&per_page=500", recordType), nil)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, _ := io.ReadAll(resp.Body)
	log.Debug("Response body: ", string(body))
	if err = json.Unmarshal(body, &r); err != nil {
		return nil, err
	}

	log.Debugf("Get dns record response: %+v\n", r)
	if !r.Success {
		return nil, models.Error("Get dns unsuccess: ", r)
	}
	return r.Records, nil
}
func getRecordType(ipType string) string {
	if ipType == "" || strings.ToUpper(ipType) == IPV4 {
		return IPTypeA
	} else if strings.ToUpper(ipType) == IPV6 {
		return IPTypeAAAA
	}
	return IPTypeA
}
func (provider *CloudflareProvider) createRecord(record *DNSRecord) error {

	content, err := json.Marshal(record)
	if err != nil {
		log.Error("Encoder error: ", err)
		return err
	}

	req, client := provider.newRequest("POST", fmt.Sprintf("/zones/%s/dns_records", record.ZoneID), bytes.NewBuffer(content))
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var r DNSRecordUpdateResponse
	log.Debug("Create record response body: ", string(body))
	if err = json.Unmarshal(body, &r); err != nil {
		return err
	}

	log.Debugf("Create record response: %+v\n", r)
	if !r.Success {
		return models.Error("Create record unsuccess: ", r)
	}

	return nil
}

// Update DNS A Record with new IP.
func (provider *CloudflareProvider) updateRecord(record DNSRecord) error {
	var r DNSRecordUpdateResponse

	j, _ := json.Marshal(record)
	req, client := provider.newRequest("PUT",
		"/zones/"+record.ZoneID+"/dns_records/"+record.ID,
		bytes.NewBuffer(j),
	)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	log.Debug("Update record response body: ", string(body))
	if err = json.Unmarshal(body, &r); err != nil {
		return err
	}

	log.Debugf("Update record response: %+v\n", r)
	if !r.Success {
		return models.Error("Update record unsuccess: ", r)
	}
	return nil
}
