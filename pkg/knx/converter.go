package knx

import (
	"encoding/xml"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"

	"github.com/chr-fritz/knx-exporter/pkg/knx/export"
	"github.com/chr-fritz/knx-exporter/pkg/utils"
)

func ConvertGroupAddresses(src string, target string) error {
	addressExport, err := parseExport(src)
	if err != nil {
		return err
	}

	groupAddresses := collectGroupAddresses(addressExport.GroupRange)

	addressConfigs := convertAddresses(groupAddresses)
	cfg := Config{
		AddressConfigs: addressConfigs,
		MetricsPrefix:  "knx_",
	}

	return writeConfig(cfg, target)
}

func writeConfig(cfg Config, target string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("can not marshal config: %s", err)
	}
	targetFile, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("can not create file %s: %s", target, err)
	}
	defer utils.Close(targetFile)

	_, err = targetFile.Write(data)
	if err != nil {
		return fmt.Errorf("can not write config into %s: %s", target, err)
	}
	return nil
}

func parseExport(src string) (export.GroupAddressExport, error) {
	source, err := os.Open(src)
	if err != nil {
		return export.GroupAddressExport{}, fmt.Errorf("can not open source file '%s': %s", src, err)
	}
	defer utils.Close(source)

	decoder := xml.NewDecoder(source)
	addressExport := export.GroupAddressExport{}
	err = decoder.Decode(&addressExport)
	if err != nil {
		return export.GroupAddressExport{}, fmt.Errorf("can not parse group address export: %s", err)
	}
	return addressExport, nil
}

func collectGroupAddresses(groupRange []export.GroupRange) []export.GroupAddress {
	var addresses []export.GroupAddress

	for _, gr := range groupRange {
		for _, address := range gr.GroupAddress {
			addresses = append(addresses, address)
		}
		addresses = append(addresses, collectGroupAddresses(gr.GroupRange)...)
	}

	return addresses
}

func convertAddresses(groupAddresses []export.GroupAddress) map[GroupAddress]GroupAddressConfig {
	addressConfigs := make(map[GroupAddress]GroupAddressConfig)
	for _, ga := range groupAddresses {
		address, err := NewGroupAddress(ga.Address)
		if err != nil {
			logrus.Warnf("Can not convert address '%s': %s", ga.Address, err)
			continue
		}

		name, err := normalizeMetricName(ga.Name)
		if err != nil {
			logrus.Info("Can not normalize group address name, ", err)
		}
		cfg := GroupAddressConfig{
			Name:       name,
			Comment:    ga.Name + "\n" + ga.Description,
			DPT:        ga.DPTs,
			MetricType: "",
			Export:     false,
			ReadActive: false,
			MaxAge:     0,
		}
		addressConfigs[address] = cfg
	}
	return addressConfigs
}

var validRegex = regexp.MustCompilePOSIX("^[a-zA-Z_:][a-zA-Z0-9_:]*$")
var replaceRegex = regexp.MustCompilePOSIX("[^a-zA-Z0-9_:]")
var latin1Replacer = strings.NewReplacer("Ä", "Ae", "Ü", "Ue", "Ö", "Oe", "ä", "ae", "ü", "ue", "ö", "oe", "ß", "ss")

func normalizeMetricName(name string) (string, error) {
	if validRegex.MatchString(name) {
		return name, nil
	}

	normalized := latin1Replacer.Replace(name)
	if validRegex.MatchString(normalized) {
		return normalized, nil
	}
	normalized = replaceRegex.ReplaceAllLiteralString(normalized, "_")
	if !validRegex.MatchString(normalized) {
		return "", fmt.Errorf("the group address name \"%s\" don't matchs the following regex: [a-zA-Z_:][a-zA-Z0-9_:]*", name)
	}
	return normalized, nil
}
