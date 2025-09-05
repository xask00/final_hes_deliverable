pulsar <-> Orchestrator <-> DLMS Processor (stateless?)
               |
            Database

need a cache in process for fast processing

1 row =
device id, device ip, port, 3 keys, 



## DLMS Processor GRPC
- get data (one or more devices)
- set attr
- exec
- FOTA

1000,000 devices
  - 30 seconds/device
  - 
10 mins - 

1000,000 d/10 mins = 100,000 devices / min 
100,000 devices/min via 4 DLMS processors = 25,000 devices/min
(i.e) 25,000 threads / go routines / virtual threads per DLMS processor.

### DLMS Processor API
get (OBIS, devices[], threads=X, rampup=Y) stream ->


25,000 devices - 
500 bytes (1 KB worst case)
.5 KB
12,000 KB - 12 MB

(stream) <-|-|->(stream)


for 25,000 device lets say output is 10 KB
25,000 x 10 KB = 250 MB

### Meter device data in db

1 Million x 1 KB = 1 GB


