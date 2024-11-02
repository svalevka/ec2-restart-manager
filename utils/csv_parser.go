// utils/csv_parser.go
package utils

import (
	"encoding/csv"
	"io"
	"strings"

	"ec2-restart-manager/models"

	"github.com/jszwec/csvutil"
)

func ParseCSVToStruct(csvContent []byte) ([]models.EC2Instance, error) {
	var instances []models.EC2Instance
	reader := csv.NewReader(strings.NewReader(string(csvContent)))
	decoder, err := csvutil.NewDecoder(reader)
	if err != nil {
		return nil, err
	}

	for {
		var instance models.EC2Instance
		if err := decoder.Decode(&instance); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if instance.State == "running" {
			instances = append(instances, instance)
		}
	}
	return instances, nil
}
