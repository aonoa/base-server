package casbin

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	adminv1 "base-server/api/gen/go/admin/service/v1"
)

const serviceRegistryCacheTTL = 30 * time.Second

type serviceRegistryEntry struct {
	prefix      string
	serviceCode string
}

type serviceRegistryResolver struct {
	list func(context.Context) ([]*adminv1.ServiceRegistryItem, error)
	now  func() time.Time
	ttl  time.Duration

	mu        sync.RWMutex
	entries   []serviceRegistryEntry
	expiresAt time.Time
}

func newServiceRegistryResolver(list func(context.Context) ([]*adminv1.ServiceRegistryItem, error)) *serviceRegistryResolver {
	return &serviceRegistryResolver{
		list: list,
		now:  time.Now,
		ttl:  serviceRegistryCacheTTL,
	}
}

func (r *serviceRegistryResolver) Resolve(ctx context.Context, path string) (string, error) {
	entries, err := r.loadEntries(ctx)
	if err != nil {
		return "", err
	}
	return selectServiceCode(entries, path), nil
}

func (r *serviceRegistryResolver) loadEntries(ctx context.Context) ([]serviceRegistryEntry, error) {
	if r == nil || r.list == nil {
		return nil, nil
	}
	now := r.now()

	r.mu.RLock()
	if len(r.entries) > 0 && now.Before(r.expiresAt) {
		entries := append([]serviceRegistryEntry(nil), r.entries...)
		r.mu.RUnlock()
		return entries, nil
	}
	r.mu.RUnlock()

	items, err := r.list(ctx)
	if err != nil {
		r.mu.RLock()
		defer r.mu.RUnlock()
		if len(r.entries) > 0 {
			return append([]serviceRegistryEntry(nil), r.entries...), nil
		}
		return nil, err
	}
	entries := buildServiceRegistryEntries(items)

	r.mu.Lock()
	r.entries = entries
	r.expiresAt = now.Add(r.ttl)
	r.mu.Unlock()

	return append([]serviceRegistryEntry(nil), entries...), nil
}

func buildServiceRegistryEntries(items []*adminv1.ServiceRegistryItem) []serviceRegistryEntry {
	entries := make([]serviceRegistryEntry, 0, len(items))
	for _, item := range items {
		if item == nil || item.Status == 0 || item.ServiceCode == "" {
			continue
		}
		prefix := normalizeHTTPPrefix(item.HttpPrefix)
		if prefix == "" {
			continue
		}
		entries = append(entries, serviceRegistryEntry{
			prefix:      prefix,
			serviceCode: strings.TrimSpace(item.ServiceCode),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if len(entries[i].prefix) == len(entries[j].prefix) {
			return entries[i].prefix < entries[j].prefix
		}
		return len(entries[i].prefix) > len(entries[j].prefix)
	})
	return entries
}

func selectServiceCode(entries []serviceRegistryEntry, path string) string {
	path = normalizeHTTPPrefix(path)
	if path == "" {
		return ""
	}
	for _, entry := range entries {
		if path == entry.prefix || strings.HasPrefix(path, entry.prefix+"/") {
			return entry.serviceCode
		}
	}
	return ""
}

func normalizeHTTPPrefix(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	if len(value) > 1 {
		value = strings.TrimRight(value, "/")
	}
	return value
}
