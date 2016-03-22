# Domainalyzer

This is a simple tool which takes a list of domain names (deduped, please) and looks up

1. Is it even registered
2. Is it resolvable
3. If so, what organization name does the apex's IP resolve to

UI:

    ./domainalyzer -c 5000 -o results.tsv < input.txt

Where -c gives the level of concurrency to sustain. 

Input: STDIN, with entries that look like

    domain<:optional tag>

Output: *appends* to the file given in argv[0], so multiple runs are acceptable.

    domain [tab] status [ tab ] host org

## Implementation notes

    GeoIPASNum.dat
    http://dev.maxmind.com/geoip/legacy/geolite/

    # Requires the geoip c library, a la `brew install geoip`
    go get github.com/abh/geoip

    func (gi *GeoIP) GetOrg(ip string) string {
    func (gi *GeoIP) GetName(ip string) (name string, netmask int) {
    file := "/usr/share/GeoIP/GeoIP.dat"

    gi, err := geoip.Open(file)
    if err != nil {
        fmt.Printf("Could not open GeoIP database\n")
    }

    if gi != nil {
        country, netmask := gi.GetCountry("207.171.7.51")
    }

    func LookupIP(host string) (ips []IP, err error)
    # ip has a stringer

## Go concurrency limiting pattern

	concurrency := 500

	// semaphore channel, limits the amount of concurrency we want
	sem := make(chan bool, concurrency)

	urls := []string{"url1", "url2"}

	for _, url := range urls {
        // this blocks if the channel is full, limiting the rate
		sem <- true
		go func(url) {
            // reading from the semaphore is deferred
            // that means when the function exits, it 'frees the slot'
			defer func() { <-sem }()
			// do work on url here
		}(url)
	}
    // the loop gets exited when the very last bit of work is *submitted*
    // but not when it is *completed*. So this fills it back up to capacity...
    // which can only happen when all the pending, enqueued, work is done!
	for i := 0; i < cap(sem); i++ {
		sem <- true
	}
    // so when you get here, mission accomplished.


