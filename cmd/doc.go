// Package main is the entry point for the kith-pms binary.
//
// @title           kith API
// @version         1.0
// @description     Self-hosted personal relationship manager API.
//
// @contact.name    kith
// @contact.url     https://github.com/nhymxu/kith-pms
//
// @license.name    MIT
//
// @host            localhost:8000
// @BasePath        /v1
//
// @securityDefinitions.apikey  CookieAuth
// @in                          cookie
// @name                        kith_session
// @description                 HttpOnly session cookie set by POST /v1/auth/login
//
// @securityDefinitions.apikey  CSRFHeader
// @in                          header
// @name                        X-Requested-With
// @description                 Must be set to "kith-spa" on all mutating requests (POST/PUT/PATCH/DELETE)
package main
