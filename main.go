package main

import (
	"fmt"
	fuzz "github.com/google/gofuzz"
	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	metafuzzer "k8s.io/apimachinery/pkg/apis/meta/fuzzer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"reflect"
	"strconv"
)

/*

	GenerateName string `json:"generateName,omitempty" protobuf:"bytes,2,opt,name=generateName"`
	SelfLink string `json:"selfLink,omitempty" protobuf:"bytes,4,opt,name=selfLink"`
	UID types.UID `json:"uid,omitempty" protobuf:"bytes,5,opt,name=uid,casttype=k8s.io/kubernetes/pkg/types.UID"`
	ResourceVersion string `json:"resourceVersion,omitempty" protobuf:"bytes,6,opt,name=resourceVersion"`
	Generation int64 `json:"generation,omitempty" protobuf:"varint,7,opt,name=generation"`
	CreationTimestamp Time `json:"creationTimestamp,omitempty" protobuf:"bytes,8,opt,name=creationTimestamp"`
	DeletionTimestamp *Time `json:"deletionTimestamp,omitempty" protobuf:"bytes,9,opt,name=deletionTimestamp"`
	DeletionGracePeriodSeconds *int64 `json:"deletionGracePeriodSeconds,omitempty" protobuf:"varint,10,opt,name=deletionGracePeriodSeconds"`
	OwnerReferences []OwnerReference `json:"ownerReferences,omitempty" patchStrategy:"merge" patchMergeKey:"uid" protobuf:"bytes,13,rep,name=ownerReferences"`
	ClusterName string `json:"clusterName,omitempty" protobuf:"bytes,15,opt,name=clusterName"`
	ManagedFields []ManagedFieldsEntry `json:"managedFields,omitempty" protobuf:"bytes,17,rep,name=managedFields"`
*/

// CRDSafeFuzzerFuncs will merge the given funcLists, overriding early funcs with later ones if there first
// argument has the same type.
func CRDSafeFuzzerFuncs(funcs ...fuzzer.FuzzerFuncs) fuzzer.FuzzerFuncs {
	return fuzzer.FuzzerFuncs(func(codecs serializer.CodecFactory) []interface{} {
		result := []interface{}{}
		for _, fns := range funcs {
			if fns == nil {
				continue
			}
			for _, f := range fns(codecs) {
				if !matches(f) {
					result = append(result, f)
				}
			}
		}
		result = append(result, func(j *metav1.ObjectMeta, c fuzz.Continue) {
			c.FuzzNoCustom(j)

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
			if len(j.Finalizers) == 0 {
				j.Finalizers = nil
			}

			j.GenerateName = ""
			j.SelfLink = ""
			j.UID = ""
			j.ResourceVersion = ""
			j.Generation = 0
			j.CreationTimestamp = metav1.Time{}
			j.DeletionTimestamp = nil
			j.DeletionGracePeriodSeconds = nil
			j.OwnerReferences = nil
			j.ClusterName = ""
			j.ManagedFields = nil
		})
		return result
	})
}

func main() {
	scheme := clientsetscheme.Scheme
	codecFactory := serializer.NewCodecFactory(scheme)
	fns := CRDSafeFuzzerFuncs(metafuzzer.Funcs)(codecFactory)

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
		inV := x.In(1)
		if inV.Kind() != reflect.Struct {
			return false
		}
		if inV.PkgPath() != "github.com/google/gofuzz" || inV.Name() != "Continue" {
			return false
		}
	}
	return true
}
