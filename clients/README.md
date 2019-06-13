# Client
Small IOT core client

## Getting Started

### Configuration
Update device and project variables in env_gen.sh

### Prerequisites

Create certificates:
[See doc](https://cloud.google.com/iot/docs/how-tos/credentials/keys)
```
openssl req -x509 -nodes -newkey rsa:2048 \
    -keyout device.key.pem \
    -out device.crt.pem \
    -days 365 \
    -subj "/CN=unused"
curl https://pki.google.com/roots.pem > ./root-ca.pem
```

Register a device
```
gcloud iot devices create $IOTCORE_DEVICE \
  --project=$IOTCORE_PROJECT \
  --region=$IOTCORE_REGION \
  --registry=$IOTCORE_REGISTRY \
  --public-key path=./device.crt.pem,type=rsa-x509-pem
```

## Download dependencies
```
go get github.com/dgrijalva/jwt-go
go get github.com/eclipse/paho.mqtt.golang
```

## Run

```
go run ./main.go \
  -project $IOTCORE_PROJECT \
  -region $IOTCORE_REGION \
  -registry $IOTCORE_REGISTRY \
  -device $IOTCORE_DEVICE \
  -ca ./root-ca.pem \
  -key ./device.key.pem \
  -src "iot-core demo" \
  -events 10
```
