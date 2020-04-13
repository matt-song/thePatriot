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
```
# some command here [TBD]
```

## How it works: 
1. Get the url for speed test from VPS vendor
2. get the path of the route and test the latency,package drop rate from local to each node.
3. download the files from each site and test the speed
4. store all info into DB

## example output 
```
HostName                           speed(KB/s)     avgLossRate     maxLossRate      avgLatency      maxLatency
speedtest.london.linode.com               2976            0.00            0.00          410.47          410.47
speedtest.mumbai1.linode.com              1821            0.00            0.00          570.49          570.49
speedtest.tokyo2.linode.com                233            0.00            0.00          120.55          120.55
speedtest.syd1.linode.com                  154            0.00            0.00          242.73          242.73
speedtest.singapore.linode.com             133            0.00            0.00          190.38          190.38
speedtest.newark.linode.com                 55            0.00            0.00          365.82          365.82
speedtest.toronto1.linode.com               53            0.00            0.00          356.20          356.20
speedtest.dallas.linode.com                 51            0.00            0.00          303.25          303.25
speedtest.fremont.linode.com                46            0.00            0.00          328.64          328.64
speedtest.atlanta.linode.com                 6            0.00            0.00          339.58          339.58
speedtest.frankfurt.linode.com               5            0.00            0.00          419.28          419.28```
```

## to query the historical data 
```
# select * from final_report where vendor = 'linode' and hostname = 'speedtest.tokyo2.linode.com' order by 1;
      testdate       | vendor |          hostname           | speed | avg_lossrate | max_lossrate | avg_latency | max_latency
---------------------+--------+-----------------------------+-------+--------------+--------------+-------------+-------------
 2020-04-13 12:18:00 | linode | speedtest.tokyo2.linode.com |   181 |         0.00 |         0.00 |      235.79 |      235.79
 2020-04-13 12:30:00 | linode | speedtest.tokyo2.linode.com |   140 |         0.00 |         0.00 |      121.50 |      121.50
 2020-04-13 12:33:00 | linode | speedtest.tokyo2.linode.com |   123 |         0.00 |         0.00 |      120.48 |      120.48
(3 rows)
```
