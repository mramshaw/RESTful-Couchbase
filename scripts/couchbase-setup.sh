#!/bin/bash

# Print commands as they are executed
set -x

# Enable job control
set -m

# Start couchbase server in background
/entrypoint.sh couchbase-server &

# Allow couchbase time to start up cleanly
sleep 10

# Create cluster, minimal settings for testing
/opt/couchbase/bin/couchbase-cli cluster-init -c localhost --cluster-username halcouch --cluster-password couchpass --services data,index,query

# Create bucket, minimal settings for testing
/opt/couchbase/bin/couchbase-cli bucket-create -c localhost --username halcouch --password couchpass --bucket recipes --bucket-type couchbase --bucket-ramsize 100 --enable-flush=1

# Return to the couchbase server
fg 1
