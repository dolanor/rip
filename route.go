package rip

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/dolanor/rip/internal/ripreflect"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
)

type EntityRoute[
	Ent any,
	EP EntityProvider[Ent],
] struct {
	path          string
	provider      EP
	openAPISchema *openapi3.T
	generator     *openapi3gen.Generator
}

func (er *EntityRoute[Ent, EP]) OpenAPISchema() *openapi3.T {
	return er.openAPISchema
}

func NewEntityRoute[
	Ent any,
	EP EntityProvider[Ent],
](path string, ep EP, options *RouteOptions) (string, http.HandlerFunc, *openapi3.T) {

	generator := openapi3gen.NewGenerator(
		openapi3gen.UseAllExportedFields(),
	)
	// just a base that we can merge with other entity routes on the router
	oaSpec := newOpenApiSpec("", "should not be use as is", "")
	rt := EntityRoute[Ent, EP]{
		path:          path,
		provider:      ep,
		openAPISchema: &oaSpec,
		generator:     generator,
	}

	rt.generateOperation()
	path, handler := rt.HandleEntities(path, ep, options)
	return path, handler, rt.openAPISchema
}

func (rt *EntityRoute[Ent, EP]) HandleEntities(
	urlPath string,
	ep EP,
	options *RouteOptions,
) (path string, handler http.HandlerFunc) {
	// end HandleEntities OMIT
	if options == nil {
		options = defaultOptions
	}

	if len(options.codecs.Codecs) == 0 {
		err := fmt.Sprintf("no codecs defined on route: %s", urlPath)
		panic(err)
	}

	return handleEntityWithPath(urlPath, ep.Create, ep.Get, ep.Update, ep.Delete, ep.List, options)
}

func (rt *EntityRoute[Ent, EP]) generateOperation() {

	var ent Ent
	tag, ok := ripreflect.TagFromType(ent)
	if !ok {
		panic("generate OpenAPI operation: can not get type tag")
	}

	// we register the /{entity}/{id} path ID parameter once
	{
		entityPath := path.Join(rt.path, "{id}")
		param := openapi3.NewPathParameter("id")
		param.Description = "id of the " + tag
		param.Schema = openapi3.NewStringSchema().NewRef()

		rt.openAPISchema.Paths.Set(entityPath, &openapi3.PathItem{
			Parameters: []*openapi3.ParameterRef{
				{
					Value: param,
				},
			},
		})
	}

	for _, method := range []string{
		http.MethodPost,
		http.MethodGet,
		http.MethodPut,
		http.MethodDelete,
	} {
		op := openapi3.NewOperation()
		op.Tags = append(op.Tags, tag)
		if (method == http.MethodPost || method == http.MethodPut) && tag != "string" {
			bodySchema, ok := rt.openAPISchema.Components.Schemas[tag]
			if !ok {
				var err error
				// TODO test with the entity or a pointer to it
				bodySchema, err = rt.generator.NewSchemaRefForValue(ent, rt.openAPISchema.Components.Schemas)
				if err != nil {
					// there is no point of going further, and silently failing would be bad.
					panic("generate OpenAPI operation: can not generate schema ref for request value: " + fmt.Sprintf("%+v: %v", ent, err))
				}
				rt.openAPISchema.Components.Schemas[tag] = bodySchema
			}

			requestBody := openapi3.NewRequestBody().
				WithRequired(true).
				WithDescription("Request body for " + tag)

			if bodySchema != nil {
				// TODO add route options multiple encoding
				content := openapi3.NewContentWithSchema(bodySchema.Value, []string{"application/json"})
				content["application/json"].Schema.Ref = "#/components/schemas/" + tag
				requestBody.WithContent(content)
			}

			// add request body to operation
			op.RequestBody = &openapi3.RequestBodyRef{
				Value: requestBody,
			}
		}

		// Response body
		responseSchema, ok := rt.openAPISchema.Components.Schemas[tag]
		if !ok {
			var err error
			// TODO test with the entity ent or a pointer to it
			responseSchema, err = rt.generator.NewSchemaRefForValue(ent, rt.openAPISchema.Components.Schemas)
			if err != nil {
				// there is no point of going further, and silently failing would be bad.
				panic("generate OpenAPI operation: can not generate schema ref for response value: " + fmt.Sprintf("%+v: %v", ent, err))
			}
			rt.openAPISchema.Components.Schemas[tag] = responseSchema
		}

		response := openapi3.NewResponse().WithDescription("OK")
		if responseSchema != nil {
			content := openapi3.NewContentWithSchema(responseSchema.Value, []string{"application/json"})
			content["application/json"].Schema.Ref = "#/components/schemas/" + tag
			response.WithContent(content)
		}
		op.AddResponse(200, response)

		entityPath := rt.path
		switch method {
		case http.MethodGet,
			http.MethodPut,
			http.MethodDelete:
			entityPath = path.Join(rt.path, "{id}")
		}
		rt.openAPISchema.AddOperation(entityPath, method, op)
		dumpSchema("full schema", rt.openAPISchema)
	}
}

func dumpSchema(title string, schema any) {
	b, _ := json.Marshal(schema)
	fmt.Print(string(b))
	fmt.Fprintln(os.Stderr)
}
