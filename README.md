# rdns-fs

This binary parses a reverse DNS json file from [Rapid 7](https://opendata.rapid7.com/sonar.rdns_v2/) and splits it onto the local file system based on the first octect of the IP. In addition, it will include IPs that match your CIDR filter. This is useful for specifying the IP range of a cloud provider for example. This partitioning allows for faster searches of records by IP.

The output files will look like:

    rdns.<firstOctet>.0.0.0.json

with each timestamp being a copy  of the PTR record at that point in time e.g.

    {"timestamp":"1609324571","name":"1.113.142.231","value":"em1-113-142-231.pool.e-mobile.ne.jp","type":"ptr"}

To execute:
    
    ./rdns-fs -file <path-to-your>-rdns.json -cdir <path-to-your-cidrs>.txt -workers <number of concurrent workers, default 50>

I am using RDAP to help discover the AWS CIDR range e.g.
    
    curl -s https://rdap.arin.net/registry/entity/AT-88-Z | jq -cr '.networks[].cidr0_cidrs[] | .v4prefix + "/" + (.length|tostring)' > cidrs.txt
    curl -s https://rdap.arin.net/registry/entity/AMAZO-4 | jq -cr '.networks[].cidr0_cidrs[] | .v4prefix + "/" + (.length|tostring)' >> cidrs.txt

From there on, it is up to your imagination as to what you want to use this information for...
