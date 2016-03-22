# Domainalyzer

This is a simple tool which takes a list of domain names (deduped, please) and looks up

1. Is it even registered
2. Is it resolvable
3. If so, what organization name does the apex's IP resolve to

UI:

    ./domainalyzer -c 5000 -o results.tsv < input.txt

Where -c gives the level of concurrency to sustain. 

*Input*: STDIN, with entries that look like

    domain

*Output*: _overwrites_ the file given in argv[0], setting the .csv header with names

## Requirements

You'll need a copy of `GeoIPASNum.dat` in your working directory, available from [MaxMind](http://dev.maxmind.com/geoip/legacy/geolite/).

To build, you also need the `geoip` C library. On OSX, `brew install geoip`.
