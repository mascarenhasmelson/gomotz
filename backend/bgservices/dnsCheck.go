package bgservices

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/mascarenhasmelson/gomotz/utils"
	"github.com/miekg/dns"
)

func buildMessage(qtype, name string) (*dns.Msg, error) {
	typeMap := map[string]uint16{
		"a": dns.TypeA, "aaaa": dns.TypeAAAA, "cname": dns.TypeCNAME,
		"mx": dns.TypeMX, "ns": dns.TypeNS, "txt": dns.TypeTXT,
		"srv": dns.TypeSRV, "soa": dns.TypeSOA, "ptr": dns.TypePTR,
		"caa": dns.TypeCAA, "dnskey": dns.TypeDNSKEY, "ds": dns.TypeDS,
		"naptr": dns.TypeNAPTR, "rrsig": dns.TypeRRSIG,
	}

	qtype = strings.ToLower(qtype)
	t, ok := typeMap[qtype]
	if !ok {
		return nil, fmt.Errorf("unsupported RR type: %s", qtype)
	}

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(name), t)
	m.RecursionDesired = true
	return m, nil
}

func parseAnswer(rrs []dns.RR) []utils.RRResponse {
	out := []utils.RRResponse{}

	for _, rr := range rrs {
		h := rr.Header()
		t := dns.TypeToString[h.Rrtype]

		switch v := rr.(type) {

		case *dns.A:
			out = append(out, utils.RRResponse{h.Name, t, h.Ttl, v.A.String()})
		case *dns.AAAA:
			out = append(out, utils.RRResponse{h.Name, t, h.Ttl, v.AAAA.String()})
		case *dns.CNAME:
			out = append(out, utils.RRResponse{h.Name, t, h.Ttl, v.Target})
		case *dns.NS:
			out = append(out, utils.RRResponse{h.Name, t, h.Ttl, v.Ns})
		case *dns.PTR:
			out = append(out, utils.RRResponse{h.Name, t, h.Ttl, v.Ptr})
		case *dns.TXT:
			for _, txt := range v.Txt {
				out = append(out, utils.RRResponse{h.Name, t, h.Ttl, txt})
			}
		case *dns.MX:
			out = append(out, utils.RRResponse{
				h.Name, t, h.Ttl,
				fmt.Sprintf("%d %s", v.Preference, v.Mx),
			})
		case *dns.SOA:
			out = append(out, utils.RRResponse{
				h.Name, t, h.Ttl,
				fmt.Sprintf("%s %s %d %d %d %d %d",
					v.Ns, v.Mbox, v.Serial, v.Refresh,
					v.Retry, v.Expire, v.Minttl),
			})
		case *dns.SRV:
			out = append(out, utils.RRResponse{
				h.Name, t, h.Ttl,
				fmt.Sprintf("%d %d %d %s", v.Priority, v.Weight, v.Port, v.Target),
			})
		case *dns.CAA:
			out = append(out, utils.RRResponse{
				h.Name, t, h.Ttl,
				fmt.Sprintf("%d %s \"%s\"", v.Flag, v.Tag, v.Value),
			})
		case *dns.NAPTR:
			out = append(out, utils.RRResponse{
				h.Name, t, h.Ttl,
				fmt.Sprintf("%d %d \"%s\" \"%s\" \"%s\" %s",
					v.Order, v.Preference, v.Flags, v.Service,
					v.Regexp, v.Replacement),
			})
		default:
			// Get the string representation and remove the header part
			rrStr := rr.String()
			if idx := strings.Index(rrStr, "\t"); idx != -1 {
				rrStr = strings.TrimSpace(rrStr[idx:])
			}
			out = append(out, utils.RRResponse{
				h.Name, t, h.Ttl, rrStr,
			})
		}
	}
	return out
}

func QueryRR(req utils.Request) ([]utils.RRResponse, error) {
	if req.Server == "" {
		req.Server = "8.8.8.8"
	}
	if net.ParseIP(req.Server) == nil {
		return nil, fmt.Errorf("invalid DNS server IP: %s", req.Server)
	}
	msg, err := buildMessage(req.Type, req.Domain)
	if err != nil {
		return nil, err
	}
	c := &dns.Client{Net: "udp", Timeout: 10 * time.Second}
	resp, _, err := c.Exchange(msg, req.Server+":53")
	if err != nil || (resp != nil && resp.Truncated) {
		c.Net = "tcp"
		resp, _, err = c.Exchange(msg, req.Server+":53")
	}
	if err != nil {
		return nil, fmt.Errorf("DNS query failed: %v", err)
	}
	if resp == nil {
		return nil, fmt.Errorf("empty DNS response")
	}
	if resp.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("DNS error: %s", dns.RcodeToString[resp.Rcode])
	}
	return parseAnswer(resp.Answer), nil
}
