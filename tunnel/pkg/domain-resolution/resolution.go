package domain_resolution

import "tunnel/pkg/config"

type IDomainResolution interface {
	CheckSubDomainExists(subDomain string) (bool, error)
	AddSubDomainRecord(rr, typ, value string) (string, error)
	UpdateSubDomainRemark(recordID, remark string) error
}
type IDNSFactory interface {
	NewDns() IDomainResolution
}

type dnsFactory struct {
	config *config.Config
}

func NewDnsFactory(config *config.Config) IDNSFactory {
	return &dnsFactory{
		config: config,
	}
}
