package utils

import (
	"io"

	"github.com/sirupsen/logrus"
)

func Close(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		logrus.Warn("Can not close stream: ", err)
	}
}
