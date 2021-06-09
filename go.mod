module github.com/ohowland/cgc_core

go 1.16

require (
	github.com/eclipse/paho.mqtt.golang v1.3.5
	github.com/go-sql-driver/mysql v1.6.0
	github.com/goburrow/modbus v0.1.0
	github.com/goburrow/serial v0.1.0 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/uuid v1.2.0
	github.com/ohowland/cgc_optimize v0.0.0-unpublished
	go.mongodb.org/mongo-driver v1.5.3
	gotest.tools v2.2.0+incompatible
	gotest.tools/v3 v3.0.3
)

replace github.com/ohowland/cgc_optimize v0.0.0-unpublished => ../cgc_optimize
