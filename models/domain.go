package models

import "fmt"

type DomainRecord struct {
	Name    string
	IP      string
	IPType  string
	Proxied bool
}

func (rec *DomainRecord) GetIP() string {
	return rec.IP
}

func (rec *DomainRecord) String() string {
	return fmt.Sprintf("%+v", *rec)
}
