# Adapting serial port paths to your environment
The serial port path will vary depending on your operating system and hardware configuration. To ensure the code runs correctly, you must adapt the serial port path to match your specific environment.

In both the client and server implementations, there is a variable SERIAL_PATH that you can modify according to your setup:
```
var SERIAL_PATH = "/dev/tty.usbmodem1201" // Example for macOS
```
# Note
The below commands has been prepared for setup when server will be run on Raspberry Pi 4 board with Raspian OS and client runs on MacOS.
## Adapting serial port paths to your environment
The serial port path will vary depending on your operating system and hardware configuration. To ensure the code runs correctly, you must adapt the serial port path to match your specific environment.

In both the client and server implementations, there is a variable SERIAL_PATH that you can modify according to your setup:
```
var SERIAL_PATH = "/dev/tty.usbmodem1201" // Example for macOS
```

# CoAP over serial
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

# DTLS over serial
## Steps to reproduce
### Cert genration
To generate CA and certs you can sue the following commands:
```
# Self-signed CA
CERT_SUBJ="/C=PL/ST=Pl/L=Pll/O=Plll/CN=example.com"
openssl ecparam -name secp224r1 -genkey -noout -out root_ca_key.pem
openssl ec -in root_ca_key.pem -pubout -out root_ca_pubkey.pem
openssl req -new -key root_ca_key.pem -x509 -nodes -days 365 -out root_ca_cert.pem -subj $CERT_SUBJ

# Server
openssl ecparam -name secp224r1 -genkey -noout -out server_key.pem
openssl req -new -sha256 -key server_key.pem -subj $CERT_SUBJ -out server.csr
openssl x509 -req -in server.csr  -CA root_ca_cert.pem -CAkey root_ca_key.pem -CAcreateserial -out server_cert.pem -days 500 -sha256

# Client
CERT_SUBJ="/C=PL/ST=Pl/L=Pll/O=Plll/CN=example.com/emailAddress=client1@example.com"
openssl ecparam -name secp224r1 -genkey -noout -out client_key.pem
openssl req -new -sha256 -key client_key.pem -subj $CERT_SUBJ -out client.csr
openssl x509 -req -in client.csr  -CA root_ca_cert.pem -CAkey root_ca_key.pem -CAcreateserial -out client_cert.pem -days 500 -sha256
```

All neccessary keys and certs need to be placed in certs dir in the same location, when binary is started.

### Build and deploy
- Build server app
```
GOOS=linux GOARCH=arm go build -o dist/dtls_server cmd/dtls_server/main.go
```
- Copy server binary to destination board
```
scp dist/dtls_server pi@raspberrypi.local:
```

### Run (order of operations matter) 
#### On destination board
```
./dtls_server
```

#### On local machine
```
go run cmd/dtls_client/main.go
```
