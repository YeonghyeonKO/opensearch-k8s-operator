package services

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/opensearch-project/opensearch-go/opensearchutil"
	"opensearch.opster.io/opensearch-gateway/requests"
	"opensearch.opster.io/opensearch-gateway/responses"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	K8sAttributeField = "k8s-uid"

	ROLES         = "roles"
	INTERNALUSERS = "internalusers"
	ROLESMAPPING  = "rolesmapping"
	ACTIONGROUPS  = "actiongroups"
)

func ShouldUpdateUser(
	ctx context.Context,
	service *OsClusterClient,
	username string,
	user requests.User,
) (bool, error) {
	resp, err := service.GetSecurityResource(ctx, INTERNALUSERS, username)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return true, nil
	} else if resp.IsError() {
		return false, fmt.Errorf("response from API is %s", resp.Status())
	}

	// Blank the password for a fair comparison
	user.Password = ""

	userResponse := responses.GetUserResponse{}

	err = json.NewDecoder(resp.Body).Decode(&userResponse)
	if err != nil {
		return false, err
	}

	existingUID, ok := userResponse[username].Attributes[K8sAttributeField]
	if !ok {
		return false, fmt.Errorf("user resource not currently managed by kubernetes")
	}

	if existingUID != user.Attributes[K8sAttributeField] {
		return false, fmt.Errorf("kubernetes resource conflict; uids don't match")
	}

	if reflect.DeepEqual(user, userResponse[username]) {
		return false, nil
	}

	lg := log.FromContext(ctx).WithValues("os_service", "security")
	lg.Info("user requires update")
	return true, nil
}

func UserExists(ctx context.Context, service *OsClusterClient, username string) (bool, error) {
	resp, err := service.GetSecurityResource(ctx, INTERNALUSERS, username)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return false, nil
	} else if resp.IsError() {
		return false, fmt.Errorf("response from API is %s", resp.Status())
	}
	return true, nil
}

func UserUIDMatches(ctx context.Context, service *OsClusterClient, username string, uid string) (bool, error) {
	resp, err := service.GetSecurityResource(ctx, INTERNALUSERS, username)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return false, fmt.Errorf("response from API is %s", resp.Status())
	}

	userResponse := responses.GetUserResponse{}

	err = json.NewDecoder(resp.Body).Decode(&userResponse)
	if err != nil {
		return false, err
	}

	existingUID, ok := userResponse[username].Attributes[K8sAttributeField]

	return ok && existingUID == uid, nil
}

func CreateOrUpdateUser(
	ctx context.Context,
	service *OsClusterClient,
	username string,
	user requests.User,
) error {
	resp, err := service.PutSecurityResource(ctx, INTERNALUSERS, username, opensearchutil.NewJSONReader(user))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return fmt.Errorf("failed to create user: %s", resp.String())
	}
	return nil
}

func DeleteUser(ctx context.Context, service *OsClusterClient, username string) error {
	resp, err := service.DeleteSecurityResource(ctx, INTERNALUSERS, username)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("response from API is %s", resp.Status())
	}
	return nil
}

func RoleExists(ctx context.Context, service *OsClusterClient, rolename string) (bool, error) {
	resp, err := service.GetSecurityResource(ctx, ROLES, rolename)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return false, nil
	} else if resp.IsError() {
		return false, fmt.Errorf("response from API is %s", resp.Status())
	}
	return true, nil
}

func ShouldUpdateRole(
	ctx context.Context,
	service *OsClusterClient,
	rolename string,
	role requests.Role,
) (bool, error) {
	resp, err := service.GetSecurityResource(ctx, ROLES, rolename)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return true, nil
	} else if resp.IsError() {
		return false, fmt.Errorf("response from API is %s", resp.Status())
	}

	roleResponse := responses.GetRoleResponse{}

	err = json.NewDecoder(resp.Body).Decode(&roleResponse)
	if err != nil {
		return false, err
	}

	if reflect.DeepEqual(role, roleResponse[rolename]) {
		return false, nil
	}

	lg := log.FromContext(ctx).WithValues("os_service", "security")
	lg.V(1).Info(fmt.Sprintf("exsiting role: %+v", roleResponse[rolename]))
	lg.V(1).Info(fmt.Sprintf("new role: %+v", role))
	lg.Info("role requires update")
	return true, nil
}

func CreateOrUpdateRole(
	ctx context.Context,
	service *OsClusterClient,
	rolename string,
	role requests.Role,
) error {
	resp, err := service.PutSecurityResource(ctx, ROLES, rolename, opensearchutil.NewJSONReader(role))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return fmt.Errorf("failed to create role: %s", resp.String())
	}
	return nil
}

func DeleteRole(ctx context.Context, service *OsClusterClient, rolename string) error {
	resp, err := service.DeleteSecurityResource(ctx, ROLES, rolename)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("response from API is %s", resp.Status())
	}
	return nil
}

func RoleMappingExists(
	ctx context.Context,
	service *OsClusterClient,
	rolename string,
) (bool, error) {
	resp, err := service.GetSecurityResource(ctx, ROLESMAPPING, rolename)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return false, nil
	} else if resp.IsError() {
		return false, fmt.Errorf("response from API is %s", resp.Status())
	}
	return true, nil
}

func FetchExistingRoleMapping(
	ctx context.Context,
	service *OsClusterClient,
	rolename string,
) (requests.RoleMapping, error) {
	resp, err := service.GetSecurityResource(ctx, ROLESMAPPING, rolename)
	if err != nil {
		return requests.RoleMapping{}, err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return requests.RoleMapping{}, fmt.Errorf("response from API is %s", resp.Status())
	}

	mappingResp := responses.GetRoleMappingReponse{}
	err = json.NewDecoder(resp.Body).Decode(&mappingResp)
	if err != nil {
		return requests.RoleMapping{}, err
	}

	return mappingResp[rolename], nil
}

func CreateOrUpdateRoleMapping(
	ctx context.Context,
	service *OsClusterClient,
	rolename string,
	mapping requests.RoleMapping,
) error {
	resp, err := service.PutSecurityResource(ctx, ROLESMAPPING, rolename, opensearchutil.NewJSONReader(mapping))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return fmt.Errorf("failed to create role mapping: %s", resp.String())
	}
	return nil
}

func DeleteRoleMapping(ctx context.Context, service *OsClusterClient, rolename string) error {
	resp, err := service.DeleteSecurityResource(ctx, ROLESMAPPING, rolename)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("response from API is %s", resp.Status())
	}
	return nil
}

// ActionGroupExists checks if the passed actionGroup already exists or not
func ActionGroupExists(ctx context.Context, service *OsClusterClient, actionGroupName string) (bool, error) {
	resp, err := service.GetSecurityResource(ctx, ACTIONGROUPS, actionGroupName)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return false, nil
	} else if resp.IsError() {
		return false, fmt.Errorf("response from API is %s", resp.Status())
	}
	return true, nil
}

// ShouldUpdateActionGroup checks whether a previously created actiongroup needs an update or not
func ShouldUpdateActionGroup(
	ctx context.Context,
	service *OsClusterClient,
	actionGroupName string,
	actionGroup requests.ActionGroup,
) (bool, error) {
	resp, err := service.GetSecurityResource(ctx, ACTIONGROUPS, actionGroupName)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return true, nil
	} else if resp.IsError() {
		return false, fmt.Errorf("response from API is %s", resp.Status())
	}

	actionGroupResponse := responses.GetActionGroupResponse{}

	err = json.NewDecoder(resp.Body).Decode(&actionGroupResponse)
	if err != nil {
		return false, err
	}

	if reflect.DeepEqual(actionGroup, actionGroupResponse[actionGroupName]) {
		return false, nil
	}

	lg := log.FromContext(ctx).WithValues("os_service", "security")
	lg.V(1).Info(fmt.Sprintf("exsiting actiongroup: %+v", actionGroupResponse[actionGroupName]))
	lg.V(1).Info(fmt.Sprintf("new actiongroup: %+v", actionGroup))
	lg.Info("actiongroup requires update")
	return true, nil
}

// CreateOrUpdateActionGroup creates a new action group or updates a previously created action group
func CreateOrUpdateActionGroup(
	ctx context.Context,
	service *OsClusterClient,
	actionGroupName string,
	actionGroup requests.ActionGroup,
) error {
	resp, err := service.PutSecurityResource(ctx, ACTIONGROUPS, actionGroupName, opensearchutil.NewJSONReader(actionGroup))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return fmt.Errorf("failed to create actiongroup: %s", resp.String())
	}
	return nil
}

// DeleteActionGroup deletes a previously created action group
func DeleteActionGroup(ctx context.Context, service *OsClusterClient, actionGroupName string) error {
	resp, err := service.DeleteSecurityResource(ctx, ACTIONGROUPS, actionGroupName)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return fmt.Errorf("response from API is %s", resp.Status())
	}
	return nil
}
