// Package registry lets generated extensions hook into the HTTP server without
// editing the hand-written server setup.
package registry

import (
	khttp "github.com/go-kratos/kratos/v3/transport/http"
)

// Registrar installs an extension's routes on the HTTP server.
type Registrar func(*khttp.Server)

var registrars []Registrar

// Add records an extension registrar. It is called from extension init().
func Add(r Registrar) {
	if r != nil {
		registrars = append(registrars, r)
	}
}

// RegisterAll installs every registered extension.
func RegisterAll(server *khttp.Server) {
	for _, r := range registrars {
		r(server)
	}
}
