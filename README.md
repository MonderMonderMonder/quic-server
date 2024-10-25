# A Simple QUIC Server-Client in Go

This repository contains a basic implementation of a server and client using the QUIC protocol.

## Requirements

- Go 1.18 or later
- [quic-go](https://github.com/lucas-clemente/quic-go)

## File Structure

- `server/main.go` and `client/main.go` - Server and client logic using the QUIC protocol.
- `build.sh` - Builds both server and client applications.
- `run-server.sh` and `run-client.sh` - Starts the QUIC server and client respectively.
  
## Usage

### Build the Applications

To build both the server and client, run:
```bash
./build.sh
```

### Run the Server

Start the server with:
```bash
./run-server.sh
```

This will start the server and listen for incoming connections over QUIC.

### Run the Client

After the server is running, initiate a client connection with:
```bash
./run-client.sh
```

The client will connect to the server over QUIC and execute the protocol as defined in `main.go`.

## QUIC Protocol Overview

QUIC (Quick UDP Internet Connections) is a modern transport protocol over UDP, offering low-latency communication with enhanced security and reliability.

## Notes

- This implementation here is intended for local testing and development purposes.
- For production deployments, consider setting up TLS certificates and handling secure connections as per `quic-go` best practices.
