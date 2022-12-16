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

func BuildPage(path string, data interface{}) (*bytes.Buffer, error) {
	var bodyBuffer bytes.Buffer
	fmt.Println(1, data)
	parsedFiles, err := template.New("").Funcs(template.FuncMap{
		"mul": Mul,
	}).ParseFiles(path, "templates/base.html")
	if err != nil {
		return &bodyBuffer, err
	}
	t := template.Must(parsedFiles, err)
	fmt.Println(2)
	err = t.ExecuteTemplate(&bodyBuffer, "base", data)
	fmt.Println(3)
	return &bodyBuffer, err
}
func AlphSortPeople(people []database.Person) (alphSortedPeople []database.Person) {
	alphSortedPeople = make([]database.Person, len(people))
	copy(alphSortedPeople, people)
	sort.Slice(alphSortedPeople, func(i, j int) bool {
		return alphSortedPeople[i].Name < alphSortedPeople[j].Name
	})
	return alphSortedPeople

}

func GetPersonIdToNameMap(people []database.Person) (personIdToNameMap map[string]string) {
	personIdToNameMap = make(map[string]string)
	for _, person := range people {
		personIdToNameMap[person.Id] = person.Name
	}
	return personIdToNameMap

}
