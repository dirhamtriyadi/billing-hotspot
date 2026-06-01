// Package docs holds the Swagger specification. This is a minimal placeholder
// so the project compiles out of the box; run `make swagger` (swag init) to
// regenerate it with the full set of annotated routes.
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
  "swagger": "2.0",
  "info": {
    "title": "Billing Hotspot API",
    "description": "Backend billing, package & voucher API. Provisions FreeRADIUS credentials via the radius-api and accepts payments via Xendit, Midtrans and Tripay.",
    "version": "1.0"
  },
  "basePath": "/api/v1",
  "paths": {}
}`

// SwaggerInfo registers the spec metadata with swaggo.
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "Billing Hotspot API",
	Description:      "Backend billing, package & voucher API for a FreeRADIUS/Mikrotik hotspot.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
