package database

type Person struct {
	Name   string
	Id     int64
	FaIcon string
	Wins   int64
	Losses int64
	Score  float64
}
type Game struct {
	Id              int64
	ReporterName    string
	ReporterId      int64
	OtherPersonName string
	OtherPersonId   int64
	Wins            int64
	Losses          int64
}
