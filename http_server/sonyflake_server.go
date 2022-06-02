package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/heipacker/sonyflake"
	"github.com/zpatrick/go-config"
)

var sf *sonyflake.Sonyflake

func doInit(c *config.Config) {
	var st sonyflake.Settings
	mid, err := c.Int("global.mid")
	if err != nil {
		panic("get mid error")
	}
	log.Println("Server id is :" + strconv.Itoa(mid))
	st.MachineID = func() (uint16, error) {
		return uint16(mid), nil
	}
	sf = sonyflake.NewSonyflake(st)
	if sf == nil {
		panic("sonyflake not created")
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	id, err := sf.NextID()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(sonyflake.Decompose(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header()["Content-Type"] = []string{"application/json; charset=utf-8"}
	w.Write(body)
}

func batchHandler(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.ParseInt(r.URL.Query().Get("c"), 10, 64)
	var result []map[string]uint64
	for i := 0; i < int(count); i++ {
		id, err := sf.NextID()
		if err != nil {
			continue
		}
		result = append(result, sonyflake.Decompose(id))
	}
	body, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header()["Content-Type"] = []string{"application/json; charset=utf-8"}
	w.Write(body)
}

func handlerStr(w http.ResponseWriter, r *http.Request) {
	id, err := sf.NextID()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	decompose := sonyflake.Decompose(id)
	dataMap := make(map[string]interface{})
	dataMap["id"] = strconv.FormatInt(int64(decompose["id"]), 10)
	dataMap["msb"] = decompose["msb"]
	dataMap["time"] = decompose["time"]
	dataMap["sequence"] = decompose["sequence"]
	dataMap["machine-id"] = decompose["machine-id"]
	body, err := json.Marshal(dataMap)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header()["Content-Type"] = []string{"application/json; charset=utf-8"}
	w.Write(body)
}

func main() {
	iniFile := config.NewINIFile("config.ini")
	c := config.NewConfig([]config.Provider{iniFile})
	if err := c.Load(); err != nil {
		panic("load config error")
	}
	port, err := c.Int("global.port")
	if err != nil {
		panic("get port config error")
	}
	doInit(c)
	log.Println("Server is at :" + strconv.Itoa(port))

	http.HandleFunc("/", handler)
	http.HandleFunc("/batch", batchHandler)
	http.HandleFunc("/str", handlerStr)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
