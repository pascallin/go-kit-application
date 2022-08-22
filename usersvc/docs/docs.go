// Package docs GENERATED BY SWAG; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/user/v1/register": {
            "post": {
                "security": [
                    {
                        "ServiceApiKey": []
                    }
                ],
                "description": "create subscription",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "user register",
                "parameters": [
                    {
                        "description": "data",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/endpoints.RegisterResponse"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/endpoints.RegisterRequest"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "endpoints.RegisterRequest": {
            "type": "object",
            "properties": {
                "username": {
                    "type": "string"
                }
            }
        },
        "endpoints.RegisterResponse": {
            "type": "object",
            "properties": {
                "err": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "ServiceApiKey": {
            "type": "apiKey",
            "name": "x-api-key",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "user service",
	Description:      "user service",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}