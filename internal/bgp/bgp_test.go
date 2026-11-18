package bgp

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudapp3/vmlg/internal/config"
)

func TestBirdUnixSocketProtocolParser(t *testing.T) {
	sockPath := filepath.Join(t.TempDir(), "bird.ctl")
	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				defer conn.Close()
				_, _ = fmt.Fprint(conn, "0001 BIRD 2.15.1 ready.\n")
				sc := bufio.NewScanner(conn)
				for sc.Scan() {
					cmd := strings.TrimSpace(sc.Text())
					switch {
					case cmd == "restrict":
						_, _ = fmt.Fprint(conn, "0016 Access restricted\n")
					case strings.HasPrefix(cmd, "show route for "):
						_, _ = fmt.Fprint(conn, "1007-Table master4:\n")
						_, _ = fmt.Fprint(conn, " 1.1.1.0/24           unicast [bgp_cn2 2026-11-01 from 59.43.187.1] * (100) [AS13335i]\n")
						_, _ = fmt.Fprint(conn, "\tvia 59.43.187.1 on eth0\n")
						_, _ = fmt.Fprint(conn, "1012-\tBGP.origin: IGP\n")
						_, _ = fmt.Fprint(conn, " \tBGP.as_path: 4809 13335\n")
						_, _ = fmt.Fprint(conn, " \tBGP.local_pref: 200\n")
						_, _ = fmt.Fprint(conn, " \tBGP.community: (64512,100) (65535,65281)\n")
						_, _ = fmt.Fprint(conn, " \tBGP.large_community: (64512, 1000, 1)\n")
						_, _ = fmt.Fprint(conn, "0000 \n")
					case cmd == "show protocols":
						_, _ = fmt.Fprint(conn, "1002-Name       Proto      Table      State  Since         Info\n")
						_, _ = fmt.Fprint(conn, " bgp_cn2    BGP        ---        up     2026-11-01    Established\n")
						_, _ = fmt.Fprint(conn, "0000 \n")
					}
				}
			}(c)
		}
	}()

	drv := NewDriver(config.BGPConfig{
		Enabled:    true,
		Driver:     "bird",
		RouterASN:  "64512",
		BirdSocket: sockPath,
		Communities: map[string]string{
			"64512:1000:1": "HKIX Route",
		},
	})

	if ok, note := drv.Available(); !ok {
		t.Fatalf("expected bird driver available, got: %s", note)
	}

	res, err := drv.LookupRoute(context.Background(), "1.1.1.1")
	if err != nil {
		t.Fatalf("LookupRoute failed: %v", err)
	}
	if res.Prefix != "1.1.1.0/24" || len(res.Paths) != 1 {
		t.Fatalf("unexpected route result: %+v", res)
	}
	p := res.Paths[0]
	if !p.Best || p.LocalPref != 200 || len(p.ASPath) != 2 || p.ASPath[0].ASN != "AS4809" {
		t.Fatalf("unexpected parsed BIRD path: %+v", p)
	}
	if len(p.Communities) != 3 || p.Communities[2].Label != "HKIX Route" {
		t.Fatalf("unexpected communities: %+v", p.Communities)
	}

	sum, err := drv.Summary(context.Background())
	if err != nil || len(sum.Peers) != 1 || sum.Peers[0].Name != "bgp_cn2" {
		t.Fatalf("unexpected BIRD summary: %+v err=%v", sum, err)
	}

	// Command injection attempt must fail validation
	if _, err := drv.LookupRoute(context.Background(), "1.1.1.1; configure"); err != ErrInvalidPrefix {
		t.Fatalf("expected ErrInvalidPrefix, got %v", err)
	}
}

func TestFRRJSONParser(t *testing.T) {
	sample := []byte(`{
		"prefix": "8.8.8.0/24",
		"paths": [{
			"locPrf": 180,
			"metric": 10,
			"origin": "IGP",
			"bestpath": {"overall": true},
			"aspath": {"string": "9929 15169"},
			"nexthops": [{"ip": "218.105.1.1", "interfaceName": "eth0"}],
			"community": {"list": ["65535:65281"]}
		}]
	}`)

	res, err := parseFRRRouteJSON("8.8.8.8", "AS64512", nil, sample)
	if err != nil {
		t.Fatal(err)
	}
	if res.Prefix != "8.8.8.0/24" || len(res.Paths) != 1 || !res.Paths[0].Best || res.Paths[0].ASPath[0].ASN != "AS9929" {
		t.Fatalf("unexpected FRR parse result: %+v", res)
	}
}
