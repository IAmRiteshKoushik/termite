#!/bin/bash

set -e

# Check if plumber and jq are installed
if ! command -v plumber &> /dev/null || ! command -v jq &> /dev/null;
    then
    echo "Error: 'plumber' and 'jq' are required. Please install them before running this script." >&2
    exit 1
fi

# Path to the JSON file
JSON_FILE="test/woc-sample.json"

# Check if the JSON file exists
if [ ! -f "$JSON_FILE" ]; then
    echo "Error: JSON file not found at $JSON_FILE" >&2
    exit 1
fi

# RabbitMQ settings
EXCHANGE="tentacloid.woc"
ROUTING_KEY="woc.new"

# Read the JSON file and publish each object to RabbitMQ
jq -c '.[]' "$JSON_FILE" | while IFS= read -r payload;
    do
    echo "Publishing payload: $payload"
    plumber write rabbitmq --address="amqp://guest:guest@localhost:5672/"
        --exchange="$EXCHANGE"
        --routing-key="$ROUTING_KEY"
        --input-data="$payload"
done

echo "All payloads from $JSON_FILE published successfully to exchange '$EXCHANGE' with routing key '$ROUTING_KEY'."

