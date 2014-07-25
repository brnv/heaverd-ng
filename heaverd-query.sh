#!/bin/bash

SERF=$HOME/go/bin/serf

if [[ $1 = "hostinfo" ]]; then
	$SERF query hostinfo "$2" > /dev/null
else
	RECEIVED=`cat`
	echo $RECEIVED | nc localhost 1444
fi

