package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/miekg/dns"
)

// dnsServers holds the nameservers used for DNS tests during this session.
// It is populated lazily from dns_conf_path and never written back to disk.
var dnsServers []string

func ensureDNSServersLoaded() {
	if len(dnsServers) > 0 {
		return
	}
	config, err := dns.ClientConfigFromFile(dns_conf_path)
	if err != nil {
		return
	}
	dnsServers = append(dnsServers, config.Servers...)
}

func manageDNSServers() {
	ensureDNSServersLoaded()

	const (
		actionList   = "[ 1 ] - List servernames"
		actionAdd    = "[ 2 ] - Add servername"
		actionRemove = "[ 3 ] - Remove servername"
		actionBack   = "[ 4 ] - Return"
	)
	actions := []string{actionList, actionAdd, actionRemove, actionBack}

	for {
		sel := promptui.Select{
			Label:     "Manage DNS Servernames",
			Items:     actions,
			Templates: templates,
		}

		_, opt, err := sel.Run()
		if err != nil {
			if errors.Is(err, promptui.ErrInterrupt) {
				fmt.Println("\n The operation has been interrupt by user!")
				os.Exit(0)
			}
			fmt.Println("Sorry, an exception has occurred:", err)
			return
		}

		switch opt {
		case actionList:
			if len(dnsServers) == 0 {
				fmt.Println("\n No servernames configured.")
				continue
			}
			fmt.Println("\n Configured servernames:")
			for i, s := range dnsServers {
				fmt.Printf("  %d. %s\n", i+1, s)
			}

		case actionAdd:
			prompt := promptui.Prompt{
				Label: "Servername IP (e.g. 8.8.8.8)",
				Validate: func(input string) error {
					if net.ParseIP(strings.TrimSpace(input)) == nil {
						return fmt.Errorf("invalid IP address")
					}
					return nil
				},
			}
			ip, err := prompt.Run()
			if err != nil {
				continue
			}
			dnsServers = append(dnsServers, strings.TrimSpace(ip))
			fmt.Println("\n Servername added.")

		case actionRemove:
			if len(dnsServers) == 0 {
				fmt.Println("\n No servernames configured.")
				continue
			}
			removeSel := promptui.Select{
				Label:     "Choose servername to remove",
				Items:     dnsServers,
				Templates: templates,
			}
			idx, _, err := removeSel.Run()
			if err != nil {
				continue
			}
			dnsServers = append(dnsServers[:idx], dnsServers[idx+1:]...)
			fmt.Println("\n Servername removed.")

		case actionBack:
			return
		}
	}
}

// resolveViaServer resolves host using a specific DNS server, returning the
// resolved addresses and how long the lookup took.
func resolveViaServer(server, host string) ([]string, time.Duration, error) {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: 3 * time.Second}
			return d.DialContext(ctx, "udp", net.JoinHostPort(server, "53"))
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	start := time.Now()
	addrs, err := resolver.LookupHost(ctx, host)
	elapsed := time.Since(start)
	return addrs, elapsed, err
}

func testConfiguredDNS(host string) {
	ensureDNSServersLoaded()

	if len(dnsServers) == 0 {
		fmt.Println("\n No servernames configured to test.")
		return
	}

	fmt.Printf("\n ---Testing DNS resolution for %q---\n", host)
	for _, server := range dnsServers {
		addrs, elapsed, err := resolveViaServer(server, host)
		if err != nil {
			fmt.Printf("  [%s] FAILED (%s) - %v\n", server, elapsed.Round(time.Millisecond), err)
			continue
		}
		fmt.Printf("  [%s] OK (%s) - %s\n", server, elapsed.Round(time.Millisecond), strings.Join(addrs, ", "))
	}
}
