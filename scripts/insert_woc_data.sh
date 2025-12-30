#!/bin/bash

set -e

# Check if plumber and jq are installed
if ! command -v plumber &> /dev/null || ! command -v jq &> /dev/null;
    then
    echo "Error: 'plumber' and 'jq' are required. Please install them before running this script." >&2
    exit 1
fi

# Path to the JSON file
JSON_FILE="scripts/woc-sample.json"

# Check if the JSON file exists
if [ ! -f "$JSON_FILE" ]; then
    echo "Error: JSON file not found at $JSON_FILE" >&2
    exit 1
fi

# RabbitMQ settings
EXCHANGE=""
ROUTING_KEY="woc-registrations"

# Read the JSON file and publish each object to RabbitMQ
jq -c '.[]' "$JSON_FILE" | while IFS= read -r payload;
    do
    echo "Publishing payload: $payload"
    if [ -z "$EXCHANGE" ]; then
        plumber write rabbit --address="amqp://guest:guest@localhost:5672/" --exchange-name="amq.default" --routing-key="$ROUTING_KEY" --input="$payload"
    else
        plumber write rabbit --address="amqp://guest:guest@localhost:5672/" --exchange-name="$EXCHANGE" --routing-key="$ROUTING_KEY" --input="$payload"
    fi
done

echo "All payloads from $JSON_FILE published successfully to exchange '$EXCHANGE' with routing key '$ROUTING_KEY'."

