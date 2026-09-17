package provider

import "testing"

func TestRealmClientPolicyMutexKeysAreScopedByCollectionAndRealm(t *testing.T) {
	profileRealmOne := realmClientPolicyProfilesMutexKey("realm-one")
	profileRealmTwo := realmClientPolicyProfilesMutexKey("realm-two")
	policyRealmOne := realmClientPolicyPoliciesMutexKey("realm-one")

	if profileRealmOne == profileRealmTwo {
		t.Fatal("profile mutations in different realms must not share a mutex key")
	}
	if profileRealmOne == policyRealmOne {
		t.Fatal("profile and policy collections must not share a mutex key")
	}
}
