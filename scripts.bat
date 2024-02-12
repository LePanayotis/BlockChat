kafka-topics.sh --create --bootstrap-server localhost:9092 --topic welcome --partitions 1 --replication-factor 1
kafka-topics.sh --create --bootstrap-server localhost:9092 --topic post-transaction --partitions 1 --replication-factor 1
kafka-topics.sh --create --bootstrap-server localhost:9092 --topic post-block --partitions 1 --replication-factor 1
kafka-topics.sh --create --bootstrap-server localhost:9092 --topic enter --partitions 1 --replication-factor 1

kafka-configs.sh \
  --alter \
  --entity-type topics \
  --entity-name welcome \
  --add-config retention.ms=20000


go build -o blockchat.exe .

go run . start  `
  --node-id 0 `
  --nodes 3 `
  --blockchain-path blockchain0.json `
  --database-path db0.json `
  --protocol tcp `
  --socket localhost:1500 `
  --capacity 5 `
  --broker-url localhost:9094

go run . start  `
  --node-id 1 `
  --nodes 3 `
  --blockchain-path blockchain1.json `
  --database-path db1.json `
  --protocol tcp `
  --socket localhost:1501 `
  --capacity 5 `
  --broker-url localhost:9094

go run . start  `
  --node-id 2 `
  --nodes 3 `
  --blockchain-path blockchain2.json `
  --database-path db2.json `
  --protocol tcp `
  --socket localhost:1502 `
  --capacity 5 `
  --broker-url localhost:9094