package main

import (
	"reflect"
	"sort"
)

type sortableFields []string

func (s sortableFields) Contains(str string) bool {
	for _, item := range s {
		if item == str {
			return true
		}
	}

	return false
}

var sFields sortableFields = []string{"Id", "Name", "Age", "About", "Gender"}

func getSortValueByField(user UserServer, field string) interface{} {
	val := reflect.ValueOf(user)

	var found interface{} = nil

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		if typeField.Name == field {
			found = valueField.Interface()
		}

		// fmt.Printf("\tname=%v, type=%v, value=%v sortField=%v\n",
		// 	typeField.Name,
		// 	typeField.Type.Kind(),
		// 	valueField,
		// field)
	}

	return found
}

func compareLess(left, right interface{}) bool {
	switch left.(type) {
	case int:
		lInt := left.(int)
		rInt := right.(int)
		return lInt > rInt
	case string:
		lStr := left.(string)
		rStr := right.(string)
		return lStr > rStr
	}

	return false
}

func orderUsers(users *[]UserServer, field string, by int) {
	if field != "" && by != OrderByAsIs && sFields.Contains(field) {
		sort.SliceStable(*users, func(i, j int) bool {
			left := getSortValueByField((*users)[i], field)
			right := getSortValueByField((*users)[j], field)

			if left == nil || right == nil {
				return false
			}

			result := compareLess(left, right)
			if by == OrderByAsc {
				return !result
			}

			return result
		})
	}
}
