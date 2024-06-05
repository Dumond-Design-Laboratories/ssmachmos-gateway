package cli

import (
	"encoding/json"

	"github.com/jukuly/ss_mach_mo/server/internal/model"
)

func sensorJSONToString(jsonStr []byte) (string, error) {
	s := model.Sensor{}
	err := json.Unmarshal(jsonStr, &s)
	if err != nil {
		return "", err
	}

	return s.ToString(), nil
}
