#!/bin/zsh

n=$1

for (( i = 0; i < $n; i++ ))
do
	../peer/peer --n=$n --base_ip=127.0.0.1 --base_port=8080 --i=$i &
done
