package render

//go:generate esc -o bindata/esc.go -pkg=bindata templates
import (
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"strings"
	"text/template"

	"github.com/aidziuz/gotests/internal/models"
	"github.com/aidziuz/gotests/internal/render/bindata"
	"github.com/aidziuz/gotests/templates"
)

const (
	nFile = 7
)

var (
	tmpls         *template.Template
	templateFuncs = template.FuncMap{
		"Capitalize": capitalize,
		"AddPackage": addpackage,
		"Field":      fieldName,
		"Receiver":   receiverName,
		"Param":      parameterName,
		"Want":       wantName,
		"Got":        gotName,
	}
)

func init() {
	initEmptyTmpls()
	for _, name := range bindata.AssetNames() {
		tmpls = template.Must(tmpls.Funcs(templateFuncs).Parse(bindata.FSMustString(false, name)))
	}
}

// LoadFromData allows to load from a data slice
func LoadFromData(templateData [][]byte) {
	initEmptyTmpls()
	for _, d := range templateData {
		tmpls = template.Must(tmpls.Parse(string(d)))
	}
}

// LoadCustomTemplates allows to load in custom templates from a specified path.
func LoadCustomTemplates(dir string) error {
	initEmptyTmpls()

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("ioutil.ReadDir: %v", err)
	}

	templateFiles := []string{}
	for _, f := range files {
		templateFiles = append(templateFiles, path.Join(dir, f.Name()))
	}
	tmpls, err = tmpls.ParseFiles(templateFiles...)
	if err != nil {
		return fmt.Errorf("tmpls.ParseFiles: %v", err)
	}
	return nil
}

// LoadCustomTemplatesName allows to load in custom templates of a specified name from the templates directory.
func LoadCustomTemplatesName(name string) error {
	initEmptyTmpls()

	f, err := templates.Dir(false, "/").Open(name)
	if err != nil {
		return fmt.Errorf("templates.Open: %v", err)
	}

	files, err := f.Readdir(nFile)
	if err != nil {
		return fmt.Errorf("f.Readdir: %v", err)
	}

	for _, f := range files {
		text, err := templates.FSString(false, path.Join("/", name, f.Name()))
		if err != nil {
			return fmt.Errorf("templates.FSString: %v", err)
		}

		tmpls, err = tmpls.Parse(text)
		if err != nil {
			return fmt.Errorf("tmpls.Parse: %v", err)
		}
	}

	return nil
}

func initEmptyTmpls() {
	tmpls = template.New("render").Funcs(templateFuncs)
}

func capitalize(s string) string {
	if len(s) > 1 {
		return strings.ToUpper(string(s[0])) + s[1:]
	}
	return strings.ToUpper(s)
}

func addpackage(e *models.Expression) bool {
	return !isBasicType(e.Value) && !strings.Contains(e.Value, ".")
}

func isBasicType(t string) bool {
	switch t {
	case "bool", "string", "int", "int8", "int16", "int32", "int64", "uint",
		"uint8", "uint16", "uint32", "uint64", "uintptr", "byte", "rune",
		"float32", "float64", "complex64", "complex128":
		return true
	default:
		if strings.HasPrefix(t, "map[string]") {
			return true
		}
		return false
	}
}

func fieldName(f *models.Field) string {
	var n string
	if f.IsNamed() {
		n = f.Name
	} else {
		n = f.Type.String()
	}
	return n
}

func receiverName(f *models.Receiver) string {
	var n string
	if f.IsNamed() {
		n = f.Name
	} else {
		n = f.ShortName()
	}
	if n == "name" {
		// Avoid conflict with test struct's "name" field.
		n = "n"
	} else if n == "t" {
		// Avoid conflict with test argument.
		// "tr" is short for t receiver.
		n = "tr"
	}
	return n
}

func parameterName(f *models.Field) string {
	var n string
	if f.IsNamed() {
		n = f.Name
	} else {
		n = fmt.Sprintf("in%v", f.Index)
	}
	return n
}

func wantName(f *models.Field) string {
	var n string
	if f.IsNamed() {
		n = "expected" + strings.Title(f.Name)
	} else if f.Index == 0 {
		n = "expected"
	} else {
		n = fmt.Sprintf("expected%v", f.Index)
	}
	return n
}

func gotName(f *models.Field) string {
	var n string
	if f.IsNamed() {
		n = "actual" + strings.Title(f.Name)
	} else if f.Index == 0 {
		n = "actual"
	} else {
		n = fmt.Sprintf("actual%v", f.Index)
	}
	return n
}

func Header(w io.Writer, h *models.Header) error {
	if err := tmpls.ExecuteTemplate(w, "header", h); err != nil {
		return err
	}
	_, err := w.Write(h.Code)
	return err
}

func TestFunction(w io.Writer, f *models.Function, h *models.Header, constr *models.Function, printInputs, subtests, parallel bool, templateParams map[string]interface{}) error {
	return tmpls.ExecuteTemplate(w, "function", struct {
		*models.Function
		Header         *models.Header
		Constructor    *models.Function
		PrintInputs    bool
		Subtests       bool
		Parallel       bool
		TemplateParams map[string]interface{}
	}{
		Function:       f,
		Header:         h,
		Constructor:    constr,
		PrintInputs:    printInputs,
		Subtests:       subtests,
		Parallel:       parallel,
		TemplateParams: templateParams,
	})
}
