module github.com/ohowland/cgc_core

go 1.16

require (
	github.com/eclipse/paho.mqtt.golang v1.3.5 // indirect
	github.com/goburrow/modbus v0.1.0 // indirect
	github.com/goburrow/serial v0.1.0 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/ohowland/cgc_optimize v0.0.0-unpublished
	github.com/pkg/errors v0.9.1 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	gotest.tools v2.2.0+incompatible // indirect
	example.com/theirmodule v0.0.0-unpublished
)

replace github.com/ohowland/cgc_optimize v0.0.0-unpublished => ../cgc_optimize