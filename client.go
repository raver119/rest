package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
)

var reDot = regexp.MustCompile("\\.$")

func ValidateIP(v string) bool {
	return net.ParseIP(string(v)) != nil
}

type Client struct {
	url string
	ttl uint32
}

type DomainResponse struct {
	A    []string `json:"a"`
	AAAA []string `json:"aaaa"`
}

func (z Client) Name() string {
	return "rest"
}

func BuildClient(url string, ttl uint32) (Client, error) {
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	return Client{url: url, ttl: ttl}, nil
}

func (z Client) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	qname := state.Name()

	var tt string
	switch state.QType() {
	case dns.TypeA:
		tt = "A"
	case dns.TypeAAAA:
		tt = "AAAA"
	default:
		// NXDOMAIN
		return dns.RcodeNameError, nil
	}

	domain := reDot.ReplaceAllString(qname, "")
	tt = "ALL"
	url := fmt.Sprintf("%v%v/%v", z.url, tt, domain)
	resp, err := http.Get(url)
	if err != nil {
		LogVerbose("failed to build REST request: [%v]; URL: [%v]", err, url)
		return dns.RcodeServerFailure, err
	}

	var response DomainResponse
	if resp.StatusCode == http.StatusOK {
		//
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			LogVerbose("failed to read response body: [%v];", err)
			return dns.RcodeServerFailure, err
		}

		err = json.Unmarshal(body, &response)
		if err != nil {
			LogVerbose("failed to deserialize response: [%v]; response: [%v];", err, body)
			return dns.RcodeServerFailure, err
		}

		LogVerbose("Domain [%v] resolves to {A: %v, AAAA: %v};", domain, response.A, response.AAAA)

		m := new(dns.Msg)
		switch state.QType() {
		case dns.TypeA:
			if len(response.A) == 0 {
				m.SetRcode(r, dns.RcodeSuccess)
				//m.Ns = []dns.RR{soa(n)}
				_ = w.WriteMsg(m)
				return dns.RcodeSuccess, nil
			}

			m.SetReply(r)
			m.Answer = z.buildAnswers(qname, state.QType(), response.A...)
		case dns.TypeAAAA:
			if len(response.AAAA) == 0 {
				m.SetRcode(r, dns.RcodeSuccess)
				//m.Ns = []dns.RR{soa(n)}
				_ = w.WriteMsg(m)
				return dns.RcodeSuccess, nil
			}

			m.SetReply(r)
			m.Answer = z.buildAnswers(qname, state.QType(), response.AAAA...)
		}

		m.Authoritative = true
		m.RecursionDesired = false
		m.RecursionAvailable = false
		_ = w.WriteMsg(m)
		return dns.RcodeSuccess, nil
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		LogVerbose("got non-200 response: [%v]; body: [%v];", resp.StatusCode, body)

		m := new(dns.Msg)
		m.Authoritative = true
		m.RecursionAvailable = false
		m.RecursionDesired = false
		m.SetRcode(r, dns.RcodeNameError)
		//m.Ns = []dns.RR{soa(n)}
		_ = w.WriteMsg(m)

		return dns.RcodeSuccess, nil
	}
}

func (z Client) buildAnswers(qname string, qtype uint16, ips ...string) []dns.RR {
	switch qtype {
	case dns.TypeA:
		return z.answersA(qname, ips...)
	case dns.TypeAAAA:
		return z.answersAAAA(qname, ips...)
	default:
		log.Fatalf("unexpected qtype: %v", qtype)
	}

	return []dns.RR{}
}

func (z Client) answersA(qname string, ips ...string) (answers []dns.RR) {
	answers = make([]dns.RR, len(ips))
	for i, v := range ips {
		r := new(dns.A)
		r.Hdr = dns.RR_Header{Name: qname, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: z.ttl}
		r.A = net.ParseIP(v)
		answers[i] = r
	}
	return
}

func (z Client) answersAAAA(qname string, ips ...string) (answers []dns.RR) {
	answers = make([]dns.RR, len(ips))
	for i, v := range ips {
		r := new(dns.AAAA)
		r.Hdr = dns.RR_Header{Name: qname, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: z.ttl}
		r.AAAA = net.ParseIP(v)
		answers[i] = r
	}
	return
}
