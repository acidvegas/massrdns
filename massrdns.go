package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var dnsServers []string

func loadDNSServersFromFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		server := scanner.Text()

		// Check if the server contains a port
		if strings.Contains(server, ":") {
			host, port, err := net.SplitHostPort(server)
			if err != nil {
				return fmt.Errorf("invalid IP:port format for %s", server)
			}
			if net.ParseIP(host) == nil {
				return fmt.Errorf("invalid IP address in %s", server)
			}
			if _, err := strconv.Atoi(port); err != nil {
				return fmt.Errorf("invalid port in %s", server)
			}
		} else {
			if net.ParseIP(server) == nil {
				return fmt.Errorf("invalid IP address %s", server)
			}
			server += ":53" // Default to port 53 if not specified
		}

		dnsServers = append(dnsServers, server)
	}
	return scanner.Err()
}

func reverseDNSLookup(ip string, server string) string {
	ctx := context.Background()

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, network, server)
		},
	}

	names, err := resolver.LookupAddr(ctx, ip)
	if err != nil {
		return fmt.Sprintf("%s | %s | Error: %s", time.Now().Format("03:04:05 PM"), server, err)
	}

	if len(names) == 0 {
		return fmt.Sprintf("%s | %s | No PTR records", time.Now().Format("03:04:05 PM"), server)
	}
	return fmt.Sprintf("%s | %s | %s", time.Now().Format("03:04:05 PM"), server, names[0])
}

func worker(cidr *net.IPNet, resultsChan chan string) {
	for ip := make(net.IP, len(cidr.IP)); copy(ip, cidr.IP) != 0; incrementIP(ip) {
		if !cidr.Contains(ip) {
			break
		}
		randomServer := dnsServers[rand.Intn(len(dnsServers))]
		result := reverseDNSLookup(ip.String(), randomServer)
		resultsChan <- result
	}
}

func splitCIDR(cidr string, parts int) ([]*net.IPNet, error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	maskSize, _ := ipNet.Mask.Size()
	newMaskSize := maskSize
	for ; (1 << uint(newMaskSize-maskSize)) < parts; newMaskSize++ {
		if newMaskSize > 32 {
			return nil, fmt.Errorf("too many parts; cannot split further")
		}
	}

	var subnets []*net.IPNet
	for i := 0; i < parts; i++ {
		subnets = append(subnets, &net.IPNet{
			IP:   ip,
			Mask: net.CIDRMask(newMaskSize, 32),
		})
		incrementIPBy(ip, 1<<uint(32-newMaskSize))
	}

	return subnets, nil
}

func main() {
	var cidr string
	var concurrency int
	var dnsFile string

	flag.StringVar(&cidr, "cidr", "", "IP address CIDR to perform reverse DNS lookup")
	flag.IntVar(&concurrency, "concurrency", 10, "Number of concurrent workers for reverse DNS lookup")
	flag.StringVar(&dnsFile, "dnsfile", "", "Path to the file containing DNS servers (one per line)")
	flag.Parse()

	if cidr == "" || dnsFile == "" {
		fmt.Println("Please provide a CIDR using the -cidr flag and a DNS servers file with the -dnsfile flag.")
		os.Exit(1)
	}

	if err := loadDNSServersFromFile(dnsFile); err != nil {
		fmt.Printf("Error reading DNS servers from file %s: %s\n", dnsFile, err)
		os.Exit(1)
	}

	if len(dnsServers) == 0 {
		fmt.Println("No DNS servers found in the provided file.")
		os.Exit(1)
	}

	rand.Seed(time.Now().UnixNano())

	subnets, err := splitCIDR(cidr, concurrency*10) // Create more subnets than workers
	if err != nil {
		fmt.Printf("Error splitting CIDR: %s\n", err)
		os.Exit(1)
	}

	// Create a channel to feed CIDR blocks to workers
	cidrChan := make(chan *net.IPNet, len(subnets))
	for _, subnet := range subnets {
		cidrChan <- subnet
	}
	close(cidrChan) // Close it, so workers can detect when there's no more work

	resultsChan := make(chan string, concurrency*2) // Increased buffer size for results

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for subnet := range cidrChan { // Keep working until there's no more work
				worker(subnet, resultsChan)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for result := range resultsChan {
		fmt.Println(result)
	}
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func incrementIPBy(ip net.IP, count int) {
	for count > 0 {
		incrementIP(ip)
		count--
	}
}
