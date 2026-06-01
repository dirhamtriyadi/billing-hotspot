// Package docs holds the Swagger specification for the radius-api. This minimal
// placeholder lets the project compile; run `make swagger` (swag init) to
// regenerate it from the route annotations.
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
  "swagger": "2.0",
  "info": {
    "title": "Radius API",
    "description": "Management API over the FreeRADIUS SQL schema: profiles, users, sessions, NAS clients and CoA disconnect. Authenticated with the X-API-Key header.",
    "version": "1.0"
  },
  "basePath": "/api/v1",
  "paths": {}
}`

// SwaggerInfo registers the spec metadata with swaggo.
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	BasePath:         "/api/v1",
	Title:            "Radius API",
	Description:      "Management API over the FreeRADIUS SQL schema.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
