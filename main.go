package main

import (
	_ "credctl/internal/provider/command" // Import to register providers
	_ "credctl/internal/provider/oauth2"  // Import to register OAuth2 providers
	_ "credctl/internal/provider/oidc"    // Import to register OIDC providers

	"credctl/cmd"
)

func main() {
	cmd.Execute()
}