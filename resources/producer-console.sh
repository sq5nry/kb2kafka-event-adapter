#!/bin/bash

#TODO use KillBill for that, leaving only triggering to K.B.
docker exec -it kafka wget -O- --post-data='{"some data to post..."}' --header='Content-Type:application/json' http://localhost:9000/json