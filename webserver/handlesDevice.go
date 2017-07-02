package webserver

import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/koestler/go-ve-sensor/dataflow"
)

func HandleDeviceIndex(env *Environment, w http.ResponseWriter, r *http.Request) error {
	devices := dataflow.DevicesGet()

	writeJsonHeaders(w)
	b, err := json.MarshalIndent(devices, "", "    ")
	if err != nil {
		return StatusError{500, err}
	}
	w.Write(b)

	return nil;
}

func HandleDeviceGetRoundedValues(env *Environment, w http.ResponseWriter, r *http.Request) error {
	vars := mux.Vars(r)

	device, err := dataflow.DevicesGetByName(vars["DeviceId"])
	if err != nil {
		return StatusError{404, err}
	}

	roundedValues := env.RoundedStorage.GetMap(dataflow.Filter{Devices: map[*dataflow.Device]bool{device: true}})
	roundedValuesEssential := roundedValues.ConvertToEssential()

	writeJsonHeaders(w)
	b, err := json.MarshalIndent(roundedValuesEssential, "", "    ")
	if err != nil {
		return StatusError{500, err}
	}
	w.Write(b)
	return nil
}
