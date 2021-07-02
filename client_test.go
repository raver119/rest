package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"
	"github.com/gorilla/mux"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

const url = "http://127.0.0.1:8089/rest/v1/dns"
const ttl = 300

type TestCase struct {
	name        string
	r           *dns.Msg
	tc          test.Case
	wantFailure bool
	wantAnswer  []dns.RR
}

func TestClient_ServeDNS(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/rest/v1/dns/{qtype}/{domain}", func(w http.ResponseWriter, r *http.Request) {
		qtype := mux.Vars(r)["qtype"]
		domain := mux.Vars(r)["domain"]

		if qtype != "A" && qtype != "AAAA" && qtype != "ALL" {
			http.Error(w, "Only A and AAAA supported", http.StatusBadRequest)
			return
		}

		var addr []string
		var js []byte
		var found bool
		for _, v := range testCases {
			if strings.HasPrefix(v.Qname, domain) && !strings.Contains(v.Qname, "non-existent") {
				found = true
				switch qtype {
				case "A":
					addr = append(addr, "10.0.0.1")
					js, _ = json.Marshal(addr)
				case "AAAA":
					addr = append(addr, "fe80::50cc:d1ff:fe57:8cb6")
					js, _ = json.Marshal(addr)
				case "ALL":
					js, _ = json.Marshal(DomainResponse{A: []string{"10.0.0.1"}, AAAA: []string{"fe80::50cc:d1ff:fe57:8cb6"}})
				default:
					panic("Unknown")
				}

				_, _ = w.Write(js)
				break
			}
		}

		if !found {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	})
	srv := &http.Server{Addr: ":8089", Handler: r}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		_ = srv.ListenAndServe()
	}()
	time.Sleep(1 * time.Second)

	tests := []TestCase{}

	for i, v := range testCases {
		tests = append(tests, TestCase{fmt.Sprintf("test_%v", i), v.Msg(), v, len(v.Answer) == 0, v.Answer})
	}

	for _, tt := range tests {
		ctx := context.TODO()
		rec := dnstest.NewRecorder(&test.ResponseWriter{})

		t.Run(tt.name, func(t *testing.T) {
			z, _ := BuildClient(url, ttl)

			_, _ = z.ServeDNS(ctx, rec, tt.r)

			require.Equal(t, tt.wantFailure, rec.Msg == nil)

			// validate answers
			if resp := rec.Msg; rec.Msg != nil {
				if err := test.SortAndCheck(resp, tt.tc); err != nil {
					t.Error(err)
				}
			}

			if rec.Msg != nil && len(rec.Msg.Answer) > 0 {
				require.True(t, rec.Msg.Authoritative)
			}
		})
	}

	_ = srv.Shutdown(context.TODO())
	wg.Wait()
}

var testCases = []test.Case{
	{
		Qname: "example.org.", Qtype: dns.TypeA,
		Answer: []dns.RR{
			test.A("example.org. 300	IN	A 10.0.0.1"),
		},
	},
	{
		Qname: "example.org.", Qtype: dns.TypeAAAA,
		Answer: []dns.RR{
			test.AAAA("example.org. 300	IN	AAAA fe80::50cc:d1ff:fe57:8cb6"),
		},
	},
	{
		Qname: "example.org.", Qtype: dns.TypeMX,
		Answer: []dns.RR{},
	},
	{
		Qname: "non-existent.org.", Qtype: dns.TypeA,
		Answer: []dns.RR{},
	},
}
