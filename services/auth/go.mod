module auth

go 1.26.3

require middleware v0.0.0

require (
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
)

replace middleware => ../middleware
