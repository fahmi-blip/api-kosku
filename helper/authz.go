package helper

import (
	"sort"

	"kosku-api/app/model"
)

type PermissionSet struct {
	byRole map[string]map[string]struct{}
}

func NewPermissionSet(raw map[string][]string) *PermissionSet {
	byRole := make(map[string]map[string]struct{}, len(raw))
	for role, permissions := range raw {
		set := make(map[string]struct{}, len(permissions))
		for _, permission := range permissions {
			set[permission] = struct{}{}
		}
		byRole[role] = set
	}
	return &PermissionSet{byRole: byRole}
}

func DefaultPermissionSet() *PermissionSet {
	return NewPermissionSet(map[string][]string{
		model.RolePemilik: {
			"room:read",
			"room:create",
			"room:update",
			"room:delete",
			"rental:read",
			"rental:approve",
			"rental:reject",
		},
		model.RolePenghuni: {
			"room:read",
			"rental:apply",
			"rental:read",
			"rental:cancel",
		},
	})
}

func (p *PermissionSet) Can(role, permission string) bool {
	if p == nil {
		return false
	}
	permissions, ok := p.byRole[role]
	if !ok {
		return false
	}
	_, granted := permissions[permission]
	return granted
}

func (p *PermissionSet) PermissionsOf(role string) []string {
	result := []string{}
	if p == nil {
		return result
	}
	for permission := range p.byRole[role] {
		result = append(result, permission)
	}
	sort.Strings(result)
	return result
}
