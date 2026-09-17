package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/keycloak/terraform-provider-keycloak/keycloak"
)

func resourceKeycloakRealmClientPolicyProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeycloakRealmClientPolicyProfileCreate,
		ReadContext:   resourceKeycloakRealmClientPolicyProfileRead,
		DeleteContext: resourceKeycloakRealmClientPolicyProfileDelete,
		UpdateContext: resourceKeycloakRealmClientPolicyProfileUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: resourceKeycloakRealmClientPolicyProfileImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"realm_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"executor": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"configuration": {
							Type:     schema.TypeMap,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceKeycloakRealmClientPolicyProfileImport(_ context.Context, data *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	err := importRealmClientPolicyResource(data, "realm-client-policy-profiles")
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{data}, nil
}

func importRealmClientPolicyResource(data *schema.ResourceData, resourcePath string) error {
	parts := strings.SplitN(data.Id(), "/", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] != resourcePath || parts[2] == "" {
		return fmt.Errorf("invalid import ID %q; expected {realm_id}/%s/{name}", data.Id(), resourcePath)
	}

	if err := data.Set("realm_id", parts[0]); err != nil {
		return fmt.Errorf("setting realm_id during import: %w", err)
	}
	if err := data.Set("name", parts[2]); err != nil {
		return fmt.Errorf("setting name during import: %w", err)
	}

	return nil
}

func resourceKeycloakRealmClientPolicyProfileUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	profile := mapFromDataToRealmClientPolicyProfile(data)
	realmId := profile.RealmId
	mutexKey := realmClientPolicyProfilesMutexKey(realmId)
	keycloakClient.Mutex.Lock(mutexKey)
	defer keycloakClient.Mutex.Unlock(mutexKey)
	realmClientPolicyProfiles, err := keycloakClient.GetAllRealmClientPolicyProfiles(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	for i, p := range realmClientPolicyProfiles.Profiles {
		if p.Name == profile.Name {
			realmClientPolicyProfiles.Profiles[i] = *profile
		}
	}

	err = keycloakClient.UpdateRealmClientPolicyProfiles(ctx, realmId, realmClientPolicyProfiles)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmClientPolicyProfileDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	slicedProfiles := []keycloak.RealmClientPolicyProfile{}
	profile := mapFromDataToRealmClientPolicyProfile(data)
	realmId := profile.RealmId
	mutexKey := realmClientPolicyProfilesMutexKey(realmId)
	keycloakClient.Mutex.Lock(mutexKey)
	defer keycloakClient.Mutex.Unlock(mutexKey)
	realmClientPolicyProfiles, err := keycloakClient.GetAllRealmClientPolicyProfiles(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, p := range realmClientPolicyProfiles.Profiles {
		if p.Name != profile.Name {
			slicedProfiles = append(slicedProfiles, p)
		}
	}

	realmClientPolicyProfiles.Profiles = slicedProfiles

	err = keycloakClient.UpdateRealmClientPolicyProfiles(ctx, realmId, realmClientPolicyProfiles)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceKeycloakRealmClientPolicyProfileCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	profile := mapFromDataToRealmClientPolicyProfile(data)

	realmId := profile.RealmId
	mutexKey := realmClientPolicyProfilesMutexKey(realmId)
	keycloakClient.Mutex.Lock(mutexKey)
	defer keycloakClient.Mutex.Unlock(mutexKey)
	name := profile.Name
	data.SetId(fmt.Sprintf("%s/realm-client-policy-profiles/%s", realmId, name))

	realmClientPolicyProfiles, err := keycloakClient.GetAllRealmClientPolicyProfiles(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	realmClientPolicyProfiles.Profiles = append(realmClientPolicyProfiles.Profiles, *profile)

	err = keycloakClient.UpdateRealmClientPolicyProfiles(ctx, realmId, realmClientPolicyProfiles)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceKeycloakRealmClientPolicyProfileRead(ctx, data, meta)
}

// All profile mutations share one key because Keycloak replaces the full
// profile collection on every create, update, and delete operation.
func realmClientPolicyProfilesMutexKey(realmId string) string {
	return fmt.Sprintf("resourceKeycloakRealmClientPolicyProfiles:%s", realmId)
}

func resourceKeycloakRealmClientPolicyProfileRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	keycloakClient := meta.(*keycloak.KeycloakClient)
	realmId := data.Get("realm_id").(string)
	name := data.Get("name").(string)
	realmClientPolicyProfiles, err := keycloakClient.GetAllRealmClientPolicyProfiles(ctx, realmId)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, profile := range realmClientPolicyProfiles.Profiles {
		if profile.Name == name {
			profile.RealmId = realmId
			err = mapFromRealmClientPolicyProfileToData(data, &profile)
			if err != nil {
				return diag.FromErr(err)
			}
			return nil
		}
	}

	data.SetId("")
	return nil
}

func mapFromDataToRealmClientPolicyProfile(data *schema.ResourceData) *keycloak.RealmClientPolicyProfile {
	executors := []keycloak.RealmClientPolicyProfileExecutor{}

	for _, executor := range data.Get("executor").([]interface{}) {
		executorMap := executor.(map[string]interface{})

		exec := keycloak.RealmClientPolicyProfileExecutor{
			Name: executorMap["name"].(string),
		}

		if v, ok := executorMap["configuration"]; ok {
			configurations := make(map[string]interface{})
			for key, value := range v.(map[string]interface{}) {
				// handle json objects and arrays
				if strings.HasPrefix(value.(string), "{") || strings.HasPrefix(value.(string), "[") {
					var t interface{}
					json.Unmarshal([]byte(value.(string)), &t)
					configurations[key] = t
					continue
				}
				configurations[key] = value
			}
			exec.Configuration = configurations
		}

		executors = append(executors, exec)
	}

	return &keycloak.RealmClientPolicyProfile{
		Name:        data.Get("name").(string),
		RealmId:     data.Get("realm_id").(string),
		Description: data.Get("description").(string),
		Executors:   executors,
	}
}

func mapFromRealmClientPolicyProfileToData(data *schema.ResourceData, profile *keycloak.RealmClientPolicyProfile) error {
	if err := data.Set("name", profile.Name); err != nil {
		return err
	}
	if err := data.Set("realm_id", profile.RealmId); err != nil {
		return err
	}
	if err := data.Set("description", profile.Description); err != nil {
		return err
	}

	executors := make([]interface{}, 0)
	for _, ex := range profile.Executors {

		executorMap := map[string]interface{}{
			"name": ex.Name,
		}

		if ex.Configuration != nil {
			configurations, err := flattenRealmClientPolicyConfiguration(ex.Configuration)
			if err != nil {
				return err
			}
			executorMap["configuration"] = configurations
		}
		executors = append(executors, executorMap)
	}

	if err := data.Set("executor", executors); err != nil {
		return err
	}

	return nil
}

func flattenRealmClientPolicyConfiguration(configuration map[string]interface{}) (map[string]interface{}, error) {
	flattened := make(map[string]interface{}, len(configuration))
	for key, value := range configuration {
		switch value := value.(type) {
		case map[string]interface{}, []interface{}:
			jsonValue, err := json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("marshalling client policy configuration %q: %w", key, err)
			}
			flattened[key] = string(jsonValue)
		case string:
			flattened[key] = value
		case nil:
			flattened[key] = ""
		default:
			// TypeMap values are strings in Terraform state. Keycloak can return
			// primitive JSON values for resources created outside Terraform.
			flattened[key] = fmt.Sprint(value)
		}
	}

	return flattened, nil
}
