package provider

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccKeycloakRealmClientPolicyProfile_basic(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "test-profile"
	description := "Test description"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientPolicyProfile_basic(realmName, resourceName, description),
				Check:  testAccCheckKeycloakRealmClientPolicyProfileExists(realmName, resourceName),
			},
		},
	})
}

func TestAccKeycloakRealmClientPolicyProfile_mixedMutations(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	existingProfile := "existing-profile"
	newProfile := "new-profile"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientPolicyProfile_basic(realmName, existingProfile, "Before update"),
				Check:  testAccCheckKeycloakRealmClientPolicyProfileExists(realmName, existingProfile),
			},
			{
				Config: testKeycloakRealmClientPolicyProfile_mixedMutations(realmName, existingProfile, newProfile),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmClientPolicyProfileDescription(realmName, existingProfile, "After update"),
					testAccCheckKeycloakRealmClientPolicyProfileExists(realmName, newProfile),
				),
			},
		},
	})
}

func TestAccKeycloakRealmClientPolicyProfilePolicy_mixedMutations(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	profileName := "test-profile"
	existingPolicy := "existing-policy"
	newPolicy := "new-policy"

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientPolicyProfile_basicWithPolicy(realmName, "Client policy realm", profileName, "Test profile", existingPolicy, "Before update", "any-client", "{}"),
				Check:  testAccCheckKeycloakRealmClientPolicyProfilePolicyExists(realmName, existingPolicy),
			},
			{
				Config: testKeycloakRealmClientPolicyProfilePolicy_mixedMutations(realmName, profileName, existingPolicy, newPolicy),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmClientPolicyProfilePolicyDescription(realmName, existingPolicy, "After update"),
					testAccCheckKeycloakRealmClientPolicyProfilePolicyExists(realmName, newPolicy),
				),
			},
		},
	})
}

func TestAccKeycloakRealmClientPolicyProfile_basicWithExecutor(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "test-profile-with-executor"
	description := "Test description with executor"
	executorName := "pkce-enforcer"
	configuration := map[string]interface{}{
		"auto-configure": "true",
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientPolicyProfile_basicWithExecutor(realmName, resourceName, description, executorName, testKeycloakRealmClientPolicyProfile_mapConfig(configuration)),
				Check:  testAccCheckKeycloakRealmClientPolicyProfileWithExecutorExists(realmName, resourceName, executorName),
			},
		},
	})
}

func TestAccKeycloakRealmClientPolicyProfile_basicWithExecutorAndJSON(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	resourceName := "test-profile-with-executor-and-configuration"
	description := "Test description with executor and configuration"
	executorName := "secure-client-authenticator"
	configuration := map[string]interface{}{
		"allowed-client-authenticators": []string{"client-secret", "client-secret-jwt"},
		"default-client-authenticator":  "client-secret",
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientPolicyProfile_basicWithExecutor(realmName, resourceName, description, executorName, testKeycloakRealmClientPolicyProfile_mapConfig(configuration)),
				Check:  testAccCheckKeycloakRealmClientPolicyProfileWithExecutorMatches(realmName, resourceName, executorName, configuration),
			},
			{
				ResourceName:      "keycloak_realm_client_policy_profile.profile",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/realm-client-policy-profiles/%s", realmName, resourceName),
				ImportStateVerify: true,
			},
			{
				ResourceName:  "keycloak_realm_client_policy_profile.profile",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s/realm-client-policy-profiles/does-not-exist", realmName),
				ExpectError:   regexp.MustCompile("Cannot import non-existent remote object"),
			},
		},
	})
}

func TestAccKeycloakRealmClientPolicyProfile_basicWithPolicy(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	profileName := "test-profile"
	profileDescription := "Test profile description"
	policyName := "test-policy"
	policyDescription := "Test policy description"
	conditionName := "client-attributes"
	configuration := map[string]interface{}{
		"is_negative_logic": false,
		"attributes": []map[string]string{
			{
				"key":   "test-key",
				"value": "test-value",
			},
		},
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientPolicyProfile_basicWithPolicy(realmName, "Client policy realm", profileName, profileDescription, policyName, policyDescription, conditionName, testKeycloakRealmClientPolicyProfile_mapConfig(configuration)),
				Check:  testAccCheckKeycloakRealmClientPolicyProfilePolicyExists(realmName, policyName),
			},
		},
	})
}

func TestAccKeycloakRealmClientPolicyProfile_basicWithPolicyAndJSON(t *testing.T) {
	realmName := acctest.RandomWithPrefix("tf-acc")
	profileName := "test-profile"
	profileDescription := "Test profile description"
	policyName := "test-policy"
	policyDescription := "Test policy description"
	conditionName := "client-updater-context"
	configuration := map[string]interface{}{
		"is_negative_logic":    false,
		"update-client-source": []string{"ByInitialAccessToken", "ByRegistrationAccessToken"},
	}

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testKeycloakRealmClientPolicyProfile_basicWithPolicy(realmName, "Before realm update", profileName, profileDescription, policyName, policyDescription, conditionName, testKeycloakRealmClientPolicyProfile_mapConfig(configuration)),
				Check:  testAccCheckKeycloakRealmClientPolicyProfilePolicyMatches(realmName, policyName, conditionName, configuration),
			},
			{
				Config: testKeycloakRealmClientPolicyProfile_basicWithPolicy(realmName, "After realm update", profileName, profileDescription, policyName, policyDescription, conditionName, testKeycloakRealmClientPolicyProfile_mapConfig(configuration)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeycloakRealmClientPolicyProfileExists(realmName, profileName),
					testAccCheckKeycloakRealmClientPolicyProfilePolicyMatches(realmName, policyName, conditionName, configuration),
				),
			},
			{
				ResourceName:      "keycloak_realm_client_policy_profile_policy.policy",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/realm-client-policy-profile-policies/%s", realmName, policyName),
				ImportStateVerify: true,
			},
			{
				ResourceName:  "keycloak_realm_client_policy_profile_policy.policy",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s/realm-client-policy-profile-policies/does-not-exist", realmName),
				ExpectError:   regexp.MustCompile("Cannot import non-existent remote object"),
			},
		},
	})
}

func testKeycloakRealmClientPolicyProfile_mapConfig(configuration map[string]interface{}) string {
	var s string = "{"
	for k, v := range configuration {
		switch reflect.TypeOf(v).Kind() {
		case reflect.Map, reflect.Slice:
			jsonStr, _ := json.Marshal(v)
			s += fmt.Sprintf("%s = jsonencode(%s)\n", k, string(jsonStr))
		case reflect.String:
			s += fmt.Sprintf("%s = \"%v\"\n", k, v)
		default:
			s += fmt.Sprintf("%s = %v\n", k, v)
		}
	}
	s += "}"
	return s
}

func testKeycloakRealmClientPolicyProfile_basic(realm string, name string, description string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}
	
resource "keycloak_realm_client_policy_profile" "profile" {
	realm_id      = keycloak_realm.realm.realm
	name          = "%s"
	description   = "%s"
}
	`, realm, name, description)
}

func testKeycloakRealmClientPolicyProfile_mixedMutations(realm string, existingProfile string, newProfile string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_client_policy_profile" "profile" {
	realm_id    = keycloak_realm.realm.realm
	name        = "%s"
	description = "After update"
}

resource "keycloak_realm_client_policy_profile" "new" {
	realm_id    = keycloak_realm.realm.realm
	name        = "%s"
	description = "Created alongside update"
}
`, realm, existingProfile, newProfile)
}

func testKeycloakRealmClientPolicyProfile_basicWithExecutor(realm string, name string, description string, executorName string, configuration string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm = "%s"
}

resource "keycloak_realm_client_policy_profile" "profile" {
	realm_id    = keycloak_realm.realm.realm
	name        = "%s"
	description = "%s"

	executor {
		name = "%s"
		configuration = %s
	}
}

resource "keycloak_openid_client" "policy-client" {
  client_id   = "policy-test-client"
  realm_id    = keycloak_realm.realm.realm
  access_type = "PUBLIC"

  depends_on = [
    keycloak_realm_client_policy_profile.profile
  ]
}
	`, realm, name, description, executorName, configuration)
}

func testKeycloakRealmClientPolicyProfile_basicWithPolicy(realm string, realmDisplayName string, profileName string, profileDescription string, policyName string, policyDescription string, conditionName string, configuration string) string {
	return fmt.Sprintf(`
resource "keycloak_realm" "realm" {
	realm        = "%s"
	display_name = "%s"
}

	resource "keycloak_realm_client_policy_profile" "profile" {
	realm_id    = keycloak_realm.realm.realm
	name        = "%s"
	description	= "%s"
}

resource "keycloak_realm_client_policy_profile_policy" "policy" {
	realm_id    = keycloak_realm.realm.realm
  name        = "%s"
  description = "%s"

  profiles = [
    keycloak_realm_client_policy_profile.profile.name
  ]

  condition {
    name = "%s"
    configuration = %s
  }
}

resource "keycloak_openid_client" "policy-client" {
  client_id   = "policy-test-client"
  realm_id    = keycloak_realm.realm.realm
  access_type = "PUBLIC"

  depends_on = [
    keycloak_realm_client_policy_profile.profile
  ]
}
	`, realm, realmDisplayName, profileName, profileDescription, policyName, policyDescription, conditionName, configuration)
}

func testKeycloakRealmClientPolicyProfilePolicy_mixedMutations(realm string, profileName string, existingPolicy string, newPolicy string) string {
	return testKeycloakRealmClientPolicyProfile_basicWithPolicy(realm, "Client policy realm", profileName, "Test profile", existingPolicy, "After update", "any-client", "{}") + fmt.Sprintf(`

resource "keycloak_realm_client_policy_profile_policy" "new" {
	realm_id    = keycloak_realm.realm.realm
	name        = "%s"
	description = "Created alongside update"

	profiles = [
		keycloak_realm_client_policy_profile.profile.name
	]

	condition {
		name          = "any-client"
		configuration = {}
	}
}
`, newPolicy)
}

func testAccCheckKeycloakRealmClientPolicyProfileExists(realm string, profileName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := keycloakClient.GetRealmClientPolicyProfileByName(testCtx, realm, profileName)
		if err != nil {
			return fmt.Errorf("Client policy profile not found: %s", profileName)
		}

		return nil
	}
}

func testAccCheckKeycloakRealmClientPolicyProfileDescription(realm string, profileName string, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		profile, err := keycloakClient.GetRealmClientPolicyProfileByName(testCtx, realm, profileName)
		if err != nil {
			return fmt.Errorf("client policy profile not found: %s", profileName)
		}
		if profile.Description != description {
			return fmt.Errorf("unexpected description for client policy profile %s: got %q, want %q", profileName, profile.Description, description)
		}

		return nil
	}
}

func testAccCheckKeycloakRealmClientPolicyProfileWithExecutorExists(realm string, profileName string, executorName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		profile, err := keycloakClient.GetRealmClientPolicyProfileByName(testCtx, realm, profileName)
		if err != nil {
			return fmt.Errorf("Client policy profile not found: %s", profileName)
		}

		if profile.Executors[0].Name != executorName {
			return fmt.Errorf("Client policy profile executor not found: %s", executorName)
		}

		return nil
	}
}

func testAccCheckKeycloakRealmClientPolicyProfileWithExecutorMatches(realm string, profileName string, executorName string, configuration map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		profile, err := keycloakClient.GetRealmClientPolicyProfileByName(testCtx, realm, profileName)
		if err != nil {
			return fmt.Errorf("Client policy profile not found: %s", profileName)
		}

		if profile.Executors[0].Name != executorName {
			return fmt.Errorf("Client policy profile executor not found: %s", executorName)
		}

		for k, got := range profile.Executors[0].Configuration {
			want := configuration[k]

			if !equalsIgnoreType(got, want) {
				return fmt.Errorf("Client policy profile executor configuration does not match: want %v, got %v", want, got)
			}
		}

		return nil
	}
}

func testAccCheckKeycloakRealmClientPolicyProfilePolicyExists(realm string, policyName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := keycloakClient.GetRealmClientPolicyProfilePolicyByName(testCtx, realm, policyName)
		if err != nil {
			return fmt.Errorf("Client policy profile policy not found: %s", policyName)
		}

		return nil
	}
}

func testAccCheckKeycloakRealmClientPolicyProfilePolicyDescription(realm string, policyName string, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		policy, err := keycloakClient.GetRealmClientPolicyProfilePolicyByName(testCtx, realm, policyName)
		if err != nil {
			return fmt.Errorf("client policy profile policy not found: %s", policyName)
		}
		if policy.Description != description {
			return fmt.Errorf("unexpected description for client policy profile policy %s: got %q, want %q", policyName, policy.Description, description)
		}

		return nil
	}
}

func testAccCheckKeycloakRealmClientPolicyProfilePolicyMatches(realm string, policyName string, conditionName string, configuration map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		policy, err := keycloakClient.GetRealmClientPolicyProfilePolicyByName(testCtx, realm, policyName)
		if err != nil {
			return fmt.Errorf("Client policy profile policy not found: %s", policyName)
		}

		if policy.Conditions[0].Name != conditionName {
			return fmt.Errorf("Client policy profile policy condition not found: %s", conditionName)
		}

		for k, got := range policy.Conditions[0].Configuration {
			want := configuration[k]

			if !equalsIgnoreType(got, want) {
				return fmt.Errorf("Client policy profile policy condition configuration does not match: want %v, got %v", want, got)
			}
		}

		return nil
	}
}
