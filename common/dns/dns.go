package dns

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

type (
	Type          = dnsmessage.Type
	Message       = dnsmessage.Message
	Resource      = dnsmessage.Resource
	AResource     = dnsmessage.AResource
	AAAAResource  = dnsmessage.AAAAResource
	CNAMEResource = dnsmessage.CNAMEResource
	MXResource    = dnsmessage.MXResource
	NSResource    = dnsmessage.NSResource
	OPTResource   = dnsmessage.OPTResource
	PTRResource   = dnsmessage.PTRResource
	SOAResource   = dnsmessage.SOAResource
	SRVResource   = dnsmessage.SRVResource
	SVCBResource  = dnsmessage.SVCBResource
	TXTResource   = dnsmessage.TXTResource
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

func ParseTypes(name string) []Type {
	parts := strings.Split(name, "+")

	types := make([]Type, 0, len(parts))
	for _, part := range parts {
		types = append(types, GetType(part))
	}

	return types
}

var (
	ErrNotDialed = errors.New("dial before query")
)

func CanonicalizeName(name string) string {
	if len(name) == 0 {
		return "."
	}

	if name[len(name)-1] != '.' {
		return name + "."
	}

	return name
}

func MergeAnswers(messages ...*Message) *Message {
	result := &Message{}

	for _, msg := range messages {
		result.Answers = append(result.Answers, msg.Answers...)
		result.Authorities = append(result.Authorities, msg.Authorities...)
		result.Additionals = append(result.Additionals, msg.Additionals...)
	}

	return result
}

func ResourceKey(r Resource) string {
	result := "UNKNOWN"
	switch t := r.Body.(type) {
	case *AResource:
		result = fmt.Sprintf("A:%s", net.IP(t.A[:]))

	case *AAAAResource:
		result = fmt.Sprintf("AAAA:%s", net.IP(t.AAAA[:]))

	case *CNAMEResource:
		result = fmt.Sprintf("CNAME:%s:%s", r.Header.Name, t.CNAME)

	case *MXResource:
		result = fmt.Sprintf("MX:%s:%d", t.MX, t.Pref)

	case *NSResource:
		result = fmt.Sprintf("NS:%s:%s", r.Header.Name, t.NS)

	case *PTRResource:
		result = fmt.Sprintf("PTR:%s:%s", r.Header.Name, t.PTR)

	case *SOAResource:
		result = fmt.Sprintf("SOA:%s:%s %s", r.Header.Name, t.NS, t.MBox)

	case *SRVResource:
		result = fmt.Sprintf("SRV:%s:%d %d %d", r.Header.Name, t.Port, t.Priority, t.Weight)

	case *SVCBResource:
		result = fmt.Sprintf("SVCB:%s:%d %s", r.Header.Name, t.Priority, t.Target)

	case *TXTResource:
		result = fmt.Sprintf("TXT:%s:%s", r.Header.Name, strings.Join(t.TXT, ","))

	}

	return result
}
