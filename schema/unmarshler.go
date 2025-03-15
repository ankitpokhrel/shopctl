package schema

import (
	"encoding/json"
	"strconv"
)

func (m *MoneyV2) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	m.CurrencyCode = CurrencyCode(raw["currencyCode"].(string))

	if val, isFloat := raw["amount"].(float64); isFloat {
		m.Amount = val
	} else {
		if val, ok := raw["amount"]; ok {
			amount, err := strconv.ParseFloat(val.(string), 64)
			if err != nil {
				return err
			}
			m.Amount = amount
		}
	}
	return nil
}
