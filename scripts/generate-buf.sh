#!/bin/bash

set -e

# Get current user and group IDs
local_uid=$(id -u)
local_gid=$(id -g)


# First, update the buf dependencies if `buf.lock` is not present
if [ ! -f "buf.lock" ]; then
    echo "Updating buf dependencies..."
    docker-compose run --rm buf-mod-update
fi

echo "Generating buf stubs..."

# Run the buf generation container as root
docker-compose run --rm buf-generate

# Fix ownership of generated files
echo "Fixing file ownership..."
sudo chown -R $local_uid:$local_gid go/api
sudo chown $local_uid:$local_gid buf.lock
chmod 644 buf.lock

echo "Buf generation complete!"
