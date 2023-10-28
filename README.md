# Distributed Websocket Server

This is a POC/template for creating a distributed service for handling websocket connections using Redis as the central pub/sub provider.

### The aim of the POC:

- To be able to create multiple servers connecting to a single Redis instance where all the messages from websockets are published to and subscribed from.

- If the `--redis` flag isn't provided, then the websocket server shouldn't rely on a Redis server and maintain all the messages in memory until they're written to the sockets

### Why Redis?

We're using Redis here for it's use case as a multiplexer to which all the backend servers connect to and the system is horizontally scalable without creating a stickiness to a single server for websocket connections.

We can also use any other pub/sub provider in place of Redis.
