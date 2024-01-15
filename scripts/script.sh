#!/bin/bash

# Check if the number of rows is provided as an argument
if [ $# -eq 0 ]; then
    echo "Usage: $0 <number_of_rows>"
    exit 1
fi

# Assign the provided argument to the variable n
n=$1

# Generate n rows with user names and sample alert messages
for ((i = 1; i <= n; i++)); do
    echo "user$i,Order has been placed" >> ./alert_initiator/testdata/alerts.csv
done

echo "CSV file 'alerts.csv' created with $n rows."
