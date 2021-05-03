package main

import (
	"fmt"
	fuzz "github.com/google/gofuzz"
	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"strconv"

	metafuzzer "k8s.io/apimachinery/pkg/apis/meta/fuzzer"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
)

// CRDSafeFuzzerFuncs will merge the given funcLists, overriding early funcs with later ones if there first
// argument has the same type.
func CRDSafeFuzzerFuncs(funcs ...fuzzer.FuzzerFuncs) fuzzer.FuzzerFuncs {
	return fuzzer.FuzzerFuncs(func(codecs serializer.CodecFactory) []interface{} {
		result := []interface{}{}
		for _, f := range funcs {
			if f != nil && !matches(f) {
				result = append(result, f(codecs)...)
			}
		}
		result = append(result, 		func(j *metav1.ObjectMeta, c fuzz.Continue) {
			c.FuzzNoCustom(j)

			j.ResourceVersion = strconv.FormatUint(c.RandUint64(), 10)
			j.UID = types.UID(c.RandString())

			var sec, nsec int64
			c.Fuzz(&sec)
			c.Fuzz(&nsec)
			j.CreationTimestamp = metav1.Unix(sec, nsec).Rfc3339Copy()

			if j.DeletionTimestamp != nil {
				c.Fuzz(&sec)
				c.Fuzz(&nsec)
				t := metav1.Unix(sec, nsec).Rfc3339Copy()
				j.DeletionTimestamp = &t
			}

			if len(j.Labels) == 0 {
				j.Labels = nil
			} else {
				delete(j.Labels, "")
			}
			if len(j.Annotations) == 0 {
				j.Annotations = nil
			} else {
				delete(j.Annotations, "")
			}
			if len(j.OwnerReferences) == 0 {
				j.OwnerReferences = nil
			}
			if len(j.Finalizers) == 0 {
				j.Finalizers = nil
			}
		})
		return result
	})
}

func main() {
	scheme := clientsetscheme.Scheme
	codecFactory := serializer.NewCodecFactory(scheme)
	fns := metafuzzer.Funcs(codecFactory)


	n := 0
	for _, fn := range fns {
		if !matches(fn) {
			fns[n] = fn
			n++
		}
	}
	fns = fns[:n]

	for _, f := range fns {
		x := reflect.TypeOf(f)
		if x.Kind() != reflect.Func {
			continue
		}
		fmt.Println("Method:", x.String())
		fmt.Println("Variadic:", x.IsVariadic()) // Used (<type> ...) ?
		fmt.Println("Package:", x.PkgPath())

		// 		func(j *metav1.ObjectMeta, c fuzz.Continue) {
		numIn := x.NumIn() // count number of parameters
		numOut := x.NumOut() // count number of return values

		for i := 0; i < numIn; i++ {
			inV := x.In(i)
			in_Kind := inV.Kind() //func
			if in_Kind == reflect.Ptr {
				inU := inV.Elem()
				fmt.Println(inV.String(), inU.Name(), inU.String(), inU.PkgPath())
			} else if in_Kind == reflect.Struct {
				fmt.Println(inV.Name(), inV.String(), inV.PkgPath())
			}

			fmt.Printf("\nParameter IN: "+strconv.Itoa(i)+"\nKind: %v\nName: %v\n-----------\n",in_Kind,inV.Name())
		}
		for o := 0; o < numOut; o++ {
			returnV := x.Out(0)
			return_Kind := returnV.Kind()
			fmt.Printf("\nParameter OUT: "+strconv.Itoa(o)+"\nKind: %v\nName: %v\n",return_Kind,returnV.Name())
		}
	}
}

func matches(f interface{}) bool {
	x := reflect.TypeOf(f)
	if x.Kind() != reflect.Func {
		return false
	}
	if x.IsVariadic() {
		return false
	}

	/*
		*v1.ObjectMeta ObjectMeta v1.ObjectMeta k8s.io/apimachinery/pkg/apis/meta/v1
		Continue fuzz.Continue github.com/google/gofuzz
	*/

	numIn := x.NumIn() // count number of parameters
	numOut := x.NumOut() // count number of return values

	if numIn != 2 {
		return false
	}
	if numOut != 0 {
		return false
	}
	{
		inV := x.In(0)
		if inV.Kind() != reflect.Ptr {
			return false
		}
		inU := inV.Elem()
		if inU.PkgPath() != "k8s.io/apimachinery/pkg/apis/meta/v1" || inU.Name() != "ObjectMeta" {
			return false
		}
	}
	{
		inV := x.In(0)
		if inV.Kind() != reflect.Struct {
			return false
		}
		if inV.PkgPath() != "github.com/google/gofuzz" || inV.Name() != "Continue" {
			return false
		}
	}
	return true
}
