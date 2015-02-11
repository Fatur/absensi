package absensi

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

type Payload struct {
	Data []Event
}

var inMemoryData Payload
var attandances map[AttandanceId]Attandance
var uploadFileCsv = func(render render.Render, log *log.Logger) {
	render.HTML(200, "uploadCsv", nil)
}
var uploadFileCsvHandler = func(w http.ResponseWriter, r *http.Request) (int, string) {
	log.Println("parsing form")
	err := r.ParseMultipartForm(100000)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}

	files := r.MultipartForm.File["files"]
	for i, _ := range files {
		log.Println("getting handle to file")

		file, err := files[i].Open()
		defer file.Close()
		if err != nil {
			return http.StatusInternalServerError, err.Error()
		}

		reader := csv.NewReader(file)
		reader.FieldsPerRecord = -1

		rawCsvdata, err := reader.ReadAll()
		if err != nil {
			return http.StatusInternalServerError, err.Error()
		}

		for _, each := range rawCsvdata {
			log.Printf("Id : %s and Location : %s  Time : %s  Type : %s\n", each[0], each[1], each[2], each[3])
			t, err1 := time.Parse(time.RFC3339Nano, each[2])
			if err1 != nil {
				return http.StatusInternalServerError, err.Error()
			}
			tipe, err2 := strconv.ParseInt(each[3], 0, 16)
			if err2 != nil {
				return http.StatusInternalServerError, err.Error()
			}
			evt := Event{each[0], each[1], t, EventType(tipe)}
			_, err3 := inMemoryData.Add(evt)
			if err3 != nil {
				return http.StatusInternalServerError, err.Error()
			}
			calculateAttandance(evt)
		}
	}

	return 200, "ok"

}
var getAllAttandance = func() string {
	arrAtd := convertAttandanceToArray()
	response, err := json.MarshalIndent(arrAtd, "", " ")
	if err != nil {
		panic(err)
	}
	return string(response)
}
var getAllItem = func() string {
	return getAll(inMemoryData)
}
var addNew = func(w http.ResponseWriter, r *http.Request) (int, string) {
	newEvt := decodeNewRequest(r)
	id, err := inMemoryData.Add(newEvt)
	if err != nil {
		panic(err)
	}
	calculateAttandance(newEvt)
	w.Header().Set("Location", fmt.Sprintf("/%d", id))
	return http.StatusCreated, "OK"
}

func convertAttandanceToArray() []Attandance {
	attd := make([]Attandance, 0, len(attandances))
	for k := range attandances {
		attd = append(attd, attandances[k])
	}
	return attd
}
func init() {
	attandances = make(map[AttandanceId]Attandance)
	loadDataTo(&inMemoryData)
	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Directory:  "templates",
		Extensions: []string{".html"},
	}))
	setupLogs(m)

	http.Handle("/", m)
}
func setupLogs(m *martini.ClassicMartini) {
	m.Group("/logs", func(r martini.Router) {
		r.Get("", getAllItem)
		r.Post("", addNew)
	})
	m.Group("/attandances", func(r martini.Router) {
		r.Get("", getAllAttandance)
	})
	m.Group("/upload", func(r martini.Router) {
		r.Get("", uploadFileCsv)
		r.Post("", uploadFileCsvHandler)
	})

}
func getAll(data Payload) string {
	return data.ToJson()
}

func (data *Payload) ToJson() string {
	response, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		panic(err)
	}
	return string(response)
}
func (repo *Payload) Add(evt Event) (string, error) {
	repo.Data = append(repo.Data, evt)
	return evt.Id, nil
}
func decodeNewRequest(r *http.Request) Event {
	decoder := json.NewDecoder(r.Body)
	var evt Event
	err := decoder.Decode(&evt)
	if err != nil {
		panic(err)
	}
	return evt
}
func loadDataTo(repo *Payload) {

	evt1 := Event{"001", "5:6", time.Now(), In}
	evt2 := Event{"002", "5:7", time.Now(), In}
	evt3 := Event{"001", "5:7", time.Now().Add(8 * time.Hour), Out}

	repo.Add(evt1)
	repo.Add(evt2)
	repo.Add(evt3)

	calculateAttandance(evt1)

	calculateAttandance(evt2)

	calculateAttandance(evt3)

}
func calculateAttandance(evt Event) {
	id := evt.CreateAttandanceId()

	attd, ok := attandances[id]
	if !ok {
		attandances[id] = evt.CreateAttandance()
	} else {
		attd.Calculate(evt)
		attandances[id] = attd
	}
}
