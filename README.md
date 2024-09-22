# CoAP over serial

## Note
The below commands has been prepared for setup when server will be run on Raspberry Pi 4 board with Raspian OS and client runs on MacOS.

## Adapting serial port paths to your environment
The serial port path will vary depending on your operating system and hardware configuration. To ensure the code runs correctly, you must adapt the serial port path to match your specific environment.

In both the client and server implementations, there is a variable SERIAL_PATH that you can modify according to your setup:
```
var SERIAL_PATH = "/dev/tty.usbmodem1201" // Example for macOS
```

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