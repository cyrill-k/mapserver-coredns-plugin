// Package mapserver is a CoreDNS plugin that prints "example" to stdout on every packet received.
//
// It serves as an example CoreDNS plugin with numerous code comments.
package mapserver

import (
	"context"
	"crypto"
	"fmt"
	"log"
	"net/url"

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
	MapPK           crypto.PublicKey
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
	m, err := e.generateProofResponse(e.MapserverDomain, request)
	if err != nil {
		return 0, err
	}
	w.WriteMsg(m)
	return 0, nil
}

func (e Mapserver) retrieveProofsFromMapServer(domains []string) {
	mapClient := tclient.NewMapClient(e.MapAddress.Hostname()+":"+e.MapAddress.Port(), common.LoadPK(e.MapPK).(crypto.PublicKey))
	defer mapClient.Close()

	proofs, err := mapClient.GetProofForDomains(e.MapID, mapClient.GetMapPK(), domains)
	common.LogError("Couldn't retrieve proofs for all domains: %s", err)

	log.Print("Entries in map server...")
	for i, proof := range proofs {
		err := proof.Validate(e.MapID, mapClient.GetMapPK(), common.DefaultTreeNonce)
		if err != nil {
			log.Fatalf("Entry %d (%s): Validate failed: %s", i, proof.GetDomain(), err)
		}
		log.Printf("Entry %d (%s): %s", i, proof.GetDomain(), proof.ToString())
	}
}

func (e Mapserver) generateProofResponse(domain string, r request.Request) (*dns.Msg, error) {
	// get proof from map server using map_client
	e.retrieveProofsFromMapServer([]string{domain})

	var proof []byte
	for i := 0; i < 256; i++ {
		proof = append(proof, byte(i))
	}

	// encode proof into Txt records
	proofStrings := bytesToStrings(proof)

	// add Txt records to DNS response
	setupEdns0Opt(r.Req)
	m := new(dns.Msg)
	m.SetReply(r.Req)

	hdr := dns.RR_Header{Name: r.QName(), Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0}
	m.Answer = append(m.Answer, &dns.TXT{Hdr: hdr, Txt: proofStrings})
	//

	return m, nil
}

func bytesToStrings(in []byte) []string {
	var out []string
	for i := range in {
		if i%255 == 0 {
			out = append(out, "")
		}
		out[len(out)-1] += fmt.Sprintf("\\%d%d%d", in[i]/100, in[i]/10%10, in[i]%10)
	}
	return out
}

// Name implements the Handler interface.
func (e Mapserver) Name() string { return "mapserver" }

// setupEdns0Opt will retrieve the EDNS0 OPT or create it if it does not exist.
func setupEdns0Opt(r *dns.Msg) *dns.OPT {
	o := r.IsEdns0()
	if o == nil {
		r.SetEdns0(4096, false)
		o = r.IsEdns0()
	}
	return o
}
