package batch

// BatchUserRoleClient is the API Client for BatchUserRoleClient
// for testing.
type BatchUserRoleClient interface {
	ListUsers() (*[]string, error)
	ListRolesForUser(name string) (*[]string, error)
	ListAllRoles() (*[]string, error)
	UpdatePassword(name string, password string) error
	CreateUser(name string, password string) error
	CreateRole(name string) error
	AssignRolesToUser(name string, roles ...string) error
	RemoveRolesFromUser(name string, roles ...string) error
	DeleteUsers(name ...string) error
	DeleteFiles(file ...string) error
}

type batchUserRoleClientForValidation struct {
	client BatchUserRoleClient
}

func (t *batchUserRoleClientForValidation) ListUsers() (*[]string, error) {
	return t.client.ListUsers()
}
func (t *batchUserRoleClientForValidation) ListRolesForUser(name string) (*[]string, error) {
	return t.client.ListRolesForUser(name)
}
func (t *batchUserRoleClientForValidation) ListAllRoles() (*[]string, error) {
	return t.client.ListAllRoles()
}
func (t *batchUserRoleClientForValidation) UpdatePassword(name string, password string) error {
	return nil
}
func (t *batchUserRoleClientForValidation) CreateUser(name string, password string) error {
	return nil
}
func (t *batchUserRoleClientForValidation) CreateRole(name string) error {
	return nil
}
func (t *batchUserRoleClientForValidation) AssignRolesToUser(name string, roles ...string) error {
	return nil
}
func (t *batchUserRoleClientForValidation) RemoveRolesFromUser(name string, roles ...string) error {
	return nil
}
func (t *batchUserRoleClientForValidation) DeleteUsers(name ...string) error {
	return nil
}
func (t *batchUserRoleClientForValidation) DeleteFiles(file ...string) error {
	return nil
}
func NewBatchUserRoleClientForValidation(client BatchUserRoleClient) BatchUserRoleClient {
	return &batchUserRoleClientForValidation{client}
}
