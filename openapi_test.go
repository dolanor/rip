package rip

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEntity is a sample entity for testing OpenAPI generation
type TestEntity struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Count       int       `json:"count"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	Tags        []string  `json:"tags,omitempty"`
}

// Type returns the reflect Type of the entity for OpenAPI generation
func (te TestEntity) Type() reflect.Type {
	return reflect.TypeOf(te)
}

// Address is a nested struct for testing complex schema generation
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	ZipCode string `json:"zip_code"`
}

// ComplexTestEntity includes nested structures for testing advanced schema generation
type ComplexTestEntity struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Addresses    []Address `json:"addresses"`
	MainAddress  *Address  `json:"main_address,omitempty"`
}

// MockEntityProvider implements the EntityProvider interface for testing
type MockEntityProvider struct{}

func (p *MockEntityProvider) Create(ctx context.Context, ent TestEntity) (TestEntity, error) {
	return ent, nil
}

func (p *MockEntityProvider) Get(ctx context.Context, id string) (TestEntity, error) {
	return TestEntity{ID: id}, nil
}

func (p *MockEntityProvider) Update(ctx context.Context, ent TestEntity) error {
	return nil
}

func (p *MockEntityProvider) Delete(ctx context.Context, id string) error {
	return nil
}

func (p *MockEntityProvider) List(ctx context.Context, offset, limit int) ([]TestEntity, error) {
	return []TestEntity{}, nil
}

// Helper function to generate and parse OpenAPI spec for testing
func generateAndParseOpenAPI(t *testing.T, setupFunc func()) (*openapi3.T, string, error) {
	// Set up the test
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "openapi.json")
	
	// Run the setup function
	setupFunc()
	
	// Create router and generate OpenAPI
	router := NewRouter(http.NewServeMux())
	config := OpenAPIConfig{
		Info: OpenAPIInfo{
			Title:   "Test API",
			Version: "1.0.0",
		},
		OutputPath: outputPath,
	}
	
	err := router.GenerateOpenAPI(config)
	if err != nil {
		return nil, "", err
	}
	
	// Read and parse the OpenAPI spec
	data, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, "", err
	}
	
	var spec openapi3.T
	err = json.Unmarshal(data, &spec)
	return &spec, string(data), err
}

// Test path conversion from rip format to OpenAPI format
func TestConvertPathToOpenAPI(t *testing.T) {
	testCases := []struct {
		name     string
		ripPath  string
		expected string
	}{
		{
			name:     "Simple path",
			ripPath:  "/users",
			expected: "/users",
		},
		{
			name:     "Path with trailing slash",
			ripPath:  "/users/",
			expected: "/users/",
		},
		{
			name:     "Path with parameter",
			ripPath:  "/users/:id",
			expected: "/users/{id}",
		},
		{
			name:     "Path with multiple parameters",
			ripPath:  "/orgs/:orgID/users/:userID",
			expected: "/orgs/{orgID}/users/{userID}",
		},
		{
			name:     "Path with parameter and subpath",
			ripPath:  "/users/:id/profile",
			expected: "/users/{id}/profile",
		},
		{
			name:     "Root path",
			ripPath:  "/",
			expected: "/",
		},
		{
			name:     "Root with parameter",
			ripPath:  "/:id",
			expected: "/{id}",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := convertPathToOpenAPI(tc.ripPath)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Test schema generation for different entity types
func TestEntityToSchemaRef(t *testing.T) {
	testCases := []struct {
		name          string
		entityType    reflect.Type
		expectedProps []string
		expectedTypes map[string]string
	}{
		{
			name:          "TestEntity",
			entityType:    reflect.TypeOf(TestEntity{}),
			expectedProps: []string{"id", "name", "count", "is_active", "created_at", "tags"},
			expectedTypes: map[string]string{
				"id":         "string",
				"count":      "integer",
				"is_active":  "boolean",
				"tags":       "array",
			},
		},
		{
			name:          "Address",
			entityType:    reflect.TypeOf(Address{}),
			expectedProps: []string{"street", "city", "zip_code"},
			expectedTypes: map[string]string{
				"street":  "string",
				"city":    "string",
				"zip_code": "string",
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schemaRef, err := EntityToSchemaRef(tc.entityType)
			require.NoError(t, err)
			require.NotNil(t, schemaRef.Value)
			
			// Verify properties
			for _, prop := range tc.expectedProps {
				assert.Contains(t, schemaRef.Value.Properties, prop)
			}
			
			// Verify property types
			for prop, expectedType := range tc.expectedTypes {
				propSchema := schemaRef.Value.Properties[prop]
				require.NotNil(t, propSchema.Value)
				assert.Contains(t, propSchema.Value.Type.Slice(), expectedType)
			}
		})
	}
}

// Test schema generation for complex entities with nested structures
func TestComplexEntitySchema(t *testing.T) {
	schemaRef, err := EntityToSchemaRef(reflect.TypeOf(ComplexTestEntity{}))
	require.NoError(t, err)
	require.NotNil(t, schemaRef.Value)
	
	// Basic entity properties
	assert.Contains(t, schemaRef.Value.Properties, "id")
	assert.Contains(t, schemaRef.Value.Properties, "name")
	
	// Verify array of objects (addresses)
	require.Contains(t, schemaRef.Value.Properties, "addresses")
	addrProp := schemaRef.Value.Properties["addresses"]
	assert.NotNil(t, addrProp.Value)
	assert.Equal(t, "array", addrProp.Value.Type.Slice()[0])
	
	// Verify nested object (main_address)
	require.Contains(t, schemaRef.Value.Properties, "main_address")
	mainAddrProp := schemaRef.Value.Properties["main_address"]
	assert.NotNil(t, mainAddrProp.Value)
}

// Test automatic route registration via HandleEntities
func TestHandleEntitiesRegistration(t *testing.T) {
	// Clear any existing routes
	ClearRegisteredRoutes()
	
	// Use HandleEntities to register a route
	provider := &MockEntityProvider{}
	_, _ = HandleEntities[TestEntity](
		"/auto-register",
		provider,
		nil,
	)
	
	// Verify the route was registered in the global registry
	require.Contains(t, globalRouteRegistry, "/auto-register")
	metadata := globalRouteRegistry["/auto-register"]
	assert.Equal(t, reflect.TypeOf(TestEntity{}), metadata.EntityType)
	
	// Generate OpenAPI to verify it works
	spec, _, err := generateAndParseOpenAPI(t, func() {
		// Route already registered by HandleEntities
	})
	require.NoError(t, err)
	
	// Verify the schema was generated
	require.NotNil(t, spec.Components.Schemas["TestEntity"])
	
	// Verify paths were created
	collectionPath := spec.Paths.Find("/auto-register/")
	require.NotNil(t, collectionPath)
	
	itemPath := spec.Paths.Find("/auto-register/{id}")
	require.NotNil(t, itemPath)
}

// Test RegisterEntityRoute for manual route registration
func TestRegisterEntityRoute(t *testing.T) {
	// Clear any existing routes
	ClearRegisteredRoutes()
	
	// Register a test entity route
	testSetup := func() {
		RegisterEntityRoute("/manual-register", reflect.TypeOf(TestEntity{}), []string{"create", "get", "update", "delete", "list"})
	}
	
	// Generate OpenAPI spec
	spec, _, err := generateAndParseOpenAPI(t, testSetup)
	require.NoError(t, err)
	
	// Verify the schema was generated
	require.NotNil(t, spec.Components.Schemas["TestEntity"])
	
	// Verify paths were created
	collectionPath := spec.Paths.Find("/manual-register/")
	require.NotNil(t, collectionPath)
	
	itemPath := spec.Paths.Find("/manual-register/{id}")
	require.NotNil(t, itemPath)
}

// Test error handling and edge cases
func TestOpenAPIGenerationErrors(t *testing.T) {
	testCases := []struct {
		name          string
		setupFunc     func()
		expectError   bool
		errorContains string
	}{
		{
			name: "Empty route registry",
			setupFunc: func() {
				ClearRegisteredRoutes()
			},
			expectError: false,
		},
		{
			name: "Invalid output path",
			setupFunc: func() {
				ClearRegisteredRoutes()
				RegisterEntityRoute("/invalid-path", reflect.TypeOf(TestEntity{}), []string{"get"})
			},
			expectError:   true,
			errorContains: "failed to write OpenAPI spec",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up the test
			tc.setupFunc()
			
			// Create router and generate OpenAPI
			router := NewRouter(http.NewServeMux())
			config := OpenAPIConfig{
				Info: OpenAPIInfo{
					Title:   "Error Test",
					Version: "1.0.0",
				},
			}
			
			if tc.name == "Invalid output path" {
				// Use an invalid path for this test
				config.OutputPath = "/nonexistent/directory/openapi.json"
			}
			
			err := router.GenerateOpenAPI(config)
			if tc.expectError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test the generation of different operations (GET, POST, PUT, DELETE)
func TestOperationGeneration(t *testing.T) {
	// Clear any existing routes
	ClearRegisteredRoutes()
	
	// Register a test entity route
	RegisterEntityRoute("/operations", reflect.TypeOf(TestEntity{}), []string{"create", "get", "update", "delete", "list"})
	
	// Generate OpenAPI spec
	spec, specJSON, err := generateAndParseOpenAPI(t, func() {
		// Route already registered above
	})
	require.NoError(t, err)
	
	// Use subtests for different parts of the specification
	t.Run("SchemaComponents", func(t *testing.T) {
		require.NotNil(t, spec.Components, "Components should not be nil")
		require.NotNil(t, spec.Components.Schemas, "Schemas should not be nil")
		
		// Check if TestEntity schema exists
		testEntitySchema, exists := spec.Components.Schemas["TestEntity"]
		require.True(t, exists, "TestEntity schema should exist")
		require.NotNil(t, testEntitySchema.Value, "TestEntity schema value should not be nil")
		
		// Check schema properties
		props := testEntitySchema.Value.Properties
		require.NotNil(t, props, "Properties should not be nil")
		
		// Check specific property types
		propTests := map[string]string{
			"id":         "string",
			"name":       "string",
			"count":      "integer",
			"is_active":  "boolean",
			"tags":       "array",
		}
		
		for prop, expectedType := range propTests {
			require.Contains(t, props, prop, "Property %s should exist", prop)
			propSchema := props[prop]
			require.NotNil(t, propSchema.Value, "Property %s schema should not be nil", prop)
			require.NotNil(t, propSchema.Value.Type, "Property %s type should not be nil", prop)
			assert.Contains(t, propSchema.Value.Type.Slice(), expectedType, "Property %s should be of type %s", prop, expectedType)
		}
	})
	
	t.Run("CollectionOperations", func(t *testing.T) {
		collectionPath := spec.Paths.Find("/operations/")
		require.NotNil(t, collectionPath, "Collection path should exist")
		
		// Test POST (Create)
		t.Run("Create", func(t *testing.T) {
			require.NotNil(t, collectionPath.Post, "POST operation should exist")
			assert.Contains(t, collectionPath.Post.Tags, "testentity")
			
			// Check request body
			require.NotNil(t, collectionPath.Post.RequestBody, "Request body should exist")
			require.NotNil(t, collectionPath.Post.RequestBody.Value, "Request body value should exist")
			require.Contains(t, collectionPath.Post.RequestBody.Value.Content, "application/json")
			
			// Check responses
			require.NotNil(t, collectionPath.Post.Responses.Value("201"))
			require.NotNil(t, collectionPath.Post.Responses.Value("400"))
			require.NotNil(t, collectionPath.Post.Responses.Value("500"))
		})
		
		// Test GET (List)
		t.Run("List", func(t *testing.T) {
			require.NotNil(t, collectionPath.Get, "GET operation should exist")
			assert.Contains(t, collectionPath.Get.Tags, "testentity")
			
			// Check pagination parameters
			foundOffset := false
			foundLimit := false
			
			for _, param := range collectionPath.Get.Parameters {
				if param.Value.Name == "offset" {
					foundOffset = true
				} else if param.Value.Name == "limit" {
					foundLimit = true
				}
			}
			
			assert.True(t, foundOffset, "Offset parameter should exist")
			assert.True(t, foundLimit, "Limit parameter should exist")
			
			// Check responses
			require.NotNil(t, collectionPath.Get.Responses.Value("200"))
			require.NotNil(t, collectionPath.Get.Responses.Value("500"))
		})
	})
	
	t.Run("ItemOperations", func(t *testing.T) {
		itemPath := spec.Paths.Find("/operations/{id}")
		require.NotNil(t, itemPath, "Item path should exist")
		
		// Test GET (Get)
		t.Run("Get", func(t *testing.T) {
			require.NotNil(t, itemPath.Get, "GET operation should exist")
			assert.Contains(t, itemPath.Get.Tags, "testentity")
			
			// Check ID parameter
			foundIDParam := false
			for _, param := range itemPath.Get.Parameters {
				if param.Value.Name == "id" {
					foundIDParam = true
					break
				}
			}
			assert.True(t, foundIDParam, "ID parameter should exist")
			
			// Check responses
			require.NotNil(t, itemPath.Get.Responses.Value("200"))
			require.NotNil(t, itemPath.Get.Responses.Value("404"))
			require.NotNil(t, itemPath.Get.Responses.Value("500"))
		})
		
		// Test PUT (Update)
		t.Run("Update", func(t *testing.T) {
			require.NotNil(t, itemPath.Put, "PUT operation should exist")
			assert.Contains(t, itemPath.Put.Tags, "testentity")
			
			// Check request body
			require.NotNil(t, itemPath.Put.RequestBody, "Request body should exist")
			require.NotNil(t, itemPath.Put.RequestBody.Value, "Request body value should exist")
			require.Contains(t, itemPath.Put.RequestBody.Value.Content, "application/json")
			
			// Check responses
			require.NotNil(t, itemPath.Put.Responses.Value("200"))
			require.NotNil(t, itemPath.Put.Responses.Value("400"))
			require.NotNil(t, itemPath.Put.Responses.Value("404"))
			require.NotNil(t, itemPath.Put.Responses.Value("500"))
		})
		
		// Test DELETE (Delete)
		t.Run("Delete", func(t *testing.T) {
			require.NotNil(t, itemPath.Delete, "DELETE operation should exist")
			assert.Contains(t, itemPath.Delete.Tags, "testentity")
			
			// Check responses
			require.NotNil(t, itemPath.Delete.Responses.Value("204"))
			require.NotNil(t, itemPath.Delete.Responses.Value("404"))
			require.NotNil(t, itemPath.Delete.Responses.Value("500"))
		})
	})
	
	// For debugging - Log the full JSON spec
	t.Logf("OpenAPI spec: %s", specJSON)
}
