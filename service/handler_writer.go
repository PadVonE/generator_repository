package service

import (
	"errors"
	"github.com/gin-gonic/gin"
	"html/template"
	"log"
)

var (
	ContextResult = "Result"
)

// EndpointResult
type (
	EndpointResult struct {
		Status   interface{}   `json:"status"`
		Meta     *EndpointMeta `json:"_meta,omitempty"`
		Payload  interface{}   `json:"payload"`
		Template string        `json:"-"`
	}

	HtmlResult struct {
	}

	EndpointMeta struct {
		IsLastPage bool    `json:"is_last_page"`
		TotalCount int     `json:"total"`
		TotalClick int     `json:"total_click"`
		TotalViews int     `json:"total_views"`
		TotalCtr   float64 `json:"total_ctr"`
	}

	ViewData struct {
		Head template.HTML
		Body template.HTML
	}
)

// Result to get current context result.
func Result(ctx *gin.Context) *EndpointResult {
	return ctx.MustGet(ContextResult).(*EndpointResult)
}

// Error to check for error existence in context.
func Error(ctx *gin.Context) error {
	if value, isError := Result(ctx).Status.(error); isError && value != nil {
		return value
	}

	return nil
}

func SetPayload(ctx *gin.Context, value interface{}) {
	Result(ctx).Payload = value
}

func SetTemplate(ctx *gin.Context, value string) {
	Result(ctx).Template = value
}

// AbortWith aborts request processing with error.
func AbortWith(ctx *gin.Context, err error) {
	Result(ctx).Status = err
	ctx.Abort()
}

// ResponseWriter provides result structure to context,
// normalize, format and sends response to a client.
func (s *Service) ResponseHtmlWriter(ctx *gin.Context) {
	ctx.Set(ContextResult, &EndpointResult{})
	ctx.Next()

	result := Result(ctx)

	viewData, ok := result.Payload.(ViewData)

	if !ok {
		log.Printf("%#v", result.Payload)
		//raven.CaptureError(err, nil)
		AbortWith(ctx, errors.New("Payload"))
	}
	tmpl := template.Must(template.ParseFiles(
		"./templates/index.html",
	))

	err := tmpl.Execute(ctx.Writer, viewData)

	if err != nil {
		log.Println(err)
		//	raven.CaptureError(err, nil)
	}

}
