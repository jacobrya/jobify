package handler

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthRequests_JSON(t *testing.T) {
	t.Run("registerRequest", func(t *testing.T) {
		req := registerRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		data, err := json.Marshal(req)
		assert.NoError(t, err)

		var decoded registerRequest
		err = json.Unmarshal(data, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, req.Email, decoded.Email)
		assert.Equal(t, req.Password, decoded.Password)
	})

	t.Run("loginRequest", func(t *testing.T) {
		req := loginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		data, err := json.Marshal(req)
		assert.NoError(t, err)

		var decoded loginRequest
		err = json.Unmarshal(data, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, req.Email, decoded.Email)
		assert.Equal(t, req.Password, decoded.Password)
	})

	t.Run("tokenRequest", func(t *testing.T) {
		req := tokenRequest{
			RefreshToken: "some_token",
		}
		data, err := json.Marshal(req)
		assert.NoError(t, err)

		var decoded tokenRequest
		err = json.Unmarshal(data, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, req.RefreshToken, decoded.RefreshToken)
	})
}
