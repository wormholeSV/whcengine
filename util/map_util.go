package util

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

func To_Map(t interface{}) map[string]interface{} {
	mt := make(map[string]interface{})
	j, _ := json.Marshal(t)
	json.Unmarshal(j, &mt)
	return mt
}

func Merge_Map(t map[string]interface{}, s map[string]interface{}) map[string]interface{} {
	for k, v := range s {
		t[k] = v
	}

	return t
}

func FixDecimal(base map[string]interface{}, key string) {
	val := base[key]
	if val != nil && val.(string) != "" {
		fixVal, _ := decimal.NewFromString(val.(string))
		base[key] = fixVal
	}
}
