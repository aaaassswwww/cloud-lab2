#!/usr/bin/bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo "Script directory: $script_dir"
cd "$script_dir/.."
echo "Current directory: $(pwd)"

services=(
    "cart"
    "checkout"
    "email"
    "frontend"
    "order"
    "payment"
    "product"
    "user"
)
# Build images for all services
for service in "${services[@]}"; do
    echo "Building image for $service and pushing to Docker Hub..."
    docker buildx build \
        --platform linux/arm64,linux/amd64 \
        -t "buwandocker/$service:lab2" \
        -f app/$service/Dockerfile \
        --push .
    echo "Image for $service built successfully."

    echo "For product microservice, building an unhealthy image and pushing to Docker Hub..."
    if [[ "$service" == "product" ]]; then
        docker buildx build \
            --platform linux/arm64,linux/amd64 \
            -t "buwandocker/$service:lab2-unhealthy" \
            -f app/$service/Dockerfile.unhealthy \
            --push .
        echo "Unhealthy image for $service built successfully."
    fi
done