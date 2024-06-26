// ECode generated by swaggo/swag. DO NOT EDIT.

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
        "/api/v1/shortLink": {
            "post": {
                "description": "submit information to create shortLink",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "shortLink"
                ],
                "summary": "create shortLink",
                "parameters": [
                    {
                        "description": "shortLink information",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/types.CreateShortLinkRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/types.CreateShortLinkRespond"
                        }
                    }
                }
            }
        },
        "/api/v1/shortLink/condition": {
            "post": {
                "description": "get shortLink by condition",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "shortLink"
                ],
                "summary": "get shortLink by condition",
                "parameters": [
                    {
                        "description": "query condition",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/github_com_zhufuyi_sponge_internal_types.Conditions"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/types.GetShortLinkByConditionRespond"
                        }
                    }
                }
            }
        },
        "/api/v1/shortLink/delete/ids": {
            "post": {
                "description": "delete shortLinks by batch id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "shortLink"
                ],
                "summary": "delete shortLinks",
                "parameters": [
                    {
                        "description": "id array",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/types.DeleteShortLinksByIDsRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/types.DeleteShortLinksByIDsRespond"
                        }
                    }
                }
            }
        },
        "/api/v1/shortLink/list": {
            "get": {
                "description": "list of shortLinks by last id and limit",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "shortLink"
                ],
                "summary": "list of shortLinks by last id and limit",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "last id, default is MaxInt64",
                        "name": "lastID",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "default": 10,
                        "description": "size in each page",
                        "name": "limit",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "default": "-id",
                        "description": "sort by column name of table, and the ",
                        "name": "sort",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/types.ListShortLinksRespond"
                        }
                    }
                }
            },
            "post": {
                "description": "list of shortLinks by paging and conditions",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "shortLink"
                ],
                "summary": "list of shortLinks by query parameters",
                "parameters": [
                    {
                        "description": "query parameters",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/github_com_zhufuyi_sponge_internal_types.Params"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/types.ListShortLinksRespond"
                        }
                    }
                }
            }
        },
        "/api/v1/shortLink/list/ids": {
            "post": {
                "description": "list of shortLinks by batch id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "shortLink"
                ],
                "summary": "list of shortLinks by batch id",
                "parameters": [
                    {
                        "description": "id array",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/types.ListShortLinksByIDsRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/types.ListShortLinksByIDsRespond"
                        }
                    }
                }
            }
        },
        "/api/v1/shortLink/{id}": {
            "get": {
                "description": "get shortLink detail by id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "shortLink"
                ],
                "summary": "get shortLink detail",
                "parameters": [
                    {
                        "type": "string",
                        "description": "id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/types.GetShortLinkByIDRespond"
                        }
                    }
                }
            },
            "put": {
                "description": "update shortLink information by id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "shortLink"
                ],
                "summary": "update shortLink",
                "parameters": [
                    {
                        "type": "string",
                        "description": "id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "shortLink information",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/types.UpdateShortLinkByIDRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/types.UpdateShortLinkByIDRespond"
                        }
                    }
                }
            },
            "delete": {
                "description": "delete shortLink by id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "shortLink"
                ],
                "summary": "delete shortLink",
                "parameters": [
                    {
                        "type": "string",
                        "description": "id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/types.DeleteShortLinkByIDRespond"
                        }
                    }
                }
            }
        },
        "/codes": {
            "get": {
                "description": "list error codes info",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "system"
                ],
                "summary": "list error codes info",
                "responses": {}
            }
        },
        "/config": {
            "get": {
                "description": "show config info",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "system"
                ],
                "summary": "show config info",
                "responses": {}
            }
        },
        "/health": {
            "get": {
                "description": "check health",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "system"
                ],
                "summary": "check health",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/handlerfunc.checkHealthResponse"
                        }
                    }
                }
            }
        },
        "/ping": {
            "get": {
                "description": "ping",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "system"
                ],
                "summary": "ping",
                "responses": {}
            }
        }
    },
    "definitions": {
        "github_com_zhufuyi_sponge_internal_types.Column": {
            "type": "object",
            "properties": {
                "exp": {
                    "description": "expressions, which default to = when the value is null, have =, !=, \u003e, \u003e=, \u003c, \u003c=, like",
                    "type": "string"
                },
                "logic": {
                    "description": "logical type, defaults to and when value is null, only \u0026(and), ||(or)",
                    "type": "string"
                },
                "name": {
                    "description": "column name",
                    "type": "string"
                },
                "value": {
                    "description": "column value"
                }
            }
        },
        "github_com_zhufuyi_sponge_internal_types.Conditions": {
            "type": "object",
            "properties": {
                "columns": {
                    "description": "columns info",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_zhufuyi_sponge_internal_types.Column"
                    }
                }
            }
        },
        "github_com_zhufuyi_sponge_internal_types.Params": {
            "type": "object",
            "properties": {
                "columns": {
                    "description": "query conditions",
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/github_com_zhufuyi_sponge_internal_types.Column"
                    }
                },
                "page": {
                    "description": "page number, starting from page 0",
                    "type": "integer"
                },
                "size": {
                    "description": "lines per page",
                    "type": "integer"
                },
                "sort": {
                    "description": "sorted fields, multi-column sorting separated by commas",
                    "type": "string"
                }
            }
        },
        "handlerfunc.checkHealthResponse": {
            "type": "object",
            "properties": {
                "hostname": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                }
            }
        },
        "types.CreateShortLinkRequest": {
            "type": "object",
            "properties": {
                "age": {
                    "description": "age",
                    "type": "integer"
                },
                "avatar": {
                    "description": "avatar",
                    "type": "string",
                    "minLength": 5
                },
                "email": {
                    "description": "email",
                    "type": "string"
                },
                "gender": {
                    "description": "gender, 1:Male, 2:Female, other values:unknown",
                    "type": "integer",
                    "maximum": 2,
                    "minimum": 0
                },
                "name": {
                    "description": "username",
                    "type": "string",
                    "minLength": 2
                },
                "password": {
                    "description": "password",
                    "type": "string"
                },
                "phone": {
                    "description": "phone number, e164 rules, e.g. +8612345678901",
                    "type": "string"
                }
            }
        },
        "types.CreateShortLinkRespond": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "return code",
                    "type": "integer"
                },
                "data": {
                    "description": "return data",
                    "type": "object",
                    "properties": {
                        "id": {
                            "description": "id",
                            "type": "integer"
                        }
                    }
                },
                "msg": {
                    "description": "return information description",
                    "type": "string"
                }
            }
        },
        "types.DeleteShortLinkByIDRespond": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "return code",
                    "type": "integer"
                },
                "data": {
                    "description": "return data"
                },
                "msg": {
                    "description": "return information description",
                    "type": "string"
                }
            }
        },
        "types.DeleteShortLinksByIDsRequest": {
            "type": "object",
            "properties": {
                "ids": {
                    "description": "id list",
                    "type": "array",
                    "minItems": 1,
                    "items": {
                        "type": "integer"
                    }
                }
            }
        },
        "types.DeleteShortLinksByIDsRespond": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "return code",
                    "type": "integer"
                },
                "data": {
                    "description": "return data"
                },
                "msg": {
                    "description": "return information description",
                    "type": "string"
                }
            }
        },
        "types.GetShortLinkByConditionRespond": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "return code",
                    "type": "integer"
                },
                "data": {
                    "description": "return data",
                    "type": "object",
                    "properties": {
                        "shortLink": {
                            "$ref": "#/definitions/types.ShortLinkObjDetail"
                        }
                    }
                },
                "msg": {
                    "description": "return information description",
                    "type": "string"
                }
            }
        },
        "types.GetShortLinkByIDRespond": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "return code",
                    "type": "integer"
                },
                "data": {
                    "description": "return data",
                    "type": "object",
                    "properties": {
                        "shortLink": {
                            "$ref": "#/definitions/types.ShortLinkObjDetail"
                        }
                    }
                },
                "msg": {
                    "description": "return information description",
                    "type": "string"
                }
            }
        },
        "types.ListShortLinksByIDsRequest": {
            "type": "object",
            "properties": {
                "ids": {
                    "description": "id list",
                    "type": "array",
                    "minItems": 1,
                    "items": {
                        "type": "integer"
                    }
                }
            }
        },
        "types.ListShortLinksByIDsRespond": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "return code",
                    "type": "integer"
                },
                "data": {
                    "description": "return data",
                    "type": "object",
                    "properties": {
                        "shortLinks": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/types.ShortLinkObjDetail"
                            }
                        }
                    }
                },
                "msg": {
                    "description": "return information description",
                    "type": "string"
                }
            }
        },
        "types.ListShortLinksRespond": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "return code",
                    "type": "integer"
                },
                "data": {
                    "description": "return data",
                    "type": "object",
                    "properties": {
                        "shortLinks": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/types.ShortLinkObjDetail"
                            }
                        }
                    }
                },
                "msg": {
                    "description": "return information description",
                    "type": "string"
                }
            }
        },
        "types.UpdateShortLinkByIDRequest": {
            "type": "object",
            "properties": {
                "age": {
                    "description": "age",
                    "type": "integer"
                },
                "avatar": {
                    "description": "avatar",
                    "type": "string"
                },
                "email": {
                    "description": "email",
                    "type": "string"
                },
                "gender": {
                    "description": "gender, 1:Male, 2:Female, other values:unknown",
                    "type": "integer"
                },
                "id": {
                    "description": "id",
                    "type": "integer"
                },
                "name": {
                    "description": "username",
                    "type": "string"
                },
                "password": {
                    "description": "password",
                    "type": "string"
                },
                "phone": {
                    "description": "phone number",
                    "type": "string"
                }
            }
        },
        "types.UpdateShortLinkByIDRespond": {
            "type": "object",
            "properties": {
                "code": {
                    "description": "return code",
                    "type": "integer"
                },
                "data": {
                    "description": "return data"
                },
                "msg": {
                    "description": "return information description",
                    "type": "string"
                }
            }
        },
        "types.ShortLinkObjDetail": {
            "type": "object",
            "properties": {
                "age": {
                    "description": "age",
                    "type": "integer"
                },
                "avatar": {
                    "description": "avatar",
                    "type": "string"
                },
                "createdAt": {
                    "description": "create time",
                    "type": "string"
                },
                "email": {
                    "description": "email",
                    "type": "string"
                },
                "gender": {
                    "description": "gender, 1:Male, 2:Female, other values:unknown",
                    "type": "integer"
                },
                "id": {
                    "description": "id",
                    "type": "string"
                },
                "loginAt": {
                    "description": "login timestamp",
                    "type": "integer"
                },
                "name": {
                    "description": "username",
                    "type": "string"
                },
                "phone": {
                    "description": "phone number",
                    "type": "string"
                },
                "status": {
                    "description": "account status, 1:inactive, 2:activated, 3:blocked",
                    "type": "integer"
                },
                "updatedAt": {
                    "description": "update time",
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "BearerAuth": {
            "description": "Type \"Bearer your-jwt-token\" to Value",
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "2.0",
	Host:             "localhost:8080",
	BasePath:         "",
	Schemes:          []string{"http", "https"},
	Title:            "SnapLink api docs",
	Description:      "http server api docs",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
