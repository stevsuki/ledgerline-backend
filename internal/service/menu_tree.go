package service

import (
	"sort"

	"github.com/google/uuid"

	"github.com/stevensuki/ledgerline-backend/internal/domain"
)

// visibleMenuTree: flat menus -> two-level sidebar tree, dropping groups with no readable child.
func visibleMenuTree(menus []domain.Menu) []domain.Menu {
	roots := make([]domain.Menu, 0, len(menus))
	hasChildren := make(map[uuid.UUID]bool)
	readable := make(map[uuid.UUID][]domain.Menu)

	for _, m := range menus {
		if m.ParentID == nil {
			roots = append(roots, m)
			continue
		}
		hasChildren[*m.ParentID] = true
		if m.Access.CanRead {
			readable[*m.ParentID] = append(readable[*m.ParentID], m)
		}
	}

	out := make([]domain.Menu, 0, len(roots))
	for _, root := range roots {
		if !hasChildren[root.ID] {
			// A root with no children at all is a page, not a group.
			if root.Access.CanRead {
				out = append(out, root)
			}
			continue
		}

		children := readable[root.ID]
		if len(children) == 0 {
			continue
		}
		sortMenus(children)
		root.Children = children
		out = append(out, root)
	}

	sortMenus(out)
	return out
}

// sortMenus: display order, with code as tie-breaker so equal sort_order stays stable.
func sortMenus(menus []domain.Menu) {
	sort.SliceStable(menus, func(i, j int) bool {
		if menus[i].SortOrder != menus[j].SortOrder {
			return menus[i].SortOrder < menus[j].SortOrder
		}
		return menus[i].Code < menus[j].Code
	})
}
