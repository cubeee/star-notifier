# Star notifier

Sends Discord webhooks when shooting stars are scouted, along with a precise map image if mapped.

![Example webhook](/misc/example.png)

## Usage
Run in a Docker container with at least the required environment variables set.

## Environment variables
See [lib/env.go](/lib/env.go)

## TODO
* Replace hardcoded `StarLocations` in `lib/stars.go` with ability to read from a file
