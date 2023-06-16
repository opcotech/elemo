package service

import "errors"

var (
	ErrEmailSend                       = errors.New("failed to send email")                           // failed to send email
	ErrInvalidEmail                    = errors.New("invalid email address")                          // invalid email address
	ErrInvalidPaginationParams         = errors.New("invalid pagination parameters")                  // invalid pagination parameters
	ErrLicenseGet                      = errors.New("failed to get license")                          // failed to get license
	ErrLicensePing                     = errors.New("failed to ping license")                         // failed to ping license
	ErrNoLicenseService                = errors.New("no license service provided")                    // no license service provided
	ErrNoOrganizationRepository        = errors.New("no organization repository provided")            // no organization repository provided
	ErrNoPatchData                     = errors.New("no patch data provided")                         // no patch data provided
	ErrNoPermission                    = errors.New("no permission")                                  // no permission
	ErrNoPermissionRepository          = errors.New("no permission repository provided")              // no permission repository provided
	ErrNoPermissionService             = errors.New("no permission service provided")                 // no permission service provided
	ErrNoResources                     = errors.New("no resources provided")                          // no resources provided
	ErrNoRoleRepository                = errors.New("no role repository provided")                    // no role repository provided
	ErrNoTodoRepository                = errors.New("no todo repository provided")                    // no todo repository provided
	ErrNoUser                          = errors.New("no user provided")                               // no user provided
	ErrNoUserRepository                = errors.New("no user repository provided")                    // no user repository provided
	ErrNoVersionInfo                   = errors.New("no version info provided")                       // no version info provided
	ErrOrganizationCreate              = errors.New("failed to create organization")                  // failed to create organization
	ErrOrganizationDelete              = errors.New("failed to delete organization")                  // failed to delete organization
	ErrOrganizationGet                 = errors.New("failed to get organization")                     // failed to get organization
	ErrOrganizationGetAll              = errors.New("failed to get organizations")                    // failed to get organizations
	ErrOrganizationMemberAdd           = errors.New("failed to add member to organization")           // failed to add member to organization
	ErrOrganizationMemberRemove        = errors.New("failed to remove member from organization")      // failed to remove member from organization
	ErrOrganizationMembersGet          = errors.New("failed to get members of organization")          // failed to get members of organization
	ErrOrganizationUpdate              = errors.New("failed to update organization")                  // failed to update organization
	ErrPermissionCreate                = errors.New("failed to create permission")                    // failed to create permission
	ErrPermissionDelete                = errors.New("failed to delete permission")                    // failed to delete permission
	ErrPermissionGet                   = errors.New("failed to get permission")                       // failed to get permission
	ErrPermissionGetBySubject          = errors.New("failed to get permission by subject")            // failed to get permission by subject
	ErrPermissionGetBySubjectAndTarget = errors.New("failed to get permission by subject and target") // failed to get permission by subject and target
	ErrPermissionGetByTarget           = errors.New("failed to get permission by target")             // failed to get permission by target
	ErrPermissionHasAnyRelation        = errors.New("failed to check if permission has any relation") // failed to check if permission has any relation
	ErrPermissionHasPermission         = errors.New("failed to check if permission has permission")   // failed to check if permission has permission
	ErrPermissionHasSystemRole         = errors.New("failed to check if permission has system role")  // failed to check if permission has system role
	ErrPermissionUpdate                = errors.New("failed to update permission")                    // failed to update permission
	ErrQuotaExceeded                   = errors.New("quota exceeded")                                 // quota exceeded
	ErrQuotaInvalid                    = errors.New("invalid quota")                                  // invalid quota
	ErrQuotaUsageGet                   = errors.New("failed to get usage of quota")                   // failed to get usage of quota
	ErrRoleAddMember                   = errors.New("failed to add member to role")                   // failed to add member to role
	ErrRoleCreate                      = errors.New("failed to create role")                          // failed to create role
	ErrRoleDelete                      = errors.New("failed to delete role")                          // failed to delete role
	ErrRoleGet                         = errors.New("failed to get role")                             // failed to get role
	ErrRoleGetBelongsTo                = errors.New("failed to get roles that belongs to")            // failed to get roles that belongs to
	ErrRoleRemoveMember                = errors.New("failed to remove member from role")              // failed to remove member from role
	ErrRoleUpdate                      = errors.New("failed to update role")                          // failed to update role
	ErrSystemHealthCheck               = errors.New("system health check failed")                     // system health check failed
	ErrTodoCreate                      = errors.New("failed to create todo")                          // failed to create todo
	ErrTodoDelete                      = errors.New("failed to delete todo")                          // failed to delete todo
	ErrTodoGet                         = errors.New("failed to get todo")                             // failed to get todo
	ErrTodoGetAll                      = errors.New("failed to get todos")                            // failed to get todos
	ErrTodoUpdate                      = errors.New("failed to update todo")                          // failed to update todo
	ErrUserCreate                      = errors.New("failed to create user")                          // failed to create user
	ErrUserDelete                      = errors.New("failed to delete user")                          // failed to delete user
	ErrUserGet                         = errors.New("failed to get user")                             // failed to get user
	ErrUserGetAll                      = errors.New("failed to get users")                            // failed to get users
	ErrUserUpdate                      = errors.New("failed to update user")                          // failed to update user
)
