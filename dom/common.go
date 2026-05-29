package dom

import (
	"fmt"
	"strings"
	"time"
)

type CommonModel struct {
	Id        int       `gorm:"primary_key" query:"Id"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (*CommonModel) WhereBuild(conditionMap map[string][]any) (whereSql string, value []any) {
	conditionArr := make([]string, 0, 10)
	valueArr := make([]any, 0, 10)
	for field, v := range conditionMap {
		switch v[1].(type) {
		case string:
			if v[1].(string) == "" {
				continue
			}
		case int:
			if v[1].(int) == 0 {
				continue
			}
		default:

		}
		cond := ""
		condition := v[0]
		conditionValues := v[1:]
		switch condition {
		case "between":
			cond = fmt.Sprintf("%v BETWEEN ? AND ?", field)
		case "contains":
			cond = fmt.Sprintf("CONTAINS(%v, ?)", field)
		case "json_contains":
			cond = fmt.Sprintf("JSON_CONTAINS(%v, ?, '$')", field)
		case "json_search":
			cond = fmt.Sprintf("JSON_SEARCH(%v, 'all', ?)", field)
		case "like":
			if len(v) == 3 {
				if v[2] == "prefix" {
					conditionValues[0] = conditionValues[0].(string) + "%"
					cond = fmt.Sprintf("%v LIKE ?", field)
				} else if v[2] == "suffix" {
					conditionValues[0] = "%" + conditionValues[0].(string)
					cond = fmt.Sprintf("%v LIKE ?", field)
				} else {
					conditionValues[0] = "%" + conditionValues[0].(string) + "%"
					cond = fmt.Sprintf("%v LIKE ?", field)
				}
			} else {
				conditionValues[0] = "%" + conditionValues[0].(string) + "%"
				cond = fmt.Sprintf("%v LIKE ?", field)
			}
		case "in":
			cond = fmt.Sprintf("%v IN (?)", field)
		case "<":
			cond = fmt.Sprintf("%v < ?", field)
		case ">":
			cond = fmt.Sprintf("%v > ?", field)
		default:
			cond = fmt.Sprintf("%v %v ?", field, condition)
		}
		conditionArr = append(conditionArr, cond)
		valueArr = append(valueArr, conditionValues...)
	}
	whereSql = strings.Join(conditionArr, " AND ")
	return whereSql, valueArr
}

func (cm *CommonModel) NodeProps() map[string]any {
	return map[string]any{
		"Id":        cm.Id,
		"CreatedAt": cm.CreatedAt.String(),
		"UpdatedAt": cm.UpdatedAt.String(),
	}
}
