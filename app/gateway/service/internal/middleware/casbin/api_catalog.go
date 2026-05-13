package casbin

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	adminv1 "base-server/api/gen/go/admin/service/v1"
)

const apiCatalogCacheTTL = 30 * time.Second

type apiCatalogEntry struct {
	path          string
	method        string
	resourceGroup string
}

type apiOwnership struct {
	ServiceCode   string
	ResourceGroup string
}

type apiCatalogResolver struct {
	list func(context.Context) ([]*adminv1.ApiListItem, error)
	now  func() time.Time
	ttl  time.Duration

	mu        sync.RWMutex
	entries   []apiCatalogEntry
	expiresAt time.Time
}

func newAPICatalogResolver(list func(context.Context) ([]*adminv1.ApiListItem, error)) *apiCatalogResolver {
	return &apiCatalogResolver{
		list: list,
		now:  time.Now,
		ttl:  apiCatalogCacheTTL,
	}
}

func (r *apiCatalogResolver) Resolve(ctx context.Context, path, method string) (apiOwnership, error) {
	entries, err := r.loadEntries(ctx)
	if err != nil {
		return apiOwnership{}, err
	}
	return selectAPIOwnership(entries, path, method), nil
}

func (r *apiCatalogResolver) loadEntries(ctx context.Context) ([]apiCatalogEntry, error) {
	if r == nil || r.list == nil {
		return nil, nil
	}
	now := r.now()

	r.mu.RLock()
	if len(r.entries) > 0 && now.Before(r.expiresAt) {
		entries := append([]apiCatalogEntry(nil), r.entries...)
		r.mu.RUnlock()
		return entries, nil
	}
	r.mu.RUnlock()

	items, err := r.list(ctx)
	if err != nil {
		r.mu.RLock()
		defer r.mu.RUnlock()
		if len(r.entries) > 0 {
			return append([]apiCatalogEntry(nil), r.entries...), nil
		}
		return nil, err
	}
	entries := buildAPICatalogEntries(items)

	r.mu.Lock()
	r.entries = entries
	r.expiresAt = now.Add(r.ttl)
	r.mu.Unlock()

	return append([]apiCatalogEntry(nil), entries...), nil
}

func buildAPICatalogEntries(items []*adminv1.ApiListItem) []apiCatalogEntry {
	entries := make([]apiCatalogEntry, 0, len(items))
	for _, item := range items {
		if item == nil || strings.TrimSpace(item.Path) == "" || strings.TrimSpace(item.Method) == "" {
			continue
		}
		entries = append(entries, apiCatalogEntry{
			path:          normalizeHTTPPrefix(item.Path),
			method:        strings.ToUpper(strings.TrimSpace(item.Method)),
			resourceGroup: strings.TrimSpace(item.ResourcesGroup),
		})
	}
	return entries
}

func selectAPIOwnership(entries []apiCatalogEntry, path, method string) apiOwnership {
	path = normalizeHTTPPrefix(path)
	method = strings.ToUpper(strings.TrimSpace(method))
	for _, entry := range entries {
		if entry.method != method {
			continue
		}
		if entry.path == path || matchAPIPath(path, entry.path) {
			return apiOwnership{
				ResourceGroup: entry.resourceGroup,
			}
		}
	}
	return apiOwnership{}
}

var apiCatalogPathParam = regexp.MustCompile(`\{[^/]+\}`)

func matchAPIPath(path, pattern string) bool {
	if idx := strings.Index(path, "?"); idx >= 0 {
		path = path[:idx]
	}
	pattern = strings.ReplaceAll(pattern, "/*", "/.*")
	pattern = apiCatalogPathParam.ReplaceAllString(pattern, `[^/]+`)
	re, err := regexp.Compile("^" + pattern + "$")
	if err != nil {
		return false
	}
	return re.MatchString(path)
}
