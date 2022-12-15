package helpers

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"sort"
	"tischtennis/database"
)

var VERSION = os.Getenv("VERSION")
var BASE_PATH = os.Getenv("BASE_PATH")

func Mul(param1 float64, param2 float64) string {
	return fmt.Sprintf("%.2f", param1*param2)
}

func BuildPage(path string, data interface{}) *bytes.Buffer {
	var bodyBuffer bytes.Buffer
	t := template.Must(template.New("").Funcs(template.FuncMap{
		"mul": Mul,
	}).ParseFiles(path, "templates/base.html"))
	t.ExecuteTemplate(&bodyBuffer, "base", data)
	return &bodyBuffer
}
func AlphSortPeople(people []database.Person) (alphSortedPeople []database.Person) {
	alphSortedPeople = make([]database.Person, len(people))
	copy(alphSortedPeople, people)
	sort.Slice(alphSortedPeople, func(i, j int) bool {
		return alphSortedPeople[i].Name < alphSortedPeople[j].Name
	})
	return alphSortedPeople

}
