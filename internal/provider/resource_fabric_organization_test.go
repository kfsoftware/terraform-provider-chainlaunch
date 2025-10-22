package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganizationResource(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-test")
	mspID := acctest.RandomWithPrefix("TestMSP")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccOrganizationResourceConfig(rName, mspID, "Test organization"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("chainlaunch_organization.test", "name", rName),
					resource.TestCheckResourceAttr("chainlaunch_organization.test", "msp_id", mspID),
					resource.TestCheckResourceAttr("chainlaunch_organization.test", "description", "Test organization"),
					resource.TestCheckResourceAttrSet("chainlaunch_organization.test", "id"),
					resource.TestCheckResourceAttrSet("chainlaunch_organization.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "chainlaunch_organization.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccOrganizationResourceConfig(rName, mspID, "Updated test organization"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("chainlaunch_organization.test", "description", "Updated test organization"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccOrganizationResourceConfig(name, mspID, description string) string {
	return fmt.Sprintf(`
resource "chainlaunch_organization" "test" {
  name        = %[1]q
  msp_id      = %[2]q
  description = %[3]q
}
`, name, mspID, description)
}
