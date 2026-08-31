package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/opcotech/elemo/internal/pkg/validate"
)

const (
	PluginStatusUnknown   PluginStatus = iota // unknown
	PluginStatusInstalled                     // installed
	PluginStatusStarting                      // starting
	PluginStatusActive                        // active
	PluginStatusDisabling                     // disabling
	PluginStatusDisabled                      // disabled
	PluginStatusFailed                        // failed
)

const (
	PluginGraphPropertyTypeUnknown  PluginGraphPropertyType = iota // unknown
	PluginGraphPropertyTypeStr                                     // string
	PluginGraphPropertyTypeInteger                                 // integer
	PluginGraphPropertyTypeDecimal                                 // decimal
	PluginGraphPropertyTypeBoolean                                 // boolean
	PluginGraphPropertyTypeDate                                    // date
	PluginGraphPropertyTypeDateTime                                // datetime
)

const (
	PluginGraphCardinalityUnknown    PluginGraphCardinality = iota // unknown
	PluginGraphCardinalityOneToOne                                 // one-to-one
	PluginGraphCardinalityOneToMany                                // one-to-many
	PluginGraphCardinalityManyToOne                                // many-to-one
	PluginGraphCardinalityManyToMany                               // many-to-many
)

const (
	PluginConfigFieldTypeUnknown      PluginConfigFieldType = iota // unknown
	PluginConfigFieldTypeStr                                       // string
	PluginConfigFieldTypeInteger                                   // integer
	PluginConfigFieldTypeBoolean                                   // boolean
	PluginConfigFieldTypeGraphBinding                              // graph_binding
)

const (
	PluginGraphRelationDirectionUnknown  PluginGraphRelationDirection = iota // unknown
	PluginGraphRelationDirectionOutgoing                                     // outgoing
	PluginGraphRelationDirectionIncoming                                     // incoming
	PluginGraphRelationDirectionBoth                                         // both
)

const (
	PluginSchemaVersionV1      = 1
	PluginAPIVersionV1         = "1"
	PluginManifestFileName     = "plugin.yaml"
	PluginBackendWASMPath      = "backend/plugin.wasm"
	PluginFrontendEntryDefault = "frontend/index.js"
)

const (
	pluginIDMinLen      = 3
	pluginIDMaxLen      = 128
	pluginNameMinLen    = 2
	pluginNameMaxLen    = 120
	pluginVersionMaxLen = 64
	pluginKindMaxLen    = 64
)

var (
	pluginIDPattern   = regexp.MustCompile(`^[a-z][a-z0-9]*(\.[a-z0-9]+)+$`)
	pluginKindPattern = regexp.MustCompile(`^[A-Z][A-Za-z0-9_]*$`)
	pluginPropPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

var reservedGraphPropertyNames = map[string]struct{}{
	"id":         {},
	"plugin_id":  {},
	"kind":       {},
	"created_at": {},
	"updated_at": {},
}

var coreGraphScopeTypes = map[string]ResourceType{
	"Organization": ResourceTypeOrganization,
	"Namespace":    ResourceTypeNamespace,
	"Project":      ResourceTypeProject,
	"Issue":        ResourceTypeIssue,
	"Document":     ResourceTypeDocument,
	"Folder":       ResourceTypeFolder,
	"User":         ResourceTypeUser,
	"Team":         ResourceTypeTeam,
}

var coreEdgeKindNames = map[string]struct{}{
	"ASSIGNED_TO":    {},
	"BELONGS_TO":     {},
	"COMMENTED":      {},
	"CREATED":        {},
	"HAS_ATTACHMENT": {},
	"HAS_COMMENT":    {},
	"HAS_LABEL":      {},
	"HAS_NAMESPACE":  {},
	"HAS_PERMISSION": {},
	"HAS_PROJECT":    {},
	"HAS_TEAM":       {},
	"INVITED":        {},
	"INVITED_TO":     {},
	"KIND_OF":        {},
	"MEMBER_OF":      {},
	"RELATED_TO":     {},
	"SPEAKS":         {},
	"WATCHES":        {},
	"SCOPED_TO":      {},
	"LOCATED_IN":     {},
	"IN_SCOPE_OF":    {},
	"GRANTED":        {},
	"DEFINES_ROLE":   {},
}

// PluginStatus is the lifecycle state of an installation or runtime instance.
//
//go:generate go tool enumer -type=PluginStatus -text -sql -transform=noop -linecomment -output=plugin_status_gen.go
type PluginStatus uint8

// PluginGraphPropertyType is a closed set of scalar property types.
//
//go:generate go tool enumer -type=PluginGraphPropertyType -text -transform=noop -linecomment -output=plugin_graph_property_type_gen.go
type PluginGraphPropertyType uint8

// PluginGraphCardinality is a closed set of relation cardinalities.
//
//go:generate go tool enumer -type=PluginGraphCardinality -text -transform=noop -linecomment -output=plugin_graph_cardinality_gen.go
type PluginGraphCardinality uint8

// PluginConfigFieldType is a closed set of activation config field types.
//
//go:generate go tool enumer -type=PluginConfigFieldType -text -transform=noop -linecomment -output=plugin_config_field_type_gen.go
type PluginConfigFieldType uint8

// PluginGraphRelationDirection is the list direction for plugin domain edges.
//
//go:generate go tool enumer -type=PluginGraphRelationDirection -text -transform=noop -linecomment -output=plugin_graph_relation_direction_gen.go
type PluginGraphRelationDirection uint8

// PluginCapability is a closed host-API capability string.
type PluginCapability string

const (
	CapabilityIssuesRead         PluginCapability = "issues.read"
	CapabilityIssuesUpdate       PluginCapability = "issues.update"
	CapabilityProjectsRead       PluginCapability = "projects.read"
	CapabilityUsersRead          PluginCapability = "users.read"
	CapabilityPermissionsCheck   PluginCapability = "permissions.check"
	CapabilityPluginStorageRead  PluginCapability = "plugin.storage.read"
	CapabilityPluginStorageWrite PluginCapability = "plugin.storage.write"
	CapabilityGraphRead          PluginCapability = "graph.read"
	CapabilityGraphWrite         PluginCapability = "graph.write"
	CapabilityEventsPublish      PluginCapability = "events.publish"
)

// PluginCapabilities is the closed set of host-API capabilities.
var PluginCapabilities = []PluginCapability{
	CapabilityIssuesRead,
	CapabilityIssuesUpdate,
	CapabilityProjectsRead,
	CapabilityUsersRead,
	CapabilityPermissionsCheck,
	CapabilityPluginStorageRead,
	CapabilityPluginStorageWrite,
	CapabilityGraphRead,
	CapabilityGraphWrite,
	CapabilityEventsPublish,
}

var pluginCapabilitySet = func() map[PluginCapability]struct{} {
	out := make(map[PluginCapability]struct{}, len(PluginCapabilities))
	for _, capability := range PluginCapabilities {
		out[capability] = struct{}{}
	}
	return out
}()

// Valid reports whether c is a known capability.
func (c PluginCapability) Valid() bool {
	_, ok := pluginCapabilitySet[c]
	return ok
}

func (c PluginCapability) String() string {
	return string(c)
}

// PluginUISlot is a host UI contribution point.
type PluginUISlot string

const (
	PluginSlotIssueSidebar         PluginUISlot = "issue.sidebar"
	PluginSlotIssueActions         PluginUISlot = "issue.actions"
	PluginSlotIssueActivity        PluginUISlot = "issue.activity"
	PluginSlotOrganizationSettings PluginUISlot = "organization.settings"
	PluginSlotProjectSettings      PluginUISlot = "project.settings"
	PluginSlotProjectSidebar       PluginUISlot = "project.sidebar"
)

var pluginSlotSet = map[PluginUISlot]struct{}{
	PluginSlotIssueSidebar:         {},
	PluginSlotIssueActions:         {},
	PluginSlotIssueActivity:        {},
	PluginSlotOrganizationSettings: {},
	PluginSlotProjectSettings:      {},
	PluginSlotProjectSidebar:       {},
}

func (s PluginUISlot) Valid() bool {
	_, ok := pluginSlotSet[s]
	return ok
}

func (s PluginUISlot) String() string {
	return string(s)
}

// PluginEventType is a domain event plugins may subscribe to.
type PluginEventType string

const (
	PluginEventIssueCreated     PluginEventType = "issue.created"
	PluginEventIssueUpdated     PluginEventType = "issue.updated"
	PluginEventIssueDeleted     PluginEventType = "issue.deleted"
	PluginEventProjectCreated   PluginEventType = "project.created"
	PluginEventProjectUpdated   PluginEventType = "project.updated"
	PluginEventExtensionCreated PluginEventType = "extension.created"
	PluginEventExtensionUpdated PluginEventType = "extension.updated"
	PluginEventExtensionDeleted PluginEventType = "extension.deleted"
)

var pluginEventSet = map[PluginEventType]struct{}{
	PluginEventIssueCreated:     {},
	PluginEventIssueUpdated:     {},
	PluginEventIssueDeleted:     {},
	PluginEventProjectCreated:   {},
	PluginEventProjectUpdated:   {},
	PluginEventExtensionCreated: {},
	PluginEventExtensionUpdated: {},
	PluginEventExtensionDeleted: {},
}

func (e PluginEventType) Valid() bool {
	_, ok := pluginEventSet[e]
	return ok
}

func (e PluginEventType) String() string {
	return string(e)
}

// PluginGraphPropertyDecl is a declared scalar on a plugin node kind.
type PluginGraphPropertyDecl struct {
	Name     string                  `json:"name" yaml:"name"`
	Type     PluginGraphPropertyType `json:"type" yaml:"type"`
	Required bool                    `json:"required" yaml:"required"`
}

// PluginGraphNodeScope names the IN_SCOPE_OF parent for a node kind.
type PluginGraphNodeScope struct {
	Parent string `json:"parent" yaml:"parent"`
}

// PluginGraphNodeDecl is a plugin-defined graph node kind.
type PluginGraphNodeDecl struct {
	Kind       string                    `json:"kind" yaml:"kind"`
	Scope      PluginGraphNodeScope      `json:"scope" yaml:"scope"`
	Properties []PluginGraphPropertyDecl `json:"properties" yaml:"properties"`
}

// PluginGraphForeignDecl is a cross-plugin kind alias resolved at activation.
type PluginGraphForeignDecl struct {
	Name       string                    `json:"name" yaml:"name"`
	Parent     string                    `json:"parent" yaml:"parent"`
	Properties []PluginGraphPropertyDecl `json:"properties" yaml:"properties"`
}

// PluginGraphRelationDecl is a plugin-defined domain relation.
type PluginGraphRelationDecl struct {
	Kind        string                 `json:"kind" yaml:"kind"`
	From        string                 `json:"from" yaml:"from"`
	To          string                 `json:"to" yaml:"to"`
	Cardinality PluginGraphCardinality `json:"cardinality" yaml:"cardinality"`
}

// PluginGraphSchema is the optional manifest graph declaration.
type PluginGraphSchema struct {
	Nodes     []PluginGraphNodeDecl     `json:"nodes" yaml:"nodes"`
	Foreign   []PluginGraphForeignDecl  `json:"foreign,omitempty" yaml:"foreign,omitempty"`
	Relations []PluginGraphRelationDecl `json:"relations" yaml:"relations"`
}

// PluginGraphBindingValue is the stored value of a graph_binding config field.
type PluginGraphBindingValue struct {
	PluginID string `json:"plugin_id"`
	Kind     string `json:"kind"`
}

// PluginConfigFieldDecl is a per-activation config field on the manifest.
type PluginConfigFieldDecl struct {
	Name     string                `json:"name" yaml:"name"`
	Type     PluginConfigFieldType `json:"type" yaml:"type"`
	Foreign  string                `json:"foreign,omitempty" yaml:"foreign,omitempty"`
	Required bool                  `json:"required" yaml:"required"`
}

// PluginRequires declares host compatibility.
type PluginRequires struct {
	Elemo     string `json:"elemo" yaml:"elemo"`
	PluginAPI string `json:"pluginApi" yaml:"pluginApi"`
}

// PluginBackendDecl describes the WASM guest.
type PluginBackendDecl struct {
	Entry string `json:"entry" yaml:"entry"`
}

// PluginFrontendDecl describes the ESM frontend package.
type PluginFrontendDecl struct {
	Entry  string `json:"entry" yaml:"entry"`
	Module string `json:"module" yaml:"module"`
}

// PluginManifest is the versioned plugin.yaml document.
type PluginManifest struct {
	SchemaVersion int                     `json:"schemaVersion" yaml:"schemaVersion"`
	ID            string                  `json:"id" yaml:"id"`
	Name          string                  `json:"name" yaml:"name"`
	Version       string                  `json:"version" yaml:"version"`
	Requires      PluginRequires          `json:"requires" yaml:"requires"`
	Backend       *PluginBackendDecl      `json:"backend,omitempty" yaml:"backend,omitempty"`
	Frontend      *PluginFrontendDecl     `json:"frontend,omitempty" yaml:"frontend,omitempty"`
	Capabilities  []PluginCapability      `json:"capabilities" yaml:"capabilities"`
	Events        []PluginEventType       `json:"events" yaml:"events"`
	Slots         []PluginUISlot          `json:"slots" yaml:"slots"`
	Config        []PluginConfigFieldDecl `json:"config,omitempty" yaml:"config,omitempty"`
	Graph         *PluginGraphSchema      `json:"graph,omitempty" yaml:"graph,omitempty"`
}

// PluginInstallation is the instance-wide package record.
type PluginInstallation struct {
	ID        string         `json:"id"`
	PluginID  string         `json:"plugin_id"`
	Version   string         `json:"version"`
	Status    PluginStatus   `json:"status"`
	Manifest  PluginManifest `json:"manifest"`
	Error     string         `json:"error,omitempty"`
	CreatedAt *time.Time     `json:"created_at,omitempty"`
	UpdatedAt *time.Time     `json:"updated_at,omitempty"`
}

// PluginActivation is a scoped enablement of an installed plugin.
type PluginActivation struct {
	PluginID  string          `json:"plugin_id"`
	ScopeID   ID              `json:"scope_id"`
	Enabled   bool            `json:"enabled"`
	Config    json.RawMessage `json:"config,omitempty"`
	CreatedAt *time.Time      `json:"created_at,omitempty"`
	UpdatedAt *time.Time      `json:"updated_at,omitempty"`
}

// PluginStorageEntry is a private JSONB key/value for a plugin and scope.
type PluginStorageEntry struct {
	PluginID  string     `json:"plugin_id"`
	ScopeID   ID         `json:"scope_id"`
	Key       string     `json:"key"`
	Value     []byte     `json:"value"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// Extension is a plugin-defined graph node (Neo4j label Extension).
type Extension struct {
	ID         ID             `json:"id"`
	PluginID   string         `json:"plugin_id"`
	Kind       string         `json:"kind"`
	Properties map[string]any `json:"properties"`
	Parent     *ID            `json:"parent,omitempty"`
	CreatedAt  *time.Time     `json:"created_at,omitempty"`
	UpdatedAt  *time.Time     `json:"updated_at,omitempty"`
}

// ExtensionRelation is a namespaced domain edge involving an Extension.
type ExtensionRelation struct {
	ID        string     `json:"id"`
	Kind      string     `json:"kind"`
	From      ID         `json:"from"`
	To        ID         `json:"to"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

func (s PluginStatus) IsHealthy() bool {
	return s == PluginStatusInstalled || s == PluginStatusActive || s == PluginStatusDisabled
}

// ServesFrontend reports whether frontend discovery and assets may be served.
func (s PluginStatus) ServesFrontend() bool {
	switch s {
	case PluginStatusInstalled, PluginStatusStarting, PluginStatusActive, PluginStatusFailed:
		return true
	default:
		return false
	}
}

func (m *PluginManifest) Validate() error {
	if m == nil {
		return ErrInvalidPluginManifest
	}
	if m.SchemaVersion != PluginSchemaVersionV1 {
		return errors.Join(ErrInvalidPluginManifest, ErrPluginSchemaVersion)
	}
	if !pluginIDPattern.MatchString(m.ID) || len(m.ID) < pluginIDMinLen || len(m.ID) > pluginIDMaxLen {
		return errors.Join(ErrInvalidPluginManifest, ErrInvalidPluginID)
	}
	if n := len(strings.TrimSpace(m.Name)); n < pluginNameMinLen || n > pluginNameMaxLen {
		return errors.Join(ErrInvalidPluginManifest, errors.New("invalid plugin name"))
	}
	if m.Version == "" || len(m.Version) > pluginVersionMaxLen {
		return errors.Join(ErrInvalidPluginManifest, errors.New("invalid plugin version"))
	}
	if err := validatePluginAPIRequirement(m.Requires.PluginAPI); err != nil {
		return errors.Join(ErrInvalidPluginManifest, err)
	}
	if m.Backend == nil && m.Frontend == nil {
		return errors.Join(ErrInvalidPluginManifest, errors.New("backend or frontend is required"))
	}
	if m.Backend != nil && strings.TrimSpace(m.Backend.Entry) == "" {
		m.Backend.Entry = PluginBackendWASMPath
	}
	if m.Frontend != nil && strings.TrimSpace(m.Frontend.Entry) == "" {
		m.Frontend.Entry = PluginFrontendEntryDefault
	}
	seenCap := make(map[PluginCapability]struct{}, len(m.Capabilities))
	for _, capability := range m.Capabilities {
		if !capability.Valid() {
			return errors.Join(ErrInvalidPluginManifest, ErrInvalidPluginCapability)
		}
		if _, dup := seenCap[capability]; dup {
			return errors.Join(ErrInvalidPluginManifest, ErrInvalidPluginCapability)
		}
		seenCap[capability] = struct{}{}
	}
	for _, evt := range m.Events {
		if !evt.Valid() {
			return errors.Join(ErrInvalidPluginManifest, ErrInvalidPluginEvent)
		}
	}
	for _, slot := range m.Slots {
		if !slot.Valid() {
			return errors.Join(ErrInvalidPluginManifest, ErrInvalidPluginSlot)
		}
	}
	if m.Graph != nil {
		if err := m.Graph.Validate(); err != nil {
			return errors.Join(ErrInvalidPluginManifest, err)
		}
	}
	if err := validateConfigFields(m.Config, m.Graph); err != nil {
		return errors.Join(ErrInvalidPluginManifest, err)
	}
	return nil
}

func validatePluginAPIRequirement(req string) error {
	trimmed := strings.TrimSpace(req)
	if trimmed == "" {
		return ErrPluginAPIIncompatible
	}
	if trimmed == PluginAPIVersionV1 || trimmed == "^"+PluginAPIVersionV1 || trimmed == "^1" {
		return nil
	}
	return ErrPluginAPIIncompatible
}

func (s *PluginGraphSchema) Validate() error {
	if s == nil {
		return nil
	}
	kinds := make(map[string]*PluginGraphNodeDecl, len(s.Nodes))
	for i := range s.Nodes {
		node := &s.Nodes[i]
		if err := node.validate(); err != nil {
			return err
		}
		if _, dup := kinds[node.Kind]; dup {
			return fmt.Errorf("duplicate graph kind %s", node.Kind)
		}
		kinds[node.Kind] = node
	}
	foreign := make(map[string]*PluginGraphForeignDecl, len(s.Foreign))
	for i := range s.Foreign {
		decl := &s.Foreign[i]
		if err := decl.validate(); err != nil {
			return err
		}
		if _, core := coreGraphScopeTypes[decl.Name]; core {
			return fmt.Errorf("foreign name %s collides with a core type", decl.Name)
		}
		if _, local := kinds[decl.Name]; local {
			return fmt.Errorf("foreign name %s collides with a local kind", decl.Name)
		}
		if _, dup := foreign[decl.Name]; dup {
			return fmt.Errorf("duplicate foreign name %s", decl.Name)
		}
		if strings.TrimSpace(decl.Parent) == "" {
			return fmt.Errorf("foreign %s is missing parent", decl.Name)
		}
		if _, ok := coreGraphScopeTypes[decl.Parent]; !ok {
			if _, local := kinds[decl.Parent]; !local {
				return fmt.Errorf("foreign %s parent %s is unknown", decl.Name, decl.Parent)
			}
		}
		foreign[decl.Name] = decl
	}
	for _, node := range s.Nodes {
		if err := validateScopeParentChain(node.Kind, node.Scope.Parent, kinds, nil); err != nil {
			return err
		}
	}
	relKinds := make(map[string]struct{}, len(s.Relations))
	for _, rel := range s.Relations {
		if err := rel.validate(kinds, foreign); err != nil {
			return err
		}
		if _, dup := relKinds[rel.Kind]; dup {
			return fmt.Errorf("duplicate relation kind %s", rel.Kind)
		}
		relKinds[rel.Kind] = struct{}{}
	}
	return nil
}

func (n *PluginGraphNodeDecl) validate() error {
	if !pluginKindPattern.MatchString(n.Kind) || len(n.Kind) > pluginKindMaxLen {
		return fmt.Errorf("invalid graph kind %q", n.Kind)
	}
	if strings.TrimSpace(n.Scope.Parent) == "" {
		return fmt.Errorf("graph kind %s is missing scope.parent", n.Kind)
	}
	seen := make(map[string]struct{}, len(n.Properties))
	for _, prop := range n.Properties {
		if !pluginPropPattern.MatchString(prop.Name) {
			return fmt.Errorf("invalid property name %q", prop.Name)
		}
		if _, reserved := reservedGraphPropertyNames[prop.Name]; reserved {
			return fmt.Errorf("property name %q is reserved", prop.Name)
		}
		if !prop.Type.IsAPluginGraphPropertyType() || prop.Type == PluginGraphPropertyTypeUnknown {
			return fmt.Errorf("invalid property type for %s", prop.Name)
		}
		if _, dup := seen[prop.Name]; dup {
			return fmt.Errorf("duplicate property %s", prop.Name)
		}
		seen[prop.Name] = struct{}{}
	}
	return nil
}

func (f *PluginGraphForeignDecl) validate() error {
	if f == nil {
		return fmt.Errorf("invalid foreign kind")
	}
	if !pluginKindPattern.MatchString(f.Name) || len(f.Name) > pluginKindMaxLen {
		return fmt.Errorf("invalid foreign name %q", f.Name)
	}
	seen := make(map[string]struct{}, len(f.Properties))
	for _, prop := range f.Properties {
		if !pluginPropPattern.MatchString(prop.Name) {
			return fmt.Errorf("invalid property name %q", prop.Name)
		}
		if _, reserved := reservedGraphPropertyNames[prop.Name]; reserved {
			return fmt.Errorf("property name %q is reserved", prop.Name)
		}
		if !prop.Type.IsAPluginGraphPropertyType() || prop.Type == PluginGraphPropertyTypeUnknown {
			return fmt.Errorf("invalid property type for %s", prop.Name)
		}
		if _, dup := seen[prop.Name]; dup {
			return fmt.Errorf("duplicate property %s", prop.Name)
		}
		seen[prop.Name] = struct{}{}
	}
	return nil
}

func (r PluginGraphRelationDecl) validate(
	kinds map[string]*PluginGraphNodeDecl,
	foreign map[string]*PluginGraphForeignDecl,
) error {
	if !pluginKindPattern.MatchString(r.Kind) {
		return fmt.Errorf("invalid relation kind %q", r.Kind)
	}
	if _, reserved := coreEdgeKindNames[r.Kind]; reserved {
		return fmt.Errorf("relation kind %s collides with a core edge", r.Kind)
	}
	if r.Cardinality == PluginGraphCardinalityUnknown || !r.Cardinality.IsAPluginGraphCardinality() {
		return fmt.Errorf("invalid cardinality for relation %s", r.Kind)
	}
	if !isGraphEndpoint(r.From, kinds, foreign) {
		return fmt.Errorf("relation %s from %s is not a core type, plugin kind, or foreign alias", r.Kind, r.From)
	}
	if !isGraphEndpoint(r.To, kinds, foreign) {
		return fmt.Errorf("relation %s to %s is not a core type, plugin kind, or foreign alias", r.Kind, r.To)
	}
	return nil
}

func isGraphEndpoint(
	name string,
	kinds map[string]*PluginGraphNodeDecl,
	foreign map[string]*PluginGraphForeignDecl,
) bool {
	if _, ok := coreGraphScopeTypes[name]; ok {
		return true
	}
	if _, ok := kinds[name]; ok {
		return true
	}
	_, ok := foreign[name]
	return ok
}

func validateConfigFields(fields []PluginConfigFieldDecl, graph *PluginGraphSchema) error {
	seen := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		if !pluginPropPattern.MatchString(field.Name) {
			return fmt.Errorf("invalid config field name %q", field.Name)
		}
		if _, dup := seen[field.Name]; dup {
			return fmt.Errorf("duplicate config field %s", field.Name)
		}
		seen[field.Name] = struct{}{}
		if !field.Type.IsAPluginConfigFieldType() || field.Type == PluginConfigFieldTypeUnknown {
			return fmt.Errorf("invalid config field type for %s", field.Name)
		}
		if field.Type == PluginConfigFieldTypeGraphBinding {
			if strings.TrimSpace(field.Foreign) == "" {
				return fmt.Errorf("config field %s is missing foreign", field.Name)
			}
			if graph == nil {
				return fmt.Errorf("config field %s references foreign %s without a graph", field.Name, field.Foreign)
			}
			if _, ok := graph.ForeignKind(field.Foreign); !ok {
				return fmt.Errorf("config field %s references unknown foreign %s", field.Name, field.Foreign)
			}
		} else if field.Foreign != "" {
			return fmt.Errorf("config field %s cannot declare foreign", field.Name)
		}
	}
	return nil
}

func validateScopeParentChain(kind, parent string, kinds map[string]*PluginGraphNodeDecl, seen map[string]struct{}) error {
	if _, ok := coreGraphScopeTypes[parent]; ok {
		return nil
	}
	next, ok := kinds[parent]
	if !ok {
		return fmt.Errorf("graph kind %s parent %s is unknown", kind, parent)
	}
	if seen == nil {
		seen = make(map[string]struct{})
	}
	if _, cycle := seen[kind]; cycle {
		return fmt.Errorf("graph scope cycle involving %s", kind)
	}
	seen[kind] = struct{}{}
	if _, cycle := seen[parent]; cycle {
		return fmt.Errorf("graph scope cycle involving %s", parent)
	}
	return validateScopeParentChain(parent, next.Scope.Parent, kinds, seen)
}

// CoreScopeType returns the ResourceType for a core graph parent name.
func CoreScopeType(name string) (ResourceType, bool) {
	rt, ok := coreGraphScopeTypes[name]
	return rt, ok
}

// IsCoreGraphType reports whether name is a core allowlisted graph type.
func IsCoreGraphType(name string) bool {
	_, ok := coreGraphScopeTypes[name]
	return ok
}

// IsReservedGraphProperty reports whether name is reserved on Extension nodes.
func IsReservedGraphProperty(name string) bool {
	_, ok := reservedGraphPropertyNames[name]
	return ok
}

// IsReservedEdgeKind reports whether name collides with a core EdgeKind.
func IsReservedEdgeKind(name string) bool {
	_, ok := coreEdgeKindNames[name]
	return ok
}

// HasCapability reports whether the manifest grants the capability.
func (m PluginManifest) HasCapability(capability PluginCapability) bool {
	for _, c := range m.Capabilities {
		if c == capability {
			return true
		}
	}
	return false
}

// NodeKind returns the declared node kind, if any.
func (s *PluginGraphSchema) NodeKind(kind string) (PluginGraphNodeDecl, bool) {
	if s == nil {
		return PluginGraphNodeDecl{}, false
	}
	for _, node := range s.Nodes {
		if node.Kind == kind {
			return node, true
		}
	}
	return PluginGraphNodeDecl{}, false
}

// RelationKind returns the declared relation, if any.
func (s *PluginGraphSchema) RelationKind(kind string) (PluginGraphRelationDecl, bool) {
	if s == nil {
		return PluginGraphRelationDecl{}, false
	}
	for _, rel := range s.Relations {
		if rel.Kind == kind {
			return rel, true
		}
	}
	return PluginGraphRelationDecl{}, false
}

// ForeignKind returns the declared foreign alias, if any.
func (s *PluginGraphSchema) ForeignKind(name string) (PluginGraphForeignDecl, bool) {
	if s == nil {
		return PluginGraphForeignDecl{}, false
	}
	for _, decl := range s.Foreign {
		if decl.Name == name {
			return decl, true
		}
	}
	return PluginGraphForeignDecl{}, false
}

// ConfigField returns the declared config field, if any.
func (m PluginManifest) ConfigField(name string) (PluginConfigFieldDecl, bool) {
	for _, field := range m.Config {
		if field.Name == name {
			return field, true
		}
	}
	return PluginConfigFieldDecl{}, false
}

// GraphBinding returns the bound foreign source for alias, if configured.
func GraphBinding(config json.RawMessage, fields []PluginConfigFieldDecl, alias string) (PluginGraphBindingValue, bool) {
	if len(config) == 0 {
		return PluginGraphBindingValue{}, false
	}
	var values map[string]json.RawMessage
	if err := json.Unmarshal(config, &values); err != nil {
		return PluginGraphBindingValue{}, false
	}
	for _, field := range fields {
		if field.Type != PluginConfigFieldTypeGraphBinding || field.Foreign != alias {
			continue
		}
		raw, ok := values[field.Name]
		if !ok || len(raw) == 0 || string(raw) == "null" {
			return PluginGraphBindingValue{}, false
		}
		var binding PluginGraphBindingValue
		if err := json.Unmarshal(raw, &binding); err != nil {
			return PluginGraphBindingValue{}, false
		}
		if binding.PluginID == "" || binding.Kind == "" {
			return PluginGraphBindingValue{}, false
		}
		return binding, true
	}
	return PluginGraphBindingValue{}, false
}

// MatchesForeign reports whether node satisfies the foreign alias shape.
func (f PluginGraphForeignDecl) MatchesKind(node PluginGraphNodeDecl) bool {
	if f.Parent != node.Scope.Parent {
		return false
	}
	props := make(map[string]PluginGraphPropertyDecl, len(node.Properties))
	for _, p := range node.Properties {
		props[p.Name] = p
	}
	for _, want := range f.Properties {
		got, ok := props[want.Name]
		if !ok || got.Type != want.Type {
			return false
		}
	}
	return true
}

// BindingMatches reports whether ownerPluginID+kind is a configured binding.
func BindingMatches(config json.RawMessage, fields []PluginConfigFieldDecl, ownerPluginID, kind string) bool {
	if ownerPluginID == "" || kind == "" {
		return false
	}
	var values map[string]json.RawMessage
	if err := json.Unmarshal(config, &values); err != nil && len(config) > 0 {
		return false
	}
	for _, field := range fields {
		if field.Type != PluginConfigFieldTypeGraphBinding {
			continue
		}
		raw, ok := values[field.Name]
		if !ok || len(raw) == 0 || string(raw) == "null" {
			continue
		}
		var binding PluginGraphBindingValue
		if err := json.Unmarshal(raw, &binding); err != nil {
			continue
		}
		if binding.PluginID == ownerPluginID && binding.Kind == kind {
			return true
		}
	}
	return false
}

// NamespacedRelationType returns the Neo4j rel type for a plugin domain edge.
func NamespacedRelationType(pluginID, kind string) string {
	sanitized := strings.ReplaceAll(pluginID, ".", "_")
	sanitized = strings.ReplaceAll(sanitized, "-", "_")
	return "EXT__" + sanitized + "__" + kind
}

func (e *Extension) Validate() error {
	if e == nil {
		return ErrInvalidExtensionDetails
	}
	if err := validate.Struct(e); err != nil {
		return errors.Join(ErrInvalidExtensionDetails, err)
	}
	if err := e.ID.Validate(); err != nil {
		return errors.Join(ErrInvalidExtensionDetails, err)
	}
	if e.ID.Type != ResourceTypeExtension {
		return errors.Join(ErrInvalidExtensionDetails, ErrInvalidID)
	}
	if !pluginIDPattern.MatchString(e.PluginID) {
		return errors.Join(ErrInvalidExtensionDetails, ErrInvalidPluginID)
	}
	if !pluginKindPattern.MatchString(e.Kind) {
		return errors.Join(ErrInvalidExtensionDetails, errors.New("invalid extension kind"))
	}
	return nil
}

// NewExtension allocates an Extension with a real xid.
func NewExtension(pluginID, kind string, properties map[string]any) (*Extension, error) {
	if properties == nil {
		properties = map[string]any{}
	}
	ext := &Extension{
		ID:         MustNewID(ResourceTypeExtension),
		PluginID:   pluginID,
		Kind:       kind,
		Properties: properties,
	}
	if err := ext.Validate(); err != nil {
		return nil, err
	}
	return ext, nil
}

// IsPluginActivationScopeType reports whether rt may own a plugin activation.
func IsPluginActivationScopeType(rt ResourceType) bool {
	switch rt {
	case ResourceTypeOrganization, ResourceTypeNamespace, ResourceTypeProject:
		return true
	default:
		return false
	}
}
