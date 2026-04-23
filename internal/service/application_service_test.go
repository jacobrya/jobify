package service

import (
	"context"
	"errors"
	"testing"

	"github.com/abzalserikbay/jobify/internal/domain"
	"github.com/google/uuid"
)

type mockAppRepo struct {
	createFn       func(ctx context.Context, a *domain.Application) error
	getByIDFn      func(ctx context.Context, id uuid.UUID) (*domain.Application, error)
	getByUserIDFn  func(ctx context.Context, userID uuid.UUID) ([]domain.Application, error)
	updateStatusFn func(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus) error
	deleteFn       func(ctx context.Context, id uuid.UUID) error

	lastStatus domain.ApplicationStatus
}

func (m *mockAppRepo) Create(ctx context.Context, a *domain.Application) error {
	return m.createFn(ctx, a)
}

func (m *mockAppRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Application, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockAppRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Application, error) {
	return m.getByUserIDFn(ctx, userID)
}

func (m *mockAppRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus) error {
	m.lastStatus = status
	if m.updateStatusFn == nil {
		return nil
	}
	return m.updateStatusFn(ctx, id, status)
}

func (m *mockAppRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.deleteFn(ctx, id)
}

func TestApplicationService_Create(t *testing.T) {
	userID := uuid.New()
	jobID := uuid.New()

	repo := &mockAppRepo{
		createFn: func(ctx context.Context, a *domain.Application) error { return nil },
	}
	svc := NewApplicationService(repo)

	app, err := svc.Create(context.Background(), userID, jobID, "interesting role")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.UserID != userID || app.JobID != jobID {
		t.Errorf("ids: got user=%v job=%v, want user=%v job=%v", app.UserID, app.JobID, userID, jobID)
	}
	if app.Status != domain.StatusSaved {
		t.Errorf("status: got %q, want %q", app.Status, domain.StatusSaved)
	}
	if app.Note != "interesting role" {
		t.Errorf("note: got %q, want %q", app.Note, "interesting role")
	}
	if app.ID == uuid.Nil {
		t.Error("ID must be assigned")
	}
}

func TestApplicationService_UpdateStatus(t *testing.T) {
	ownerID := uuid.New()
	strangerID := uuid.New()
	appID := uuid.New()
	errDB := errors.New("db down")

	tests := []struct {
		name       string
		callerID   uuid.UUID
		newStatus  domain.ApplicationStatus
		current    domain.ApplicationStatus
		getByIDErr error
		wantErr    error
	}{
		{
			name:      "saved to applied",
			callerID:  ownerID,
			newStatus: domain.StatusApplied,
			current:   domain.StatusSaved,
			wantErr:   nil,
		},
		{
			name:      "applied to interview",
			callerID:  ownerID,
			newStatus: domain.StatusInterview,
			current:   domain.StatusApplied,
			wantErr:   nil,
		},
		{
			name:      "interview to offer",
			callerID:  ownerID,
			newStatus: domain.StatusOffer,
			current:   domain.StatusInterview,
			wantErr:   nil,
		},
		{
			name:      "rejected allowed",
			callerID:  ownerID,
			newStatus: domain.StatusRejected,
			current:   domain.StatusApplied,
			wantErr:   nil,
		},
		{
			name:      "non-owner forbidden",
			callerID:  strangerID,
			newStatus: domain.StatusApplied,
			current:   domain.StatusSaved,
			wantErr:   domain.ErrForbidden,
		},
		{
			name:      "unknown status rejected",
			callerID:  ownerID,
			newStatus: domain.ApplicationStatus("bogus"),
			current:   domain.StatusSaved,
			wantErr:   domain.ErrInvalidInput,
		},
		{
			name:       "application not found",
			callerID:   ownerID,
			newStatus:  domain.StatusApplied,
			getByIDErr: domain.ErrNotFound,
			wantErr:    domain.ErrNotFound,
		},
		{
			name:       "repo error",
			callerID:   ownerID,
			newStatus:  domain.StatusApplied,
			getByIDErr: errDB,
			wantErr:    errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAppRepo{
				getByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Application, error) {
					if tt.getByIDErr != nil {
						return nil, tt.getByIDErr
					}
					return &domain.Application{ID: id, UserID: ownerID, Status: tt.current}, nil
				},
			}
			svc := NewApplicationService(repo)

			err := svc.UpdateStatus(context.Background(), appID, tt.callerID, tt.newStatus)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err: got %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && repo.lastStatus != tt.newStatus {
				t.Errorf("persisted status: got %q, want %q", repo.lastStatus, tt.newStatus)
			}
		})
	}
}

func TestApplicationService_Delete(t *testing.T) {
	ownerID := uuid.New()
	strangerID := uuid.New()
	appID := uuid.New()

	tests := []struct {
		name       string
		callerID   uuid.UUID
		getByIDErr error
		wantErr    error
		wantDelete bool
	}{
		{
			name:       "owner deletes",
			callerID:   ownerID,
			wantErr:    nil,
			wantDelete: true,
		},
		{
			name:     "non-owner forbidden",
			callerID: strangerID,
			wantErr:  domain.ErrForbidden,
		},
		{
			name:       "not found",
			callerID:   ownerID,
			getByIDErr: domain.ErrNotFound,
			wantErr:    domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deleteCalled := false
			repo := &mockAppRepo{
				getByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.Application, error) {
					if tt.getByIDErr != nil {
						return nil, tt.getByIDErr
					}
					return &domain.Application{ID: id, UserID: ownerID, Status: domain.StatusApplied}, nil
				},
				deleteFn: func(ctx context.Context, id uuid.UUID) error {
					deleteCalled = true
					return nil
				},
			}
			svc := NewApplicationService(repo)

			err := svc.Delete(context.Background(), appID, tt.callerID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err: got %v, want %v", err, tt.wantErr)
			}
			if deleteCalled != tt.wantDelete {
				t.Errorf("delete called: got %v, want %v", deleteCalled, tt.wantDelete)
			}
		})
	}
}
