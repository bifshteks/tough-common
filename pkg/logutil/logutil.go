package logutil

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

var DummyLogger = &logrus.Logger{Out: ioutil.Discard}
