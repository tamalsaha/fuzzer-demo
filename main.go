package main

import (
	"fmt"
	fuzz "github.com/google/gofuzz"
	"gomodules.xyz/refutil"
	metafuzzer "k8s.io/apimachinery/pkg/apis/meta/fuzzer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
)

func main() {
	scheme := clientsetscheme.Scheme
	codecFactory := serializer.NewCodecFactory(scheme)
	fns := metafuzzer.Funcs(codecFactory)
	idx := refutil.Index(fns, func(j *metav1.ObjectMeta, c fuzz.Continue) {})
	fmt.Println(idx)
}
