package models

type DomainRecord struct {
	Name    string
	IP      string
	IPType  string
	Proxied bool
}
