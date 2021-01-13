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
    
    ./rdns-fs -file <path-to-your>-rdns.json -workers <number of concurrent workers, default 500>

This will go as fast as your disk subsystem allows. I'm running it on a t3.large with 512GB SSD (GP2) EBS volume which should take 5 days to complete :D

    $ ./rdns-fs -file 2020-12-30-1609286699-rdns.json -workers 500
    220398620 / 138778972343 [>------------------------------------------------------------]   0.16% 5d6h37m37s

    $ iostat -m
    Linux 4.14.209-160.339.amzn2.x86_64              01/13/2021      _x86_64_        (2 CPU)

    avg-cpu:  %user   %nice %system %iowait  %steal   %idle
    4.62    0.00   14.88   62.39    9.65    8.46

    Device:            tps    MB_read/s    MB_wrtn/s    MB_read    MB_wrtn
    nvme1n1        2778.97         2.85        13.30       2666      12424
    nvme0n1           7.43         0.22         0.01        205         13