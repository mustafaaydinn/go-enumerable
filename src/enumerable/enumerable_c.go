package enumerable

import (
	"reflect"
)

type Val struct {
	Index int
	Val   reflect.Value
}

func runGoroutines(size int, fv reflect.Value) (chan Val, chan Val) {
	chin := make(chan Val)
	chout := make(chan Val)
	if size < 1 {
		size = 1
	}
	for i := 0; i < size; i++ {
		go func(i int, chin chan Val, chout chan Val, fv reflect.Value) {
			for in := range chin {
				chout <- Val{in.Index, fv.Call([]reflect.Value{in.Val})[0]}
			}
		}(i, chin, chout, fv)
	}
	return chin, chout
}

func send(chin chan Val, list reflect.Value) {
	size := list.Len()
	go func(chin chan Val) {
		for i := 0; i < size; i++ {
			chin <- Val{i, list.Index(i)}
		}
		close(chin)
	}(chin)
}

func MakeMapC(fptr interface{}, f interface{}, gsize int) {
	fn := reflect.ValueOf(fptr).Elem()
	out := fn.Type().Out(0)
	fv := reflect.ValueOf(f)
	fr := func(in []reflect.Value) []reflect.Value {
		list := in[0]
		l := list.Len()
		s := reflect.MakeSlice(out, l, l)
		chin, chout := runGoroutines(gsize, fv)
		send(chin, list)
		for i := 0; i < l; i++ {
			v := <-chout
			s.Index(v.Index).Set(v.Val)
		}
		return []reflect.Value{s}
	}
	fn.Set(reflect.MakeFunc(fn.Type(), fr))
}

func MakeFilterC(fptr interface{}, f interface{}, gsize int) {
	fn := reflect.ValueOf(fptr).Elem()
	out := fn.Type().Out(0)
	fv := reflect.ValueOf(f)
	fr := func(in []reflect.Value) []reflect.Value {
		list := in[0]
		l := list.Len()
		s := reflect.MakeSlice(out, 0, l)
		chin, chout := runGoroutines(gsize, fv)
		send(chin, list)
		for i := 0; i < l; i++ {
			v := <-chout
			if v.Val.Bool() {
				s = reflect.Append(s, list.Index(v.Index))
			}
		}
		return []reflect.Value{s}
	}
	fn.Set(reflect.MakeFunc(fn.Type(), fr))
}

func MakeSomeC(fptr interface{}, f interface{}, gsize int) {
	fn := reflect.ValueOf(fptr).Elem()
	fv := reflect.ValueOf(f)
	fr := func(in []reflect.Value) []reflect.Value {
		list := in[0]
		l := list.Len()
		chin, chout := runGoroutines(gsize, fv)
		send(chin, list)
		for i := 0; i < l; i++ {
			v := <-chout
			if v.Val.Bool() {
				return []reflect.Value{reflect.ValueOf(true)}
			}
		}
		return []reflect.Value{reflect.ValueOf(false)}
	}
	fn.Set(reflect.MakeFunc(fn.Type(), fr))
}

func MakeEveryC(fptr interface{}, f interface{}, gsize int) {
	fn := reflect.ValueOf(fptr).Elem()
	fv := reflect.ValueOf(f)
	fr := func(in []reflect.Value) []reflect.Value {
		list := in[0]
		l := list.Len()
		chin, chout := runGoroutines(gsize, fv)
		send(chin, list)
		for i := 0; i < l; i++ {
			v := <-chout
			if !v.Val.Bool() {
				return []reflect.Value{reflect.ValueOf(false)}
			}
		}
		return []reflect.Value{reflect.ValueOf(true)}
	}
	fn.Set(reflect.MakeFunc(fn.Type(), fr))
}

func MakeFirst(fptr interface{}, f interface{}) {
	fn := reflect.ValueOf(fptr).Elem()
	fv := reflect.ValueOf(f)
	fr := func(in []reflect.Value) []reflect.Value {
		list := in[0]
		chout := make(chan Val)
		for i := 0; i < list.Len(); i++ {
			go func(i int, v reflect.Value) {
				chout <- Val{i, fv.Call([]reflect.Value{v})[0]}
			}(i, list.Index(i))
		}
		v := <-chout
		return []reflect.Value{v.Val}
	}
	fn.Set(reflect.MakeFunc(fn.Type(), fr))
}