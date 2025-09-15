package lightmux

import (
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

func isValidMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete,
		http.MethodPatch, http.MethodHead, http.MethodOptions, http.MethodConnect, http.MethodTrace:
		return true
	default:
		return false
	}
}

// getFuncName returns the name of the function for the given handler or middleware.
func getFuncName(h any) string {
	return runtime.FuncForPC(
		reflect.ValueOf(h).Pointer(),
	).Name()
}

func allowedMethodsJoin(mp map[string]http.Handler) string {
	var methods []string
	for method := range mp {
		methods = append(methods, method)
	}
	
	return strings.Join(methods, ", ")
}