package dns

import (
	"errors"

	"golang.org/x/net/dns/dnsmessage"
)

type (
	Type     = dnsmessage.Type
	Resource = dnsmessage.Resource
)

const (
	TypeA     = dnsmessage.TypeA
	TypeNS    = dnsmessage.TypeNS
	TypeCNAME = dnsmessage.TypeCNAME
	TypeSOA   = dnsmessage.TypeSOA
	TypePTR   = dnsmessage.TypePTR
	TypeMX    = dnsmessage.TypeMX
	TypeTXT   = dnsmessage.TypeTXT
	TypeAAAA  = dnsmessage.TypeAAAA
	TypeSRV   = dnsmessage.TypeSRV
	TypeOPT   = dnsmessage.TypeOPT
	TypeSVCB  = dnsmessage.TypeSVCB
	TypeHTTPS = dnsmessage.TypeHTTPS
	TypeWKS   = dnsmessage.TypeWKS
	TypeHINFO = dnsmessage.TypeHINFO
	TypeAXFR  = dnsmessage.TypeAXFR
	TypeALL   = dnsmessage.TypeALL
)

var typeNames = map[string]Type{
	"A":     TypeA,
	"NS":    TypeNS,
	"CNAME": TypeCNAME,
	"SOA":   TypeSOA,
	"PTR":   TypePTR,
	"MX":    TypeMX,
	"TXT":   TypeTXT,
	"AAAA":  TypeAAAA,
	"SRV":   TypeSRV,
	"OPT":   TypeOPT,
	"SVCB":  TypeSVCB,
	"HTTPS": TypeHTTPS,
	"WKS":   TypeWKS,
	"HINFO": TypeHINFO,
	"AXFR":  TypeAXFR,
	"ALL":   TypeALL,
}

func GetType(name string) Type {
	if t, ok := typeNames[name]; ok {
		return t
	}

	return Type(0)
}

var (
	ErrNotDialed = errors.New("dial before query")
)
