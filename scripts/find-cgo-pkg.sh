#!/bin/bash

go list -f "{{if .CgoFiles}}{{.ImportPath}}{{end}}" $(go list -f "{{.ImportPath}}{{range .Deps}} {{.}}{{end}}")
