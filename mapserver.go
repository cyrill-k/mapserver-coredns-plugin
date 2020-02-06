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
	Next            plugin.Handler
	MapserverDomain string
	MapID           int64
	MapPK           string
	MapAddress      url.URL
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

	log.Printf("ServeDNS(%s -> %s)", request.QName(), requestedDomain)
	m, err := e.generateProofResponse(requestedDomain, request)
	if err != nil {
		return 0, err
	}
	w.WriteMsg(m)
	return 0, nil
}

func (e Mapserver) retrieveProofsFromMapServer(domains []string) ([]tclient.Proof, error) {
	mapClient := tclient.NewMapClient(e.MapAddress.Hostname()+":"+e.MapAddress.Port(), e.MapPK)
	defer mapClient.Close()

	proofs, err := mapClient.GetProofForDomains(e.MapID, mapClient.GetMapPK(), domains)
	if err != nil {
		return nil, fmt.Errorf("Couldn't retrieve proofs for all domains: %s", err)
	}

	log.Printf("Entries in map server...")
	for i, proof := range proofs {
		log.Printf("Entry (%s): %s", proof.GetDomain(), proof.ToString())
		err = proof.Validate(e.MapID, mapClient.GetMapPK(), common.DefaultTreeNonce, domains[i])
		if err != nil {
			return nil, fmt.Errorf("Entry %d (%s): Validate failed: %s", i, proof.GetDomain(), err)
		}
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
	log.Printf("Entry (%s): %s", proof.GetDomain(), proof.ToString())

	proofBytes, err := proof.MarshalBinary()

	// encode proof into Txt records
	proofStrings := common.BytesToStrings(proofBytes)

	log.Printf("r.size = %+v", r.Size())
	// add Txt records to DNS response
	// common.SetupEdns0Opt(r.Req, 4002)
	m := new(dns.Msg)
	m.SetReply(r.Req)
	// common.SetupEdns0Opt(m, 4003)

	hdr := dns.RR_Header{Name: r.QName(), Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0}
	m.Answer = append(m.Answer, &dns.TXT{Hdr: hdr, Txt: proofStrings})
	//
	// log.Printf("resp = %+v", m)

	return m, nil
}

// Name implements the Handler interface.
func (e Mapserver) Name() string { return "mapserver" }
