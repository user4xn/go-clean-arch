package genx

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Route struct {
	Method      string
	Path        string
	HandlerName string
	Tag         string
}

// GenerateSwagger scans all router.go and inject swagger annotations to handler.go
func GenerateSwagger() error {
	base := "internal/app"

	return filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, "router.go") {
			fmt.Println("Scanning:", path)
			routes := parseRouterFile(path)

			if len(routes) > 0 {
				handlerPath := strings.Replace(path, "router.go", "handler.go", 1)
				return injectSwagger(handlerPath, routes)
			}
		}

		return nil
	})
}

func parseRouterFile(path string) []Route {
	content, _ := ioutil.ReadFile(path)
	text := string(content)

	module := extractModule(path)

	// regex: g.POST("login", h.Login)
	re := regexp.MustCompile(`g\.(GET|POST|PUT|DELETE)\("([^"]+)"\s*,\s*h\.(\w+)`)
	matches := re.FindAllStringSubmatch(text, -1)

	routes := []Route{}
	for _, m := range matches {
		routes = append(routes, Route{
			Method:      strings.ToLower(m[1]),
			Path:        m[2],
			HandlerName: m[3],
			Tag:         module,
		})
	}
	return routes
}

func extractModule(path string) string {
	parts := strings.Split(path, "/")
	for i := range parts {
		if parts[i] == "app" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "App"
}

func injectSwagger(handlerPath string, routes []Route) error {
	content, _ := ioutil.ReadFile(handlerPath)
	lines := strings.Split(string(content), "\n")

	output := []string{}

	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "func (h *handler)") {
			fnName := extractFuncName(trim)

			for _, r := range routes {
				if r.HandlerName == fnName {
					output = append(output, buildAnnotation(r))
					break
				}
			}
		}
		output = append(output, line)
	}

	return ioutil.WriteFile(handlerPath, []byte(strings.Join(output, "\n")), 0644)
}

func extractFuncName(line string) string {
	// func (h *handler) Login(c *gin.Context) {
	re := regexp.MustCompile(`\)\s*(\w+)\(`)
	match := re.FindStringSubmatch(line)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func buildAnnotation(r Route) string {
	fullPath := fmt.Sprintf("/api/v1/%s/%s", r.Tag, r.Path)

	return fmt.Sprintf(`
// %s godoc
// @Summary %s
// @Description %s endpoint
// @Tags %s
// @Accept json
// @Produce json
// @Success 200 {object} util.Response
// @Failure 400 {object} util.Response
// @Router %s [%s]
`,
		r.HandlerName,
		r.HandlerName,
		r.HandlerName,
		strings.Title(r.Tag),
		fullPath,
		r.Method,
	)
}
