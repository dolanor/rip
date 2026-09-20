package rip

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"

	ripjson "github.com/dolanor/rip/encoding/json"
	"github.com/dolanor/rip/internal/ripreflect"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
)

type Route interface {
	Path() string
	Handler() http.HandlerFunc
	OpenAPISchema() *openapi3.T
}

type EntityRoute[
	Ent any,
	EP EntityProvider[Ent],
] struct {
	path        string
	handlerFunc http.HandlerFunc

	provider EP

	openAPISchema *openapi3.T
	generator     *openapi3gen.Generator
}

func (er *EntityRoute[Ent, EP]) Path() string {
	return er.path
}

func (er *EntityRoute[Ent, EP]) Handler() http.HandlerFunc {
	return er.handlerFunc
}

func (er *EntityRoute[Ent, EP]) OpenAPISchema() *openapi3.T {
	return er.openAPISchema
}

// NewEntityRoute generates an endpoint that will handle all the methods for the entityProvider at the
// given path and with specific options.
// It returns the path, the endpoint handler and the schema of this route.
func NewEntityRoute[
	Ent any,
	EP EntityProvider[Ent],
](path string, entityProvider EP, options ...EntityRouteOption) *EntityRoute[Ent, EP] {

	var cfg entityRouteConfig
	for _, o := range options {
		o(&cfg)
	}

	generator := openapi3gen.NewGenerator(
		openapi3gen.UseAllExportedFields(),
	)

	// just a base that we can merge with other entity routes on the router
	oaSpec := newOpenApiSpec("", "should not be use as is", "")
	rt := EntityRoute[Ent, EP]{
		path:          path,
		provider:      entityProvider,
		openAPISchema: &oaSpec,
		generator:     generator,
	}

	rt.generateOperation()

	_, handler := rt.createEntityHandler(path, entityProvider, cfg)
	rt.handlerFunc = handler
	return &rt
}

func (rt *EntityRoute[Ent, EP]) createEntityHandler(
	urlPath string,
	ep EP,
	cfg entityRouteConfig,
) (path string, handler http.HandlerFunc) {
	if len(cfg.codecs.Codecs) == 0 {
		err := fmt.Sprintf("no codecs defined on route: %s", urlPath)
		panic(err)
	}

	cfg = setEntityRouteConfigDefaults(cfg)

	return handleEntityWithPath(urlPath, ep.Create, ep.Get, ep.Update, ep.Delete, ep.List, cfg)
}

func (rt *EntityRoute[Ent, EP]) generateOperation() {
	var ent Ent
	tag, ok := ripreflect.TagFromType(ent)
	if !ok {
		panic("generate OpenAPI operation: can not get type tag")
	}

	// we register the /{entity}/{id} path ID parameter once
	// and save it as an OpenAPI  path parameter (so we don't have to duplicate it on
	// every OpenAPI operation.
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
				// TODO: add encoding types from registered codecs
				content := openapi3.NewContentWithSchema(bodySchema.Value, []string{"application/json", "text/yaml"})

				content["application/json"].Schema.Ref = "#/components/schemas/" + tag
				content["text/yaml"].Schema.Ref = "#/components/schemas/" + tag
				requestBody.WithContent(content)
			}

			// add request body to operation
			op.RequestBody = &openapi3.RequestBodyRef{
				Value: requestBody,
			}
		}

		// TODO: handle more API responses.
		response := openapi3.NewResponse().WithDescription("OK")

		if method != http.MethodDelete {
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
			if responseSchema == nil {
				panic("could not find response schema: " + tag)
			}

			// TODO: add encoding types from registered codecs
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

	}

	rt.generateList()
}

func (rt *EntityRoute[Ent, EP]) generateList() {
	method := http.MethodGet
	var ent Ent
	tag, ok := ripreflect.TagFromType(ent)
	if !ok {
		panic("generate OpenAPI operation: can not get type tag")
	}

	op := openapi3.NewOperation()
	op.Tags = append(op.Tags, tag)

	// Response body
	itemResponseSchema, ok := rt.openAPISchema.Components.Schemas[tag]
	if !ok {
		var err error
		// TODO test with the entity ent or a pointer to it
		itemResponseSchema, err = rt.generator.NewSchemaRefForValue(ent, rt.openAPISchema.Components.Schemas)
		if err != nil {
			// there is no point of going further, and silently failing would be bad.
			panic("generate OpenAPI operation: can not generate schema ref for response value: " + fmt.Sprintf("%+v: %v", ent, err))
		}
		rt.openAPISchema.Components.Schemas[tag] = itemResponseSchema
	}

	// TODO: handle more API responses.
	if itemResponseSchema == nil {
		panic("could not find response schema: " + tag)
	}

	itemsResponseSchema := openapi3.NewArraySchema().WithItems(itemResponseSchema.Value)
	content := openapi3.NewContentWithSchema(itemsResponseSchema, []string{"application/json"})

	response := openapi3.NewResponse().WithDescription("OK").WithContent(content)

	op.AddResponse(200, response)

	entityPath := rt.path
	rt.openAPISchema.AddOperation(entityPath, method, op)
}

func dumpSchema(title string, schema any) {
	b, _ := json.Marshal(schema)
	fmt.Print(string(b))
	fmt.Fprintln(os.Stderr)
}

func setEntityRouteConfigDefaults(cfg entityRouteConfig) entityRouteConfig {
	if len(cfg.codecs.Codecs) == 0 {
		cfg.codecs.Register(ripjson.Codec)
	}

	if cfg.listPageSize == 0 {
		cfg.listPageSize = 20
	}

	if cfg.listPageSizeMax == 0 {
		cfg.listPageSizeMax = 100
	}

	return cfg
}
