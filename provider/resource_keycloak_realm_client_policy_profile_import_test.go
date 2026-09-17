package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func TestResourceKeycloakRealmClientPolicyProfileImport(t *testing.T) {
	testRealmClientPolicyImport(
		t,
		resourceKeycloakRealmClientPolicyProfile(),
		"test-realm/realm-client-policy-profiles/test-profile",
		"test-realm",
		"test-profile",
	)
}

func TestResourceKeycloakRealmClientPolicyProfilePolicyImport(t *testing.T) {
	testRealmClientPolicyImport(
		t,
		resourceKeycloakRealmClientPolicyProfilePolicy(),
		"test-realm/realm-client-policy-profile-policies/test-policy",
		"test-realm",
		"test-policy",
	)
}

func TestResourceKeycloakRealmClientPolicyImportRejectsInvalidIDs(t *testing.T) {
	resources := map[string]*schema.Resource{
		"profile": resourceKeycloakRealmClientPolicyProfile(),
		"policy":  resourceKeycloakRealmClientPolicyProfilePolicy(),
	}

	invalidIDs := []string{
		"",
		"test-realm/test-name",
		"/realm-client-policy-profiles/test-profile",
		"test-realm/realm-client-policy-profiles/",
		"test-realm/wrong-resource/test-name",
	}

	for resourceName, resourceSchema := range resources {
		for _, id := range invalidIDs {
			t.Run(resourceName+"/"+id, func(t *testing.T) {
				data := schema.TestResourceDataRaw(t, resourceSchema.Schema, nil)
				data.SetId(id)

				_, err := resourceSchema.Importer.StateContext(context.Background(), data, nil)
				if err == nil {
					t.Fatalf("expected import ID %q to be rejected", id)
				}
			})
		}
	}
}

func testRealmClientPolicyImport(t *testing.T, resourceSchema *schema.Resource, id, wantRealm, wantName string) {
	t.Helper()

	data := schema.TestResourceDataRaw(t, resourceSchema.Schema, nil)
	data.SetId(id)

	states, err := resourceSchema.Importer.StateContext(context.Background(), data, nil)
	if err != nil {
		t.Fatalf("unexpected import error: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("expected one imported state, got %d", len(states))
	}
	if got := states[0].Get("realm_id"); got != wantRealm {
		t.Errorf("unexpected realm_id: got %q, want %q", got, wantRealm)
	}
	if got := states[0].Get("name"); got != wantName {
		t.Errorf("unexpected name: got %q, want %q", got, wantName)
	}
	if got := states[0].Id(); got != id {
		t.Errorf("unexpected resource ID: got %q, want %q", got, id)
	}
}

func TestRealmClientPolicyImportFlattensPrimitiveConfigurationValues(t *testing.T) {
	profileData := schema.TestResourceDataRaw(t, resourceKeycloakRealmClientPolicyProfile().Schema, nil)
	profile := &keycloak.RealmClientPolicyProfile{
		Name:    "test-profile",
		RealmId: "test-realm",
		Executors: []keycloak.RealmClientPolicyProfileExecutor{{
			Name: "pkce-enforcer",
			Configuration: map[string]interface{}{
				"enabled": true,
				"retries": float64(3),
			},
		}},
	}

	if err := mapFromRealmClientPolicyProfileToData(profileData, profile); err != nil {
		t.Fatalf("mapping profile import state: %v", err)
	}

	executors := profileData.Get("executor").([]interface{})
	configuration := executors[0].(map[string]interface{})["configuration"].(map[string]interface{})
	if got := configuration["enabled"]; got != "true" {
		t.Errorf("unexpected boolean configuration value: got %q, want %q", got, "true")
	}
	if got := configuration["retries"]; got != "3" {
		t.Errorf("unexpected numeric configuration value: got %q, want %q", got, "3")
	}
}
