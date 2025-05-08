package web

import (
	"bytes"
	"embed"
	"errors"
	"github.com/gin-gonic/gin"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed templates/*.html
var templatesEmbed embed.FS

var mainTemplateSet *HTMLTemplateSet

func init() {
	var err error
	if mainTemplateSet, err = NewHTMLTemplateSet(
		templatesEmbed, "templates", "base.html",
	); err != nil {
		panic(err)
	}
}

type HTMLTemplateSet struct {
	templates map[string]*template.Template
}

func NewHTMLTemplateSet(fs fs.ReadDirFS, path string, baseFile string) (templateSet *HTMLTemplateSet, err error) {
	path = strings.TrimSuffix(path, "/")

	templateSet = &HTMLTemplateSet{
		templates: make(map[string]*template.Template),
	}

	dir, err := fs.ReadDir(path)
	if err != nil {
		return
	}

	for _, entry := range dir {
		if entry.Name() == baseFile {
			continue
		}

		t := template.Must(template.ParseFS(fs, path+"/"+baseFile, path+"/"+entry.Name()))
		templateName, _ := strings.CutSuffix(entry.Name(), ".html")
		templateSet.templates[templateName] = t
	}

	return
}

func (d *HTMLTemplateSet) LoadTemplate(key string) (t *template.Template, err error) {
	t, exists := d.templates[key]
	if !exists {
		err = errors.New("template does not exist")
		return
	}

	return
}

func (d *HTMLTemplateSet) FormatTemplate(key string, data any) (buf bytes.Buffer, err error) {
	t, err := d.LoadTemplate(key)
	if err != nil {
		return
	}

	if err = t.Execute(&buf, data); err != nil {
		return
	}

	return
}

func (d *HTMLTemplateSet) WriteTemplate(c *gin.Context, httpStatus int, key string, data any) {
	buf, err := d.FormatTemplate(key, data)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", "text/html")
	c.Status(httpStatus)
	io.Copy(c.Writer, &buf)

	return
}
