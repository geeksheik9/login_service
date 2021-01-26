#!/usr/bin/env bash
. ./scripts/version.sh

echo "Please enter Mongo url:"
read mongourl

docker run -d -t -i -p 3001:3000 -e LOCAL_MONGO="$mongourl" geeksheik9/login-service:$login_service_version