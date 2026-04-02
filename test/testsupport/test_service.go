package testsupport

import (
	sync "sync"

	"github.com/go-faker/faker/v4"
)

var Service *FakeService = NewFakeService()

// FakeService is a struct to hold our in-memory "database" and mock methods.
type FakeService struct {
	mu        sync.RWMutex
	resources map[int32]*Resource
	tenants   map[int32]*Tenant
}

// NewFakeService creates and returns a new instance of MockService.
func NewFakeService() *FakeService {
	s := &FakeService{
		resources: make(map[int32]*Resource),
		tenants:   make(map[int32]*Tenant),
	}
	s.seedData()
	return s
}

// ListResources returns all Resources.
func (s *FakeService) ListResources(tenantID int32) []*Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Resource
	for _, res := range s.resources {
		if res.TenantID == tenantID {
			result = append(result, res)
		}
	}
	return result
}

// GetResource retrieves a Resource by ID.
func (s *FakeService) GetResource(id int32) (*Resource, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res, exists := s.resources[id]
	return res, exists
}

// PutResource adds or updates a Resource.
func (s *FakeService) PutResource(resource *Resource) *Resource {
	s.mu.Lock()
	defer s.mu.Unlock()

	if resource.ResourceID == 0 {
		resource.ResourceID = int32(len(s.resources) + 1)
	}

	s.resources[resource.ResourceID] = resource
	return resource
}

// ListTenants returns all Tenants.
func (s *FakeService) ListTenants() []*Tenant {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Tenant
	for _, tenant := range s.tenants {
		result = append(result, tenant)
	}
	return result
}

// GetTenant retrieves a Tenant by ID.
func (s *FakeService) GetTenant(id int32) (*Tenant, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tenant, exists := s.tenants[id]
	return tenant, exists
}

// PutTenant adds or updates a Tenant.
func (s *FakeService) PutTenant(tenant *Tenant) *Tenant {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tenants[tenant.TenantID] = tenant
	return tenant
}

// DeleteTenant removes a tenant if it exists and returns true when deleted.
func (s *FakeService) DeleteTenant(id int32) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tenants[id]; !ok {
		return false
	}
	delete(s.tenants, id)
	return true
}

// Reset clears accumulated data and reinitializes with default seed data.
// This is useful for benchmarks to prevent unbounded data growth.
func (s *FakeService) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources = make(map[int32]*Resource)
	s.tenants = make(map[int32]*Tenant)
	// Re-seed with initial data (unlocked helper)
	s.seedData()
}

// seedData populates initial test data. Must be called while holding the lock.
func (s *FakeService) seedData() {
	for i := 0; i < 10; i++ {
		tenantID := int32(i)
		s.tenants[tenantID] = &Tenant{
			TenantID: tenantID,
			Name:     faker.DomainName(),
			Plan:     "diamond",
		}
		for j := 0; j < 10; j++ {
			resourceID := int32(len(s.resources) + 1)
			s.resources[resourceID] = &Resource{
				ResourceID: resourceID,
				TenantID:   tenantID,
				Name:       faker.MacAddress(),
				Type:       "resource",
			}
		}
	}
}
