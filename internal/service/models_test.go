package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenPair_JSON(t *testing.T) {
	pair := TokenPair{
		AccessToken:  "access",
		RefreshToken: "refresh",
	}

	data, err := json.Marshal(pair)
	assert.NoError(t, err)

	var decoded TokenPair
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, pair.AccessToken, decoded.AccessToken)
	assert.Equal(t, pair.RefreshToken, decoded.RefreshToken)
}
