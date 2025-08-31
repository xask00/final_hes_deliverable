# Starting pulsar
`docker compose up -d`


# Running the producer
`go run cmd/producer/main.go`

# Runnin the consumer
`go run cmd/consumer/main.go`




# Running pulsar manually
```
docker run -it \
-e PULSAR_PREFIX_xxx=yyy \
-p 6650:6650  \
-p 8080:8080 \
--mount source=pulsardata,target=/pulsar/data \
--mount source=pulsarconf,target=/pulsar/conf \
apachepulsar/pulsar:4.0.6 sh \
-c "bin/apply-config-from-env.py \
conf/standalone.conf && \
bin/pulsar standalone"
```