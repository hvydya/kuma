package resolver

import (
	"sync"

	"github.com/miekg/dns"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/dns/vips"
)

type DNSResolver interface {
	GetDomain() string
	SetVIPs(list vips.List)
	SetVIPsChangedHandler(handler vips.ChangeHandler)

	ForwardLookup(service string) (string, error)
	ForwardLookupFQDN(name string) (string, error)
	ReverseLookup(ip string) (string, error)
}

type dnsResolver struct {
	sync.RWMutex
	domain  string
	viplist vips.List
	handler vips.ChangeHandler
}

var _ DNSResolver = &dnsResolver{}

func NewDNSResolver(domain string) DNSResolver {
	return &dnsResolver{
		domain: domain,
	}
}

func (d *dnsResolver) GetDomain() string {
	return d.domain
}

func (s *dnsResolver) SetVIPs(list vips.List) {
	s.Lock()
	defer s.Unlock()
	s.viplist = list
	if s.handler != nil {
		s.handler(s.viplist)
	}
}

func (s *dnsResolver) SetVIPsChangedHandler(handler vips.ChangeHandler) {
	s.Lock()
	defer s.Unlock()
	s.handler = handler
}

func (s *dnsResolver) ForwardLookup(service string) (string, error) {
	s.RLock()
	defer s.RUnlock()

	ip, found := s.viplist[service]

	if !found {
		return "", errors.Errorf("service [%s] not found in domain [%s].", service, s.domain)
	}
	return ip, nil
}

func (s *dnsResolver) ForwardLookupFQDN(name string) (string, error) {
	s.RLock()
	defer s.RUnlock()
	domain, err := s.domainFromName(name)
	if err != nil {
		return "", err
	}

	if domain != s.domain {
		return "", errors.Errorf("domain [%s] not found.", domain)
	}

	service, err := s.serviceFromName(name)
	if err != nil {
		return "", err
	}

	ip, found := s.viplist[service]
	if !found {
		return "", errors.Errorf("service [%s] not found in domain [%s].", service, domain)
	}

	return ip, nil
}

func (s *dnsResolver) ReverseLookup(ip string) (string, error) {
	s.RLock()
	defer s.RUnlock()

	for service, serviceIP := range s.viplist {
		if serviceIP == ip {
			return service + "." + s.domain, nil
		}
	}

	return "", errors.Errorf("IP [%s] not found", ip)
}

func (s *dnsResolver) domainFromName(name string) (string, error) {
	split := dns.SplitDomainName(name)
	if len(split) < 1 {
		return "", errors.Errorf("wrong DNS name: %s", name)
	}

	return split[len(split)-1], nil
}

func (s *dnsResolver) serviceFromName(name string) (string, error) {
	split := dns.SplitDomainName(name)
	if len(split) < 1 {
		return "", errors.Errorf("wrong DNS name: %s", name)
	}

	service := split[0]

	return service, nil
}
