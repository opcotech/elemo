package model

import "errors"

var (
	ErrInvalidAssignmentDetails     = errors.New("invalid assignment details")              // the assignment details are invalid
	ErrInvalidAssignmentKind        = errors.New("invalid assigned to kind")                // the assigned to kind is invalid
	ErrInvalidAttachmentDetails     = errors.New("invalid attachment details")              // the attachment details are invalid
	ErrInvalidCommentDetails        = errors.New("invalid comment details")                 // the comment details are invalid
	ErrInvalidDocumentDetails       = errors.New("invalid document details")                // the document details are invalid
	ErrInvalidHealthStatus          = errors.New("invalid health status")                   // health status is invalid
	ErrInvalidID                    = errors.New("invalid id")                              // the id is invalid
	ErrInvalidIssueDetails          = errors.New("invalid issue details")                   // the issue details are invalid
	ErrInvalidIssueKind             = errors.New("invalid issue kind")                      // the issue kind is invalid
	ErrInvalidIssuePriority         = errors.New("invalid issue priority")                  // the issue priority is invalid
	ErrInvalidIssueRelationDetails  = errors.New("invalid issue relation details")          // the issue relation details are invalid
	ErrInvalidIssueRelationKind     = errors.New("invalid issue relation kind")             // the issue relation kind is invalid
	ErrInvalidIssueResolution       = errors.New("invalid issue resolution")                // the issue resolution is invalid
	ErrInvalidIssueStatus           = errors.New("invalid issue status")                    // the issue status is invalid
	ErrInvalidLabelDetails          = errors.New("invalid Type details")                    // the Type details are invalid
	ErrInvalidLanguage              = errors.New("invalid language code")                   // Language is not valid
	ErrInvalidNamespaceDetails      = errors.New("invalid namespace details")               // the namespace details are invalid
	ErrInvalidNotificationDetails   = errors.New("invalid notification details")            // the notification details are invalid
	ErrInvalidNotificationRecipient = errors.New("invalid notification recipient")          // the notification recipient is invalid
	ErrInvalidOrganizationDetails   = errors.New("invalid organization details")            // the organization details are invalid
	ErrInvalidOrganizationStatus    = errors.New("invalid organization status")             // the organization status is invalid
	ErrInvalidPermissionDetails     = errors.New("invalid permission details")              // the permission details are invalid
	ErrInvalidPermissionKind        = errors.New("invalid permission kind")                 // the permission kind is invalid
	ErrInvalidProjectDetails        = errors.New("invalid project details")                 // the project details are invalid
	ErrInvalidProjectStatus         = errors.New("invalid project status")                  // the project status is invalid
	ErrInvalidResourceType          = errors.New("invalid resource type")                   // the resource type is invalid
	ErrInvalidRoleDetails           = errors.New("invalid role details")                    // the role details are invalid
	ErrInvalidSystemRole            = errors.New("invalid system role")                     // the system role is invalid
	ErrInvalidTodoDetails           = errors.New("invalid todo details")                    // the todo details are invalid
	ErrInvalidTodoPriority          = errors.New("invalid todo priority")                   // the todo priority is invalid
	ErrInvalidUserDetails           = errors.New("invalid user details")                    // the user details are invalid
	ErrInvalidUserStatus            = errors.New("invalid user status")                     // the user status is invalid
	ErrPermissionSubjectTargetEqual = errors.New("permission subject and target are equal") // the permission subject and target are equal
)
