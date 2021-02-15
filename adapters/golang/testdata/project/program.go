package project

import (
	"fmt"

	"github.com/bazelbuild/buildtools/build"
)

func FakeFunction() {
	var x *build.File
	fmt.Printf(" %s \n", x.Path)
}
