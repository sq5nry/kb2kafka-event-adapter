how to run:

* verify env variables in ```docker_compose.yml``` & ```.env``` files
* run ```docker-compose up```
* summary of variables
```
KAFKA_TOPIC_NAME=test-topic
KAFKA_BOOTSTRAP_SERVER=localhost:9092
KAFKA_GROUP_ID=group-id-1
KAFKA_RESPONSE_TOPIC_NAME=response-test-topic
```

```
LISTENER_ADDRESS=:8082
LISTENER_PATH=/callmeback
KAFKA_BROKER_HEALTHCHECK_URL=http://kafkahealthcheck:8000
KAFKA_CLUSTER_HEALTHCHECK_URL=http://kafkahealthcheck:8000/cluster
```
