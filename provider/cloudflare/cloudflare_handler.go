package cloudflare

import (
	"context"
	"fmt"
	"strings"

	"github.com/0987363/dns-failover/models"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"

	cf "github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/dns"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/cloudflare/cloudflare-go/v4/zones"
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
	Email string
	Key   string
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
	IPType  string `json:"type"`
	ZoneID  string `json:"zone_id"`
	TTL     int32  `json:"ttl"`
}

func (rec *DNSRecord) GetIP() string {
	return rec.IP
}

func (rec *DNSRecord) String() string {
	return fmt.Sprintf("%+v", *rec)
}

// ZoneResponse is a wrapper for Zones.
type ZoneResponse struct {
	Zones   []Zone `json:"result"`
	Success bool   `json:"success"`
}

type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (provider *CloudflareProvider) UpdateDNS(domain string, drs []*models.DomainRecord) error {
	eTLD, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		return err
	}
	zoneID, err := provider.getZone(eTLD)
	if err != nil {
		return err
	}

	currents, err := provider.searchDNSRecords(zoneID, domain)
	if err != nil {
		return err
	}
	log.Infof("Found dns records: %+v", currents)

	add, del := models.CompareObjectsByIP(currents, drs)
	if len(add) > 0 {
		for _, rec := range []*models.DomainRecord(add) {
			log.Infof("Create dns records: %+v", rec)
			if err := provider.createRecord(zoneID, rec); err != nil {
				return err
			}
		}
	} else {
		for _, rec := range []*DNSRecord(del) {
			log.Infof("Delete dns records: %+v", rec)
			if err := provider.deleteRecord(rec); err != nil {
				return err
			}
		}
	}

	return nil
}

func (provider *CloudflareProvider) getZone(domain string) (string, error) {
	client := provider.newClient()
	res, err := client.Zones.List(context.TODO(), zones.ZoneListParams{
		Name: cf.F(domain),
	})
	if err != nil {
		return "", err
	}
	log.Debugf("Found zones: %+v\n", res.Result)

	if len(res.Result) > 0 {
		return res.Result[0].ID, nil
	}
	return "", models.Error("Could not found domain:", domain)
}

func (provider *CloudflareProvider) searchDNSRecords(zoneID, domain string) ([]*DNSRecord, error) {
	client := provider.newClient()
	res, err := client.DNS.Records.List(context.TODO(), dns.RecordListParams{
		ZoneID: cf.F(zoneID),
		Name: cf.F(dns.RecordListParamsName{
			Exact: cf.F(domain),
		}),
	})
	if err != nil {
		return nil, err
	}

	if len(res.Result) == 0 {
		return nil, models.Error("Could not found records:", zoneID)
	}

	records := []*DNSRecord{}
	for _, rec := range res.Result {
		rec := &DNSRecord{
			ID:      rec.ID,
			Name:    rec.Name,
			IP:      rec.Content,
			Proxied: rec.Proxied,
			IPType:  string(rec.Type),
			ZoneID:  zoneID,
		}
		records = append(records, rec)
	}
	log.Debugf("Found records: %+v\n", records)

	return records, nil
}

/*
func (provider *CloudflareProvider) getDNSRecords(zoneID string, drs []*models.DomainRecord) ([]*DNSRecord, error) {
	client := provider.newClient()
	res, err := client.DNS.Records.List(context.TODO(), dns.RecordListParams{
		ZoneID: cf.F(zoneID),
		Name:   cf.F(),
	})
	if err != nil {
		return nil, err
	}

	log.Infof("Found records: %+v\n", res.result)
	if len(res.result) == 0 {
		return nil, models.Error("Could not found records:", zoneID)
	}

	records := []*DNSRecord{}
	for _, rec := range res.Result {
		for _, dr := range drs {
			if rec.Name == dr.Name && rec.Content == dr.IP && rec.Type == dr.IPType {
				rec := &DNSRecord{
					ID:      rec.ID,
					Name:    dr.Name,
					IP:      dr.IP,
					Proxied: dr.Proxied,
					IPType:  dr.IPType,
					ZoneID:  zoneID,
				}
				records = append(records, rec)
				continue
			}
		}
	}

	return records, nil
}
*/

func getRecordType(ipType string) string {
	if ipType == "" || strings.ToUpper(ipType) == IPV4 {
		return IPTypeA
	} else if strings.ToUpper(ipType) == IPV6 {
		return IPTypeAAAA
	}
	return IPTypeA
}
func (provider *CloudflareProvider) createRecord(zoneID string, record *models.DomainRecord) error {
	client := provider.newClient()

	res, err := client.DNS.Records.New(context.TODO(), dns.RecordNewParams{
		ZoneID: cf.F(zoneID),
		Record: dns.ARecordParam{
			Name:    cf.F(record.Name),
			Content: cf.F(record.IP),
			Proxied: cf.F(record.Proxied),
			TTL:     cf.F(dns.TTL(60)),
			Type:    cf.F(dns.ARecordType(record.IPType)),
		},
	})

	log.Debugf("Record create response: %+v\n", res)

	return err
}

/*
// Update DNS A Record with new IP.
func (provider *CloudflareProvider) updateRecord(record *DNSRecord) error {
	client := provider.newClient()
	res, err := client.DNS.Records.Update(
		context.TODO(),
		record.ID,
		dns.RecordEditParams{
			ZoneID: cf.F(record.ZoneID),
			Record: dns.RecordParam{
				Name:    cf.F(record.IP),
				Proxied: cf.F(record.Proxied),
				TTL:     cf.F(dns.TTL(record.TTL)),
				Type:    cf.F(record.Type),
			},
		},
	)
	log.Infof("Record update response: %+v\n", res)

	return err
}
*/

func (provider *CloudflareProvider) deleteRecord(record *DNSRecord) error {
	client := provider.newClient()
	res, err := client.DNS.Records.Delete(
		context.TODO(),
		record.ID,
		dns.RecordDeleteParams{
			ZoneID: cf.F(record.ZoneID),
		},
	)
	log.Infof("Record delete response: %+v\n", res)
	return err
}

func (provider *CloudflareProvider) newClient() *cf.Client {
	if provider.Key != "" {
		return cf.NewClient(
			option.WithAPIToken(provider.Key),
		)
	}
	return cf.NewClient()
}
