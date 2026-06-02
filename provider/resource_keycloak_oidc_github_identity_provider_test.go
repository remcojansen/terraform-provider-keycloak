package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

/*
	FIXME(raito): check that?
	note: we cannot use parallel tests for this resource as only one instance of a Github identity provider can be created
	for a realm.
*/

func TestAccKeycloakOidcGithubIdentityProvider_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcGithubIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcGithubIdentityProvider_basic(),
				Check:  testAccCheckKeycloakOidcGithubIdentityProviderExists("keycloak_oidc_github_identity_provider.github"),
			},
		},
	})
}

func TestAccKeycloakOidcGithubIdentityProvider_customAlias(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcGithubIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_github_identity_provider" "github" {
	realm             = data.keycloak_realm.realm.id
	client_id         = "example_id"
	client_secret     = "example_token"

	alias = "example"
}
	`, testAccRealm.Realm),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOidcGithubIdentityProviderExists("keycloak_oidc_github_identity_provider.github"),
					resource.TestCheckResourceAttr("keycloak_oidc_github_identity_provider.github", "alias", "example"),
				),
			},
		},
	})
}

func TestAccKeycloakOidcGithubIdentityProvider_customDisplayName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcGithubIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_github_identity_provider" "github" {
	realm             = data.keycloak_realm.realm.id
	client_id         = "example_id"
	client_secret     = "example_token"

	display_name = "Example Github"
}
	`, testAccRealm.Realm),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOidcGithubIdentityProviderExists("keycloak_oidc_github_identity_provider.github"),
					resource.TestCheckResourceAttr("keycloak_oidc_github_identity_provider.github", "display_name", "Example Github"),
				),
			},
		},
	})
}

func TestAccKeycloakOidcGithubIdentityProvider_extraConfig(t *testing.T) {
	customConfigValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcGithubIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcGithubIdentityProvider_customConfig("dummyConfig", customConfigValue),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOidcGithubIdentityProviderExists("keycloak_oidc_github_identity_provider.github_custom"),
					testAccCheckKeycloakOidcGithubIdentityProviderHasCustomConfigValue("keycloak_oidc_github_identity_provider.github_custom", customConfigValue),
				),
			},
		},
	})
}

// ensure that extra_config keys which are covered by top-level attributes are not allowed
func TestAccKeycloakOidcGithubIdentityProvider_extraConfigInvalid(t *testing.T) {
	customConfigValue := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcGithubIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config:      testKeycloakOidcGithubIdentityProvider_customConfig("syncMode", customConfigValue),
				ExpectError: regexp.MustCompile("extra_config key \"syncMode\" is not allowed"),
			},
		},
	})
}

func TestAccKeycloakOidcGithubIdentityProvider_linkOrganization(t *testing.T) {

	organizationName := acctest.RandomWithPrefix("tf-acc")

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcGithubIdentityProvider_linkOrganization(organizationName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakOidcGithubIdentityProviderExists("keycloak_oidc_github_identity_provider.github"),
					testAccCheckKeycloakOidcGithubIdentityProviderLinkOrganization("keycloak_oidc_github_identity_provider.github"),
				),
			},
		},
	})
}

func TestAccKeycloakOidcGithubIdentityProvider_createAfterManualDestroy(t *testing.T) {
	var idp = &keycloak.IdentityProvider{}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcGithubIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcGithubIdentityProvider_basic(),
				Check:  testAccCheckKeycloakOidcGithubIdentityProviderFetch("keycloak_oidc_github_identity_provider.github", idp),
			},
			{
				PreConfig: func() {
					err := keycloakClient.DeleteIdentityProvider(testCtx, idp.Realm, idp.Alias)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testKeycloakOidcGithubIdentityProvider_basic(),
				Check:  testAccCheckKeycloakOidcGithubIdentityProviderExists("keycloak_oidc_github_identity_provider.github"),
			},
		},
	})
}

func TestAccKeycloakOidcGithubIdentityProvider_basicUpdateAll(t *testing.T) {
	firstEnabled := randomBool()
	firstHideOnLogin := randomBool()

	firstOidc := &keycloak.IdentityProvider{
		Alias:       acctest.RandString(10),
		Enabled:     firstEnabled,
		HideOnLogin: firstHideOnLogin,
		Config: &keycloak.IdentityProviderConfig{
			ClientId:     acctest.RandString(10),
			ClientSecret: acctest.RandString(10),
			GuiOrder:     strconv.Itoa(acctest.RandIntRange(1, 3)),
			SyncMode:     randomStringInSlice(syncModes),
		},
	}

	secondOidc := &keycloak.IdentityProvider{
		Alias:       acctest.RandString(10),
		Enabled:     !firstEnabled,
		HideOnLogin: !firstHideOnLogin,
		Config: &keycloak.IdentityProviderConfig{
			ClientId:     acctest.RandString(10),
			ClientSecret: acctest.RandString(10),
			GuiOrder:     strconv.Itoa(acctest.RandIntRange(1, 3)),
			SyncMode:     randomStringInSlice(syncModes),
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcGithubIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testKeycloakOidcGithubIdentityProvider_basicFromInterface(firstOidc),
				Check:  testAccCheckKeycloakOidcGithubIdentityProviderExists("keycloak_oidc_github_identity_provider.github"),
			},
			{
				Config: testKeycloakOidcGithubIdentityProvider_basicFromInterface(secondOidc),
				Check:  testAccCheckKeycloakOidcGithubIdentityProviderExists("keycloak_oidc_github_identity_provider.github"),
			},
		},
	})
}

func TestAccKeycloakOidcGithubIdentityProvider_clientSecretWriteOnly(t *testing.T) {
	t.Parallel()

	oidcName := acctest.RandomWithPrefix("tf-acc")
	clientSecretWO := acctest.RandomWithPrefix("tf-acc")
	clientSecretWOVersion := 1

	// the keycloak client is obfuscating the client_secret value, therefore we can't assert its value
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		CheckDestroy:             testAccCheckKeycloakOidcGithubIdentityProviderDestroy(),
		Steps: []resource.TestStep{
			{
				// test CREATION of the client_secret via write-only attribute
				Config: testKeycloakOidcGithubIdentityProvider_clientSecretWriteOnly(oidcName, clientSecretWO, clientSecretWOVersion),
				Check: resource.ComposeTestCheckFunc(
					// assert openid client against the Keycloak's API response (value SHOULD be the new one)
					testAccCheckKeycloakOidcGithubIdentityProviderExists("keycloak_oidc_github_identity_provider.oidc"),

					// assert openid client against the Terraform state (client_secret value SHOULD NOT be stored in state)
					resource.TestCheckNoResourceAttr("keycloak_oidc_github_identity_provider.oidc", "client_secret"),
					resource.TestCheckResourceAttr("keycloak_oidc_github_identity_provider.oidc", "client_secret_wo_version", strconv.Itoa(clientSecretWOVersion)),
				),
			},
		},
	})
}

func testAccCheckKeycloakOidcGithubIdentityProviderExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getKeycloakOidcGithubIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckKeycloakOidcGithubIdentityProviderFetch(resourceName string, idp *keycloak.IdentityProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedOidc, err := getKeycloakOidcGithubIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		idp.Alias = fetchedOidc.Alias
		idp.Realm = fetchedOidc.Realm

		return nil
	}
}

func testAccCheckKeycloakOidcGithubIdentityProviderHasCustomConfigValue(resourceName, customConfigValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedOidc, err := getKeycloakOidcGithubIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		if fetchedOidc.Config.ExtraConfig["dummyConfig"].(string) != customConfigValue {
			return fmt.Errorf("expected custom oidc provider to have config with a custom key 'dummyConfig' with a value %s, but value was %s", customConfigValue, fetchedOidc.Config.ExtraConfig["dummyConfig"].(string))
		}

		return nil
	}
}

func testAccCheckKeycloakOidcGithubIdentityProviderLinkOrganization(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fetchedOidc, err := getKeycloakOidcIdentityProviderFromState(s, resourceName)
		if err != nil {
			return err
		}

		if fetchedOidc.OrganizationId == "" {
			return fmt.Errorf("expected custom oidc provider to be linked with an organization, but it was not")
		}

		return nil
	}
}

func testAccCheckKeycloakOidcGithubIdentityProviderDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "keycloak_oidc_github_identity_provider" {
				continue
			}

			id := rs.Primary.ID
			realm := rs.Primary.Attributes["realm"]

			idp, _ := keycloakClient.GetIdentityProvider(testCtx, realm, id)
			if idp != nil {
				return fmt.Errorf("oidc config with id %s still exists", id)
			}
		}

		return nil
	}
}

func getKeycloakOidcGithubIdentityProviderFromState(s *terraform.State, resourceName string) (*keycloak.IdentityProvider, error) {
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return nil, fmt.Errorf("resource not found: %s", resourceName)
	}

	realm := rs.Primary.Attributes["realm"]
	alias := rs.Primary.Attributes["alias"]

	idp, err := keycloakClient.GetIdentityProvider(testCtx, realm, alias)
	if err != nil {
		return nil, fmt.Errorf("error getting oidc identity provider config with alias %s: %s", alias, err)
	}

	return idp, nil
}

func testKeycloakOidcGithubIdentityProvider_basic() string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_github_identity_provider" "github" {
	realm             = data.keycloak_realm.realm.id
	client_id         = "example_id"
	client_secret     = "example_token"
}
	`, testAccRealm.Realm)
}

func testKeycloakOidcGithubIdentityProvider_customConfig(configKey, configValue string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_github_identity_provider" "github_custom" {
	realm             = data.keycloak_realm.realm.id
	provider_id       = "github"
	client_id         = "example_id"
	client_secret     = "example_token"
	extra_config      = {
		%s = "%s"
	}
}
	`, testAccRealm.Realm, configKey, configValue)
}

func testKeycloakOidcGithubIdentityProvider_basicFromInterface(idp *keycloak.IdentityProvider) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_github_identity_provider" "github" {
	realm             						= data.keycloak_realm.realm.id
	enabled           						= %t
	api_url                                 = "%s"
	base_url                                = "%s"
	github_json_format                      = %t
	client_id         						= "%s"
	client_secret     						= "%s"
	gui_order                               = %s
	sync_mode                               = "%s"
	hide_on_login_page                      = %t
}
	`, testAccRealm.Realm, idp.Enabled, idp.Config.ApiUrl, idp.Config.BaseUrl, idp.Config.GithubJsonFormat, idp.Config.ClientId, idp.Config.ClientSecret, idp.Config.GuiOrder, idp.Config.SyncMode, idp.HideOnLogin)
}

func testKeycloakOidcGithubIdentityProvider_linkOrganization(organizationName string) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_organization" "org" {
	realm   = data.keycloak_realm.realm.id
	name    = "%s"
	enabled = true

	domain {
		name     = "example.com"
		verified = true
 	}
}

resource "keycloak_oidc_github_identity_provider" "github" {
	realm             = data.keycloak_realm.realm.id
	client_id         = "example_id"
	client_secret     = "example_token"

	organization_id   				= keycloak_organization.org.id
	org_domain		  				= "example.com"
	org_redirect_mode_email_matches = true
}
	`, testAccRealm.Realm, organizationName)
}

func testKeycloakOidcGithubIdentityProvider_clientSecretWriteOnlyFromComputedValue(oidc string) string {
	// 'client_secret_wo_version' is always 1
	// we are making it conditional to make the value unknown during the validation
	// which is the same situation as when the value is not hardcoded, but comes from other module
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_openid_client" "secret_source" {
	realm_id    = data.keycloak_realm.realm.id
	client_id   = "%s-secret-source"
	access_type = "CONFIDENTIAL"
}

resource "keycloak_oidc_github_identity_provider" "oidc" {
	realm                    = data.keycloak_realm.realm.id
	alias                    = "%s"
	authorization_url        = "https://example.com/auth"
	token_url                = "https://example.com/token"
	client_id                = "example_id"
	client_secret_wo         = keycloak_openid_client.secret_source.client_secret
	client_secret_wo_version = keycloak_openid_client.secret_source.id != "" ? 1 : 0
}
	`, testAccRealm.Realm, oidc, oidc)
}

func testKeycloakOidcGithubIdentityProvider_clientSecretWriteOnly(oidc, clientSecretWriteOnly string, clientSecretWriteOnlyVersion int) string {
	return fmt.Sprintf(`
data "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_oidc_github_identity_provider" "oidc" {
	realm             		 = data.keycloak_realm.realm.id
	alias             		 = "%s"
	authorization_url 		 = "https://example.com/auth"
	token_url         		 = "https://example.com/token"
	client_id         		 = "example_id"
	client_secret_wo         = "%s"
	client_secret_wo_version = "%d"

	issuer = "hello"
}
	`, testAccRealm.Realm, oidc, clientSecretWriteOnly, clientSecretWriteOnlyVersion)
}
