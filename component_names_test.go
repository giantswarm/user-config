package userconfig_test

import (
	"reflect"
	"testing"

	"github.com/giantswarm/user-config"
)

func Test_ComponentNames_ContainAny_AllMatching(t *testing.T) {
	current := userconfig.ComponentNames{
		userconfig.ComponentName("a"),
		userconfig.ComponentName("b"),
		userconfig.ComponentName("c"),
		userconfig.ComponentName("d"),
		userconfig.ComponentName("e"),
	}

	names := userconfig.ComponentNames{
		userconfig.ComponentName("a"),
		userconfig.ComponentName("b"),
		userconfig.ComponentName("c"),
		userconfig.ComponentName("d"),
		userconfig.ComponentName("e"),
	}

	if !current.ContainAny(names) {
		t.Fatalf("ContainAny should return true")
	}
}

func Test_ComponentNames_ContainAny_OneMatching(t *testing.T) {
	current := userconfig.ComponentNames{
		userconfig.ComponentName("a"),
		userconfig.ComponentName("b"),
		userconfig.ComponentName("c"),
		userconfig.ComponentName("d"),
		userconfig.ComponentName("e"),
	}

	names := userconfig.ComponentNames{
		userconfig.ComponentName("c"),
	}

	if !current.ContainAny(names) {
		t.Fatalf("ContainAny should return true")
	}
}

func Test_ComponentNames_ContainAny_NoneMatching(t *testing.T) {
	current := userconfig.ComponentNames{
		userconfig.ComponentName("a"),
		userconfig.ComponentName("b"),
		userconfig.ComponentName("c"),
		userconfig.ComponentName("d"),
		userconfig.ComponentName("e"),
	}

	names := userconfig.ComponentNames{
		userconfig.ComponentName("x"),
		userconfig.ComponentName("y"),
		userconfig.ComponentName("z"),
	}

	if current.ContainAny(names) {
		t.Fatalf("ContainAny should return false")
	}
}

func Test_ComponentNames_Unique_AllUnique(t *testing.T) {
	names := userconfig.ComponentNames{
		userconfig.ComponentName("a"),
		userconfig.ComponentName("b"),
		userconfig.ComponentName("c"),
		userconfig.ComponentName("d"),
		userconfig.ComponentName("e"),
	}

	expected := userconfig.ComponentNames{
		userconfig.ComponentName("a"),
		userconfig.ComponentName("b"),
		userconfig.ComponentName("c"),
		userconfig.ComponentName("d"),
		userconfig.ComponentName("e"),
	}

	result := names.Unique()

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Unique generated wrong result")
	}
}

func Test_ComponentNames_Unique_AllDuplicated(t *testing.T) {
	names := userconfig.ComponentNames{
		userconfig.ComponentName("a"),
		userconfig.ComponentName("a"),
		userconfig.ComponentName("a"),
		userconfig.ComponentName("b"),
		userconfig.ComponentName("b"),
		userconfig.ComponentName("c"),
		userconfig.ComponentName("c"),
		userconfig.ComponentName("d"),
		userconfig.ComponentName("d"),
		userconfig.ComponentName("d"),
		userconfig.ComponentName("d"),
		userconfig.ComponentName("d"),
		userconfig.ComponentName("e"),
		userconfig.ComponentName("e"),
	}

	expected := userconfig.ComponentNames{
		userconfig.ComponentName("a"),
		userconfig.ComponentName("b"),
		userconfig.ComponentName("c"),
		userconfig.ComponentName("d"),
		userconfig.ComponentName("e"),
	}

	result := names.Unique()

	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("Unique generated wrong result")
	}
}
