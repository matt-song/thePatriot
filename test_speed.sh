#!/bin/bash

report=/tmp/speed_test_report.txt
today=`date +%F`
mail_list='sxiaobo@vmware.com'

cd /root/thePatriot/src
# go run generate_test_report_multiThread.go -b /usr/sbin -h aio1 -u gpadmin -p 5432 -o ~/temp_csv
go run generate_test_report_multiThread.go -b /usr/sbin -h 127.0.0.1 -u postgres -p 5432 -o ~/temp_csv 
psql thePatriot -U postgres -c "select * from final_report where testdate = (select testdate from final_report where testdate::text ~ '$today' order by 1 desc limit 1) order by speed  desc;" > $report
cat $report | mail -s "Speed test report of [$today]" $mail_list
