package main

import (
	_ "credctl/internal/provider/command"     // Import to register providers
	_ "credctl/internal/provider/oauth2"      // Import to register OAuth2 provider
	_ "credctl/internal/provider/oauth2proxy" // Import to register OAuth2 Proxy provider

	"credctl/cmd"
)

func main() {
	cmd.Execute()
}
