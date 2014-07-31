#!/bin/bash
# TODO переделать

SERF=./serf

if [[ $1 = "notify" ]]; then
	$SERF query host-notify "$2"  > /dev/null
else
	RECEIVED=`cat`
	echo $RECEIVED | nc localhost 1444
fi

