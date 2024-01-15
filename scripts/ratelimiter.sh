#!/bin/bash

# Run curl 10 times
for i in {1..10}; do
   curl --location 'http://localhost:3334/webhook' --header 'Content-Type: application/json' --data '{
    "userId": "125",
    "alertMessage": "Order has been placed"
}' &
done

# Wait for all background processes to finish
wait

