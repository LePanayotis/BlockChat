kafka-topics.sh --create --bootstrap-server localhost:9092 --topic welcome --partitions 1 --replication-factor 1 
kafka-topics.sh --create --bootstrap-server localhost:9092 --topic post-transaction --partitions 1 --replication-factor 1
kafka-topics.sh --create --bootstrap-server localhost:9092 --topic post-block --partitions 1 --replication-factor 1
kafka-topics.sh --create --bootstrap-server localhost:9092 --topic enter --partitions 1 --replication-factor 1

kafka-topics.sh --bootstrap-server localhost:9094 --list

kafka-configs.sh \
  --alter \
  --entity-type topics \
  --entity-name welcome \
  --add-config retention.ms=20000
