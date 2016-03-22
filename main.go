package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/abh/geoip"
	"github.com/bogdanovich/dns_resolver"
)

var resolver *dns_resolver.DnsResolver
var gi *geoip.GeoIP

func init() {
	var err error
	resolver, err = dns_resolver.NewFromResolvConf("/etc/resolv.conf")
	if err != nil {
		log.Fatal("Couldn't build a resolver", err)
	}
	resolver.RetryTimes = 5

	file := "GeoIPASNum.dat"
	gi, err = geoip.Open(file)
	if err != nil {
		log.Fatal("Couldn't open the geoip database", err)
	}
}

// DomainInfo tracks the name, organization, IP and lookup state
type DomainInfo struct {
	domain string
	ok     bool
	ip     string
	org    string
}

func (d DomainInfo) String() string {
	return fmt.Sprintf("domain: %s ip: %s org: %s ok? %t", d.domain, d.ip, d.org, d.ok)
}

// lookupDomain takes a domain name string, looks it up, and writes
// the resulting object to the output channel
func lookupDomain(domain string, sem <-chan bool, out chan<- DomainInfo) {
	// grab the concurrency limiter value
	defer func() { <-sem }()
	//log.Println("Looking up domain ", domain)
	rv := DomainInfo{domain: domain}

	ips, err := resolver.LookupHost(domain)
	if err != nil || len(ips) == 0 {
		out <- rv
		return
	}

	rv.ip = ips[0].String()
	rv.org = gi.GetOrg(rv.ip)
	rv.ok = true
	//log.Println("Looked up results:", rv)
	out <- rv
	return
}

// outputWriter pulls results from the channel and writes them
// to the file in CSV format.
func outputWriter(filename string, results <-chan DomainInfo, done chan<- bool) {
	defer func() { done <- true }()
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString("domain,ip,org,ok\n")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Writer: Waiting on results")
	for d := range results {
		// log.Println("Writer: got result", d)
		_, err = f.WriteString(fmt.Sprintf("%s,%s,%s,%t\n", d.domain, d.ip, d.org, d.ok))
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Println("Writer: exiting")
}

func main() {

	concurrency := flag.Int("c", 500, "concurrency")
	outfile := flag.String("o", "results.csv", "output .csv file")
	flag.Parse()

	sem := make(chan bool, *concurrency)
	outchan := make(chan DomainInfo)
	writerdone := make(chan bool)

	go outputWriter(*outfile, outchan, writerdone)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// blocks if the channel is full, limiting the rate
		sem <- true
		go lookupDomain(strings.TrimSpace(scanner.Text()), sem, outchan)
	}

	// the loop gets exited when the very last bit of work is *submitted*
	// but not when it is *completed*. So this fills it back up to capacity...
	// which can only happen when all the pending, enqueued, work is done!
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
	log.Println("Completed all concurrent work. Waiting on writer to finish")
	close(outchan)
	<-writerdone
}
