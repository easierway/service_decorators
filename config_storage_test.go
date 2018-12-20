package service_decorators

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/hashicorp/consul/api"
)

func TestGetAndSetForConsul(t *testing.T) {
	storage, err := CreateConsulConfigStorage(&api.Config{})
	if err != nil {
		t.Log(err)
	}
	keyName := "ChaosExp"
	configStr := `{
    "IsTestingByChaosResponseFn" : true,
    "AdditionalResponseTime" : 500,
    "ChaosRate" : 20
   }`
	storage.Set(keyName, []byte(configStr))
	v, err := storage.Get(keyName)
	if err != nil {
		if strings.Contains(err.Error(), "connection") {
			t.Log("Warning: The failure is caused by Consul connection.")
		} else {
			t.Error(err)
		}
	}
	t.Log(string(v))
	config := ChaosEngineeringConfig{}
	err = json.Unmarshal([]byte(v), &config)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%v", config)
}
