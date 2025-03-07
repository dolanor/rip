package rip

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// OpenAPIInfo contains metadata for the API documentation
type OpenAPIInfo struct {
	Title          string
	Description    string
	Version        string
	TermsOfService string
	ContactName    string
	ContactEmail   string
	ContactURL     string
	LicenseName    string
	LicenseURL     string
}

// OpenAPIConfig contains configuration for OpenAPI generation
type OpenAPIConfig struct {
	Info       OpenAPIInfo
	OutputPath string
	ServerURLs []string
}

// RouteMetadata stores information about registered routes for OpenAPI generation
type RouteMetadata struct {
	Method      string
	Pattern     string
	EntityType  reflect.Type
	Operations  []string // "create", "get", "update", "delete", "list"
	PathParams  map[string]reflect.Type
	QueryParams map[string]reflect.Type
}

var globalRouteRegistry = make(map[string]RouteMetadata)

// RegisterRouteMetadata registers metadata for a route pattern
func RegisterRouteMetadata(pattern string, metadata RouteMetadata) {
	globalRouteRegistry[pattern] = metadata
}

// ClearRegisteredRoutes clears all registered routes - primarily used for testing
func ClearRegisteredRoutes() {
	globalRouteRegistry = make(map[string]RouteMetadata)
}

// TypeOf returns the reflect.Type of a generic type T
// This is a helper function for use with RegisterEntityRoute
func TypeOf[T any]() reflect.Type {
	var t T
	return reflect.TypeOf(t)
}

// RegisterEntityRoute registers an entity type for a route pattern
func RegisterEntityRoute(pattern string, entityType reflect.Type, operations []string) {
	metadata, exists := globalRouteRegistry[pattern]
	if !exists {
		metadata = RouteMetadata{
			Pattern:     pattern,
			Operations:  operations,
			PathParams:  make(map[string]reflect.Type),
			QueryParams: make(map[string]reflect.Type),
		}
	}
	metadata.EntityType = entityType
	globalRouteRegistry[pattern] = metadata
}

// GenerateOpenAPI generates OpenAPI documentation based on registered routes
func (rt *Router) GenerateOpenAPI(config OpenAPIConfig) error {
	// Initialize components
	components := &openapi3.Components{
		Schemas: make(openapi3.Schemas),
	}

	// Initialize the OpenAPI specification
	spec := openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:          config.Info.Title,
			Description:    config.Info.Description,
			Version:        config.Info.Version,
			TermsOfService: config.Info.TermsOfService,
			Contact: &openapi3.Contact{
				Name:  config.Info.ContactName,
				Email: config.Info.ContactEmail,
				URL:   config.Info.ContactURL,
			},
			License: &openapi3.License{
				Name: config.Info.LicenseName,
				URL:  config.Info.LicenseURL,
			},
		},
		Paths:      openapi3.NewPaths(),
		Components: components,
	}

	// Add servers if provided
	for _, url := range config.ServerURLs {
		spec.Servers = append(spec.Servers, &openapi3.Server{URL: url})
	}

	// Track entities to avoid duplicate schema generation
	schemaRefs := make(map[reflect.Type]*openapi3.SchemaRef)

	// Process each registered route
	for pattern, metadata := range globalRouteRegistry {
		entityType := metadata.EntityType
		if _, exists := schemaRefs[entityType]; !exists {
			typeName := entityType.Name()
			schemaRef, err := EntityToSchemaRef(entityType)
			if err != nil {
				return fmt.Errorf("failed to generate schema for %s: %w", typeName, err)
			}
			schemaRefs[entityType] = schemaRef

			// Add schema to components
			spec.Components.Schemas[typeName] = schemaRef
		}

		// Add paths for this route
		err := addPathsFromMetadata(&spec, metadata, schemaRefs[entityType])
		if err != nil {
			return fmt.Errorf("failed to add paths for %s: %w", pattern, err)
		}
	}

	// Write to file if output path is provided
	if config.OutputPath != "" {
		err := writeOpenAPISpec(spec, config.OutputPath)
		if err != nil {
			return fmt.Errorf("failed to write OpenAPI spec: %w", err)
		}
	}

	return nil
}

// EntityToSchemaRef converts an entity type to an OpenAPI schema reference
func EntityToSchemaRef(t reflect.Type) (*openapi3.SchemaRef, error) {
	typeName := t.Name()
	schema := openapi3.NewObjectSchema()
	schema.Title = typeName
	schema.Description = "Schema for " + typeName

	// Iterate through fields to build properties
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Extract the field name from json tag
		fieldName := strings.Split(jsonTag, ",")[0]

		// Create a schema for the field based on its type
		var propSchema *openapi3.Schema
		switch field.Type.Kind() {
		case reflect.String:
			propSchema = openapi3.NewStringSchema()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			propSchema = openapi3.NewIntegerSchema()
		case reflect.Float32, reflect.Float64:
			propSchema = openapi3.NewFloat64Schema()
		case reflect.Bool:
			propSchema = openapi3.NewBoolSchema()
		case reflect.Slice, reflect.Array:
			arraySchema := openapi3.NewArraySchema()
			arraySchema.Items = openapi3.NewSchemaRef("", openapi3.NewStringSchema())
			propSchema = arraySchema
		default:
			propSchema = openapi3.NewObjectSchema()
		}

		// Add property to schema
		if schema.Properties == nil {
			schema.Properties = make(openapi3.Schemas)
		}
		schema.Properties[fieldName] = openapi3.NewSchemaRef("", propSchema)
	}

	return openapi3.NewSchemaRef("", schema), nil
}

// addPathsFromMetadata adds OpenAPI paths based on route metadata
func addPathsFromMetadata(spec *openapi3.T, metadata RouteMetadata, schemaRef *openapi3.SchemaRef) error {
	entityName := metadata.EntityType.Name()
	pattern := metadata.Pattern

	// Check if the pattern ends with a trailing slash
	collectionPattern := pattern
	if !strings.HasSuffix(pattern, "/") {
		collectionPattern = pattern + "/"
	}

	// Convert to OpenAPI path format
	openAPIPath := convertPathToOpenAPI(collectionPattern)

	// Create collection path item if it doesn't exist
	collectionPathItem := spec.Paths.Find(openAPIPath)
	if collectionPathItem == nil {
		collectionPathItem = &openapi3.PathItem{}
		spec.Paths.Set(openAPIPath, collectionPathItem)
	}

	// Item path (with ID parameter)
	itemPath := convertPathToOpenAPI(strings.TrimSuffix(pattern, "/") + "/{id}")
	itemPathItem := spec.Paths.Find(itemPath)
	if itemPathItem == nil {
		itemPathItem = &openapi3.PathItem{}
		spec.Paths.Set(itemPath, itemPathItem)
	}

	// Add operations based on metadata
	for _, op := range metadata.Operations {
		switch op {
		case "create":
			collectionPathItem.Post = createOperation(op, entityName, schemaRef, "201")
		case "list":
			collectionPathItem.Get = createOperation(op, entityName, schemaRef, "200")
			// Add pagination parameters
			addPaginationParams(collectionPathItem.Get)
		case "get":
			itemPathItem.Get = createOperation(op, entityName, schemaRef, "200")
			// Add ID parameter
			addIDParam(itemPathItem.Get)
		case "update":
			itemPathItem.Put = createOperation(op, entityName, schemaRef, "200")
			// Add ID parameter
			addIDParam(itemPathItem.Put)
		case "delete":
			itemPathItem.Delete = createOperation(op, entityName, schemaRef, "204")
			// Add ID parameter
			addIDParam(itemPathItem.Delete)
		}
	}

	return nil
}

// createOperation creates an OpenAPI operation
func createOperation(opType, entityName string, schemaRef *openapi3.SchemaRef, successCode string) *openapi3.Operation {
	operation := openapi3.NewOperation()

	// Set operation details
	operation.Tags = []string{getTagFromEntityName(entityName)}
	operation.Summary = getSummaryForOperation(opType, entityName)
	operation.Description = getDescriptionForOperation(opType, entityName)
	operation.OperationID = opType + entityName

	// Setup responses
	responses := openapi3.NewResponses()

	// Success response
	successResp := openapi3.NewResponse()
	successResp.Description = stringPtr(getResponseDescription(opType, entityName))

	// Add response schema if appropriate for this operation
	respSchema := getResponseSchema(opType, schemaRef)
	if respSchema != nil {
		successResp.Content = openapi3.NewContentWithJSONSchemaRef(respSchema)
	}

	responses.Set(successCode, &openapi3.ResponseRef{Value: successResp})

	// Error responses
	responses.Set("400", &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: stringPtr("Bad request"),
		},
	})
	responses.Set("404", &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: stringPtr("Not found"),
		},
	})
	responses.Set("500", &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: stringPtr("Internal server error"),
		},
	})
	responses.Set("default", &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: stringPtr(""),
		},
	})

	operation.Responses = responses

	// Set request body for operations that need it
	if opType == "create" || opType == "update" {
		reqBody := openapi3.NewRequestBody()
		reqBody.Description = entityName + " object that needs to be " + getPastTenseForOperation(opType)
		reqBody.Required = true
		reqBody.Content = openapi3.NewContentWithJSONSchemaRef(schemaRef)
		operation.RequestBody = &openapi3.RequestBodyRef{Value: reqBody}
	}

	// Add parameters based on operation type
	if opType == "list" {
		addPaginationParams(operation)
	}

	if opType == "get" || opType == "update" || opType == "delete" {
		addIDParam(operation)
	}

	return operation
}

// Helper function to get response schema based on operation type
func getResponseSchema(opType string, entitySchemaRef *openapi3.SchemaRef) *openapi3.SchemaRef {
	switch opType {
	case "get", "create", "update":
		return entitySchemaRef
	case "list":
		arraySchema := openapi3.NewArraySchema()
		arraySchema.Items = entitySchemaRef
		return openapi3.NewSchemaRef("", arraySchema)
	default:
		return nil
	}
}

// Helper function to add pagination parameters to an operation
func addPaginationParams(operation *openapi3.Operation) {
	// Add offset parameter
	offsetParam := &openapi3.Parameter{
		Name:        "offset",
		In:          "query",
		Description: "Starting position for fetching results",
		Schema:      openapi3.NewSchemaRef("", openapi3.NewIntegerSchema().WithDefault(0)),
	}
	operation.Parameters = append(operation.Parameters, &openapi3.ParameterRef{Value: offsetParam})

	// Add limit parameter
	limitParam := &openapi3.Parameter{
		Name:        "limit",
		In:          "query",
		Description: "Maximum number of items to return",
		Schema:      openapi3.NewSchemaRef("", openapi3.NewIntegerSchema().WithDefault(100)),
	}
	operation.Parameters = append(operation.Parameters, &openapi3.ParameterRef{Value: limitParam})
}

// Helper function to add ID parameter to an operation
func addIDParam(operation *openapi3.Operation) {
	idParam := &openapi3.Parameter{
		Name:        "id",
		In:          "path",
		Required:    true,
		Description: "ID of the resource",
		Schema:      openapi3.NewSchemaRef("", openapi3.NewStringSchema()),
	}
	operation.Parameters = append(operation.Parameters, &openapi3.ParameterRef{Value: idParam})
}

// Helper functions to create pointers for primitive types
func stringPtr(s string) *string {
	return &s
}

// Helper functions for operation metadata

func getTagFromEntityName(entityName string) string {
	return strings.ToLower(entityName)
}

func getSummaryForOperation(opType string, entityName string) string {
	switch opType {
	case "create":
		return fmt.Sprintf("Create a new %s", entityName)
	case "get":
		return fmt.Sprintf("Get a %s by ID", entityName)
	case "update":
		return fmt.Sprintf("Update an existing %s", entityName)
	case "delete":
		return fmt.Sprintf("Delete a %s", entityName)
	case "list":
		return fmt.Sprintf("List %ss", entityName)
	default:
		return ""
	}
}

func getDescriptionForOperation(opType string, entityName string) string {
	switch opType {
	case "create":
		return fmt.Sprintf("Creates a new %s in the system", entityName)
	case "get":
		return fmt.Sprintf("Retrieves a %s by its unique identifier", entityName)
	case "update":
		return fmt.Sprintf("Updates an existing %s with new information", entityName)
	case "delete":
		return fmt.Sprintf("Removes a %s from the system", entityName)
	case "list":
		return fmt.Sprintf("Returns a paginated list of %ss", entityName)
	default:
		return ""
	}
}

func getResponseDescription(opType string, entityName string) string {
	switch opType {
	case "create":
		return fmt.Sprintf("The created %s", entityName)
	case "get":
		return fmt.Sprintf("The requested %s", entityName)
	case "update":
		return fmt.Sprintf("The updated %s", entityName)
	case "delete":
		return "No content"
	case "list":
		return fmt.Sprintf("A list of %ss", entityName)
	default:
		return "Successful operation"
	}
}

func getPastTenseForOperation(opType string) string {
	switch opType {
	case "create":
		return "created"
	case "update":
		return "updated"
	default:
		return opType + "d"
	}
}

// convertPathToOpenAPI converts a rip URL path to OpenAPI path format
// e.g., "/albums/:id" -> "/albums/{id}"
func convertPathToOpenAPI(ripPath string) string {
	parts := strings.Split(ripPath, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + strings.TrimPrefix(part, ":") + "}"
		}
	}
	return strings.Join(parts, "/")
}

// Helper to write OpenAPI spec to file
func writeOpenAPISpec(spec openapi3.T, outputPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(outputPath, data, 0644)
}
