package userconfig

import (
	"reflect"
	"sort"
)

type DiffType string

const (
	// DiffInfoServiceNameUpdated
	DiffInfoServiceNameUpdated DiffType = "service-name-updated"

	// DiffInfoComponentAdded
	DiffInfoComponentAdded DiffType = "component-added"

	// DiffInfoComponentRemoved
	DiffInfoComponentRemoved DiffType = "component-removed"

	// DiffInfoComponentUpdated
	DiffInfoComponentUpdated DiffType = "component-updated"
)

type DiffInfo struct {
	Type DiffType

	Old string
	New string
}

func Diff(oldDef, newDef V2AppDefinition) []DiffInfo {
	diffInfos := []DiffInfo{}

	diffInfos = append(diffInfos, diffServiceNameUpdated(oldDef.AppName, newDef.AppName)...)
	diffInfos = append(diffInfos, diffComponentAdded(oldDef.Components, newDef.Components)...)
	diffInfos = append(diffInfos, diffComponentUpdated(oldDef.Components, newDef.Components)...)
	diffInfos = append(diffInfos, diffComponentRemoved(oldDef.Components, newDef.Components)...)

	return diffInfos
}

func diffServiceNameUpdated(oldDef, newDef AppName) []DiffInfo {
	diffInfos := []DiffInfo{}

	if !newDef.Equals(oldDef) {
		diffInfos = append(diffInfos, DiffInfo{
			Type: DiffInfoServiceNameUpdated,
			Old:  oldDef.String(),
			New:  newDef.String(),
		})
	}

	return diffInfos
}

func diffComponentAdded(oldDef, newDef ComponentDefinitions) []DiffInfo {
	diffInfos := []DiffInfo{}

	for _, orderedName := range orderedComponentKeys(newDef) {
		newName := ComponentName(orderedName)

		if _, ok := oldDef[newName]; !ok {
			diffInfos = append(diffInfos, DiffInfo{
				Type: DiffInfoComponentAdded,
				New:  newName.String(),
			})
		}
	}

	return diffInfos
}

func diffComponentUpdated(oldDef, newDef ComponentDefinitions) []DiffInfo {
	diffInfos := []DiffInfo{}

	for _, orderedName := range orderedComponentKeys(oldDef) {
		oldName := ComponentName(orderedName)
		oldComponent := oldDef[oldName]

		if newComponent, ok := newDef[oldName]; ok {
			if !reflect.DeepEqual(newComponent, oldComponent) {
				diffInfos = append(diffInfos, DiffInfo{
					Type: DiffInfoComponentUpdated,
					Old:  oldName.String(),
					New:  oldName.String(),
				})
			}
		}
	}

	return diffInfos
}

func diffComponentRemoved(oldDef, newDef ComponentDefinitions) []DiffInfo {
	diffInfos := []DiffInfo{}

	for _, orderedName := range orderedComponentKeys(oldDef) {
		oldName := ComponentName(orderedName)

		if _, ok := newDef[oldName]; !ok {
			diffInfos = append(diffInfos, DiffInfo{
				Type: DiffInfoComponentRemoved,
				Old:  oldName.String(),
			})
		}
	}

	return diffInfos
}

func orderedComponentKeys(oldDef ComponentDefinitions) []string {
	keys := []string{}

	for oldName, _ := range oldDef {
		keys = append(keys, oldName.String())
	}
	sort.Strings(keys)

	return keys
}
