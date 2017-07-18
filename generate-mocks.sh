#!/bin/bash

mkdir -p mock_batch
mockgen -source batch/client.go -destination mock_batch/client.go
mkdir -p mock_client
mockgen -source client/logger.go -destination mock_client/logger.go