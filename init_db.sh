#!/bin/bash
# /usr/bin/mongod --bind_ip_all --replSet dbrs

# DELAY=10
# sleep $DELAY

mongosh <<EOF
config = {
  	"_id" : "dbrs",
  	"members" : [
  		{
  			"_id" : 0,
  			"host" : "mongo1:27017",
            "priority": 2
  		},
  		{
  			"_id" : 1,
  			"host" : "mongo2:27017",
            "priority": 1
  		},
  		{
  			"_id" : 2,
  			"host" : "mongo3:27017",
            "priority": 1
  		}
  	]
  }
rs.initiate(config, { force: true });
EOF

# echo "****** Waiting for ${DELAY} seconds for replicaset configuration to be applied ******"

# sleep $DELAY
