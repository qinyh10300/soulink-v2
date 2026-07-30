package gate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"soulink-backend/internal/store"
)

const GateKeyMatchingWrite = "matching_write"

type Status struct {
	GateKey  string
	StartsAt time.Time
	EndsAt   time.Time
	Reason   string
	Active   bool
}

type Service struct {
	store   *store.Store
	nowFunc func() time.Time
}

// NewService 创建运行时接口开关服务。
func NewService(dataStore *store.Store) *Service {
	return &Service{store: dataStore, nowFunc: time.Now}
}

// Check 返回指定 gate 当前是否处于关闭状态。
func (s *Service) Check(ctx context.Context, gateKey string) (Status, error) {
	if s == nil || s.store == nil {
		return Status{GateKey: gateKey}, nil
	}
	gateKey = strings.TrimSpace(gateKey)
	if gateKey == "" {
		return Status{}, fmt.Errorf("gate key is required")
	}

	window, err := s.store.GetInterfaceGateWindow(ctx, nil, gateKey)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return Status{GateKey: gateKey}, nil
		}
		return Status{}, err
	}

	now := s.currentTime()
	return Status{
		GateKey:  gateKey,
		StartsAt: window.StartsAt,
		EndsAt:   window.EndsAt,
		Reason:   strings.TrimSpace(window.Reason),
		Active:   isWindowActive(window.StartsAt, window.EndsAt, now),
	}, nil
}

// SetWindow 设置指定 gate 的关闭时间窗。
func (s *Service) SetWindow(ctx context.Context, gateKey string, startsAt, endsAt time.Time, reason string) (store.InterfaceGateWindow, error) {
	if s == nil || s.store == nil {
		return store.InterfaceGateWindow{}, fmt.Errorf("gate service is not ready")
	}
	gateKey = strings.TrimSpace(gateKey)
	if gateKey == "" {
		return store.InterfaceGateWindow{}, fmt.Errorf("gate key is required")
	}
	if !endsAt.After(startsAt) {
		return store.InterfaceGateWindow{}, fmt.Errorf("gate end time must be after start time")
	}

	now := s.currentTime()
	return s.store.UpsertInterfaceGateWindow(ctx, nil, store.UpsertInterfaceGateWindowParams{
		GateKey:   gateKey,
		StartsAt:  startsAt,
		EndsAt:    endsAt,
		Reason:    strings.TrimSpace(reason),
		CreatedAt: now,
		UpdatedAt: now,
	})
}

// ClearWindow 清除指定 gate 的关闭时间窗。
func (s *Service) ClearWindow(ctx context.Context, gateKey string) error {
	if s == nil || s.store == nil {
		return fmt.Errorf("gate service is not ready")
	}
	gateKey = strings.TrimSpace(gateKey)
	if gateKey == "" {
		return fmt.Errorf("gate key is required")
	}
	err := s.store.DeleteInterfaceGateWindow(ctx, nil, gateKey)
	if errors.Is(err, store.ErrNotFound) {
		return nil
	}
	return err
}

func (s *Service) currentTime() time.Time {
	if s != nil && s.nowFunc != nil {
		return s.nowFunc()
	}
	return time.Now()
}

func isWindowActive(startsAt, endsAt, now time.Time) bool {
	return !now.Before(startsAt) && now.Before(endsAt)
}
