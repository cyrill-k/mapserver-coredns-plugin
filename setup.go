package mapserver

import (
	"fmt"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"

	"github.com/caddyserver/caddy"
)

// init registers this plugin.
func init() { plugin.Register("mapserver", setup) }

// setup is the function that gets called when the config parser see the token "mapserver". Setup is responsible
// for parsing any extra options the example plugin may have. The first token this function sees is "example".
func setup(c *caddy.Controller) error {
	c.Next() // Ignore "example" and give us the next token.
	args := c.RemainingArgs()
	var mapID int64
	var mapPK string
	var mapAddress *url.URL
	if len(args) == 3 {
		dat, err := ioutil.ReadFile(args[0])
		if err != nil {
			return plugin.Error("mapserver", c.ArgErr())
		}
		mapIDString := strings.TrimSuffix(string(dat), "\n")
		mapIDInt, err := strconv.Atoi(mapIDString)
		if err != nil {
			return plugin.Error("mapserver", c.ArgErr())
		}
		mapID = int64(mapIDInt)
		mapPK = args[1]
		mapAddress, err = url.Parse(args[2])
		if err != nil {
			return plugin.Error("mapserver", c.ArgErr())
		}
	} else {
		return plugin.Error("mapserver", c.ArgErr())
	}

	u, err := url.Parse(c.Key)
	if err != nil {
		fmt.Printf("Error parsing '%s': %s", c.Key, u.Hostname())
	}
	fmt.Printf("setup mapserver at %s", strings.TrimSuffix(u.Hostname(), "."))

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return Mapserver{Next: next, MapserverDomain: strings.TrimSuffix(u.Hostname(), "."), MapID: mapID, MapPK: mapPK, MapAddress: *mapAddress}
	})

	// All OK, return a nil error.
	return nil
}
