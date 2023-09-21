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
var failureCounts = make(map[string]int)
var showErrors bool

func loadDNSServersFromFile(filePath string) ([]string, error) {
	var servers []string

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		server := scanner.Text()

		if strings.Contains(server, ":") {
			host, port, err := net.SplitHostPort(server)
			if err != nil {
				return nil, fmt.Errorf("invalid IP:port format for %s", server)
			}
			if net.ParseIP(host) == nil {
				return nil, fmt.Errorf("invalid IP address in %s", server)
			}
			if _, err := strconv.Atoi(port); err != nil {
				return nil, fmt.Errorf("invalid port in %s", server)
			}
		} else {
			if net.ParseIP(server) == nil {
				return nil, fmt.Errorf("invalid IP address %s", server)
			}
			server += ":53"
		}

		servers = append(servers, server)
	}
	return servers, scanner.Err()
}

func reverseDNSLookup(ip string, server string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", server)
		},
	}

	names, err := resolver.LookupAddr(ctx, ip)
	if err != nil {
		if isNetworkError(err) {
			return "", err
		}
		return "", err
	}

	if len(names) == 0 {
		return fmt.Sprintf("%s | %-18s | No PTR records", time.Now().Format("03:04:05 PM"), server), nil
	}
	return fmt.Sprintf("%s | %-18s | %s", time.Now().Format("03:04:05 PM"), server, names[0]), nil
}

func isNetworkError(err error) bool {
	errorString := err.Error()
	return strings.Contains(errorString, "timeout") || strings.Contains(errorString, "connection refused")
}

func pickRandomServer(servers []string, triedServers map[string]bool) string {
	for _, i := range rand.Perm(len(servers)) {
		if !triedServers[servers[i]] {
			return servers[i]
		}
	}
	return ""
}

func removeFromList(servers []string, server string) []string {
	var newList []string
	for _, s := range servers {
		if s != server {
			newList = append(newList, s)
		}
	}
	return newList
}

func splitCIDR(cidr string, parts int) ([]*net.IPNet, error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	startIP := make(net.IP, len(ip))
	copy(startIP, ip)

	maskSize, _ := ipNet.Mask.Size()

	maxParts := 1 << uint(32-maskSize)
	if parts > maxParts {
		parts = maxParts
	}

	newMaskSize := maskSize
	for ; (1 << uint(newMaskSize-maskSize)) < parts; newMaskSize++ {
		if newMaskSize > 32 {
			return nil, fmt.Errorf("too many parts; cannot split further")
		}
	}

	var subnets []*net.IPNet
	for i := 0; i < parts; i++ {
		subnets = append(subnets, &net.IPNet{
			IP:   make(net.IP, len(startIP)),
			Mask: net.CIDRMask(newMaskSize, 32),
		})
		copy(subnets[i].IP, startIP)
		incrementIPBy(startIP, 1<<uint(32-newMaskSize))
	}

	return subnets, nil
}

func worker(cidr *net.IPNet, resultsChan chan string) {
	for ip := make(net.IP, len(cidr.IP)); copy(ip, cidr.IP) != 0; incrementIP(ip) {
		if !cidr.Contains(ip) {
			break
		}

		triedServers := make(map[string]bool)
		retries := 10
		success := false

		for retries > 0 {
			randomServer := pickRandomServer(dnsServers, triedServers)
			if randomServer == "" {
				break
			}

			result, err := reverseDNSLookup(ip.String(), randomServer)

			if err != nil {
				if showErrors {
					resultsChan <- fmt.Sprintf("%s | %-18s | Error: %s", time.Now().Format("03:04:05 PM"), randomServer, err)
				}

				if isNetworkError(err) {
					failureCounts[randomServer]++
					if failureCounts[randomServer] > 10 {
						dnsServers = removeFromList(dnsServers, randomServer)
						delete(failureCounts, randomServer)
					}
				}

				triedServers[randomServer] = true
				retries--
				continue
			} else {
				resultsChan <- result
				success = true
				break
			}
		}

		if !success && showErrors {
			resultsChan <- fmt.Sprintf("%s | %-18s | Max retries reached", time.Now().Format("03:04:05 PM"), ip)
		}
	}
}

func main() {
	var cidr string
	var concurrency int
	var dnsFile string

	flag.StringVar(&cidr, "cidr", "", "IP address CIDR to perform reverse DNS lookup")
	flag.IntVar(&concurrency, "concurrency", 10, "Number of concurrent workers for reverse DNS lookup")
	flag.StringVar(&dnsFile, "dnsfile", "", "Path to the file containing DNS servers (one per line)")
	flag.BoolVar(&showErrors, "errors", false, "Display errors in the output") // New flag
	flag.Parse()

	if cidr == "" || dnsFile == "" {
		fmt.Println("Please provide a CIDR using the -cidr flag and a DNS servers file with the -dnsfile flag.")
		os.Exit(1)
	}

	var err error
	dnsServers, err = loadDNSServersFromFile(dnsFile)
	if err != nil {
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

	if len(subnets) < concurrency {
		concurrency = len(subnets) // Limit concurrency to number of subnets
	}

	cidrChan := make(chan *net.IPNet, len(subnets))
	for _, subnet := range subnets {
		cidrChan <- subnet
	}
	close(cidrChan)

	resultsChan := make(chan string, concurrency*2)

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for subnet := range cidrChan {
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
