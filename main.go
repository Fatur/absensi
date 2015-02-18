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
	"appengine"
	"appengine/datastore"
)

type Payload struct {
	Data []Event
}

var uploadFileCsv = func(render render.Render, log *log.Logger) {
	render.HTML(200, "uploadCsv", nil)
}
var uploadFileCsvHandler = func(w http.ResponseWriter, r *http.Request) (int, string) {
	c := appengine.NewContext(r)

	log.Println("parsing form")
	err := r.ParseMultipartForm(100000)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}

	files := r.MultipartForm.File["files"]
	for i, _ := range files {
		log.Println("getting handle to file")

		file, errOpenFile := files[i].Open()
		defer file.Close()
		if errOpenFile != nil {
			return http.StatusInternalServerError, errOpenFile.Error()
		}

		reader := csv.NewReader(file)
		reader.FieldsPerRecord = -1

		rawCsvdata, errReadFile := reader.ReadAll()
		if errReadFile != nil {
			return http.StatusInternalServerError, errReadFile.Error()
		}

		for _, each := range rawCsvdata {
			log.Printf("Id : %s and Location : %s  Time : %s  Type : %s\n", each[0], each[1], each[2], each[3])
			t, parseTimeError := time.Parse(time.RFC3339Nano, each[2])
			if parseTimeError != nil {
				return http.StatusInternalServerError, parseTimeError.Error()
			}
			tipe, parseTypeError := strconv.ParseInt(each[3], 0, 16)
			if parseTypeError != nil {
				return http.StatusInternalServerError, parseTypeError.Error()
			}
			evt := Event{each[0], each[1], t, EventType(tipe)}
			_, saveEvtError := saveEvent(c, evt)
			if saveEvtError != nil {
				return http.StatusInternalServerError, saveEvtError.Error()
			}
			errAttd := calculateAttandance(c, evt)
			if errAttd != nil {
				return http.StatusInternalServerError, errAttd.Error()
			}
		}
	}

	return 200, "ok"

}

func saveEvent(context appengine.Context, evt Event) (*datastore.Key, error) {
	key := datastore.NewIncompleteKey(context, "Event", nil)
	return datastore.Put(context, key, &evt)
}
func calculateAttandance(context appengine.Context, evt Event) error {
	id := evt.CreateAttandanceId()
	attd, err := findAttandance(context, id)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			attd = evt.CreateAttandance()
		} else {
			return err
		}
	} else {
		attd.Calculate(evt)
	}

	_, errPut := saveAttandance(context, attd)

	return errPut
}
func saveAttandance(context appengine.Context, attd Attandance) (*datastore.Key, error) {
	key := datastore.NewKey(context, "Attandance", attd.Id.ToKey(), 0, nil)
	return datastore.Put(context, key, &attd)

}
func findAttandance(context appengine.Context, attandanceId AttandanceId) (Attandance, error) {
	key := datastore.NewKey(context, "Attandance", attandanceId.ToKey(), 0, nil)
	var attandance Attandance
	err := datastore.Get(context, key, &attandance)
	return attandance, err

}

var addNew = func(w http.ResponseWriter, r *http.Request) (int, string) {
	newEvt := decodeNewRequest(r)
	c := appengine.NewContext(r)
	id, err := saveEvent(c, newEvt)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}
	errAttd := calculateAttandance(c, newEvt)
	if errAttd != nil {
		return http.StatusInternalServerError, errAttd.Error()
	}
	w.Header().Set("Location", fmt.Sprintf("/%d", id.Encode()))
	return http.StatusCreated, "OK"
}
var getAllAttandance = func(r *http.Request) (int, string) {
	context := appengine.NewContext(r)
	atandances, errFind := findAllAttandance(context)
	if errFind != nil {
		return http.StatusInternalServerError, errFind.Error()
	}
	response, errAttd := json.MarshalIndent(atandances, "", " ")
	if errAttd != nil {
		return http.StatusInternalServerError, errAttd.Error()
	}
	return 200, string(response)
}

func findAllAttandance(context appengine.Context) ([]Attandance, error) {
	q := datastore.NewQuery("Attandance").Order("Id.EmployeeId").Order("Id.Date")
	var attds []Attandance
	_, err := q.GetAll(context, &attds)

	return attds, err
}

var getAllById = func(r *http.Request, params martini.Params) (int, string) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Event").Filter("Id =", params["id"]).Order("Id").Order("Time")
	var evts []Event
	_, err := q.GetAll(c, &evts)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}

	var inMemoryData Payload
	inMemoryData.Data = evts

	return 200, getAll(inMemoryData)

}
var getAllItem = func(r *http.Request) (int, string) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("Event").Order("Id").Order("Time")
	var evts []Event
	_, err := q.GetAll(c, &evts)
	if err != nil {
		return http.StatusInternalServerError, err.Error()
	}
	var inMemoryData Payload
	inMemoryData.Data = evts
	return 200, getAll(inMemoryData)
}

func init() {

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
		r.Get("/:id", getAllById)
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

func decodeNewRequest(r *http.Request) Event {
	decoder := json.NewDecoder(r.Body)
	var evt Event
	err := decoder.Decode(&evt)
	if err != nil {
		panic(err)
	}
	return evt
}
