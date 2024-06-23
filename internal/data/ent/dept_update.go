// Code generated by ent, DO NOT EDIT.

package ent

import (
	"base-server/internal/data/ent/dept"
	"base-server/internal/data/ent/predicate"
	"base-server/internal/data/ent/role"
	"base-server/internal/data/ent/user"
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// DeptUpdate is the builder for updating Dept entities.
type DeptUpdate struct {
	config
	hooks    []Hook
	mutation *DeptMutation
}

// Where appends a list predicates to the DeptUpdate builder.
func (du *DeptUpdate) Where(ps ...predicate.Dept) *DeptUpdate {
	du.mutation.Where(ps...)
	return du
}

// SetUpdateTime sets the "update_time" field.
func (du *DeptUpdate) SetUpdateTime(t time.Time) *DeptUpdate {
	du.mutation.SetUpdateTime(t)
	return du
}

// SetName sets the "name" field.
func (du *DeptUpdate) SetName(s string) *DeptUpdate {
	du.mutation.SetName(s)
	return du
}

// SetSort sets the "sort" field.
func (du *DeptUpdate) SetSort(i int) *DeptUpdate {
	du.mutation.ResetSort()
	du.mutation.SetSort(i)
	return du
}

// AddSort adds i to the "sort" field.
func (du *DeptUpdate) AddSort(i int) *DeptUpdate {
	du.mutation.AddSort(i)
	return du
}

// SetStatus sets the "status" field.
func (du *DeptUpdate) SetStatus(b bool) *DeptUpdate {
	du.mutation.SetStatus(b)
	return du
}

// SetDesc sets the "desc" field.
func (du *DeptUpdate) SetDesc(s string) *DeptUpdate {
	du.mutation.SetDesc(s)
	return du
}

// SetExtension sets the "extension" field.
func (du *DeptUpdate) SetExtension(s string) *DeptUpdate {
	du.mutation.SetExtension(s)
	return du
}

// SetDom sets the "dom" field.
func (du *DeptUpdate) SetDom(i int64) *DeptUpdate {
	du.mutation.ResetDom()
	du.mutation.SetDom(i)
	return du
}

// AddDom adds i to the "dom" field.
func (du *DeptUpdate) AddDom(i int64) *DeptUpdate {
	du.mutation.AddDom(i)
	return du
}

// SetPid sets the "pid" field.
func (du *DeptUpdate) SetPid(i int64) *DeptUpdate {
	du.mutation.SetPid(i)
	return du
}

// SetNillablePid sets the "pid" field if the given value is not nil.
func (du *DeptUpdate) SetNillablePid(i *int64) *DeptUpdate {
	if i != nil {
		du.SetPid(*i)
	}
	return du
}

// ClearPid clears the value of the "pid" field.
func (du *DeptUpdate) ClearPid() *DeptUpdate {
	du.mutation.ClearPid()
	return du
}

// AddUserIDs adds the "users" edge to the User entity by IDs.
func (du *DeptUpdate) AddUserIDs(ids ...uuid.UUID) *DeptUpdate {
	du.mutation.AddUserIDs(ids...)
	return du
}

// AddUsers adds the "users" edges to the User entity.
func (du *DeptUpdate) AddUsers(u ...*User) *DeptUpdate {
	ids := make([]uuid.UUID, len(u))
	for i := range u {
		ids[i] = u[i].ID
	}
	return du.AddUserIDs(ids...)
}

// SetRolesID sets the "roles" edge to the Role entity by ID.
func (du *DeptUpdate) SetRolesID(id int64) *DeptUpdate {
	du.mutation.SetRolesID(id)
	return du
}

// SetNillableRolesID sets the "roles" edge to the Role entity by ID if the given value is not nil.
func (du *DeptUpdate) SetNillableRolesID(id *int64) *DeptUpdate {
	if id != nil {
		du = du.SetRolesID(*id)
	}
	return du
}

// SetRoles sets the "roles" edge to the Role entity.
func (du *DeptUpdate) SetRoles(r *Role) *DeptUpdate {
	return du.SetRolesID(r.ID)
}

// SetParentID sets the "parent" edge to the Dept entity by ID.
func (du *DeptUpdate) SetParentID(id int64) *DeptUpdate {
	du.mutation.SetParentID(id)
	return du
}

// SetNillableParentID sets the "parent" edge to the Dept entity by ID if the given value is not nil.
func (du *DeptUpdate) SetNillableParentID(id *int64) *DeptUpdate {
	if id != nil {
		du = du.SetParentID(*id)
	}
	return du
}

// SetParent sets the "parent" edge to the Dept entity.
func (du *DeptUpdate) SetParent(d *Dept) *DeptUpdate {
	return du.SetParentID(d.ID)
}

// AddChildIDs adds the "children" edge to the Dept entity by IDs.
func (du *DeptUpdate) AddChildIDs(ids ...int64) *DeptUpdate {
	du.mutation.AddChildIDs(ids...)
	return du
}

// AddChildren adds the "children" edges to the Dept entity.
func (du *DeptUpdate) AddChildren(d ...*Dept) *DeptUpdate {
	ids := make([]int64, len(d))
	for i := range d {
		ids[i] = d[i].ID
	}
	return du.AddChildIDs(ids...)
}

// Mutation returns the DeptMutation object of the builder.
func (du *DeptUpdate) Mutation() *DeptMutation {
	return du.mutation
}

// ClearUsers clears all "users" edges to the User entity.
func (du *DeptUpdate) ClearUsers() *DeptUpdate {
	du.mutation.ClearUsers()
	return du
}

// RemoveUserIDs removes the "users" edge to User entities by IDs.
func (du *DeptUpdate) RemoveUserIDs(ids ...uuid.UUID) *DeptUpdate {
	du.mutation.RemoveUserIDs(ids...)
	return du
}

// RemoveUsers removes "users" edges to User entities.
func (du *DeptUpdate) RemoveUsers(u ...*User) *DeptUpdate {
	ids := make([]uuid.UUID, len(u))
	for i := range u {
		ids[i] = u[i].ID
	}
	return du.RemoveUserIDs(ids...)
}

// ClearRoles clears the "roles" edge to the Role entity.
func (du *DeptUpdate) ClearRoles() *DeptUpdate {
	du.mutation.ClearRoles()
	return du
}

// ClearParent clears the "parent" edge to the Dept entity.
func (du *DeptUpdate) ClearParent() *DeptUpdate {
	du.mutation.ClearParent()
	return du
}

// ClearChildren clears all "children" edges to the Dept entity.
func (du *DeptUpdate) ClearChildren() *DeptUpdate {
	du.mutation.ClearChildren()
	return du
}

// RemoveChildIDs removes the "children" edge to Dept entities by IDs.
func (du *DeptUpdate) RemoveChildIDs(ids ...int64) *DeptUpdate {
	du.mutation.RemoveChildIDs(ids...)
	return du
}

// RemoveChildren removes "children" edges to Dept entities.
func (du *DeptUpdate) RemoveChildren(d ...*Dept) *DeptUpdate {
	ids := make([]int64, len(d))
	for i := range d {
		ids[i] = d[i].ID
	}
	return du.RemoveChildIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (du *DeptUpdate) Save(ctx context.Context) (int, error) {
	du.defaults()
	return withHooks(ctx, du.sqlSave, du.mutation, du.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (du *DeptUpdate) SaveX(ctx context.Context) int {
	affected, err := du.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (du *DeptUpdate) Exec(ctx context.Context) error {
	_, err := du.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (du *DeptUpdate) ExecX(ctx context.Context) {
	if err := du.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (du *DeptUpdate) defaults() {
	if _, ok := du.mutation.UpdateTime(); !ok {
		v := dept.UpdateDefaultUpdateTime()
		du.mutation.SetUpdateTime(v)
	}
}

func (du *DeptUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(dept.Table, dept.Columns, sqlgraph.NewFieldSpec(dept.FieldID, field.TypeInt64))
	if ps := du.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := du.mutation.UpdateTime(); ok {
		_spec.SetField(dept.FieldUpdateTime, field.TypeTime, value)
	}
	if value, ok := du.mutation.Name(); ok {
		_spec.SetField(dept.FieldName, field.TypeString, value)
	}
	if value, ok := du.mutation.Sort(); ok {
		_spec.SetField(dept.FieldSort, field.TypeInt, value)
	}
	if value, ok := du.mutation.AddedSort(); ok {
		_spec.AddField(dept.FieldSort, field.TypeInt, value)
	}
	if value, ok := du.mutation.Status(); ok {
		_spec.SetField(dept.FieldStatus, field.TypeBool, value)
	}
	if value, ok := du.mutation.Desc(); ok {
		_spec.SetField(dept.FieldDesc, field.TypeString, value)
	}
	if value, ok := du.mutation.Extension(); ok {
		_spec.SetField(dept.FieldExtension, field.TypeString, value)
	}
	if value, ok := du.mutation.Dom(); ok {
		_spec.SetField(dept.FieldDom, field.TypeInt64, value)
	}
	if value, ok := du.mutation.AddedDom(); ok {
		_spec.AddField(dept.FieldDom, field.TypeInt64, value)
	}
	if du.mutation.UsersCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   dept.UsersTable,
			Columns: dept.UsersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := du.mutation.RemovedUsersIDs(); len(nodes) > 0 && !du.mutation.UsersCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   dept.UsersTable,
			Columns: dept.UsersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := du.mutation.UsersIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   dept.UsersTable,
			Columns: dept.UsersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if du.mutation.RolesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   dept.RolesTable,
			Columns: []string{dept.RolesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(role.FieldID, field.TypeInt64),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := du.mutation.RolesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   dept.RolesTable,
			Columns: []string{dept.RolesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(role.FieldID, field.TypeInt64),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if du.mutation.ParentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   dept.ParentTable,
			Columns: []string{dept.ParentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(dept.FieldID, field.TypeInt64),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := du.mutation.ParentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   dept.ParentTable,
			Columns: []string{dept.ParentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(dept.FieldID, field.TypeInt64),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if du.mutation.ChildrenCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   dept.ChildrenTable,
			Columns: []string{dept.ChildrenColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(dept.FieldID, field.TypeInt64),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := du.mutation.RemovedChildrenIDs(); len(nodes) > 0 && !du.mutation.ChildrenCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   dept.ChildrenTable,
			Columns: []string{dept.ChildrenColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(dept.FieldID, field.TypeInt64),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := du.mutation.ChildrenIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   dept.ChildrenTable,
			Columns: []string{dept.ChildrenColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(dept.FieldID, field.TypeInt64),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, du.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{dept.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	du.mutation.done = true
	return n, nil
}

// DeptUpdateOne is the builder for updating a single Dept entity.
type DeptUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *DeptMutation
}

// SetUpdateTime sets the "update_time" field.
func (duo *DeptUpdateOne) SetUpdateTime(t time.Time) *DeptUpdateOne {
	duo.mutation.SetUpdateTime(t)
	return duo
}

// SetName sets the "name" field.
func (duo *DeptUpdateOne) SetName(s string) *DeptUpdateOne {
	duo.mutation.SetName(s)
	return duo
}

// SetSort sets the "sort" field.
func (duo *DeptUpdateOne) SetSort(i int) *DeptUpdateOne {
	duo.mutation.ResetSort()
	duo.mutation.SetSort(i)
	return duo
}

// AddSort adds i to the "sort" field.
func (duo *DeptUpdateOne) AddSort(i int) *DeptUpdateOne {
	duo.mutation.AddSort(i)
	return duo
}

// SetStatus sets the "status" field.
func (duo *DeptUpdateOne) SetStatus(b bool) *DeptUpdateOne {
	duo.mutation.SetStatus(b)
	return duo
}

// SetDesc sets the "desc" field.
func (duo *DeptUpdateOne) SetDesc(s string) *DeptUpdateOne {
	duo.mutation.SetDesc(s)
	return duo
}

// SetExtension sets the "extension" field.
func (duo *DeptUpdateOne) SetExtension(s string) *DeptUpdateOne {
	duo.mutation.SetExtension(s)
	return duo
}

// SetDom sets the "dom" field.
func (duo *DeptUpdateOne) SetDom(i int64) *DeptUpdateOne {
	duo.mutation.ResetDom()
	duo.mutation.SetDom(i)
	return duo
}

// AddDom adds i to the "dom" field.
func (duo *DeptUpdateOne) AddDom(i int64) *DeptUpdateOne {
	duo.mutation.AddDom(i)
	return duo
}

// SetPid sets the "pid" field.
func (duo *DeptUpdateOne) SetPid(i int64) *DeptUpdateOne {
	duo.mutation.SetPid(i)
	return duo
}

// SetNillablePid sets the "pid" field if the given value is not nil.
func (duo *DeptUpdateOne) SetNillablePid(i *int64) *DeptUpdateOne {
	if i != nil {
		duo.SetPid(*i)
	}
	return duo
}

// ClearPid clears the value of the "pid" field.
func (duo *DeptUpdateOne) ClearPid() *DeptUpdateOne {
	duo.mutation.ClearPid()
	return duo
}

// AddUserIDs adds the "users" edge to the User entity by IDs.
func (duo *DeptUpdateOne) AddUserIDs(ids ...uuid.UUID) *DeptUpdateOne {
	duo.mutation.AddUserIDs(ids...)
	return duo
}

// AddUsers adds the "users" edges to the User entity.
func (duo *DeptUpdateOne) AddUsers(u ...*User) *DeptUpdateOne {
	ids := make([]uuid.UUID, len(u))
	for i := range u {
		ids[i] = u[i].ID
	}
	return duo.AddUserIDs(ids...)
}

// SetRolesID sets the "roles" edge to the Role entity by ID.
func (duo *DeptUpdateOne) SetRolesID(id int64) *DeptUpdateOne {
	duo.mutation.SetRolesID(id)
	return duo
}

// SetNillableRolesID sets the "roles" edge to the Role entity by ID if the given value is not nil.
func (duo *DeptUpdateOne) SetNillableRolesID(id *int64) *DeptUpdateOne {
	if id != nil {
		duo = duo.SetRolesID(*id)
	}
	return duo
}

// SetRoles sets the "roles" edge to the Role entity.
func (duo *DeptUpdateOne) SetRoles(r *Role) *DeptUpdateOne {
	return duo.SetRolesID(r.ID)
}

// SetParentID sets the "parent" edge to the Dept entity by ID.
func (duo *DeptUpdateOne) SetParentID(id int64) *DeptUpdateOne {
	duo.mutation.SetParentID(id)
	return duo
}

// SetNillableParentID sets the "parent" edge to the Dept entity by ID if the given value is not nil.
func (duo *DeptUpdateOne) SetNillableParentID(id *int64) *DeptUpdateOne {
	if id != nil {
		duo = duo.SetParentID(*id)
	}
	return duo
}

// SetParent sets the "parent" edge to the Dept entity.
func (duo *DeptUpdateOne) SetParent(d *Dept) *DeptUpdateOne {
	return duo.SetParentID(d.ID)
}

// AddChildIDs adds the "children" edge to the Dept entity by IDs.
func (duo *DeptUpdateOne) AddChildIDs(ids ...int64) *DeptUpdateOne {
	duo.mutation.AddChildIDs(ids...)
	return duo
}

// AddChildren adds the "children" edges to the Dept entity.
func (duo *DeptUpdateOne) AddChildren(d ...*Dept) *DeptUpdateOne {
	ids := make([]int64, len(d))
	for i := range d {
		ids[i] = d[i].ID
	}
	return duo.AddChildIDs(ids...)
}

// Mutation returns the DeptMutation object of the builder.
func (duo *DeptUpdateOne) Mutation() *DeptMutation {
	return duo.mutation
}

// ClearUsers clears all "users" edges to the User entity.
func (duo *DeptUpdateOne) ClearUsers() *DeptUpdateOne {
	duo.mutation.ClearUsers()
	return duo
}

// RemoveUserIDs removes the "users" edge to User entities by IDs.
func (duo *DeptUpdateOne) RemoveUserIDs(ids ...uuid.UUID) *DeptUpdateOne {
	duo.mutation.RemoveUserIDs(ids...)
	return duo
}

// RemoveUsers removes "users" edges to User entities.
func (duo *DeptUpdateOne) RemoveUsers(u ...*User) *DeptUpdateOne {
	ids := make([]uuid.UUID, len(u))
	for i := range u {
		ids[i] = u[i].ID
	}
	return duo.RemoveUserIDs(ids...)
}

// ClearRoles clears the "roles" edge to the Role entity.
func (duo *DeptUpdateOne) ClearRoles() *DeptUpdateOne {
	duo.mutation.ClearRoles()
	return duo
}

// ClearParent clears the "parent" edge to the Dept entity.
func (duo *DeptUpdateOne) ClearParent() *DeptUpdateOne {
	duo.mutation.ClearParent()
	return duo
}

// ClearChildren clears all "children" edges to the Dept entity.
func (duo *DeptUpdateOne) ClearChildren() *DeptUpdateOne {
	duo.mutation.ClearChildren()
	return duo
}

// RemoveChildIDs removes the "children" edge to Dept entities by IDs.
func (duo *DeptUpdateOne) RemoveChildIDs(ids ...int64) *DeptUpdateOne {
	duo.mutation.RemoveChildIDs(ids...)
	return duo
}

// RemoveChildren removes "children" edges to Dept entities.
func (duo *DeptUpdateOne) RemoveChildren(d ...*Dept) *DeptUpdateOne {
	ids := make([]int64, len(d))
	for i := range d {
		ids[i] = d[i].ID
	}
	return duo.RemoveChildIDs(ids...)
}

// Where appends a list predicates to the DeptUpdate builder.
func (duo *DeptUpdateOne) Where(ps ...predicate.Dept) *DeptUpdateOne {
	duo.mutation.Where(ps...)
	return duo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (duo *DeptUpdateOne) Select(field string, fields ...string) *DeptUpdateOne {
	duo.fields = append([]string{field}, fields...)
	return duo
}

// Save executes the query and returns the updated Dept entity.
func (duo *DeptUpdateOne) Save(ctx context.Context) (*Dept, error) {
	duo.defaults()
	return withHooks(ctx, duo.sqlSave, duo.mutation, duo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (duo *DeptUpdateOne) SaveX(ctx context.Context) *Dept {
	node, err := duo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (duo *DeptUpdateOne) Exec(ctx context.Context) error {
	_, err := duo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (duo *DeptUpdateOne) ExecX(ctx context.Context) {
	if err := duo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (duo *DeptUpdateOne) defaults() {
	if _, ok := duo.mutation.UpdateTime(); !ok {
		v := dept.UpdateDefaultUpdateTime()
		duo.mutation.SetUpdateTime(v)
	}
}

func (duo *DeptUpdateOne) sqlSave(ctx context.Context) (_node *Dept, err error) {
	_spec := sqlgraph.NewUpdateSpec(dept.Table, dept.Columns, sqlgraph.NewFieldSpec(dept.FieldID, field.TypeInt64))
	id, ok := duo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Dept.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := duo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, dept.FieldID)
		for _, f := range fields {
			if !dept.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != dept.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := duo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := duo.mutation.UpdateTime(); ok {
		_spec.SetField(dept.FieldUpdateTime, field.TypeTime, value)
	}
	if value, ok := duo.mutation.Name(); ok {
		_spec.SetField(dept.FieldName, field.TypeString, value)
	}
	if value, ok := duo.mutation.Sort(); ok {
		_spec.SetField(dept.FieldSort, field.TypeInt, value)
	}
	if value, ok := duo.mutation.AddedSort(); ok {
		_spec.AddField(dept.FieldSort, field.TypeInt, value)
	}
	if value, ok := duo.mutation.Status(); ok {
		_spec.SetField(dept.FieldStatus, field.TypeBool, value)
	}
	if value, ok := duo.mutation.Desc(); ok {
		_spec.SetField(dept.FieldDesc, field.TypeString, value)
	}
	if value, ok := duo.mutation.Extension(); ok {
		_spec.SetField(dept.FieldExtension, field.TypeString, value)
	}
	if value, ok := duo.mutation.Dom(); ok {
		_spec.SetField(dept.FieldDom, field.TypeInt64, value)
	}
	if value, ok := duo.mutation.AddedDom(); ok {
		_spec.AddField(dept.FieldDom, field.TypeInt64, value)
	}
	if duo.mutation.UsersCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   dept.UsersTable,
			Columns: dept.UsersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := duo.mutation.RemovedUsersIDs(); len(nodes) > 0 && !duo.mutation.UsersCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   dept.UsersTable,
			Columns: dept.UsersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := duo.mutation.UsersIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   dept.UsersTable,
			Columns: dept.UsersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if duo.mutation.RolesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   dept.RolesTable,
			Columns: []string{dept.RolesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(role.FieldID, field.TypeInt64),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := duo.mutation.RolesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: false,
			Table:   dept.RolesTable,
			Columns: []string{dept.RolesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(role.FieldID, field.TypeInt64),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if duo.mutation.ParentCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   dept.ParentTable,
			Columns: []string{dept.ParentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(dept.FieldID, field.TypeInt64),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := duo.mutation.ParentIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   dept.ParentTable,
			Columns: []string{dept.ParentColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(dept.FieldID, field.TypeInt64),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if duo.mutation.ChildrenCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   dept.ChildrenTable,
			Columns: []string{dept.ChildrenColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(dept.FieldID, field.TypeInt64),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := duo.mutation.RemovedChildrenIDs(); len(nodes) > 0 && !duo.mutation.ChildrenCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   dept.ChildrenTable,
			Columns: []string{dept.ChildrenColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(dept.FieldID, field.TypeInt64),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := duo.mutation.ChildrenIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   dept.ChildrenTable,
			Columns: []string{dept.ChildrenColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(dept.FieldID, field.TypeInt64),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_node = &Dept{config: duo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, duo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{dept.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	duo.mutation.done = true
	return _node, nil
}
