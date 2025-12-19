// Created by SR @ 2025
// MIT License

package mdns_test

import (
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sergereinov/mdns"
)

const (
	_INSTANCE        = "test"
	_SERVICE         = "_foobar._tcp"
	_PORT            = 12345
	_INFO            = "test info"
	_RESOLVE_TIMEOUT = 50 * time.Millisecond
)

func TestClient_BindToIPv4(t *testing.T) {
	// Start the server on the loopback interface and enable the IP_MULTICAST_LOOP socket option
	loopIfi := findInterfaceByAddr("127.0.0.")
	someIp := net.ParseIP("10.2.3.4")
	server, err := startServer(_INSTANCE, _SERVICE, _PORT, _INFO, &someIp, loopIfi, true)
	if err != nil {
		t.Fatalf("startServer err: %v", err)
	}
	defer func() {
		if err := server.Shutdown(); err != nil {
			t.Fatalf("server.Shutdown err: %v", err)
		}
	}()

	// Try to find _SERVICE via the loopback subnet
	loopIp4 := net.ParseIP("127.0.0.2")
	var remote *mdns.ServiceEntry
	ok := resolve(_SERVICE, nil, &loopIp4, func(entry *mdns.ServiceEntry) {
		remote = entry
	})
	if !ok {
		t.Errorf("resolve failed")
	} else if remote == nil {
		t.Errorf("remote found but got nil")
	} else if remote.Info != _INFO || remote.AddrV4.String() != someIp.String() || remote.Port != _PORT {
		t.Errorf("remote found but got %+v", remote)
	}
}

func findInterfaceByAddr(prefix string) *net.Interface {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, ifi := range ifaces {
		addrs, err := ifi.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if strings.HasPrefix(addr.String(), prefix) {
				return &ifi
			}
		}
	}
	return nil
}

func resolve(service string, ifi *net.Interface, bindToIp4 *net.IP, cb func(*mdns.ServiceEntry)) (found bool) {
	entriesCh := make(chan *mdns.ServiceEntry, 4)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for entry := range entriesCh {
			found = true
			cb(entry)
		}
	}()

	params := mdns.DefaultParams(service)
	params.Interface = ifi
	params.BindToIPv4 = bindToIp4
	params.DisableIPv6 = true
	params.Entries = entriesCh
	params.Timeout = _RESOLVE_TIMEOUT
	mdns.Query(params)

	close(entriesCh)
	wg.Wait()
	return
}

func startServer(instance, service string, port int, info string, ip *net.IP, ifi *net.Interface, ipMulticastLoop bool) (*mdns.Server, error) {
	var ips []net.IP
	if ip != nil {
		ips = []net.IP{*ip}
	}
	zone, err := mdns.NewMDNSService(instance, service, "", "", port, ips, []string{info})
	if err != nil {
		return nil, err
	}
	server, err := mdns.NewServer(&mdns.Config{
		Zone:            zone,
		Iface:           ifi,
		IpMulticastLoop: ipMulticastLoop,
	})
	return server, err
}
