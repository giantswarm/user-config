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

	diffInfos = append(diffInfos, diffServiceNameUpdated(newDef.AppName, oldDef.AppName)...)
	diffInfos = append(diffInfos, diffComponentAdded(newDef.Components, oldDef.Components)...)
	diffInfos = append(diffInfos, diffComponentRemoved(newDef.Components, oldDef.Components)...)
	diffInfos = append(diffInfos, diffComponentUpdated(newDef.Components, oldDef.Components)...)

	return diffInfos
}

func diffServiceNameUpdated(a, b AppName) []DiffInfo {
	diffInfos := []DiffInfo{}

	if !a.Equals(b) {
		diffInfos = append(diffInfos, DiffInfo{
			Type: DiffInfoServiceNameUpdated,
			Old:  b.String(),
			New:  a.String(),
		})
	}

	return diffInfos
}

func diffComponentAdded(a, b ComponentDefinitions) []DiffInfo {
	diffInfos := []DiffInfo{}

	for _, orderedName := range orderedComponentKeys(a) {
		aName := ComponentName(orderedName)

		if _, ok := b[aName]; !ok {
			diffInfos = append(diffInfos, DiffInfo{
				Type: DiffInfoComponentAdded,
				New:  aName.String(),
			})
		}
	}

	return diffInfos
}

func diffComponentRemoved(a, b ComponentDefinitions) []DiffInfo {
	diffInfos := []DiffInfo{}

	for _, orderedName := range orderedComponentKeys(b) {
		bName := ComponentName(orderedName)

		if _, ok := a[bName]; !ok {
			diffInfos = append(diffInfos, DiffInfo{
				Type: DiffInfoComponentRemoved,
				Old:  bName.String(),
			})
		}
	}

	return diffInfos
}

func diffComponentUpdated(a, b ComponentDefinitions) []DiffInfo {
	diffInfos := []DiffInfo{}

	for _, orderedName := range orderedComponentKeys(b) {
		bName := ComponentName(orderedName)
		bComponent := b[bName]

		if aComponent, ok := a[bName]; ok {
			if !reflect.DeepEqual(aComponent, bComponent) {
				diffInfos = append(diffInfos, DiffInfo{
					Type: DiffInfoComponentUpdated,
					Old:  bName.String(),
					New:  bName.String(),
				})
			}
		}
	}

	return diffInfos
}

func orderedComponentKeys(a ComponentDefinitions) []string {
	keys := []string{}

	for aName, _ := range a {
		keys = append(keys, aName.String())
	}
	sort.Strings(keys)

	return keys
}
