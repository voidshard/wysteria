package wysteria_client

import (
	wyc "github.com/voidshard/wysteria/common"
	"reflect"
	"testing"
)

func TestMultiOr(t *testing.T) {
	// Arrange
	testSearch := newTestQuery()
	funcs := []func(string) *search{
		testSearch.Id,
		testSearch.ChildOf,
		testSearch.LinkSource,
		testSearch.LinkDestination,
		testSearch.ItemType,
		testSearch.ItemVariant,
		testSearch.Name,
		testSearch.ResourceLocation,
	}

	cases := []struct {
		CallFuncs []func(string) *search
	}{
		{[]func(string) *search{funcs[0], funcs[3], funcs[5], funcs[7]}},
		{[]func(string) *search{funcs[1], funcs[5], funcs[6], funcs[2]}},
		{[]func(string) *search{funcs[2], funcs[0], funcs[4], funcs[7]}},
		{[]func(string) *search{funcs[3], funcs[1], funcs[1], funcs[1]}},
		{[]func(string) *search{funcs[4], funcs[2], funcs[7], funcs[2]}},
		{[]func(string) *search{funcs[5], funcs[7], funcs[6], funcs[3]}},
		{[]func(string) *search{funcs[6], funcs[6], funcs[5], funcs[5]}},
		{[]func(string) *search{funcs[7], funcs[5], funcs[4]}},
		{[]func(string) *search{funcs[0], funcs[4], funcs[1]}},
		{[]func(string) *search{funcs[1], funcs[3], funcs[2]}},
		{[]func(string) *search{funcs[2], funcs[2], funcs[3]}},
		{[]func(string) *search{funcs[3], funcs[1], funcs[1]}},
		{[]func(string) *search{funcs[4], funcs[0], funcs[5]}},
		{[]func(string) *search{funcs[5], funcs[1], funcs[2]}},
		{[]func(string) *search{funcs[6], funcs[2], funcs[7]}},
		{[]func(string) *search{funcs[7], funcs[4], funcs[3]}},
		{[]func(string) *search{funcs[0], funcs[7]}},
		{[]func(string) *search{funcs[1], funcs[1]}},
		{[]func(string) *search{funcs[2], funcs[1]}},
		{[]func(string) *search{funcs[3], funcs[3]}},
		{[]func(string) *search{funcs[4], funcs[1]}},
		{[]func(string) *search{funcs[5], funcs[0]}},
		{[]func(string) *search{funcs[6], funcs[1]}},
		{[]func(string) *search{funcs[7], funcs[0]}},
	}

	for _, tst := range cases {
		testSearch.Clear()

		// act
		for _, f := range tst.CallFuncs {
			f("foobarbaz")
			testSearch.Or()
		}

		// assert
		if len(testSearch.query) != len(tst.CallFuncs) {
			t.Error("Expected len of search Query to be ", len(tst.CallFuncs), " got", len(testSearch.query))
		}
	}
}

func TestSingleOr(t *testing.T) {
	// Arrange
	testSearch := newTestQuery()
	funcs := []func(string) *search{
		testSearch.Id,
		testSearch.ChildOf,
		testSearch.LinkSource,
		testSearch.LinkDestination,
		testSearch.ItemType,
		testSearch.ItemVariant,
		testSearch.Name,
		testSearch.ResourceLocation,
	}

	cases := []struct {
		CallFuncs []func(string) *search
	}{
		{[]func(string) *search{funcs[0], funcs[3], funcs[5], funcs[7]}},
		{[]func(string) *search{funcs[1], funcs[5], funcs[6], funcs[2]}},
		{[]func(string) *search{funcs[2], funcs[0], funcs[4], funcs[7]}},
		{[]func(string) *search{funcs[3], funcs[1], funcs[1], funcs[1]}},
		{[]func(string) *search{funcs[4], funcs[2], funcs[7], funcs[2]}},
		{[]func(string) *search{funcs[5], funcs[7], funcs[6], funcs[3]}},
		{[]func(string) *search{funcs[6], funcs[6], funcs[5], funcs[5]}},
		{[]func(string) *search{funcs[7], funcs[5], funcs[4]}},
		{[]func(string) *search{funcs[0], funcs[4], funcs[1]}},
		{[]func(string) *search{funcs[1], funcs[3], funcs[2]}},
		{[]func(string) *search{funcs[2], funcs[2], funcs[3]}},
		{[]func(string) *search{funcs[3], funcs[1], funcs[1]}},
		{[]func(string) *search{funcs[4], funcs[0], funcs[5]}},
		{[]func(string) *search{funcs[5], funcs[1], funcs[2]}},
		{[]func(string) *search{funcs[6], funcs[2], funcs[7]}},
		{[]func(string) *search{funcs[7], funcs[4], funcs[3]}},
		{[]func(string) *search{funcs[0], funcs[7]}},
		{[]func(string) *search{funcs[1], funcs[1]}},
		{[]func(string) *search{funcs[2], funcs[1]}},
		{[]func(string) *search{funcs[3], funcs[3]}},
		{[]func(string) *search{funcs[4], funcs[1]}},
		{[]func(string) *search{funcs[5], funcs[0]}},
		{[]func(string) *search{funcs[6], funcs[1]}},
		{[]func(string) *search{funcs[7], funcs[0]}},
	}

	for _, tst := range cases {
		testSearch.Clear()
		for _, f := range tst.CallFuncs {
			f("foobarbaz")
		}

		// act
		testSearch.Or()

		// assert
		if len(testSearch.query) != 1 {
			t.Error("Expected len of search Query to be 1, got", len(testSearch.query))
		}
	}
}

func TestHasFacets(t *testing.T) {
	// Arrange
	testSearch := newTestQuery()
	cases := []struct {
		Facets map[string]string
	}{
		{map[string]string{"foo": "bar"}},
		{map[string]string{"foo": "bar", "baz": "moo"}},
		{map[string]string{"foo": "bar", "blah": "/some/path"}},
		{map[string]string{"symbols": "yaaa*!@&#(!@&"}},
		{map[string]string{"something": "123"}},
	}

	for _, tst := range cases {
		testSearch.Clear()

		// act
		testSearch.HasFacets(tst.Facets)

		// assert
		if !reflect.DeepEqual(testSearch.nextQuery.Facets, tst.Facets) {
			t.Error("Expected facets to be set. Expected", tst.Facets, "got", testSearch.nextQuery.Facets)
		}

		if !testSearch.nextQValid {
			t.Error("Expected next query to be sat 'valid'")
		}
	}
}

func TestIntTerms(t *testing.T) {
	// Arrange
	testSearch := newTestQuery()
	terms := []int32{
		-1,
		0,
		100000,
		-100000,
		1273901287,
	}
	cases := []struct {
		CallFunc  func(int32) *search
		CheckFunc func() int32
		Name      string
	}{
		{testSearch.VersionNumber, testSearch.nqVersion, "VersionNumber"},
	}

	for _, tst := range cases {
		for _, input := range terms {
			// act
			testSearch.Clear()
			tst.CallFunc(input)

			// assert
			result := tst.CheckFunc()
			if result != input {
				t.Error("Running test", tst.Name, "expected", input, "got", result)
			}

			if !testSearch.nextQValid {
				t.Error("Running test", tst.Name, "expected nextQuery to be set valid")
			}
		}
	}
}

func TestStringTerms(t *testing.T) {
	// Arrange
	testSearch := newTestQuery()
	terms := []string{
		"abc",
		"(*&",
		"q ah g",
		"123awd",
		"/this/is/a/path.something.####.foo",
	}
	cases := []struct {
		CallFunc  func(string) *search
		CheckFunc func() string
		Name      string
	}{
		{testSearch.Id, testSearch.nqId, "Id"},
		{testSearch.ChildOf, testSearch.nqChildOf, "ChildOf"},
		{testSearch.LinkSource, testSearch.nqSrc, "Src"},
		{testSearch.LinkDestination, testSearch.nqDst, "Dst"},
		{testSearch.ItemType, testSearch.nqIType, "ItemType"},
		{testSearch.ItemVariant, testSearch.nqIVariant, "ItemVariant"},
		{testSearch.Name, testSearch.nqName, "Name"},
		{testSearch.ResourceLocation, testSearch.nqLocation, "Location"},
	}

	for _, tst := range cases {
		for _, input := range terms {
			// act
			testSearch.Clear()
			tst.CallFunc(input)

			// assert
			result := tst.CheckFunc()
			if result != input {
				t.Error("Running test", tst.Name, "expected", input, "got", result)
			}

			if !testSearch.nextQValid {
				t.Error("Running test", tst.Name, "expected nextQuery to be set valid")
			}
		}
	}
}

// Test helper funcs
func (i *search) nqId() string {
	return i.nextQuery.Id
}

func (i *search) nqChildOf() string {
	return i.nextQuery.Parent
}

func (i *search) nqSrc() string {
	return i.nextQuery.LinkSrc
}

func (i *search) nqDst() string {
	return i.nextQuery.LinkDst
}

func (i *search) nqIType() string {
	return i.nextQuery.ItemType
}

func (i *search) nqIVariant() string {
	return i.nextQuery.Variant
}

func (i *search) nqName() string {
	return i.nextQuery.Name
}

func (i *search) nqLocation() string {
	return i.nextQuery.Location
}

func newTestQuery() *search {
	return &search{
		query:     []*wyc.QueryDesc{},
		nextQuery: &wyc.QueryDesc{},
	}
}

func (i *search) nqVersion() int32 {
	return i.nextQuery.VersionNumber
}
