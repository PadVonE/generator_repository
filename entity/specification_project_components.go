package entity

import "sort"

type Property struct {
	Name     string     `json:"name"`
	Type     string     `json:"type"`
	Children []Property `json:"children"`
}

type OperationInfo struct {
	NameMethod string             `json:"name_method"`
	Request    []Property         `json:"request"`
	Responses  map[int][]Property `json:"responses"`
	Tag        string             `json:"tag"`
}

type PathInfo struct {
	Path       string          `json:"path"`
	Operations []OperationInfo `json:"operations"`
}

type SpecificationProjectComponents struct {
	Name    string     `json:"name"`
	Version string     `json:"version"`
	Path    []PathInfo `json:"path"`
}

func orderFunc(p Property) int {
	switch p.Name {
	case "ID", "Id", "id":
		return 0
	case "UUID", "Uuid", "uuid":
		return 1
	case "CreatedAt", "created_at":
		return 2
	case "UpdatedAt", "updated_at":
		return 3
	case "DeletedAt", "deleted_at":
		return 4
	default:
		switch p.Type {
		case "int32", "int64", "int":
			return 5
		case "float32", "float64":
			return 6
		case "string":
			return 7
		case "array":
			return 8
		}

		return 9
	}
}

func SortProperties(properties []Property) {
	sort.Slice(properties, func(i, j int) bool {
		orderI := orderFunc(properties[i])
		orderJ := orderFunc(properties[j])
		if orderI == orderJ {
			if properties[i].Type == properties[j].Type {
				return properties[i].Name < properties[j].Name // sort by name if types are equal
			}
			return properties[i].Type < properties[j].Type
		}
		return orderI < orderJ
	})
}
