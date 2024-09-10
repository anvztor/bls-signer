#!/bin/bash

# Compile the program
go build main.go

# Start the Proposer node
echo "Starting Proposer node..."
PROPOSER=true LIBP2P_PORT=4111 LIBP2P_BOOT_NODES="/ip4/127.0.0.1/tcp/4001/p2p/12D3KooWA5HKLduQU1SU6iKmZjqQ8Ya1WkLjerZ7mbV6Rqnn2cLh" RELAYER_PRI_KEY="0x111.pem" ./main &
PROPOSER_PID=$!

# Wait a few seconds to ensure the Proposer node has started
sleep 5

# Start the normal node
echo "Starting normal node..."
PROPOSER=false LIBP2P_PORT=4112 LIBP2P_BOOT_NODES="/ip4/127.0.0.1/tcp/4001/p2p/12D3KooWA5HKLduQU1SU6iKmZjqQ8Ya1WkLjerZ7mbV6Rqnn2cLh" RELAYER_PRI_KEY="0x222.pem" ./main &
NORMAL_PID=$!

# Wait for a while to allow the nodes to interact
echo "Waiting for node interaction..."
sleep 40

# Terminate both nodes
echo "Terminating nodes..."
kill $PROPOSER_PID
kill $NORMAL_PID

echo "Test completed"