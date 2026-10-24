package assembler

import (
	"time"

	identityv1 "github.com/example/scaffold/services/api/identity/v1"
	pluginv1 "github.com/example/scaffold/services/api/plugin/v1"
	"github.com/example/scaffold/services/internal/biz"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserToProto converts a domain user into its transport shape.
func UserToProto(u *biz.User) *identityv1.User {
	if u == nil {
		return nil
	}
	return &identityv1.User{
		Id:                 u.ID,
		Username:           u.Username,
		DisplayName:        u.DisplayName,
		Email:              u.Email,
		IsAdmin:            u.IsAdmin,
		MustChangePassword: u.MustChangePassword,
		Status:             u.Status,
		Roles:              u.Roles,
		CreatedAt:          millisToTimestamp(u.CreatedAt),
	}
}

// UserFromProto converts a transport user into the domain shape. Conversions
// stay explicit so that new fields are a compile-time reminder rather than a
// silent omission.
func UserFromProto(u *identityv1.User) *biz.User {
	if u == nil {
		return nil
	}
	return &biz.User{
		ID:          u.GetId(),
		Username:    u.GetUsername(),
		DisplayName: u.GetDisplayName(),
		Email:       u.GetEmail(),
		IsAdmin:     u.GetIsAdmin(),
		Status:      u.GetStatus(),
		Roles:       u.GetRoles(),
	}
}

// UsersToProto converts a page of domain users.
func UsersToProto(items []*biz.User) []*identityv1.User {
	out := make([]*identityv1.User, 0, len(items))
	for _, item := range items {
		out = append(out, UserToProto(item))
	}
	return out
}

// RoleToProto converts a domain role into its transport shape.
func RoleToProto(r *biz.Role) *identityv1.Role {
	if r == nil {
		return nil
	}
	return &identityv1.Role{
		Id:          r.ID,
		Name:        r.Name,
		DisplayName: r.DisplayName,
		Description: r.Description,
		IsSystem:    r.IsSystem,
		Permissions: r.Permissions,
	}
}

// RoleFromProto converts a transport role into the domain shape.
func RoleFromProto(r *identityv1.Role) *biz.Role {
	if r == nil {
		return nil
	}
	return &biz.Role{
		ID:          r.GetId(),
		Name:        r.GetName(),
		DisplayName: r.GetDisplayName(),
		Description: r.GetDescription(),
		IsSystem:    r.GetIsSystem(),
		Permissions: r.GetPermissions(),
	}
}

// RolesToProto converts a page of domain roles.
func RolesToProto(items []*biz.Role) []*identityv1.Role {
	out := make([]*identityv1.Role, 0, len(items))
	for _, item := range items {
		out = append(out, RoleToProto(item))
	}
	return out
}

// PermissionToProto converts a domain permission into its transport shape.
func PermissionToProto(p *biz.Permission) *identityv1.Permission {
	if p == nil {
		return nil
	}
	return &identityv1.Permission{
		Code:     p.Code,
		Name:     p.Name,
		Resource: p.Resource,
		Action:   p.Action,
	}
}

// PermissionsToProto converts a list of permissions.
func PermissionsToProto(items []*biz.Permission) []*identityv1.Permission {
	out := make([]*identityv1.Permission, 0, len(items))
	for _, item := range items {
		out = append(out, PermissionToProto(item))
	}
	return out
}

// PluginToProto converts a plugin manifest view into its transport shape.
func PluginToProto(p *biz.PluginInfo) *pluginv1.Plugin {
	if p == nil {
		return nil
	}
	routes := make([]*pluginv1.PluginRoute, 0, len(p.Routes))
	for _, r := range p.Routes {
		routes = append(routes, &pluginv1.PluginRoute{
			Path:       r.Path,
			Name:       r.Name,
			Component:  r.Component,
			Permission: r.Permission,
		})
	}
	navs := make([]*pluginv1.PluginNav, 0, len(p.Navigations))
	for _, n := range p.Navigations {
		navs = append(navs, &pluginv1.PluginNav{
			Label:      n.Label,
			Icon:       n.Icon,
			Order:      n.Order,
			Path:       n.Path,
			Permission: n.Permission,
		})
	}
	return &pluginv1.Plugin{
		Id:          p.ID,
		Name:        p.Name,
		Version:     p.Version,
		Description: p.Description,
		Icon:        p.Icon,
		EntryUrl:    p.EntryURL,
		Styles:      p.Styles,
		Routes:      routes,
		Navigations: navs,
		Permissions: p.Permissions,
	}
}

func millisToTimestamp(millis int64) *timestamppb.Timestamp {
	if millis <= 0 {
		return nil
	}
	return timestamppb.New(time.UnixMilli(millis))
}
