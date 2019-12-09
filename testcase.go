package testcase

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

const tagName = "tc"

type fieldValue struct {
	field int
	data  reflect.Value
}

type TestGenerator struct {
	t   *testing.T
	rv  reflect.Value
	fv  []fieldValue
	idx []int
	pos int
}

type assignable interface {
	assign(v reflect.Value) error
	typ() reflect.Type
}

type assignVal struct {
	reflect.Value
}

func (a assignVal) assign(v reflect.Value) error {
	if v.Type().ConvertibleTo(a.Type()) {
		a.Set(v.Convert(a.Type()))
		return nil
	}
	return fmt.Errorf("cannot convert value '%#v' of type '%s' to type '%s'", v, v.Type(), a)
}

func (a assignVal) typ() reflect.Type {
	return a.Value.Type()
}

type assignKV struct {
	reflect.Value
	key reflect.Value
}

func (a assignKV) assign(v reflect.Value) error {
	if a.key.Type().ConvertibleTo(a.Type().Key()) && v.Type().ConvertibleTo(a.Type().Elem()) {
		a.SetMapIndex(a.key.Convert(a.Type().Key()), v.Convert(a.Type().Elem()))
		return nil
	}
	return fmt.Errorf("cannot set map[%s]=%s of (key type %s, val type %s) to with map type %s",
		a.key, v, a.key.Type(), v.Type(), a.Type())
}

func (a assignKV) typ() reflect.Type {
	return a.Value.Type().Elem()
}

var _ assignable = &assignVal{}
var _ assignable = &assignKV{}

func (g *TestGenerator) assign(src reflect.Value, dst reflect.Value) {
	// Fast path
	if src.Type().ConvertibleTo(dst.Type()) {
		dst.Set(src.Convert(dst.Type()))
		return
	}

	type item struct {
		from reflect.Value
		to   assignable
	}

	work := []item{
		{
			from: src,
			to:   assignVal{dst},
		},
	}

	for len(work) > 0 {
		assign := work[len(work)-1]
		work = work[:len(work)-1]
		from, to := assign.from, assign.to

		switch to.typ().Kind() {
		case reflect.Slice:
			if from.Type().Kind() != reflect.Slice {
				g.t.Fatalf("Cannot assign non-slice type '%s' to slice type '%s'", from.Type(), to.typ())
			}
			slice := reflect.MakeSlice(to.typ(), from.Len(), from.Cap())
			if err := to.assign(slice); err != nil {
				g.t.Fatal(err)
			}

			for i := 0; i < from.Len(); i++ {
				work = append(work, item{
					from: from.Index(i).Elem(),
					to:   assignVal{slice.Index(i)},
				})
			}
		case reflect.Map:
			if from.Kind() != reflect.Map {
				g.t.Fatalf("Cannot assign non-map type '%s' to map type '%s'", from.Type(), to.typ())
			}
			m := reflect.MakeMap(to.typ())
			if err := to.assign(m); err != nil {
				g.t.Fatal(err)
			}

			for it := from.MapRange(); it.Next(); {
				work = append(work, item{
					from: it.Value().Elem(),
					to: assignKV{
						Value: m,
						key:   it.Key(),
					},
				})
			}
		default:
			if err := assign.to.assign(assign.from); err != nil {
				g.t.Fatal(err)
			}
		}
	}
}

func (g *TestGenerator) Next() bool {
	if len(g.fv) == 0 || g.idx[0] == g.fv[0].data.Len() {
		return false
	}

	ready := false
	for !ready && g.idx[0] < g.fv[0].data.Len() {
		val := g.fv[g.pos].data.Index(g.idx[g.pos]).Elem()
		field := g.rv.Elem().Field(g.fv[g.pos].field)
		g.assign(val, field)

		if g.pos < len(g.fv)-1 {
			g.pos++
		} else {
			ready = true
			g.idx[g.pos]++
		}

		for g.idx[g.pos] == g.fv[g.pos].data.Len() && g.pos > 0 {
			g.idx[g.pos] = 0
			g.pos--
			g.idx[g.pos]++
		}
	}

	return ready
}

func GenerateTestCases(t *testing.T, tc interface{}) *TestGenerator {
	t.Helper()

	gen := &TestGenerator{
		t:  t,
		rv: reflect.ValueOf(tc),
	}

	if gen.rv.Kind() != reflect.Ptr || gen.rv.IsNil() || gen.rv.Elem().Kind() != reflect.Struct {
		t.Fatalf("Expected ptr to struct, found %v instead", tc)
	}

	typ := gen.rv.Elem().Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if tag, ok := field.Tag.Lookup(tagName); ok {
			aT := reflect.SliceOf(field.Type)
			slice := reflect.MakeSlice(aT, 0, 0)
			dest := slice.Interface()

			if err := json.Unmarshal([]byte(tag), &dest); err != nil {
				t.Fatalf("Failed to unmarshal values for field %s: %v", field.Name, err)
			}

			gen.fv = append(gen.fv, fieldValue{
				field: i,
				data:  reflect.ValueOf(dest),
			})
		}
	}
	gen.idx = make([]int, len(gen.fv))
	return gen
}
