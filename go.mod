module github.com/chr-fritz/knx-exporter

go 1.16

require (
	github.com/coreos/go-systemd/v22 v22.3.2
	github.com/ghodss/yaml v1.0.0
	github.com/golang/glog v1.0.0
	github.com/golang/mock v1.6.0
	github.com/golangci/golangci-lint v1.43.0 // indirect
	github.com/heptiolabs/healthcheck v0.0.0-20211123025425-613501dd5deb
	github.com/kr/text v0.2.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/prometheus/client_golang v1.11.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.9.0
	github.com/stretchr/testify v1.7.0
	github.com/vapourismo/knx-go v0.0.0-20211128234507-8198fa17db36
	golang.org/x/tools v0.1.8 // indirect
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

replace github.com/golang/glog => github.com/kubermatic/glog-logrus v0.0.0-20180829085450-3fa5b9870d1d
