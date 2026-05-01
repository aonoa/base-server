package authx

import "context"

// ProjectionSource is the domain-service side adapter contract.
// A business service owns its permission truth and only needs to implement
// these normalized list methods to publish a full snapshot into auth.
type ProjectionSource interface {
	SourceService() string
	ListProjectionRoles(context.Context) ([]ProjectionRole, error)
	ListProjectionAPIs(context.Context) ([]ProjectionAPI, error)
	ListProjectionBindings(context.Context) ([]UserRoleBindingProjection, error)
}

// BuildSnapshotFromSource converts a ProjectionSource into the full snapshot
// payload used by auth snapshot registration.
func BuildSnapshotFromSource(ctx context.Context, source ProjectionSource) (*PermissionSnapshot, error) {
	roles, err := source.ListProjectionRoles(ctx)
	if err != nil {
		return nil, err
	}
	apis, err := source.ListProjectionAPIs(ctx)
	if err != nil {
		return nil, err
	}
	bindings, err := source.ListProjectionBindings(ctx)
	if err != nil {
		return nil, err
	}
	return &PermissionSnapshot{
		SourceService: source.SourceService(),
		Revision:      NewProjectionRevision(),
		Roles:         roles,
		APIs:          apis,
		Bindings:      bindings,
	}, nil
}

// RegisterSourceSnapshot is the convenience entrypoint business services should
// use on startup once they implement ProjectionSource.
func (c *ProjectionClient) RegisterSourceSnapshot(ctx context.Context, source ProjectionSource) error {
	snapshot, err := BuildSnapshotFromSource(ctx, source)
	if err != nil {
		return err
	}
	return c.RegisterSnapshot(ctx, *snapshot)
}
