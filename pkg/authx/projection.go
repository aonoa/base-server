package authx

import (
	"context"
	"errors"
	"time"

	authv1 "base-server/api/gen/go/auth/service/v1"
)

var errProjectionClientUnavailable = errors.New("auth projection client is not configured")

// ProjectionRoleResource is the normalized role-to-resource projection payload.
// It intentionally mirrors the current auth proto so different services can
// reuse one sender implementation without rebuilding auth requests by hand.
type ProjectionRoleResource struct {
	ID     string
	Type   string
	Value  string
	Method string
}

// ProjectionRole is the normalized role projection payload.
type ProjectionRole struct {
	ID        int64
	Name      string
	Value     string
	Status    bool
	Remark    string
	MenuIDs   []int32
	Resources []ProjectionRoleResource
	Service   string
	Domain    string
}

// ProjectionAPI is the normalized API projection payload.
type ProjectionAPI struct {
	ID                string
	Path              string
	Method            string
	Description       string
	Module            string
	ModuleDescription string
	ResourceGroup     string
	Service           string
	Domain            string
}

// UserRoleBindingProjection is the normalized user-role binding payload.
// The current auth runtime still projects user bindings only; a later scope-
// aware version can extend this without forcing services to assemble proto DTOs.
type UserRoleBindingProjection struct {
	ID         string
	UserID     string
	RoleID     int64
	RoleValue  string
	CreateTime string
	UpdateTime string
	Service    string
	ScopeID    string
	Domain     string
}

// PermissionSnapshot is the full projection payload a source service publishes
// into auth on startup or during recovery.
type PermissionSnapshot struct {
	SourceService string
	Revision      uint64
	Roles         []ProjectionRole
	APIs          []ProjectionAPI
	Bindings      []UserRoleBindingProjection
}

// ProjectionClient wraps the auth projection RPCs behind a reusable sender.
// Current callers can keep publishing admin-originated projections, and future
// business services can reuse the same client with their own source_service.
type ProjectionClient struct {
	client         authv1.AuthServiceClient
	sourceService  string
	statusReporter ProjectionStatusReporter
	domainCode     string
	description    string
}

func NewProjectionClient(client authv1.AuthServiceClient, sourceService string) *ProjectionClient {
	return &ProjectionClient{
		client:        client,
		sourceService: sourceService,
	}
}

func (c *ProjectionClient) SetStatusReporter(reporter ProjectionStatusReporter, domainCode, description string) *ProjectionClient {
	if c == nil {
		return nil
	}
	c.statusReporter = reporter
	c.domainCode = domainCode
	c.description = description
	return c
}

func NewProjectionRevision() uint64 {
	return uint64(time.Now().UnixMilli())
}

func (c *ProjectionClient) RegisterSnapshot(ctx context.Context, snapshot PermissionSnapshot) error {
	sourceService := c.sourceServiceOrDefault(snapshot.SourceService)
	revision := snapshot.Revision
	client, err := c.authClient()
	if err != nil {
		c.reportProjectionStatus(ctx, sourceService, ProjectionSyncModeSnapshot, revision, err)
		return err
	}
	req := &authv1.RegisterPermissionSnapshotRequest{
		SourceService: sourceService,
		Revision:      snapshot.Revision,
		Roles:         make([]*authv1.PolicyRole, 0, len(snapshot.Roles)),
		Apis:          make([]*authv1.PolicyApi, 0, len(snapshot.APIs)),
		Bindings:      make([]*authv1.PolicyUserRoleBinding, 0, len(snapshot.Bindings)),
	}
	for _, role := range snapshot.Roles {
		req.Roles = append(req.Roles, projectionRoleToProto(role))
	}
	for _, api := range snapshot.APIs {
		req.Apis = append(req.Apis, projectionAPIToProto(api))
	}
	for _, binding := range snapshot.Bindings {
		req.Bindings = append(req.Bindings, userRoleBindingProjectionToProto(binding))
	}
	ctx = ForwardAuthorizationContext(ctx)
	_, err = client.RegisterPermissionSnapshot(ctx, req)
	c.reportProjectionStatus(ctx, sourceService, ProjectionSyncModeSnapshot, revision, err)
	return err
}

func (c *ProjectionClient) ApplyRoleDelta(ctx context.Context, sourceService string, before, after *ProjectionRole) error {
	sourceService = c.sourceServiceOrDefault(sourceService)
	revision := NewProjectionRevision()
	client, err := c.authClient()
	if err != nil {
		c.reportProjectionStatus(ctx, sourceService, ProjectionSyncModeDelta, revision, err)
		return err
	}
	ctx = ForwardAuthorizationContext(ctx)
	_, err = client.ApplyRoleDelta(ctx, &authv1.ApplyRoleDeltaRequest{
		SourceService: sourceService,
		Revision:      revision,
		Before:        projectionRolePtrToProto(before),
		After:         projectionRolePtrToProto(after),
	})
	c.reportProjectionStatus(ctx, sourceService, ProjectionSyncModeDelta, revision, err)
	return err
}

func (c *ProjectionClient) ApplyAPIDelta(ctx context.Context, sourceService string, before, after *ProjectionAPI) error {
	sourceService = c.sourceServiceOrDefault(sourceService)
	revision := NewProjectionRevision()
	client, err := c.authClient()
	if err != nil {
		c.reportProjectionStatus(ctx, sourceService, ProjectionSyncModeDelta, revision, err)
		return err
	}
	ctx = ForwardAuthorizationContext(ctx)
	_, err = client.ApplyApiDelta(ctx, &authv1.ApplyApiDeltaRequest{
		SourceService: sourceService,
		Revision:      revision,
		Before:        projectionAPIPtrToProto(before),
		After:         projectionAPIPtrToProto(after),
	})
	c.reportProjectionStatus(ctx, sourceService, ProjectionSyncModeDelta, revision, err)
	return err
}

func (c *ProjectionClient) ApplyUserRoleBindingDelta(ctx context.Context, sourceService string, before, after *UserRoleBindingProjection) error {
	sourceService = c.sourceServiceOrDefault(sourceService)
	revision := NewProjectionRevision()
	client, err := c.authClient()
	if err != nil {
		c.reportProjectionStatus(ctx, sourceService, ProjectionSyncModeDelta, revision, err)
		return err
	}
	ctx = ForwardAuthorizationContext(ctx)
	_, err = client.ApplyUserRoleBindingDelta(ctx, &authv1.ApplyUserRoleBindingDeltaRequest{
		SourceService: sourceService,
		Revision:      revision,
		Before:        userRoleBindingProjectionPtrToProto(before),
		After:         userRoleBindingProjectionPtrToProto(after),
	})
	c.reportProjectionStatus(ctx, sourceService, ProjectionSyncModeDelta, revision, err)
	return err
}

func (c *ProjectionClient) authClient() (authv1.AuthServiceClient, error) {
	if c == nil || c.client == nil {
		return nil, errProjectionClientUnavailable
	}
	return c.client, nil
}

func (c *ProjectionClient) sourceServiceOrDefault(sourceService string) string {
	if sourceService != "" {
		return sourceService
	}
	return c.sourceService
}

func (c *ProjectionClient) reportProjectionStatus(ctx context.Context, sourceService, syncMode string, revision uint64, syncErr error) {
	if c == nil || c.statusReporter == nil {
		return
	}
	state := ProjectionStateSynced
	lastError := ""
	if syncErr != nil {
		state = ProjectionStateError
		lastError = syncErr.Error()
	}
	_ = c.statusReporter.ReportProjectionStatus(ctx, ProjectionStatus{
		SourceService:        sourceService,
		DomainCode:           c.domainCode,
		SyncMode:             syncMode,
		State:                state,
		LastSnapshotRevision: revision,
		LastSyncTime:         projectionStatusNow(),
		LastError:            lastError,
		Description:          c.description,
	})
}

func projectionRolePtrToProto(item *ProjectionRole) *authv1.PolicyRole {
	if item == nil {
		return nil
	}
	return projectionRoleToProto(*item)
}

func projectionRoleToProto(item ProjectionRole) *authv1.PolicyRole {
	resources := make([]*authv1.PolicyRoleResource, 0, len(item.Resources))
	for _, resource := range item.Resources {
		resources = append(resources, &authv1.PolicyRoleResource{
			Id:     resource.ID,
			Type:   resource.Type,
			Value:  resource.Value,
			Method: resource.Method,
		})
	}
	return &authv1.PolicyRole{
		Id:         item.ID,
		Name:       item.Name,
		Value:      item.Value,
		Status:     item.Status,
		Remark:     item.Remark,
		MenuIds:    append([]int32(nil), item.MenuIDs...),
		Resources:  resources,
		Service:    item.Service,
		DomainCode: item.Domain,
	}
}

func projectionAPIPtrToProto(item *ProjectionAPI) *authv1.PolicyApi {
	if item == nil {
		return nil
	}
	return projectionAPIToProto(*item)
}

func projectionAPIToProto(item ProjectionAPI) *authv1.PolicyApi {
	return &authv1.PolicyApi{
		Id:                item.ID,
		Path:              item.Path,
		Method:            item.Method,
		Description:       item.Description,
		Module:            item.Module,
		ModuleDescription: item.ModuleDescription,
		ResourcesGroup:    item.ResourceGroup,
		Service:           item.Service,
		DomainCode:        item.Domain,
	}
}

func userRoleBindingProjectionPtrToProto(item *UserRoleBindingProjection) *authv1.PolicyUserRoleBinding {
	if item == nil {
		return nil
	}
	return userRoleBindingProjectionToProto(*item)
}

func userRoleBindingProjectionToProto(item UserRoleBindingProjection) *authv1.PolicyUserRoleBinding {
	return &authv1.PolicyUserRoleBinding{
		Id:         item.ID,
		UserId:     item.UserID,
		RoleId:     item.RoleID,
		RoleValue:  item.RoleValue,
		CreateTime: item.CreateTime,
		UpdateTime: item.UpdateTime,
		Service:    item.Service,
		ScopeId:    item.ScopeID,
		DomainCode: item.Domain,
	}
}
