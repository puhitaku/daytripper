Daytripper
==========

*A distributed brute-force 2ch trip calculator*

Comparison
----------

Test environment:
 - DeskMini H310
 - Intel Core i7-8700 (6C12T)
 - 32 GB RAM
 - Debian 10 (buster)

|Name|Hash/s|Ratio|Remarks|
|:---|-----:|----:|:------|
|utripper|369519|1.0|Old trip (DES)|
|daytripper|25735409|69.6|New trip (SHA-1+Base64)|


Server
------

```
$ go build

$ ./daytripper -help
Usage of ./daytripper:
  -nr int
        Number of goroutines (default: runtime.NumCPU() * 2) (default 16)
  -remote string
        Remote daytripper host (optional for distributed calculation)

$ ./daytripper triptofind
Searching for 'triptofind' with 16 goroutines...
Deader is serving at 0.0.0.0:52313
Hashes: 16687736729 (16150194 hash/s) | Elapsed 1068 sec
```


Client
------

```
$ go build

$ ./daytripper -remote the-hostname-of-server.local triptofind
Searching for 'triptofind' with 24 goroutines...
Hashes: 31233492564 (24119686 hash/s) | Elapsed 1285 sec
```

