// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
	"github.com/open-policy-agent/opa/types"
)

func builtinObjectUnion(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	objA, err := builtins.ObjectOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	objB, err := builtins.ObjectOperand(operands[1].Value, 2)
	if err != nil {
		return err
	}

	r := mergeWithOverwrite(objA, objB)

	return iter(ast.NewTerm(r))
}

func builtinObjectRemove(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	// Expect an object and an array/set/object of keys
	obj, err := builtins.ObjectOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	// Build a set of keys to remove
	keysToRemove, err := getObjectKeysParam(operands[1].Value)
	if err != nil {
		return err
	}
	r := ast.NewObject()
	obj.Foreach(func(key *ast.Term, value *ast.Term) {
		if !keysToRemove.Contains(key) {
			r.Insert(key, value)
		}
	})

	return iter(ast.NewTerm(r))
}

func builtinObjectFilter(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	// Expect an object and an array/set/object of keys
	obj, err := builtins.ObjectOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	// Build a new object from the supplied filter keys
	keys, err := getObjectKeysParam(operands[1].Value)
	if err != nil {
		return err
	}

	filterObj := ast.NewObject()
	keys.Foreach(func(key *ast.Term) {
		filterObj.Insert(key, ast.NullTerm())
	})

	// Actually do the filtering
	r, err := obj.Filter(filterObj)
	if err != nil {
		return err
	}

	return iter(ast.NewTerm(r))
}

func builtinObjectGet(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	object, err := builtins.ObjectOperand(operands[0].Value, 1)
	if err != nil {
		return err
	}

	if ret := object.Get(operands[1]); ret != nil {
		return iter(ret)
	}

	return iter(operands[2])
}

func builtinObjectLookup(_ BuiltinContext, operands []*ast.Term, iter func(*ast.Term) error) error {
	object := operands[0]

	// opinionated: since language-indexing is '.', key-path index is also '.'
	parts := strings.Split(string(operands[1].Value.(ast.String)), ".")
	for _, part := range parts {
		if object == nil {
			break
		}

		aPart := ast.StringTerm(part)
		switch original := object.Value.(type) {
		case ast.Object:
			// plain object indexing: e.g. person.name
			o := original.(ast.Object)
			object = o.Get(aPart)

		case *ast.Array:
			// index support: e.g. people.0.name
			idx, err := toIndex(original, aPart)
			if err != nil || idx < 0 || idx > original.Len() {
				object = nil
				break
			}
			a, err := builtins.ArrayOperand(original, 1)
			if err != nil {
				object = nil
				break
			}
			object = a.Get(ast.IntNumberTerm(idx))
		default:
			//we've over-indexed this one (over-reaching path)
			object = nil
			break
		}

	}

	if object == nil {
		return iter(operands[2])
	} else {
		return iter(object)
	}
}

// getObjectKeysParam returns a set of key values
// from a supplied ast array, object, set value
func getObjectKeysParam(arrayOrSet ast.Value) (ast.Set, error) {
	keys := ast.NewSet()

	switch v := arrayOrSet.(type) {
	case *ast.Array:
		_ = v.Iter(func(f *ast.Term) error {
			keys.Add(f)
			return nil
		})
	case ast.Set:
		_ = v.Iter(func(f *ast.Term) error {
			keys.Add(f)
			return nil
		})
	case ast.Object:
		_ = v.Iter(func(k *ast.Term, _ *ast.Term) error {
			keys.Add(k)
			return nil
		})
	default:
		return nil, builtins.NewOperandTypeErr(2, arrayOrSet, ast.TypeName(types.Object{}), ast.TypeName(types.S), ast.TypeName(types.Array{}))
	}

	return keys, nil
}

func mergeWithOverwrite(objA, objB ast.Object) ast.Object {
	merged, _ := objA.MergeWith(objB, func(v1, v2 *ast.Term) (*ast.Term, bool) {
		originalValueObj, ok2 := v1.Value.(ast.Object)
		updateValueObj, ok1 := v2.Value.(ast.Object)
		if !ok1 || !ok2 {
			// If we can't merge, stick with the right-hand value
			return v2, false
		}

		// Recursively update the existing value
		merged := mergeWithOverwrite(originalValueObj, updateValueObj)
		return ast.NewTerm(merged), false
	})
	return merged
}

func init() {
	RegisterBuiltinFunc(ast.ObjectUnion.Name, builtinObjectUnion)
	RegisterBuiltinFunc(ast.ObjectRemove.Name, builtinObjectRemove)
	RegisterBuiltinFunc(ast.ObjectFilter.Name, builtinObjectFilter)
	RegisterBuiltinFunc(ast.ObjectGet.Name, builtinObjectGet)
	RegisterBuiltinFunc(ast.ObjectLookup.Name, builtinObjectLookup)
}
