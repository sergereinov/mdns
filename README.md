Fork
====
This is a fork of the [mdns](https://github.com/hashicorp/mdns) library. The original readme is below the double separation line.

### Server side
An option to enable multicast loopback traffic has been added to the server side. This is useful if you're looking for mDNS services not only on your local network but also on your local PC. It's also useful for testing.

```go
server, err := mdns.NewServer(&mdns.Config{
    // ...
    IpMulticastLoop: true,
    // ...
})
```

### Client side
Added the ability to explicitly specify local IPv4 and IPv6 addresses for binding a unicast socket used to send multicast mDNS queries. AFAIK, this is a better way to select an IP subnet for UDP multicast communication than simply binding the socket to Zero-IP. This is especially true for multihomed systems with multiple active network interfaces and multiple active IP addresses on some of them.

```go
params := mdns.DefaultParams(service)
// ...
params.BindToIPv4 = net.ParseIP("192.168.155.34") // binds to subnet via known local IP within the subnet
// ...
mdns.Query(params)
```

### Tests
The binding test can be found in `client_test.go`. Examples of server and client initialization can also be found there.

SR.

---
---

# mdns

Simple mDNS client/server library in Golang. mDNS or Multicast DNS can be
used to discover services on the local network without the use of an authoritative
DNS server. This enables peer-to-peer discovery. It is important to note that many
networks restrict the use of multicasting, which prevents mDNS from functioning.
Notably, multicast cannot be used in any sort of cloud, or shared infrastructure
environment. However it works well in most office, home, or private infrastructure
environments.

Using the library is very simple, here is an example of publishing a service entry:
```go
// Setup our service export
host, _ := os.Hostname()
info := []string{"My awesome service"}
service, _ := mdns.NewMDNSService(host, "_foobar._tcp", "", "", 8000, nil, info)

// Create the mDNS server, defer shutdown
server, _ := mdns.NewServer(&mdns.Config{Zone: service})
defer server.Shutdown()
```

Doing a lookup for service providers is also very simple:
```go
// Make a channel for results and start listening
entriesCh := make(chan *mdns.ServiceEntry, 4)
go func() {
    for entry := range entriesCh {
        fmt.Printf("Got new entry: %v\n", entry)
    }
}()

// Start the lookup
mdns.Lookup("_foobar._tcp", entriesCh)
close(entriesCh)
```
