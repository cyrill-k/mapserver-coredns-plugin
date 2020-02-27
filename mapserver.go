// Package mapserver is a CoreDNS plugin that prints "example" to stdout on every packet received.
//
// It serves as an example CoreDNS plugin with numerous code comments.
package mapserver

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/cyrill-k/trustflex/common"
	"github.com/cyrill-k/trustflex/trillian/tclient"

	"github.com/coredns/coredns/plugin"
	//	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

// Mapserver is an example plugin to show how to write a plugin.
type Mapserver struct {
	Next                  plugin.Handler
	MapserverDomain       string
	MapID                 int64
	MapPK                 string
	MapAddress            url.URL
	MaxReceiveMessageSize int
}

// ServeDNS implements the plugin.Handler interface. This method gets called when example is used
// in a Server.
func (e Mapserver) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// This function could be simpler. I.e. just fmt.Println("example") here, but we want to show
	// a slightly more complex example as to make this more interesting.
	// Here we wrap the dns.ResponseWriter in a new ResponseWriter and call the next plugin, when the
	// answer comes back, it will print "example".

	request := request.Request{W: w, Req: r}
	requestedDomain := strings.TrimSuffix(request.QName(), ".")
	requestedDomain = strings.TrimSuffix(requestedDomain, e.MapserverDomain)
	requestedDomain = strings.TrimSuffix(requestedDomain, ".")

	log.Printf("Received DNS request (%s) from %s:%s", request.QName(), request.IP(), request.Port())
	m, err := e.generateProofResponse(requestedDomain, request)
	if err != nil {
		log.Printf("Failed to generate proof response: %s", err)
		return 0, err
	}
	log.Printf("Sending reply to %s:%s", request.IP(), request.Port())
	w.WriteMsg(m)
	return 0, nil
}

func (e Mapserver) retrieveProofsFromMapServer(domains []string) ([]tclient.Proof, error) {
	mapClient := tclient.NewMapClient(e.MapAddress.Hostname()+":"+e.MapAddress.Port(), e.MapPK, e.MaxReceiveMessageSize)
	defer mapClient.Close()

	proofs, err := mapClient.GetProofForDomains(e.MapID, mapClient.GetMapPK(), domains)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve proofs for all domains: %s", err)
	}

	for i, proof := range proofs {
		log.Printf("Validating %s ...", proof.ToString())
		err = proof.Validate(e.MapID, mapClient.GetMapPK(), common.DefaultTreeNonce, domains[i])
		if err != nil {
			return nil, fmt.Errorf("Entry %d (%s): Validate failed: %s", i, proof.GetDomain(), err)
		}
		log.Print("Validation succeeded")
	}
	return proofs, nil
}

func (e Mapserver) generateProofResponse(domain string, r request.Request) (*dns.Msg, error) {
	// get proof from map server using map_client
	proofs, err := e.retrieveProofsFromMapServer([]string{domain})
	if err != nil {
		return nil, err
	}
	if len(proofs) == 0 {
		return nil, fmt.Errorf("Empty proof returned")
	}
	proof := proofs[0]

	proofBytes, err := proof.MarshalBinary()

	// encode proof into Txt records
	proofStrings := common.BytesToStrings(proofBytes)

	// add Txt records to DNS response
	m := new(dns.Msg)
	m.SetReply(r.Req)

	// ahdr := dns.RR_Header{Name: r.QName(), Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0}
	// m.Answer = append(m.Answer, &dns.A{Hdr: ahdr, A: net.ParseIP("0.0.0.0")})

	txthdr := dns.RR_Header{Name: r.QName(), Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0}
	m.Answer = append(m.Answer, &dns.TXT{Hdr: txthdr, Txt: proofStrings})

	return m, nil
}

// Name implements the Handler interface.
func (e Mapserver) Name() string { return "mapserver" }
