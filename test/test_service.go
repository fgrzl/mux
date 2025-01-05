//go:generate go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
//go:generate go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
//go:generate protoc --go_out=./ --go_opt=paths=source_relative --proto_path=./ *.proto

package test

import (
	sync "sync"

	"github.com/go-faker/faker/v4"
)

// TestService is a struct to hold our in-memory "database" and mock methods.
type TestService struct {
	mu        sync.Mutex
	resources map[int32]*Resource
	tenants   map[int32]*Tenant
}

// NewFakeService creates and returns a new instance of MockService.
func NewFakeService() *TestService {
	s := &TestService{
		resources: make(map[int32]*Resource),
		tenants:   make(map[int32]*Tenant),
	}
	for i := 0; i < 10; i++ {
		tenantId := int32(i)
		s.PutTenant(&Tenant{
			TenantId: int32(i),
			Name:     faker.DomainName(),
			Plan:     "diamond",
		})
		for i := 0; i < 10; i++ {
			s.PutResource(&Resource{TenantId: tenantId, Name: faker.MacAddress(), Type: "resource"})
		}
	}
	return s
}

// ListResources returns all Resources.
func (s *TestService) ListResources(tenantId int32) []*Resource {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result []*Resource
	for _, res := range s.resources {
		if res.TenantId == tenantId {
			result = append(result, res)
		}
	}
	return result
}

// GetResource retrieves a Resource by ID.
func (s *TestService) GetResource(id int32) (*Resource, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	res, exists := s.resources[id]
	return res, exists
}

// CreateResource adds a new Resource.
func (s *TestService) PutResource(resource *Resource) *Resource {
	s.mu.Lock()
	defer s.mu.Unlock()

	if resource.ResourceId == 0 {
		resource.ResourceId = int32(len(s.resources) + 1)
	}

	s.resources[resource.ResourceId] = resource
	return resource
}

// ListTenants returns all Tenants.
func (s *TestService) ListTenants() []*Tenant {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result []*Tenant
	for _, tenant := range s.tenants {
		result = append(result, tenant)
	}
	return result
}

// GetTenant retrieves a Tenant by ID.
func (s *TestService) GetTenant(id int32) (*Tenant, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tenant, exists := s.tenants[id]
	return tenant, exists
}

// CreateTenant adds a new Tenant.
func (s *TestService) PutTenant(tenant *Tenant) *Tenant {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tenants[tenant.TenantId] = tenant
	return tenant
}
