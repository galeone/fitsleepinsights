package app

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"strings"

	chartRender "github.com/go-echarts/go-echarts/v2/render"
)

const baseTpl = `
<div class="item" id="{{ .ChartID }}" style="width:{{ .Initialization.Width }};height:{{ .Initialization.Height }};border:1px solid rgba(204, 204, 220, 0.12)"></div>
{{- range .JSAssets.Values }}
	{{ if not (hasSuffix . "echarts.min.js") }}
	   <script src="{{ . }}"></script>
   {{ end }}
{{- end }}
<script>
    "use strict";
    let echarts_{{ .ChartID | safeJS }} = echarts.init(document.getElementById('{{ .ChartID | safeJS }}'), "{{ .Theme }}");
    let option_{{ .ChartID | safeJS }} = {{ .JSON }};
    echarts_{{ .ChartID | safeJS }}.setOption(option_{{ .ChartID | safeJS }});
    {{- range .JSFunctions.Fns }}
    {{ . | safeJS }}
    {{- end }}
</script>
`

func renderChart(c interface{}) template.HTML {
	var buf bytes.Buffer
	r := c.(chartRender.Renderer)
	err := r.Render(&buf)
	if err != nil {
		log.Printf("Failed to render chart: %s", err)
		return template.HTML("")
	}

	return template.HTML(buf.String())
}

type chartRenderer struct {
	c      interface{}
	before []func()
}

func newChartRenderer(c interface{}, before ...func()) chartRender.Renderer {
	return &chartRenderer{c: c, before: before}
}

func (r *chartRenderer) Render(w io.Writer) error {
	const tplName = "chart"
	for _, fn := range r.before {
		fn()
	}

	tpl := template.
		Must(template.New(tplName).
			Funcs(template.FuncMap{
				"safeJS": func(s interface{}) template.JS {
					return template.JS(fmt.Sprint(s))
				},
				"hasSuffix": strings.HasSuffix,
			}).
			Parse(baseTpl),
		)

	err := tpl.ExecuteTemplate(w, tplName, r.c)
	return err
}
