#!/bin/bash

# URL to be requested
URL="http://www.zzgeda.com/utilityBill/signUp/dropDownList/getUniversityAndArea"

# Number of times to execute the request
NUM_REQUESTS=1000

# Loop to execute the request
for ((i=1; i<=NUM_REQUESTS; i++))
do
  # Execute the request with curl and disable keep-alive
  curl --no-keepalive "$URL"
  
  # Optionally, you can add a delay between requests to avoid overwhelming the server
  # sleep 0.1
done

