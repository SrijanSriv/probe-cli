// Code generated by go generate; DO NOT EDIT.
// 2022-05-19 20:30:49.209515 +0200 CEST m=+0.000689210

package ooapi

//go:generate go run ./internal/generator -file cloners.go

// clonerForPsiphonConfigAPI represents any type exposing a method
// like simplePsiphonConfigAPI.WithToken.
type clonerForPsiphonConfigAPI interface {
	WithToken(token string) callerForPsiphonConfigAPI
}

// clonerForTorTargetsAPI represents any type exposing a method
// like simpleTorTargetsAPI.WithToken.
type clonerForTorTargetsAPI interface {
	WithToken(token string) callerForTorTargetsAPI
}
