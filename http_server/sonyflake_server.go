package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/sony/sonyflake"
	"github.com/zpatrick/go-config"
)

var sf *sonyflake.Sonyflake

func doInit(c *config.Config) {
	var st sonyflake.Settings
	mid, err := c.Int("global.mid")
	if err != nil {
		panic("get mid error")
	}
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
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
