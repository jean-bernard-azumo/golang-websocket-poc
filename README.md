# Websocket Test

This is a simple websocket POC for future use in the project, namely for Notifications and the admin check-in live dashboard.

## Running the server

```bash
go run main.go
```

Once running, you can open the browser and navigate to `http://localhost:8081/` to see the client connect and send pings every 10 seconds, keeping the connection alive.

## Websocat

Websocat is a command-line client for testing WebSocket servers. It's a simple tool that allows you to connect to a WebSocket server and send and receive messages.

Install it from [here](https://github.com/vi/websocat)

## Running websocat

In a terminal, run:

```bash
websocat ws://localhost:8081/ws
```

Once running, you can write `ping` to the websocat prompt to see the server respond with a pong message.
You can also send a JSON object in the format: `{"status": "connected"}` to see the server update the client's status.

## Endpoints

### `/ws`

The websocket endpoint.

### `/clients`

See the list of clients connected to the websocket.

### `/check-ins`

See the list of check-ins. You'll only see one because we are simulating a single check-in in a fake db.
