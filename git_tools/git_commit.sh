#!/bin/bash

msg=$1;
[ x"$msg" != x ] && msg+=","
DATE_NOW=`date +%F`

cd ~/thePatriot
git add *
git commit -a -m "Updated at $DATE_NOW, comments: [$msg]"
git push -u origin master
