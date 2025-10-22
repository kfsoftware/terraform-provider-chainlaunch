package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"chainlaunch": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	// Check that required environment variables are set
	if v := os.Getenv("CHAINLAUNCH_URL"); v == "" {
		t.Fatal("CHAINLAUNCH_URL must be set for acceptance tests")
	}

	// Check for either API key or username/password
	apiKey := os.Getenv("CHAINLAUNCH_API_KEY")
	username := os.Getenv("CHAINLAUNCH_USERNAME")
	password := os.Getenv("CHAINLAUNCH_PASSWORD")

	if apiKey == "" && (username == "" || password == "") {
		t.Fatal("Either CHAINLAUNCH_API_KEY or both CHAINLAUNCH_USERNAME and CHAINLAUNCH_PASSWORD must be set for acceptance tests")
	}
}

func TestProvider(t *testing.T) {
	// This is a basic test that ensures the provider can be instantiated
	provider := New("test")()
	if provider == nil {
		t.Fatal("Provider instantiation failed: New(\"test\")() returned nil")
	}
}
