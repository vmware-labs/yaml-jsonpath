package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/glyn/go-yamlpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

func main() {
	tmpl := template.New("template")
	tmpl, err := tmpl.Parse(`<style type="text/css">
.tg  {border-collapse:collapse;border-spacing:0;}
.tg td{border-color:black;border-style:solid;border-width:1px;font-family:Arial, sans-serif;font-size:14px;
  overflow:hidden;padding:10px 5px;word-break:normal;}
.tg th{border-color:black;border-style:solid;border-width:1px;font-family:Arial, sans-serif;font-size:14px;
  font-weight:normal;overflow:hidden;padding:10px 5px;word-break:normal;}
.tg .tg-zv4m{border-color:#ffffff;text-align:left;vertical-align:top}
textarea, pre, input {font-family:Consolas,monospace; font-size:14px}
h1, body, label {font-family: Lato,proxima-nova,Helvetica Neue,Arial,sans-serif}
textarea, input {
	box-sizing: border-box;
	border: 1px solid;
	background-color: #f8f8f8;
	resize: none;
  }
</style>
<h1>go-yamlpath evaluator</h1>
<table class="tg">
<thead>
  <tr valign="top">
	<th class="tg-zv4m">
<form method="POST">
<label>YAML document</label> (<a href="https://yaml.org/spec/1.2/spec.html" target="_blank">syntax</a>):<br />
<pre>
<textarea name="YAML document" cols="80" rows="30" placeholder="YAML...">{{ .YAML }}</textarea>
</pre><br /><br />
<label>JSON path</label> (<a href="https://github.com/glyn/go-yamlpath#syntax" target="_blank">syntax</a>):<br />
<pre>
<input type="text" size="80" name="JSON path" placeholder="JSON path..." value="{{ .JSONPath }}"><br />
<input type="submit" value="Evaluate">
</pre>
</form>

	</th>
	<th class="tg-zv4m">
	   &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
	   &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
	</th>
	<th class="tg-zv4m">
	<label>Output:</label><br /><br />
{{if .YAMLError}}
	<br />{{ .YAMLError }}<br />
{{end}}
{{if .JSONPathError}}
    <br />Invalid JSON path: {{ .JSONPathError }}<br />
{{end}}
<pre>
{{ .Output }}<br />
</pre>
	</th>
  </tr>
</thead>
</table>
`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		type output struct {
			YAML          string
			YAMLError     error
			JSONPath      string
			JSONPathError error
			Success       bool
			Output        string
		}

		if r.Method != http.MethodPost {
			if e := tmpl.Execute(w, nil); e != nil {
				respondWithError(w, e)
			}
			return
		}

		y := r.FormValue("YAML document")
		op := output{
			YAML: y,
		}

		problem := false

		var n yaml.Node
		err := yaml.Unmarshal([]byte(y), &n)
		if err != nil {
			problem = true
			op.YAMLError = err
		}

		j := r.FormValue("JSON path")
		op.JSONPath = j
		path, err := yamlpath.NewPath(j)
		if err != nil {
			problem = true
			op.JSONPathError = err
		}

		if problem {
			if e := tmpl.Execute(w, op); e != nil {
				respondWithError(w, e)
			}
			return
		}

		results := path.Find(&n)

		out := []string{}
		for _, a := range results {
			b, err := encode(a)
			if err != nil {
				respondWithError(w, err)
				return
			}
			out = append(out, b)
		}

		op.Success = true
		op.Output = strings.Join(out, "---\n")
		if e := tmpl.Execute(w, op); e != nil {
			respondWithError(w, e)
		}
	})

	if e := http.ListenAndServe(":8080", nil); e != nil {
		log.Fatal(e)
	}
}

func encode(a *yaml.Node) (string, error) {
	var buf bytes.Buffer
	e := yaml.NewEncoder(&buf)
	defer e.Close()
	e.SetIndent(2)

	if err := e.Encode(a); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func respondWithError(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
