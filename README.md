# CoAP over serial

## Note
The below commands has been prepared for setup when server will be run on Raspberry Pi 4 board with Raspian OS and client runs on MacOS.

## Steps to reproduce
### Build and deploy
- Build server app
```
GOOS=linux GOARCH=arm go build -o dist/serial_server cmd/serial_server/main.go
```
- Copy server binary to destination board
```
scp dist/serial_server pi@raspberrypi.local:
```

### Run (order of operations matter) 

#### On destination board
```
./serial_server
```

#### On local machine
```
go run cmd/serial_client/main.go
```