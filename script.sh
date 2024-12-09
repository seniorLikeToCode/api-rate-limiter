#!/usr/bin/env bash

URL="http://localhost:4000"

echo "Sending 500 GET requests to $URL..."

# Initialize counters before starting the loop
success_count=0
too_many_requests_count=0

for i in {1..500}; do
    # echo "Request #$i"
    
    # Send a request and capture only the HTTP status code
    RESPONSE_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$URL")
    # echo "Response Code: $RESPONSE_CODE"

    # Increment the appropriate counter based on the response code
    if [ "$RESPONSE_CODE" = "200" ]; then
        success_count=$((success_count + 1))
    elif [ "$RESPONSE_CODE" = "429" ]; then
        too_many_requests_count=$((too_many_requests_count + 1))
    fi

    # Print the counts so far
    # echo "Count of 200 responses so far: $success_count"
    # echo "Count of 429 responses so far: $too_many_requests_count"
    # echo "------------------------------"
done

echo "Final count of 200 responses: $success_count"
echo "Final count of 429 responses: $too_many_requests_count"
