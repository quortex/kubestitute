// Package gen is an empty package to generate AWS EC2 adapter client code from swagger spec
package gen

//go:generate docker run --rm -v ${PWD}:/app -w /app quay.io/goswagger/swagger:v0.25.0 generate client -f ./clients/ec2adapter/swagger.yaml -A aws-ec2-adapter -t ./clients/ec2adapter
