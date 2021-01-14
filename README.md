# rdns-fs

This binary parses a reverse DNS json file from [Rapid 7](https://opendata.rapid7.com/sonar.rdns_v2/) and splits it onto the local file system based on IP. This partitioning allows for faster searches of records by IP.

The output directory will look like:
```
rdns
└── 1
    └── 113
        └── 142
            ├── 218
            │   └── 1609289452
            ├── 231
            │   └── 1609288228
```

with each timestamp being a copy  of the PTR record at that point in time e.g.


    {"timestamp":"1609324571","name":"1.113.142.231","value":"em1-113-142-231.pool.e-mobile.ne.jp","type":"ptr"}

To execute:
    
    ./rdns-fs -file <path-to-your>-rdns.json -workers <number of concurrent workers, default 50>
