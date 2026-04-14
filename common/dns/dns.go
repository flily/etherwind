package dns

import (
	"errors"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

type (
	Type          = dnsmessage.Type
	Message       = dnsmessage.Message
	Resource      = dnsmessage.Resource
	AAAAResource  = dnsmessage.AAAAResource
	AResource     = dnsmessage.AResource
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
