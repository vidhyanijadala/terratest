package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/recoveryservices/mgmt/2016-06-01/recoveryservices"
	"github.com/Azure/azure-sdk-for-go/services/recoveryservices/mgmt/2020-02-02/backup"
)

//RecoveryServicesVaultExists indicates whether a recovery services vault exists; otherwise false.
func RecoveryServicesVaultExists(vaultName, resourceGroupName, subscriptionID string) bool {
	vault, err := GetRecoveryServicesVaultE(vaultName, resourceGroupName, subscriptionID)
	if err != nil {
		return false
	}
	return (*vault.Name == vaultName)
}

// GetRecoveryServicesVaultBackupPolicyList returns a list of backup policies for the given vault.
func GetRecoveryServicesVaultBackupPolicyList(vaultName, resourceGroupName, subscriptionID string) map[string]backup.ProtectionPolicyResource {
	list, err := GetRecoveryServicesVaultBackupPolicyListE(vaultName, resourceGroupName, subscriptionID)
	if err != nil {
		return nil
	}
	return list
}

// GetRecoveryServicesVaultBackupProtectedVMList returns a list of protected VM's on the given vault/policy.
func GetRecoveryServicesVaultBackupProtectedVMList(policyName, vaultName, resourceGroupName, subscriptionID string) map[string]backup.AzureIaaSComputeVMProtectedItem {
	list, err := GetRecoveryServicesVaultBackupProtectedVMListE(policyName, vaultName, resourceGroupName, subscriptionID)
	if err != nil {
		return nil
	}
	return list
}

// GetRecoveryServicesVaultE returns a vault instance.
// This function would fail the test if there is an error.
func GetRecoveryServicesVaultE(vaultName, resourceGroupName, subscriptionID string) (*recoveryservices.Vault, error) {
	subscriptionID, err := getTargetAzureSubscription(subscriptionID)
	if err != nil {
		return nil, err
	}
	resourceGroupName, err2 := getTargetAzureResourceGroupName((resourceGroupName))
	if err2 != nil {
		return nil, err2
	}
	client := recoveryservices.NewVaultsClient(subscriptionID)
	// setup auth and create request params
	authorizer, err := NewAuthorizer()
	if err != nil {
		return nil, err
	}
	client.Authorizer = *authorizer
	vault, err := client.Get(context.Background(), resourceGroupName, vaultName)
	if err != nil {
		return nil, err
	}
	return &vault, nil
}

// GetRecoveryServicesVaultBackupPolicyListE returns a list of backup policies for the given vault.
// This function would fail the test if there is an error.
func GetRecoveryServicesVaultBackupPolicyListE(vaultName, resourceGroupName, subscriptionID string) (map[string]backup.ProtectionPolicyResource, error) {
	subscriptionID, err := getTargetAzureSubscription(subscriptionID)
	if err != nil {
		return nil, err
	}
	resourceGroupName, err2 := getTargetAzureResourceGroupName(resourceGroupName)
	if err2 != nil {
		return nil, err2
	}
	client := backup.NewPoliciesClient(subscriptionID)
	// setup authorizer
	authorizer, err := NewAuthorizer()
	if err != nil {
		return nil, err
	}
	client.Authorizer = *authorizer
	listIter, err := client.ListComplete(context.Background(), vaultName, resourceGroupName, "")
	if err != nil {
		return nil, err
	}
	policyMap := make(map[string]backup.ProtectionPolicyResource)
	for listIter.NotDone() {
		v := listIter.Value()
		policyMap[*v.Name] = v
		err := listIter.NextWithContext(context.Background())
		if err != nil {
			return nil, err
		}

	}
	return policyMap, nil
}

// GetRecoveryServicesVaultBackupProtectedVMListE returns a list of protected VM's on the given vault/policy.
// This function would fail the test if there is an error.
func GetRecoveryServicesVaultBackupProtectedVMListE(policyName, vaultName, resourceGroupName, subscriptionID string) (map[string]backup.AzureIaaSComputeVMProtectedItem, error) {
	subscriptionID, err := getTargetAzureSubscription(subscriptionID)
	if err != nil {
		return nil, err
	}
	resourceGroupName, err2 := getTargetAzureResourceGroupName(resourceGroupName)
	if err != nil {
		return nil, err2
	}
	client := backup.NewProtectedItemsGroupClient(subscriptionID)
	// setup authorizer
	authorizer, err := NewAuthorizer()
	if err != nil {
		return nil, err
	}
	client.Authorizer = *authorizer
	// Build a filter string to narrow down results to just VM's
	filter := fmt.Sprintf("backupManagementType eq 'AzureIaasVM' and itemType eq 'VM' and policyName eq '%s'", policyName)
	listIter, err := client.ListComplete(context.Background(), vaultName, resourceGroupName, filter, "")
	if err != nil {
		return nil, err
	}
	// Prep the return container
	vmList := make(map[string]backup.AzureIaaSComputeVMProtectedItem)
	// First iterator check
	for listIter.NotDone() {
		currentVM, _ := listIter.Value().Properties.AsAzureIaaSComputeVMProtectedItem()
		vmList[*currentVM.FriendlyName] = *currentVM
		err := listIter.NextWithContext(context.Background())
		if err != nil {
			return nil, err
		}
	}
	return vmList, nil
}
