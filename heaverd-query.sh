#!/bin/bash

if [[ $1 = "send" ]]; then
	serf query receive "$2" > /dev/null
else
	RECEIVED=`cat`
	echo $RECEIVED | nc localhost 1444
fi

