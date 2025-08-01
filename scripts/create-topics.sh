#!/bin/bash
set -e

echo "🚀 Starting Kafka topic creation..."

BOOTSTRAP_SERVER=localhost:9092

# Wait for Kafka to be ready
wait_for_kafka() {
  echo "⏳ Waiting for Kafka to be ready..."
  local max_attempts=30
  local attempt=1

  while [ $attempt -le $max_attempts ]; do
    if kafka-topics.sh --bootstrap-server $BOOTSTRAP_SERVER --list > /dev/null 2>&1; then
      echo "✅ Kafka is ready!"
      return 0
    fi

    echo "Attempt $attempt/$max_attempts: Kafka not ready yet, waiting..."
    sleep 5
    attempt=$((attempt + 1))
  done

  echo "❌ Kafka failed to start within expected time"
  return 1
}

create_topic() {
  local topic_name="$1"
  local partitions="${2:-2}"
  local replication_factor="${3:-1}"
  local bootstrap_server="${4:-kafka:9092}"

  if [ -z "$topic_name" ]; then
    echo "❌ Topic name is required"
    return 1
  fi

  echo "📝 Creating topic: $topic_name"

  # Clear topic
  kafka-topics.sh --delete --topic "$topic_name" --bootstrap-server "$bootstrap_server" || true

  kafka-topics.sh --create \
    --topic "$topic_name" \
    --partitions "$partitions" \
    --replication-factor "$replication_factor" \
    --if-not-exists \
    --bootstrap-server "$bootstrap_server"


  echo "🎉 Topic '$topic_name' created successfully (partitions=$partitions, rf=$replication_factor)"
}

# Wait for Kafka to be ready
wait_for_kafka

# Create topics
echo "📋 Creating Kafka topics..."

# Example topics - customize these for your application
TOPICS=(
  "events:2:1"
)

for topic_config in "${TOPICS[@]}"; do
  IFS=':' read -r topic_name partitions replication_factor <<< "$topic_config"
  create_topic "$topic_name" "$partitions" "$replication_factor" "$BOOTSTRAP_SERVER"
done

echo "✅ All topics created successfully!"
echo "📊 Listing all topics:"
kafka-topics.sh --bootstrap-server $BOOTSTRAP_SERVER --list
