# Project Name: ThePatriot
A tool that test the network performance from your local ISP to VPS provider

## Supported platform:
- OS supported: MacOS
- OS has not tested yet: Linux, Windows
- Supported VPS provider: Vultr, Linode

## Requirement:
- Tools must be installed: mtr, curl..
- A postgres DB must be aviliable for storing the data (will add --nosql mode in future)

## How to build from source:
```# some command here [TBD]```

## How it works: 
1. Get the url for speed test from VPS vendor
2. get the path of the route and test the latency,package drop rate from local to each node.
3. download the files from each site and test the speed
4. store all info into DB

## example output 
```HostName                           speed(KB/s)     avgLossRate     maxLossRate      avgLatency      maxLatency
speedtest.dallas.linode.com                 29            0.00            0.00          294.41          294.41
speedtest.fremont.linode.com                50            0.00            0.00          324.07          324.07
speedtest.toronto1.linode.com               75            0.00            0.00          361.00          361.00
speedtest.tokyo2.linode.com                 96            0.00            0.00          129.43          129.43
speedtest.syd1.linode.com                  108            0.00            0.00          250.65          250.65
speedtest.mumbai1.linode.com              1006            0.00            0.00          546.20          546.20
speedtest.frankfurt.linode.com            2297            0.00            0.00          420.23          420.23
speedtest.atlanta.linode.com                 4            0.00            0.00          319.00          319.00
speedtest.singapore.linode.com             172            0.00            0.00          190.31          190.31
speedtest.london.linode.com                234            0.00            0.00          403.07          403.07
speedtest.newark.linode.com                 38            0.00            0.00          384.97          384.97

## to query the histrionic data 
```# select * from final_report where vendor = 'linode' and hostname = 'speedtest.tokyo2.linode.com' order by 1;
      testdate       | vendor |          hostname           | speed | avg_lossrate | max_lossrate | avg_latency | max_latency
---------------------+--------+-----------------------------+-------+--------------+--------------+-------------+-------------
 2020-04-13 12:18:00 | linode | speedtest.tokyo2.linode.com |   181 |         0.00 |         0.00 |      235.79 |      235.79
 2020-04-13 12:30:00 | linode | speedtest.tokyo2.linode.com |   140 |         0.00 |         0.00 |      121.50 |      121.50
 2020-04-13 12:33:00 | linode | speedtest.tokyo2.linode.com |   123 |         0.00 |         0.00 |      120.48 |      120.48
(3 rows)

