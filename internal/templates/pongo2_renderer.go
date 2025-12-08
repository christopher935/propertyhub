package templates

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/flosch/pongo2/v6"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

type Pongo2Render struct {
	TemplateDir string
	TemplateSet *pongo2.TemplateSet
}

type Pongo2HTML struct {
	Template *pongo2.Template
	Name     string
	Data     interface{}
}

func NewPongo2Render(templateDir string) *Pongo2Render {
	loader := pongo2.MustNewLocalFileSystemLoader(templateDir)
	templateSet := pongo2.NewSet("templates", loader)
	templateSet.Debug = gin.Mode() == gin.DebugMode

	return &Pongo2Render{
		TemplateDir: templateDir,
		TemplateSet: templateSet,
	}
}

func (r *Pongo2Render) Instance(name string, data interface{}) render.Render {
	template, err := r.TemplateSet.FromFile(name)
	if err != nil {
		panic(fmt.Sprintf("Failed to load template %s: %v", name, err))
	}
	return &Pongo2HTML{
		Template: template,
		Name:     name,
		Data:     data,
	}
}

func (p *Pongo2HTML) Render(w http.ResponseWriter) error {
	p.WriteContentType(w)

	var context pongo2.Context
	if p.Data != nil {
		switch d := p.Data.(type) {
		case pongo2.Context:
			context = d
		case gin.H:
			context = pongo2.Context{}
			for k, v := range d {
				context[k] = v
			}
		case map[string]interface{}:
			context = pongo2.Context{}
			for k, v := range d {
				context[k] = v
			}
		default:
			context = pongo2.Context{"data": p.Data}
		}
	} else {
		context = pongo2.Context{}
	}

	return p.Template.ExecuteWriter(context, w)
}

func (p *Pongo2HTML) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

func RegisterPongo2Filters() {
	pongo2.RegisterFilter("formatPrice", filterFormatPrice)
	pongo2.RegisterFilter("formatDate", filterFormatDate)
	pongo2.RegisterFilter("formatNumber", filterFormatNumber)
	pongo2.RegisterFilter("safe", filterSafe)
	pongo2.RegisterFilter("upper", filterUpper)
	pongo2.RegisterFilter("lower", filterLower)
	pongo2.RegisterFilter("title", filterTitle)
	pongo2.RegisterFilter("currentYear", filterCurrentYear)
}

func filterFormatPrice(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	switch v := in.Interface().(type) {
	case int:
		return pongo2.AsValue(fmt.Sprintf("$%s", humanize.Comma(int64(v)))), nil
	case int64:
		return pongo2.AsValue(fmt.Sprintf("$%s", humanize.Comma(v))), nil
	case float64:
		return pongo2.AsValue(fmt.Sprintf("$%s", humanize.CommafWithDigits(v, 2))), nil
	case float32:
		return pongo2.AsValue(fmt.Sprintf("$%s", humanize.CommafWithDigits(float64(v), 2))), nil
	default:
		return pongo2.AsValue(fmt.Sprintf("%v", v)), nil
	}
}

func filterFormatDate(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if t, ok := in.Interface().(time.Time); ok {
		format := "January 2, 2006"
		if !param.IsNil() && param.String() != "" {
			format = param.String()
		}
		return pongo2.AsValue(t.Format(format)), nil
	}
	if t, ok := in.Interface().(*time.Time); ok && t != nil {
		format := "January 2, 2006"
		if !param.IsNil() && param.String() != "" {
			format = param.String()
		}
		return pongo2.AsValue(t.Format(format)), nil
	}
	return in, nil
}

func filterFormatNumber(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	switch v := in.Interface().(type) {
	case int:
		return pongo2.AsValue(humanize.Comma(int64(v))), nil
	case int64:
		return pongo2.AsValue(humanize.Comma(v)), nil
	case float64:
		return pongo2.AsValue(humanize.Commaf(v)), nil
	case float32:
		return pongo2.AsValue(humanize.Commaf(float64(v))), nil
	default:
		return pongo2.AsValue(fmt.Sprintf("%v", v)), nil
	}
}

func filterSafe(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsSafeValue(in.String()), nil
}

func filterUpper(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(in.String()), nil
}

func filterLower(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(in.String()), nil
}

func filterTitle(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(in.String()), nil
}

func filterCurrentYear(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	return pongo2.AsValue(time.Now().Year()), nil
}

func GetTemplateDir() string {
	return filepath.Join("web", "templates")
}
