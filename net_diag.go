package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/go-ping/ping"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	traceMaxHops = 30
	traceTimeout = 2 * time.Second
	pingCount    = 4
)

// pingHost measures latency and packet loss against host.
// It requires raw ICMP socket access (root, or the binary built with
// 'sudo setcap cap_net_raw+ep <binary>').
func pingHost(host string) (*ping.Statistics, error) {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		return nil, fmt.Errorf("could not resolve %q: %w", host, err)
	}

	pinger.Count = pingCount
	pinger.Timeout = time.Duration(pingCount) * 2 * time.Second
	pinger.SetPrivileged(true)

	if err := pinger.Run(); err != nil {
		return nil, fmt.Errorf("ping failed, try running as root or 'sudo setcap cap_net_raw+ep' on the binary: %w", err)
	}

	return pinger.Statistics(), nil
}

// traceroute discovers the route to host by sending ICMP echo requests with
// increasing TTL, one hop per line, in the traditional traceroute format.
func traceroute(host string) ([]string, error) {
	dst, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		return nil, fmt.Errorf("could not resolve %q: %w", host, err)
	}

	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, fmt.Errorf("could not open raw socket, try running as root or 'sudo setcap cap_net_raw+ep' on the binary: %w", err)
	}
	defer conn.Close()

	pconn := conn.IPv4PacketConn()
	id := os.Getpid() & 0xffff

	hops := make([]string, 0, traceMaxHops)

	for ttl := 1; ttl <= traceMaxHops; ttl++ {
		if err := pconn.SetTTL(ttl); err != nil {
			return hops, fmt.Errorf("set ttl: %w", err)
		}

		msg := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   id,
				Seq:  ttl,
				Data: []byte("PROJETO-1-traceroute"),
			},
		}
		wb, err := msg.Marshal(nil)
		if err != nil {
			return hops, fmt.Errorf("marshal icmp message: %w", err)
		}

		start := time.Now()
		if _, err := conn.WriteTo(wb, dst); err != nil {
			return hops, fmt.Errorf("write icmp message: %w", err)
		}

		conn.SetReadDeadline(time.Now().Add(traceTimeout))
		reply := make([]byte, 1500)
		n, peer, err := conn.ReadFrom(reply)
		elapsed := time.Since(start)

		if err != nil {
			hops = append(hops, fmt.Sprintf("%2d  *  request timed out", ttl))
			continue
		}

		rm, err := icmp.ParseMessage(ipv4.ICMPTypeEcho.Protocol(), reply[:n])
		if err != nil {
			hops = append(hops, fmt.Sprintf("%2d  %s  could not parse reply", ttl, peer))
			continue
		}

		hops = append(hops, fmt.Sprintf("%2d  %s  %s", ttl, peer, elapsed.Round(time.Millisecond)))

		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			return hops, nil
		case ipv4.ICMPTypeTimeExceeded, ipv4.ICMPTypeDestinationUnreachable:
			continue
		}
	}

	return hops, fmt.Errorf("destination not reached within %d hops", traceMaxHops)
}

// runConnectionTest exercises DNS resolution, latency/packet loss (ping) and
// route discovery (traceroute) against host, printing a short report.
func runConnectionTest(host string) {
	fmt.Printf("\n=== Connection Test: %s ===\n", host)

	fmt.Println("\n--- DNS Resolution ---")
	testConfiguredDNS(host)

	fmt.Println("\n--- Latency & Packet Loss (ping) ---")
	stats, err := pingHost(host)
	if err != nil {
		fmt.Println(" ", err)
	} else {
		fmt.Printf("  %d packets transmitted, %d received, %.1f%% packet loss\n",
			stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
		fmt.Printf("  rtt min/avg/max = %v/%v/%v\n", stats.MinRtt, stats.AvgRtt, stats.MaxRtt)
	}

	fmt.Println("\n--- Route (traceroute) ---")
	hops, err := traceroute(host)
	for _, h := range hops {
		fmt.Println(" ", h)
	}
	if err != nil {
		fmt.Println(" ", err)
	}
}
