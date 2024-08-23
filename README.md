# Project Name: ThePatriot
A tool that test the network performance from your local ISP to VPS provider

## Supported platform:
- OS supported: MacOS
- OS has not tested yet: Linux, Windows
- Supported VPS provider: Vultr, Linode

## Requirement:
- Tools must be installed: mtr, curl..
- A postgres DB must be aviliable for storing the data (will add --nosql mode in future)
- below module must be installed
```
go mod init speedtest
go get github.com/lib/pq
go get github.com/pborman/getopt/v2
```

## How to build from source:
```
# some command here [TBD]
```

## How it works: 
1. Get the url for speed test from VPS vendor
2. get the path of the route and test the latency,package drop rate from local to each node.
3. download the files from each site and test the speed
4. store all info into DB

## How to use:
1. Connect your PC/Laptop with the network you want to check.
2. run the tool. use -H to check the help message
```
$ go run generate_raw_result_csv.go -H
Usage: generate_raw_result_csv [-DH] [-b value] [-d value] [-h value] [-o value] [-P value] [-p value] [-u value] [-v value] [parameters ...]
 -b value  Folder which holds the mtr binary, default:
           [/usr/local/sbin]
 -D        Display debug message
 -d value  Database to connect, default: [thePatriot]
 -H        Help
 -h value  Database to connect, default: [aio1]
 -o value  The location of the csv file, default: [.]
 -P value  the port of DB, default: [5432]
 -p value  the password of DB user, default: [abc123]
 -u value  the user of DB, default: [gpadmin]
 -v value  the vendor of the VPS, default: [vultr]
 ```
 
## Example output 
```
HostName                           speed(KB/s)     avgLossRate     maxLossRate      avgLatency      maxLatency
speedtest.mumbai1.linode.com               730            8.55           28.33          636.28         1300.90
speedtest.london.linode.com                137           13.33           50.00          402.36          441.55
speedtest.tokyo2.linode.com                116           11.92           85.00          121.68          243.73
speedtest.dallas.linode.com                 69           12.75           36.67          326.22          369.53
speedtest.toronto1.linode.com               58           10.00           28.33          365.44          414.16
speedtest.singapore.linode.com              56           10.72           81.67          190.66          211.93
speedtest.newark.linode.com                 50           11.05           30.00          375.91          408.86
speedtest.syd1.linode.com                   38            9.17           76.67          239.61          368.38
speedtest.fremont.linode.com                32           11.19           30.00          346.69          406.11
speedtest.frankfurt.linode.com               5           11.14           31.67          419.76          448.96
speedtest.atlanta.linode.com                 5           14.41           91.67          347.22          474.48
```

## To query the historical data from DB, check the table final_report
```
# select * from final_report where vendor = 'linode' and hostname = 'speedtest.tokyo2.linode.com' order by 1;
      testdate       | vendor |          hostname           | speed | avg_lossrate | max_lossrate | avg_latency | max_latency
---------------------+--------+-----------------------------+-------+--------------+--------------+-------------+-------------
 2020-04-13 12:18:00 | linode | speedtest.tokyo2.linode.com |   181 |         0.00 |         0.00 |      235.79 |      235.79
 2020-04-13 12:30:00 | linode | speedtest.tokyo2.linode.com |   140 |         0.00 |         0.00 |      121.50 |      121.50
 2020-04-13 12:33:00 | linode | speedtest.tokyo2.linode.com |   123 |         0.00 |         0.00 |      120.48 |      120.48
(3 rows)
```
