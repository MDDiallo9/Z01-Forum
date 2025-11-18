#!/bin/bash

# Build the Docker image
echo "Building Docker image..."
docker build -t forum-app .

# Check if build was successful
if [ $? -eq 0 ]; then
    echo "Docker image built successfully."
    echo "To run the container, use: docker run -p 8000:8000 forum-app"
else
    echo "Docker build failed."
    exit 1
fi
