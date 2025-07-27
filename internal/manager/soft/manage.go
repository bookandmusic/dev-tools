package soft

import (
	"context"
	"fmt"
	"sync"
)

type SoftManage interface {
	Install(ctx context.Context) error
	Uninstall(ctx context.Context) error
	Update(ctx context.Context) error
}

var (
	registry = make(map[string]SoftManage)
	mu       sync.RWMutex
)

func Register(name string, m SoftManage) {
	mu.Lock()
	defer mu.Unlock()
	registry[name] = m
}

func GetManager(name string) (SoftManage, error) {
	mu.RLock()
	defer mu.RUnlock()
	m, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("manager %s not found", name)
	}
	return m, nil
}
