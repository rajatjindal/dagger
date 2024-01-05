package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIface(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	strs := []string{"a", "b"}
	ints := []int{1, 2}
	bools := []bool{true, false}
	dirs := []*Directory{
		dag.Directory().WithNewFile("/file1", "file1"),
		dag.Directory().WithNewFile("/file2", "file2"),
	}
	impl := dag.Impl(strs, ints, bools, dirs)

	test := dag.Test()

	t.Run("void", func(t *testing.T) {
		t.Parallel()
		_, err := test.Void(ctx, impl.AsTestCustomIface())
		require.NoError(t, err)
	})

	t.Run("str", func(t *testing.T) {
		t.Parallel()
		str, err := test.Str(ctx, impl.AsTestCustomIface())
		require.NoError(t, err)
		require.Equal(t, "a", str)
	})
	t.Run("withStr", func(t *testing.T) {
		t.Parallel()
		str, err := test.WithStr(impl.AsTestCustomIface(), "c").Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "c", str)
	})
	t.Run("withOptionalTypeStr", func(t *testing.T) {
		t.Parallel()
		str, err := test.WithOptionalTypeStr(impl.AsTestCustomIface(), TestWithOptionalTypeStrOpts{
			StrArg: "d",
		}).Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "d", str)
		str, err = test.WithOptionalTypeStr(impl.AsTestCustomIface()).Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "a", str)
	})
	t.Run("withOptionalPragmaStr", func(t *testing.T) {
		t.Parallel()
		str, err := test.WithOptionalPragmaStr(impl.AsTestCustomIface(), TestWithOptionalPragmaStrOpts{
			StrArg: "d",
		}).Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "d", str)
		str, err = test.WithOptionalPragmaStr(impl.AsTestCustomIface()).Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "a", str)
	})
	t.Run("strList", func(t *testing.T) {
		t.Parallel()
		strs, err := test.StrList(ctx, impl.AsTestCustomIface())
		require.NoError(t, err)
		require.Equal(t, []string{"a", "b"}, strs)
	})
	t.Run("withStrList", func(t *testing.T) {
		t.Parallel()
		strs, err := test.WithStrList(impl.AsTestCustomIface(), []string{"c", "d"}).StrList(ctx)
		require.NoError(t, err)
		require.Equal(t, []string{"c", "d"}, strs)
	})

	t.Run("int", func(t *testing.T) {
		t.Parallel()
		i, err := test.Int(ctx, impl.AsTestCustomIface())
		require.NoError(t, err)
		require.Equal(t, 1, i)
	})
	t.Run("withInt", func(t *testing.T) {
		t.Parallel()
		i, err := test.WithInt(impl.AsTestCustomIface(), 3).Int(ctx)
		require.NoError(t, err)
		require.Equal(t, 3, i)
	})
	t.Run("intList", func(t *testing.T) {
		t.Parallel()
		ints, err := test.IntList(ctx, impl.AsTestCustomIface())
		require.NoError(t, err)
		require.Equal(t, []int{1, 2}, ints)
	})
	t.Run("withIntList", func(t *testing.T) {
		t.Parallel()
		ints, err := test.WithIntList(impl.AsTestCustomIface(), []int{3, 4}).IntList(ctx)
		require.NoError(t, err)
		require.Equal(t, []int{3, 4}, ints)
	})

	t.Run("bool", func(t *testing.T) {
		t.Parallel()
		b, err := test.Bool(ctx, impl.AsTestCustomIface())
		require.NoError(t, err)
		require.Equal(t, true, b)
	})
	t.Run("withBool", func(t *testing.T) {
		t.Parallel()
		b, err := test.WithBool(impl.AsTestCustomIface(), false).Bool(ctx)
		require.NoError(t, err)
		require.Equal(t, false, b)
	})
	t.Run("boolList", func(t *testing.T) {
		t.Parallel()
		bools, err := test.BoolList(ctx, impl.AsTestCustomIface())
		require.NoError(t, err)
		require.Equal(t, []bool{true, false}, bools)
	})
	t.Run("withBoolList", func(t *testing.T) {
		t.Parallel()
		bools, err := test.WithBoolList(impl.AsTestCustomIface(), []bool{false, true}).BoolList(ctx)
		require.NoError(t, err)
		require.Equal(t, []bool{false, true}, bools)
	})

	t.Run("obj", func(t *testing.T) {
		t.Parallel()
		dir := test.Obj(impl.AsTestCustomIface())
		dirEnts, err := dir.Entries(ctx)
		require.NoError(t, err)
		require.Contains(t, dirEnts, "file1")
	})
	t.Run("withObj", func(t *testing.T) {
		t.Parallel()
		dir := test.WithObj(impl.AsTestCustomIface(), dirs[1]).Obj()
		dirEnts, err := dir.Entries(ctx)
		require.NoError(t, err)
		require.Contains(t, dirEnts, "file2")
	})
	t.Run("withOptionalTypeObj", func(t *testing.T) {
		t.Parallel()
		obj := test.WithOptionalTypeObj(impl.AsTestCustomIface(), TestWithOptionalTypeObjOpts{
			ObjArg: dag.Directory().WithNewFile("/file3", "file3"),
		}).Obj()
		dirEnts, err := obj.Entries(ctx)
		require.NoError(t, err)
		require.Contains(t, dirEnts, "file3")

		obj = test.WithOptionalTypeObj(impl.AsTestCustomIface()).Obj()
		dirEnts, err = obj.Entries(ctx)
		require.NoError(t, err)
		require.Contains(t, dirEnts, "file1")
	})
	t.Run("objList", func(t *testing.T) {
		t.Parallel()
		dirs, err := test.ObjList(ctx, impl.AsTestCustomIface())
		require.NoError(t, err)
		require.Len(t, dirs, 2)
		dirEnts1, err := dirs[0].Entries(ctx)
		require.NoError(t, err)
		require.Contains(t, dirEnts1, "file1")
		dirEnts2, err := dirs[1].Entries(ctx)
		require.NoError(t, err)
		require.Contains(t, dirEnts2, "file2")
	})
	t.Run("withObjList", func(t *testing.T) {
		t.Parallel()
		dirs, err := test.WithObjList(impl.AsTestCustomIface(), []*Directory{
			dag.Directory().WithNewFile("/file3", "file3"),
			dag.Directory().WithNewFile("/file4", "file4"),
		}).ObjList(ctx)
		require.NoError(t, err)
		require.Len(t, dirs, 2)
		dirEnts1, err := dirs[0].Entries(ctx)
		require.NoError(t, err)
		require.Contains(t, dirEnts1, "file3")
		dirEnts2, err := dirs[1].Entries(ctx)
		require.NoError(t, err)
		require.Contains(t, dirEnts2, "file4")
	})

	t.Run("selfIface", func(t *testing.T) {
		t.Parallel()
		iface := test.SelfIface(impl.AsTestCustomIface())
		str, err := iface.Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "aself", str)
	})
	t.Run("selfIfaceList", func(t *testing.T) {
		t.Parallel()
		ifaces, err := test.SelfIfaceList(ctx, impl.AsTestCustomIface())
		require.NoError(t, err)
		require.Len(t, ifaces, 2)
		str1, err := ifaces[0].Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "aself1", str1)
		str2, err := ifaces[1].Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "aself2", str2)
	})

	t.Run("otherIface", func(t *testing.T) {
		t.Parallel()
		iface := test.OtherIface(impl.AsTestCustomIface())
		str, err := iface.Foo(ctx)
		require.NoError(t, err)
		require.Equal(t, "aother", str)
	})
	t.Run("otherIfaceList", func(t *testing.T) {
		t.Parallel()
		ifaces, err := test.OtherIfaceList(ctx, impl.AsTestCustomIface())
		require.NoError(t, err)
		require.Len(t, ifaces, 2)
		str1, err := ifaces[0].Foo(ctx)
		require.NoError(t, err)
		require.Equal(t, "aother1", str1)
		str2, err := ifaces[1].Foo(ctx)
		require.NoError(t, err)
		require.Equal(t, "aother2", str2)
	})

	t.Run("ifaceListArgs", func(t *testing.T) {
		t.Parallel()
		strs, err := test.IfaceListArgs(ctx,
			[]*TestCustomIface{
				impl.AsTestCustomIface(),
				impl.SelfIface().AsTestCustomIface(),
			},
			[]*TestOtherIface{
				impl.OtherIface().AsTestOtherIface(),
				impl.SelfIface().OtherIface().AsTestOtherIface(),
			},
		)
		require.NoError(t, err)
		require.Equal(t, []string{"a", "aself", "aother", "aselfother"}, strs)
	})

	t.Run("parentIfaceFields", func(t *testing.T) {
		t.Parallel()
		t.Run("basic", func(t *testing.T) {
			t.Parallel()
			strs, err := test.
				WithIface(impl.AsTestCustomIface()).
				WithPrivateIface(dag.Impl([]string{"private"}, []int{99}, []bool{false}, []*Directory{dag.Directory()}).AsTestCustomIface()).
				WithIfaceList([]*TestCustomIface{
					impl.AsTestCustomIface(),
					impl.SelfIface().AsTestCustomIface(),
				}).
				WithOtherIfaceList([]*TestOtherIface{
					impl.OtherIface().AsTestOtherIface(),
					impl.SelfIface().OtherIface().AsTestOtherIface(),
				}).
				ParentIfaceFields(ctx)
			require.NoError(t, err)
			require.Equal(t, []string{"a", "private", "a", "aself", "aother", "aselfother"}, strs)
		})
		t.Run("optionals", func(t *testing.T) {
			t.Parallel()
			strs, err := test.
				WithOptionalTypeIface().
				WithOptionalTypeIface(TestWithOptionalTypeIfaceOpts{Iface: impl.AsTestCustomIface()}).
				WithOptionalTypeIface().
				ParentIfaceFields(ctx)
			require.NoError(t, err)
			require.Equal(t, []string{"a"}, strs)
			strs, err = test.
				WithOptionalPragmaIface().
				WithOptionalPragmaIface(TestWithOptionalPragmaIfaceOpts{Iface: impl.AsTestCustomIface()}).
				WithOptionalPragmaIface().
				ParentIfaceFields(ctx)
			require.NoError(t, err)
			require.Equal(t, []string{"a"}, strs)
		})
	})

	t.Run("returnCustomObj", func(t *testing.T) {
		t.Parallel()
		customObj := test.ReturnCustomObj(
			[]*TestCustomIface{
				impl.AsTestCustomIface(),
				impl.SelfIface().AsTestCustomIface(),
			},
			[]*TestOtherIface{
				impl.OtherIface().AsTestOtherIface(),
				impl.SelfIface().OtherIface().AsTestOtherIface(),
			},
		)

		ifaceStr, err := customObj.Iface().Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "a", ifaceStr)

		ifaces, err := customObj.IfaceList(ctx)
		require.NoError(t, err)
		require.Len(t, ifaces, 2)
		ifaceStr1, err := ifaces[0].Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "a", ifaceStr1)
		ifaceStr2, err := ifaces[1].Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "aself", ifaceStr2)

		otherCustomObjIfaceStr, err := customObj.Other().Iface().Str(ctx)
		require.NoError(t, err)

		require.Equal(t, "a", otherCustomObjIfaceStr)
		otherCustomObjIfaces, err := customObj.Other().IfaceList(ctx)
		require.NoError(t, err)
		require.Len(t, otherCustomObjIfaces, 2)
		otherCustomObjIfaceStr1, err := otherCustomObjIfaces[0].Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "a", otherCustomObjIfaceStr1)
		otherCustomObjIfaceStr2, err := otherCustomObjIfaces[1].Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "aself", otherCustomObjIfaceStr2)

		otherPtrCustomObjIfaceStr, err := customObj.OtherPtr().Iface().Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "a", otherPtrCustomObjIfaceStr)

		otherPtrCustomObjIfaces, err := customObj.OtherPtr().IfaceList(ctx)
		require.NoError(t, err)
		require.Len(t, otherPtrCustomObjIfaces, 2)
		otherPtrCustomObjIfaceStr1, err := otherPtrCustomObjIfaces[0].Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "a", otherPtrCustomObjIfaceStr1)
		otherPtrCustomObjIfaceStr2, err := otherPtrCustomObjIfaces[1].Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "aself", otherPtrCustomObjIfaceStr2)

		otherCustomObjList, err := customObj.OtherList(ctx)
		require.NoError(t, err)
		require.Len(t, otherCustomObjList, 1)
		otherCustomObjListStr, err := otherCustomObjList[0].Iface().Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "a", otherCustomObjListStr)
		otherCustomObjListIfaces, err := otherCustomObjList[0].IfaceList(ctx)
		require.NoError(t, err)
		require.Len(t, otherCustomObjListIfaces, 2)
		otherCustomObjListIfaceStr1, err := otherCustomObjListIfaces[0].Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "a", otherCustomObjListIfaceStr1)
		otherCustomObjListIfaceStr2, err := otherCustomObjListIfaces[1].Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "aself", otherCustomObjListIfaceStr2)

		otherCustomObjPtrList, err := customObj.OtherPtrList(ctx)
		require.NoError(t, err)
		require.Len(t, otherCustomObjPtrList, 1)
		otherCustomObjPtrListStr, err := otherCustomObjPtrList[0].Iface().Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "a", otherCustomObjPtrListStr)
		otherCustomObjPtrListIfaces, err := otherCustomObjPtrList[0].IfaceList(ctx)
		require.NoError(t, err)
		require.Len(t, otherCustomObjPtrListIfaces, 2)
		otherCustomObjPtrListIfaceStr1, err := otherCustomObjPtrListIfaces[0].Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "a", otherCustomObjPtrListIfaceStr1)
		otherCustomObjPtrListIfaceStr2, err := otherCustomObjPtrListIfaces[1].Str(ctx)
		require.NoError(t, err)
		require.Equal(t, "aself", otherCustomObjPtrListIfaceStr2)
	})
}