#!/bin/sh
echo "*** docker-entorypoint.sh Start... ***"
cname=`cat ./cname`
rm /usr/bin/$cname
go build -o /usr/bin/$cname
exec /usr/bin/$cname
