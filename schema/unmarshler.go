package schema

import (
	"encoding/json"
	"strconv"
)

func (m *MoneyV2) UnmarshalJSON(data []byte) error {
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if amountStr, ok := raw["amount"]; ok {
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return err
		}
		m.Amount = amount
	}
	return nil
}
