package main

import (
	"flag"
	"net/http"
	"rs_server/lib"
	"rs_server/lib/util"
	"rs_server/lib/webapi"

	log "github.com/cihub/seelog"
)

type Configs struct {
	RSPubBind      string `json:"RSPubBind"`
	AdapterSubBind string `json:"AdapterSubBind"`
}

func NewConfigs() Configs {
	return Configs{
		RSPubBind:      "0.0.0.0:5501",
		AdapterSubBind: "0.0.0.0:5520",
	}
}

type WebAPICtrlImpl struct {
	rs       *lib.RealsenseService
	isEnable bool
}

func NewWebAPICtrlImpl(rs *lib.RealsenseService) *WebAPICtrlImpl {
	w := new(WebAPICtrlImpl)
	w.isEnable = true
	w.rs = rs
	return w
}

func (w *WebAPICtrlImpl) Enable(enable bool) {
	w.isEnable = enable
	w.rs.Enable(enable)
}
func (w *WebAPICtrlImpl) IsEnable() bool {
	return w.isEnable
}

func main() {
	configs := NewConfigs()
	util.ReadConfig(&configs)

	var (
		rsAddr      = flag.String("r", configs.RSPubBind, "Specify IP and port of Realsense server.")
		adapterAddr = flag.String("a", configs.AdapterSubBind, "Specify IP and port of Adapter server.")
	)
	flag.Parse()

	rs := lib.NewRealsenseService("tcp://"+*rsAddr, "tcp://"+*adapterAddr)
	defer rs.Destory()
	log.Info("Starting Realsense Service..")
	rs.Start()
	defer rs.Stop()

	controller := NewWebAPICtrlImpl(rs)
	webapi.SetUpWebAPIforCommon(controller)

	log.Info("Http Server 5002 Start")
	http.ListenAndServe(":5002", nil)

}
