package project

import (
	"github.com/bazelbuild/buildtools/build"
	"github.com/sirupsen/logrus"
)

func FakeFunction() {
	var x *build.File
	logrus.Infof(" %s \n", x.Path)
}
