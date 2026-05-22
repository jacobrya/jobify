package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUser_JSON(t *testing.T) {
	id := uuid.New()
	now := time.Now().Truncate(time.Second)
	user := User{
		ID:        id,
		Email:     "test@example.com",
		Password:  "secret",
		Role:      RoleDeveloper,
		CreatedAt: now,
		UpdatedAt: now,
	}

	data, err := json.Marshal(user)
	assert.NoError(t, err)

	var decoded User
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, user.ID, decoded.ID)
	assert.Equal(t, user.Email, decoded.Email)
	assert.Equal(t, "", decoded.Password) // Password should be ignored by json:"-"
	assert.Equal(t, user.Role, decoded.Role)
	assert.True(t, user.CreatedAt.Equal(decoded.CreatedAt))
}

func TestJob_JSON(t *testing.T) {
	id := uuid.New()
	job := Job{
		ID:      id,
		Title:   "Go Developer",
		Company: "Tech Inc",
		Skills:  []string{"Go", "PostgreSQL"},
	}

	data, err := json.Marshal(job)
	assert.NoError(t, err)

	var decoded Job
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, job.ID, decoded.ID)
	assert.Equal(t, job.Title, decoded.Title)
	assert.Equal(t, job.Skills, decoded.Skills)
}

func TestApplication_JSON(t *testing.T) {
	id := uuid.New()
	app := Application{
		ID:     id,
		Status: StatusApplied,
		Note:   "Excited to join!",
	}

	data, err := json.Marshal(app)
	assert.NoError(t, err)

	var decoded Application
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, app.ID, decoded.ID)
	assert.Equal(t, app.Status, decoded.Status)
	assert.Equal(t, app.Note, decoded.Note)
}
