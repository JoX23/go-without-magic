package main

import "embed"

// templatesFS contiene todos los templates embebidos en el binario.
//
//go:embed templates
var templatesFS embed.FS
