package wallet

import (
	"fmt"
)

func castRequestBodyToAmount(requestBody interface{}) (Amount, error) {
	rawAmount, ok := requestBody.(map[string]interface{})
	if !ok {
		return Amount{}, fmt.Errorf("failed to cast requestBody=%#+v to Amount", requestBody)
	}

	amount, ok := rawAmount["amount"].(float64)
	if !ok {
		return Amount{}, fmt.Errorf("failed to cast rawAmount=%#+v to float64", rawAmount)
	}

	return Amount{Amount: amount}, nil
}
