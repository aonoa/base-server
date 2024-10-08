// Code generated by ent, DO NOT EDIT.

package ent

import (
	"base-server/internal/data/ent/menu"
	"base-server/internal/data/ent/predicate"
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// MenuUpdate is the builder for updating Menu entities.
type MenuUpdate struct {
	config
	hooks    []Hook
	mutation *MenuMutation
}

// Where appends a list predicates to the MenuUpdate builder.
func (mu *MenuUpdate) Where(ps ...predicate.Menu) *MenuUpdate {
	mu.mutation.Where(ps...)
	return mu
}

// SetUpdateTime sets the "update_time" field.
func (mu *MenuUpdate) SetUpdateTime(t time.Time) *MenuUpdate {
	mu.mutation.SetUpdateTime(t)
	return mu
}

// SetPid sets the "pid" field.
func (mu *MenuUpdate) SetPid(i int64) *MenuUpdate {
	mu.mutation.ResetPid()
	mu.mutation.SetPid(i)
	return mu
}

// SetNillablePid sets the "pid" field if the given value is not nil.
func (mu *MenuUpdate) SetNillablePid(i *int64) *MenuUpdate {
	if i != nil {
		mu.SetPid(*i)
	}
	return mu
}

// AddPid adds i to the "pid" field.
func (mu *MenuUpdate) AddPid(i int64) *MenuUpdate {
	mu.mutation.AddPid(i)
	return mu
}

// SetType sets the "type" field.
func (mu *MenuUpdate) SetType(i int8) *MenuUpdate {
	mu.mutation.ResetType()
	mu.mutation.SetType(i)
	return mu
}

// SetNillableType sets the "type" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableType(i *int8) *MenuUpdate {
	if i != nil {
		mu.SetType(*i)
	}
	return mu
}

// AddType adds i to the "type" field.
func (mu *MenuUpdate) AddType(i int8) *MenuUpdate {
	mu.mutation.AddType(i)
	return mu
}

// SetStatus sets the "status" field.
func (mu *MenuUpdate) SetStatus(b bool) *MenuUpdate {
	mu.mutation.SetStatus(b)
	return mu
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableStatus(b *bool) *MenuUpdate {
	if b != nil {
		mu.SetStatus(*b)
	}
	return mu
}

// SetName sets the "name" field.
func (mu *MenuUpdate) SetName(s string) *MenuUpdate {
	mu.mutation.SetName(s)
	return mu
}

// SetNillableName sets the "name" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableName(s *string) *MenuUpdate {
	if s != nil {
		mu.SetName(*s)
	}
	return mu
}

// SetTitle sets the "title" field.
func (mu *MenuUpdate) SetTitle(s string) *MenuUpdate {
	mu.mutation.SetTitle(s)
	return mu
}

// SetNillableTitle sets the "title" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableTitle(s *string) *MenuUpdate {
	if s != nil {
		mu.SetTitle(*s)
	}
	return mu
}

// SetIcon sets the "icon" field.
func (mu *MenuUpdate) SetIcon(s string) *MenuUpdate {
	mu.mutation.SetIcon(s)
	return mu
}

// SetNillableIcon sets the "icon" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableIcon(s *string) *MenuUpdate {
	if s != nil {
		mu.SetIcon(*s)
	}
	return mu
}

// SetOrder sets the "order" field.
func (mu *MenuUpdate) SetOrder(i int32) *MenuUpdate {
	mu.mutation.ResetOrder()
	mu.mutation.SetOrder(i)
	return mu
}

// SetNillableOrder sets the "order" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableOrder(i *int32) *MenuUpdate {
	if i != nil {
		mu.SetOrder(*i)
	}
	return mu
}

// AddOrder adds i to the "order" field.
func (mu *MenuUpdate) AddOrder(i int32) *MenuUpdate {
	mu.mutation.AddOrder(i)
	return mu
}

// SetPath sets the "path" field.
func (mu *MenuUpdate) SetPath(s string) *MenuUpdate {
	mu.mutation.SetPath(s)
	return mu
}

// SetNillablePath sets the "path" field if the given value is not nil.
func (mu *MenuUpdate) SetNillablePath(s *string) *MenuUpdate {
	if s != nil {
		mu.SetPath(*s)
	}
	return mu
}

// SetComponent sets the "component" field.
func (mu *MenuUpdate) SetComponent(s string) *MenuUpdate {
	mu.mutation.SetComponent(s)
	return mu
}

// SetNillableComponent sets the "component" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableComponent(s *string) *MenuUpdate {
	if s != nil {
		mu.SetComponent(*s)
	}
	return mu
}

// SetRedirect sets the "redirect" field.
func (mu *MenuUpdate) SetRedirect(s string) *MenuUpdate {
	mu.mutation.SetRedirect(s)
	return mu
}

// SetNillableRedirect sets the "redirect" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableRedirect(s *string) *MenuUpdate {
	if s != nil {
		mu.SetRedirect(*s)
	}
	return mu
}

// SetLink sets the "link" field.
func (mu *MenuUpdate) SetLink(s string) *MenuUpdate {
	mu.mutation.SetLink(s)
	return mu
}

// SetNillableLink sets the "link" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableLink(s *string) *MenuUpdate {
	if s != nil {
		mu.SetLink(*s)
	}
	return mu
}

// SetIframeSrc sets the "iframeSrc" field.
func (mu *MenuUpdate) SetIframeSrc(s string) *MenuUpdate {
	mu.mutation.SetIframeSrc(s)
	return mu
}

// SetNillableIframeSrc sets the "iframeSrc" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableIframeSrc(s *string) *MenuUpdate {
	if s != nil {
		mu.SetIframeSrc(*s)
	}
	return mu
}

// SetActiveIcon sets the "activeIcon" field.
func (mu *MenuUpdate) SetActiveIcon(s string) *MenuUpdate {
	mu.mutation.SetActiveIcon(s)
	return mu
}

// SetNillableActiveIcon sets the "activeIcon" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableActiveIcon(s *string) *MenuUpdate {
	if s != nil {
		mu.SetActiveIcon(*s)
	}
	return mu
}

// SetActivePath sets the "activePath" field.
func (mu *MenuUpdate) SetActivePath(s string) *MenuUpdate {
	mu.mutation.SetActivePath(s)
	return mu
}

// SetNillableActivePath sets the "activePath" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableActivePath(s *string) *MenuUpdate {
	if s != nil {
		mu.SetActivePath(*s)
	}
	return mu
}

// SetMaxNumOfOpenTab sets the "maxNumOfOpenTab" field.
func (mu *MenuUpdate) SetMaxNumOfOpenTab(i int16) *MenuUpdate {
	mu.mutation.ResetMaxNumOfOpenTab()
	mu.mutation.SetMaxNumOfOpenTab(i)
	return mu
}

// SetNillableMaxNumOfOpenTab sets the "maxNumOfOpenTab" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableMaxNumOfOpenTab(i *int16) *MenuUpdate {
	if i != nil {
		mu.SetMaxNumOfOpenTab(*i)
	}
	return mu
}

// AddMaxNumOfOpenTab adds i to the "maxNumOfOpenTab" field.
func (mu *MenuUpdate) AddMaxNumOfOpenTab(i int16) *MenuUpdate {
	mu.mutation.AddMaxNumOfOpenTab(i)
	return mu
}

// SetIgnoreAuth sets the "ignoreAuth" field.
func (mu *MenuUpdate) SetIgnoreAuth(b bool) *MenuUpdate {
	mu.mutation.SetIgnoreAuth(b)
	return mu
}

// SetNillableIgnoreAuth sets the "ignoreAuth" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableIgnoreAuth(b *bool) *MenuUpdate {
	if b != nil {
		mu.SetIgnoreAuth(*b)
	}
	return mu
}

// SetKeepalive sets the "keepalive" field.
func (mu *MenuUpdate) SetKeepalive(b bool) *MenuUpdate {
	mu.mutation.SetKeepalive(b)
	return mu
}

// SetNillableKeepalive sets the "keepalive" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableKeepalive(b *bool) *MenuUpdate {
	if b != nil {
		mu.SetKeepalive(*b)
	}
	return mu
}

// SetPermission sets the "permission" field.
func (mu *MenuUpdate) SetPermission(s string) *MenuUpdate {
	mu.mutation.SetPermission(s)
	return mu
}

// SetNillablePermission sets the "permission" field if the given value is not nil.
func (mu *MenuUpdate) SetNillablePermission(s *string) *MenuUpdate {
	if s != nil {
		mu.SetPermission(*s)
	}
	return mu
}

// SetAffixTab sets the "affixTab" field.
func (mu *MenuUpdate) SetAffixTab(b bool) *MenuUpdate {
	mu.mutation.SetAffixTab(b)
	return mu
}

// SetNillableAffixTab sets the "affixTab" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableAffixTab(b *bool) *MenuUpdate {
	if b != nil {
		mu.SetAffixTab(*b)
	}
	return mu
}

// SetAffixTabOrder sets the "affixTabOrder" field.
func (mu *MenuUpdate) SetAffixTabOrder(i int64) *MenuUpdate {
	mu.mutation.ResetAffixTabOrder()
	mu.mutation.SetAffixTabOrder(i)
	return mu
}

// SetNillableAffixTabOrder sets the "affixTabOrder" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableAffixTabOrder(i *int64) *MenuUpdate {
	if i != nil {
		mu.SetAffixTabOrder(*i)
	}
	return mu
}

// AddAffixTabOrder adds i to the "affixTabOrder" field.
func (mu *MenuUpdate) AddAffixTabOrder(i int64) *MenuUpdate {
	mu.mutation.AddAffixTabOrder(i)
	return mu
}

// SetHideInMenu sets the "hideInMenu" field.
func (mu *MenuUpdate) SetHideInMenu(b bool) *MenuUpdate {
	mu.mutation.SetHideInMenu(b)
	return mu
}

// SetNillableHideInMenu sets the "hideInMenu" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableHideInMenu(b *bool) *MenuUpdate {
	if b != nil {
		mu.SetHideInMenu(*b)
	}
	return mu
}

// SetHideInTab sets the "hideInTab" field.
func (mu *MenuUpdate) SetHideInTab(b bool) *MenuUpdate {
	mu.mutation.SetHideInTab(b)
	return mu
}

// SetNillableHideInTab sets the "hideInTab" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableHideInTab(b *bool) *MenuUpdate {
	if b != nil {
		mu.SetHideInTab(*b)
	}
	return mu
}

// SetHideInBreadcrumb sets the "hideInBreadcrumb" field.
func (mu *MenuUpdate) SetHideInBreadcrumb(b bool) *MenuUpdate {
	mu.mutation.SetHideInBreadcrumb(b)
	return mu
}

// SetNillableHideInBreadcrumb sets the "hideInBreadcrumb" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableHideInBreadcrumb(b *bool) *MenuUpdate {
	if b != nil {
		mu.SetHideInBreadcrumb(*b)
	}
	return mu
}

// SetHideChildrenInMenu sets the "hideChildrenInMenu" field.
func (mu *MenuUpdate) SetHideChildrenInMenu(b bool) *MenuUpdate {
	mu.mutation.SetHideChildrenInMenu(b)
	return mu
}

// SetNillableHideChildrenInMenu sets the "hideChildrenInMenu" field if the given value is not nil.
func (mu *MenuUpdate) SetNillableHideChildrenInMenu(b *bool) *MenuUpdate {
	if b != nil {
		mu.SetHideChildrenInMenu(*b)
	}
	return mu
}

// Mutation returns the MenuMutation object of the builder.
func (mu *MenuUpdate) Mutation() *MenuMutation {
	return mu.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (mu *MenuUpdate) Save(ctx context.Context) (int, error) {
	mu.defaults()
	return withHooks(ctx, mu.sqlSave, mu.mutation, mu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (mu *MenuUpdate) SaveX(ctx context.Context) int {
	affected, err := mu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (mu *MenuUpdate) Exec(ctx context.Context) error {
	_, err := mu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (mu *MenuUpdate) ExecX(ctx context.Context) {
	if err := mu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (mu *MenuUpdate) defaults() {
	if _, ok := mu.mutation.UpdateTime(); !ok {
		v := menu.UpdateDefaultUpdateTime()
		mu.mutation.SetUpdateTime(v)
	}
}

func (mu *MenuUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(menu.Table, menu.Columns, sqlgraph.NewFieldSpec(menu.FieldID, field.TypeInt64))
	if ps := mu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := mu.mutation.UpdateTime(); ok {
		_spec.SetField(menu.FieldUpdateTime, field.TypeTime, value)
	}
	if value, ok := mu.mutation.Pid(); ok {
		_spec.SetField(menu.FieldPid, field.TypeInt64, value)
	}
	if value, ok := mu.mutation.AddedPid(); ok {
		_spec.AddField(menu.FieldPid, field.TypeInt64, value)
	}
	if value, ok := mu.mutation.GetType(); ok {
		_spec.SetField(menu.FieldType, field.TypeInt8, value)
	}
	if value, ok := mu.mutation.AddedType(); ok {
		_spec.AddField(menu.FieldType, field.TypeInt8, value)
	}
	if value, ok := mu.mutation.Status(); ok {
		_spec.SetField(menu.FieldStatus, field.TypeBool, value)
	}
	if value, ok := mu.mutation.Name(); ok {
		_spec.SetField(menu.FieldName, field.TypeString, value)
	}
	if value, ok := mu.mutation.Title(); ok {
		_spec.SetField(menu.FieldTitle, field.TypeString, value)
	}
	if value, ok := mu.mutation.Icon(); ok {
		_spec.SetField(menu.FieldIcon, field.TypeString, value)
	}
	if value, ok := mu.mutation.Order(); ok {
		_spec.SetField(menu.FieldOrder, field.TypeInt32, value)
	}
	if value, ok := mu.mutation.AddedOrder(); ok {
		_spec.AddField(menu.FieldOrder, field.TypeInt32, value)
	}
	if value, ok := mu.mutation.Path(); ok {
		_spec.SetField(menu.FieldPath, field.TypeString, value)
	}
	if value, ok := mu.mutation.Component(); ok {
		_spec.SetField(menu.FieldComponent, field.TypeString, value)
	}
	if value, ok := mu.mutation.Redirect(); ok {
		_spec.SetField(menu.FieldRedirect, field.TypeString, value)
	}
	if value, ok := mu.mutation.Link(); ok {
		_spec.SetField(menu.FieldLink, field.TypeString, value)
	}
	if value, ok := mu.mutation.IframeSrc(); ok {
		_spec.SetField(menu.FieldIframeSrc, field.TypeString, value)
	}
	if value, ok := mu.mutation.ActiveIcon(); ok {
		_spec.SetField(menu.FieldActiveIcon, field.TypeString, value)
	}
	if value, ok := mu.mutation.ActivePath(); ok {
		_spec.SetField(menu.FieldActivePath, field.TypeString, value)
	}
	if value, ok := mu.mutation.MaxNumOfOpenTab(); ok {
		_spec.SetField(menu.FieldMaxNumOfOpenTab, field.TypeInt16, value)
	}
	if value, ok := mu.mutation.AddedMaxNumOfOpenTab(); ok {
		_spec.AddField(menu.FieldMaxNumOfOpenTab, field.TypeInt16, value)
	}
	if value, ok := mu.mutation.IgnoreAuth(); ok {
		_spec.SetField(menu.FieldIgnoreAuth, field.TypeBool, value)
	}
	if value, ok := mu.mutation.Keepalive(); ok {
		_spec.SetField(menu.FieldKeepalive, field.TypeBool, value)
	}
	if value, ok := mu.mutation.Permission(); ok {
		_spec.SetField(menu.FieldPermission, field.TypeString, value)
	}
	if value, ok := mu.mutation.AffixTab(); ok {
		_spec.SetField(menu.FieldAffixTab, field.TypeBool, value)
	}
	if value, ok := mu.mutation.AffixTabOrder(); ok {
		_spec.SetField(menu.FieldAffixTabOrder, field.TypeInt64, value)
	}
	if value, ok := mu.mutation.AddedAffixTabOrder(); ok {
		_spec.AddField(menu.FieldAffixTabOrder, field.TypeInt64, value)
	}
	if value, ok := mu.mutation.HideInMenu(); ok {
		_spec.SetField(menu.FieldHideInMenu, field.TypeBool, value)
	}
	if value, ok := mu.mutation.HideInTab(); ok {
		_spec.SetField(menu.FieldHideInTab, field.TypeBool, value)
	}
	if value, ok := mu.mutation.HideInBreadcrumb(); ok {
		_spec.SetField(menu.FieldHideInBreadcrumb, field.TypeBool, value)
	}
	if value, ok := mu.mutation.HideChildrenInMenu(); ok {
		_spec.SetField(menu.FieldHideChildrenInMenu, field.TypeBool, value)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, mu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{menu.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	mu.mutation.done = true
	return n, nil
}

// MenuUpdateOne is the builder for updating a single Menu entity.
type MenuUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *MenuMutation
}

// SetUpdateTime sets the "update_time" field.
func (muo *MenuUpdateOne) SetUpdateTime(t time.Time) *MenuUpdateOne {
	muo.mutation.SetUpdateTime(t)
	return muo
}

// SetPid sets the "pid" field.
func (muo *MenuUpdateOne) SetPid(i int64) *MenuUpdateOne {
	muo.mutation.ResetPid()
	muo.mutation.SetPid(i)
	return muo
}

// SetNillablePid sets the "pid" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillablePid(i *int64) *MenuUpdateOne {
	if i != nil {
		muo.SetPid(*i)
	}
	return muo
}

// AddPid adds i to the "pid" field.
func (muo *MenuUpdateOne) AddPid(i int64) *MenuUpdateOne {
	muo.mutation.AddPid(i)
	return muo
}

// SetType sets the "type" field.
func (muo *MenuUpdateOne) SetType(i int8) *MenuUpdateOne {
	muo.mutation.ResetType()
	muo.mutation.SetType(i)
	return muo
}

// SetNillableType sets the "type" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableType(i *int8) *MenuUpdateOne {
	if i != nil {
		muo.SetType(*i)
	}
	return muo
}

// AddType adds i to the "type" field.
func (muo *MenuUpdateOne) AddType(i int8) *MenuUpdateOne {
	muo.mutation.AddType(i)
	return muo
}

// SetStatus sets the "status" field.
func (muo *MenuUpdateOne) SetStatus(b bool) *MenuUpdateOne {
	muo.mutation.SetStatus(b)
	return muo
}

// SetNillableStatus sets the "status" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableStatus(b *bool) *MenuUpdateOne {
	if b != nil {
		muo.SetStatus(*b)
	}
	return muo
}

// SetName sets the "name" field.
func (muo *MenuUpdateOne) SetName(s string) *MenuUpdateOne {
	muo.mutation.SetName(s)
	return muo
}

// SetNillableName sets the "name" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableName(s *string) *MenuUpdateOne {
	if s != nil {
		muo.SetName(*s)
	}
	return muo
}

// SetTitle sets the "title" field.
func (muo *MenuUpdateOne) SetTitle(s string) *MenuUpdateOne {
	muo.mutation.SetTitle(s)
	return muo
}

// SetNillableTitle sets the "title" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableTitle(s *string) *MenuUpdateOne {
	if s != nil {
		muo.SetTitle(*s)
	}
	return muo
}

// SetIcon sets the "icon" field.
func (muo *MenuUpdateOne) SetIcon(s string) *MenuUpdateOne {
	muo.mutation.SetIcon(s)
	return muo
}

// SetNillableIcon sets the "icon" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableIcon(s *string) *MenuUpdateOne {
	if s != nil {
		muo.SetIcon(*s)
	}
	return muo
}

// SetOrder sets the "order" field.
func (muo *MenuUpdateOne) SetOrder(i int32) *MenuUpdateOne {
	muo.mutation.ResetOrder()
	muo.mutation.SetOrder(i)
	return muo
}

// SetNillableOrder sets the "order" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableOrder(i *int32) *MenuUpdateOne {
	if i != nil {
		muo.SetOrder(*i)
	}
	return muo
}

// AddOrder adds i to the "order" field.
func (muo *MenuUpdateOne) AddOrder(i int32) *MenuUpdateOne {
	muo.mutation.AddOrder(i)
	return muo
}

// SetPath sets the "path" field.
func (muo *MenuUpdateOne) SetPath(s string) *MenuUpdateOne {
	muo.mutation.SetPath(s)
	return muo
}

// SetNillablePath sets the "path" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillablePath(s *string) *MenuUpdateOne {
	if s != nil {
		muo.SetPath(*s)
	}
	return muo
}

// SetComponent sets the "component" field.
func (muo *MenuUpdateOne) SetComponent(s string) *MenuUpdateOne {
	muo.mutation.SetComponent(s)
	return muo
}

// SetNillableComponent sets the "component" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableComponent(s *string) *MenuUpdateOne {
	if s != nil {
		muo.SetComponent(*s)
	}
	return muo
}

// SetRedirect sets the "redirect" field.
func (muo *MenuUpdateOne) SetRedirect(s string) *MenuUpdateOne {
	muo.mutation.SetRedirect(s)
	return muo
}

// SetNillableRedirect sets the "redirect" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableRedirect(s *string) *MenuUpdateOne {
	if s != nil {
		muo.SetRedirect(*s)
	}
	return muo
}

// SetLink sets the "link" field.
func (muo *MenuUpdateOne) SetLink(s string) *MenuUpdateOne {
	muo.mutation.SetLink(s)
	return muo
}

// SetNillableLink sets the "link" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableLink(s *string) *MenuUpdateOne {
	if s != nil {
		muo.SetLink(*s)
	}
	return muo
}

// SetIframeSrc sets the "iframeSrc" field.
func (muo *MenuUpdateOne) SetIframeSrc(s string) *MenuUpdateOne {
	muo.mutation.SetIframeSrc(s)
	return muo
}

// SetNillableIframeSrc sets the "iframeSrc" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableIframeSrc(s *string) *MenuUpdateOne {
	if s != nil {
		muo.SetIframeSrc(*s)
	}
	return muo
}

// SetActiveIcon sets the "activeIcon" field.
func (muo *MenuUpdateOne) SetActiveIcon(s string) *MenuUpdateOne {
	muo.mutation.SetActiveIcon(s)
	return muo
}

// SetNillableActiveIcon sets the "activeIcon" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableActiveIcon(s *string) *MenuUpdateOne {
	if s != nil {
		muo.SetActiveIcon(*s)
	}
	return muo
}

// SetActivePath sets the "activePath" field.
func (muo *MenuUpdateOne) SetActivePath(s string) *MenuUpdateOne {
	muo.mutation.SetActivePath(s)
	return muo
}

// SetNillableActivePath sets the "activePath" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableActivePath(s *string) *MenuUpdateOne {
	if s != nil {
		muo.SetActivePath(*s)
	}
	return muo
}

// SetMaxNumOfOpenTab sets the "maxNumOfOpenTab" field.
func (muo *MenuUpdateOne) SetMaxNumOfOpenTab(i int16) *MenuUpdateOne {
	muo.mutation.ResetMaxNumOfOpenTab()
	muo.mutation.SetMaxNumOfOpenTab(i)
	return muo
}

// SetNillableMaxNumOfOpenTab sets the "maxNumOfOpenTab" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableMaxNumOfOpenTab(i *int16) *MenuUpdateOne {
	if i != nil {
		muo.SetMaxNumOfOpenTab(*i)
	}
	return muo
}

// AddMaxNumOfOpenTab adds i to the "maxNumOfOpenTab" field.
func (muo *MenuUpdateOne) AddMaxNumOfOpenTab(i int16) *MenuUpdateOne {
	muo.mutation.AddMaxNumOfOpenTab(i)
	return muo
}

// SetIgnoreAuth sets the "ignoreAuth" field.
func (muo *MenuUpdateOne) SetIgnoreAuth(b bool) *MenuUpdateOne {
	muo.mutation.SetIgnoreAuth(b)
	return muo
}

// SetNillableIgnoreAuth sets the "ignoreAuth" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableIgnoreAuth(b *bool) *MenuUpdateOne {
	if b != nil {
		muo.SetIgnoreAuth(*b)
	}
	return muo
}

// SetKeepalive sets the "keepalive" field.
func (muo *MenuUpdateOne) SetKeepalive(b bool) *MenuUpdateOne {
	muo.mutation.SetKeepalive(b)
	return muo
}

// SetNillableKeepalive sets the "keepalive" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableKeepalive(b *bool) *MenuUpdateOne {
	if b != nil {
		muo.SetKeepalive(*b)
	}
	return muo
}

// SetPermission sets the "permission" field.
func (muo *MenuUpdateOne) SetPermission(s string) *MenuUpdateOne {
	muo.mutation.SetPermission(s)
	return muo
}

// SetNillablePermission sets the "permission" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillablePermission(s *string) *MenuUpdateOne {
	if s != nil {
		muo.SetPermission(*s)
	}
	return muo
}

// SetAffixTab sets the "affixTab" field.
func (muo *MenuUpdateOne) SetAffixTab(b bool) *MenuUpdateOne {
	muo.mutation.SetAffixTab(b)
	return muo
}

// SetNillableAffixTab sets the "affixTab" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableAffixTab(b *bool) *MenuUpdateOne {
	if b != nil {
		muo.SetAffixTab(*b)
	}
	return muo
}

// SetAffixTabOrder sets the "affixTabOrder" field.
func (muo *MenuUpdateOne) SetAffixTabOrder(i int64) *MenuUpdateOne {
	muo.mutation.ResetAffixTabOrder()
	muo.mutation.SetAffixTabOrder(i)
	return muo
}

// SetNillableAffixTabOrder sets the "affixTabOrder" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableAffixTabOrder(i *int64) *MenuUpdateOne {
	if i != nil {
		muo.SetAffixTabOrder(*i)
	}
	return muo
}

// AddAffixTabOrder adds i to the "affixTabOrder" field.
func (muo *MenuUpdateOne) AddAffixTabOrder(i int64) *MenuUpdateOne {
	muo.mutation.AddAffixTabOrder(i)
	return muo
}

// SetHideInMenu sets the "hideInMenu" field.
func (muo *MenuUpdateOne) SetHideInMenu(b bool) *MenuUpdateOne {
	muo.mutation.SetHideInMenu(b)
	return muo
}

// SetNillableHideInMenu sets the "hideInMenu" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableHideInMenu(b *bool) *MenuUpdateOne {
	if b != nil {
		muo.SetHideInMenu(*b)
	}
	return muo
}

// SetHideInTab sets the "hideInTab" field.
func (muo *MenuUpdateOne) SetHideInTab(b bool) *MenuUpdateOne {
	muo.mutation.SetHideInTab(b)
	return muo
}

// SetNillableHideInTab sets the "hideInTab" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableHideInTab(b *bool) *MenuUpdateOne {
	if b != nil {
		muo.SetHideInTab(*b)
	}
	return muo
}

// SetHideInBreadcrumb sets the "hideInBreadcrumb" field.
func (muo *MenuUpdateOne) SetHideInBreadcrumb(b bool) *MenuUpdateOne {
	muo.mutation.SetHideInBreadcrumb(b)
	return muo
}

// SetNillableHideInBreadcrumb sets the "hideInBreadcrumb" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableHideInBreadcrumb(b *bool) *MenuUpdateOne {
	if b != nil {
		muo.SetHideInBreadcrumb(*b)
	}
	return muo
}

// SetHideChildrenInMenu sets the "hideChildrenInMenu" field.
func (muo *MenuUpdateOne) SetHideChildrenInMenu(b bool) *MenuUpdateOne {
	muo.mutation.SetHideChildrenInMenu(b)
	return muo
}

// SetNillableHideChildrenInMenu sets the "hideChildrenInMenu" field if the given value is not nil.
func (muo *MenuUpdateOne) SetNillableHideChildrenInMenu(b *bool) *MenuUpdateOne {
	if b != nil {
		muo.SetHideChildrenInMenu(*b)
	}
	return muo
}

// Mutation returns the MenuMutation object of the builder.
func (muo *MenuUpdateOne) Mutation() *MenuMutation {
	return muo.mutation
}

// Where appends a list predicates to the MenuUpdate builder.
func (muo *MenuUpdateOne) Where(ps ...predicate.Menu) *MenuUpdateOne {
	muo.mutation.Where(ps...)
	return muo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (muo *MenuUpdateOne) Select(field string, fields ...string) *MenuUpdateOne {
	muo.fields = append([]string{field}, fields...)
	return muo
}

// Save executes the query and returns the updated Menu entity.
func (muo *MenuUpdateOne) Save(ctx context.Context) (*Menu, error) {
	muo.defaults()
	return withHooks(ctx, muo.sqlSave, muo.mutation, muo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (muo *MenuUpdateOne) SaveX(ctx context.Context) *Menu {
	node, err := muo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (muo *MenuUpdateOne) Exec(ctx context.Context) error {
	_, err := muo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (muo *MenuUpdateOne) ExecX(ctx context.Context) {
	if err := muo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (muo *MenuUpdateOne) defaults() {
	if _, ok := muo.mutation.UpdateTime(); !ok {
		v := menu.UpdateDefaultUpdateTime()
		muo.mutation.SetUpdateTime(v)
	}
}

func (muo *MenuUpdateOne) sqlSave(ctx context.Context) (_node *Menu, err error) {
	_spec := sqlgraph.NewUpdateSpec(menu.Table, menu.Columns, sqlgraph.NewFieldSpec(menu.FieldID, field.TypeInt64))
	id, ok := muo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Menu.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := muo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, menu.FieldID)
		for _, f := range fields {
			if !menu.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != menu.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := muo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := muo.mutation.UpdateTime(); ok {
		_spec.SetField(menu.FieldUpdateTime, field.TypeTime, value)
	}
	if value, ok := muo.mutation.Pid(); ok {
		_spec.SetField(menu.FieldPid, field.TypeInt64, value)
	}
	if value, ok := muo.mutation.AddedPid(); ok {
		_spec.AddField(menu.FieldPid, field.TypeInt64, value)
	}
	if value, ok := muo.mutation.GetType(); ok {
		_spec.SetField(menu.FieldType, field.TypeInt8, value)
	}
	if value, ok := muo.mutation.AddedType(); ok {
		_spec.AddField(menu.FieldType, field.TypeInt8, value)
	}
	if value, ok := muo.mutation.Status(); ok {
		_spec.SetField(menu.FieldStatus, field.TypeBool, value)
	}
	if value, ok := muo.mutation.Name(); ok {
		_spec.SetField(menu.FieldName, field.TypeString, value)
	}
	if value, ok := muo.mutation.Title(); ok {
		_spec.SetField(menu.FieldTitle, field.TypeString, value)
	}
	if value, ok := muo.mutation.Icon(); ok {
		_spec.SetField(menu.FieldIcon, field.TypeString, value)
	}
	if value, ok := muo.mutation.Order(); ok {
		_spec.SetField(menu.FieldOrder, field.TypeInt32, value)
	}
	if value, ok := muo.mutation.AddedOrder(); ok {
		_spec.AddField(menu.FieldOrder, field.TypeInt32, value)
	}
	if value, ok := muo.mutation.Path(); ok {
		_spec.SetField(menu.FieldPath, field.TypeString, value)
	}
	if value, ok := muo.mutation.Component(); ok {
		_spec.SetField(menu.FieldComponent, field.TypeString, value)
	}
	if value, ok := muo.mutation.Redirect(); ok {
		_spec.SetField(menu.FieldRedirect, field.TypeString, value)
	}
	if value, ok := muo.mutation.Link(); ok {
		_spec.SetField(menu.FieldLink, field.TypeString, value)
	}
	if value, ok := muo.mutation.IframeSrc(); ok {
		_spec.SetField(menu.FieldIframeSrc, field.TypeString, value)
	}
	if value, ok := muo.mutation.ActiveIcon(); ok {
		_spec.SetField(menu.FieldActiveIcon, field.TypeString, value)
	}
	if value, ok := muo.mutation.ActivePath(); ok {
		_spec.SetField(menu.FieldActivePath, field.TypeString, value)
	}
	if value, ok := muo.mutation.MaxNumOfOpenTab(); ok {
		_spec.SetField(menu.FieldMaxNumOfOpenTab, field.TypeInt16, value)
	}
	if value, ok := muo.mutation.AddedMaxNumOfOpenTab(); ok {
		_spec.AddField(menu.FieldMaxNumOfOpenTab, field.TypeInt16, value)
	}
	if value, ok := muo.mutation.IgnoreAuth(); ok {
		_spec.SetField(menu.FieldIgnoreAuth, field.TypeBool, value)
	}
	if value, ok := muo.mutation.Keepalive(); ok {
		_spec.SetField(menu.FieldKeepalive, field.TypeBool, value)
	}
	if value, ok := muo.mutation.Permission(); ok {
		_spec.SetField(menu.FieldPermission, field.TypeString, value)
	}
	if value, ok := muo.mutation.AffixTab(); ok {
		_spec.SetField(menu.FieldAffixTab, field.TypeBool, value)
	}
	if value, ok := muo.mutation.AffixTabOrder(); ok {
		_spec.SetField(menu.FieldAffixTabOrder, field.TypeInt64, value)
	}
	if value, ok := muo.mutation.AddedAffixTabOrder(); ok {
		_spec.AddField(menu.FieldAffixTabOrder, field.TypeInt64, value)
	}
	if value, ok := muo.mutation.HideInMenu(); ok {
		_spec.SetField(menu.FieldHideInMenu, field.TypeBool, value)
	}
	if value, ok := muo.mutation.HideInTab(); ok {
		_spec.SetField(menu.FieldHideInTab, field.TypeBool, value)
	}
	if value, ok := muo.mutation.HideInBreadcrumb(); ok {
		_spec.SetField(menu.FieldHideInBreadcrumb, field.TypeBool, value)
	}
	if value, ok := muo.mutation.HideChildrenInMenu(); ok {
		_spec.SetField(menu.FieldHideChildrenInMenu, field.TypeBool, value)
	}
	_node = &Menu{config: muo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, muo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{menu.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	muo.mutation.done = true
	return _node, nil
}
