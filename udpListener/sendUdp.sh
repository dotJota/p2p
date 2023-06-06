#!/bin/bash

nMessages=$1
destIp=$2
destPort=$3

echo "Sending {$nMessages} messages to {$destIp}:{$destPort}"

for (( i = 0; i < $nMessages; i++ ))
do
	echo "Datagram $i" > /dev/udp/$destIp/$destPort
done

echo "Finished!"
