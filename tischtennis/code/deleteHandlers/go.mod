module main

go 1.16

require (
	github.com/aws/aws-lambda-go v1.36.0
	github.com/aws/aws-sdk-go v1.44.160 // indirect
	github.com/google/uuid v1.3.0 // indirect
	tischtennis/database v0.0.0
	tischtennis/helpers v0.0.0
)

replace tischtennis/helpers v0.0.0 => ../helpers

replace tischtennis/database v0.0.0 => ../database
