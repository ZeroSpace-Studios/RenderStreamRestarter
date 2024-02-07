package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RenderStreamLayersResponse struct {
	Status struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Details []struct {
			TypeURL string `json:"type_url"`
			Value   string `json:"value"`
		} `json:"details"`
	} `json:"status"`
	Result []struct {
		UID  string `json:"uid"`
		Name string `json:"name"`
	} `json:"result"`
}

type RenderStreamLayerStatusResponse struct {
	Status struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Details []struct {
			TypeURL string `json:"type_url"`
			Value   string `json:"value"`
		} `json:"details"`
	} `json:"status"`
	Result struct {
		Reference struct {
			TNow float64 `json:"tNow"`
		} `json:"reference"`
		Workload struct {
			UID       string `json:"uid"`
			Name      string `json:"name"`
			Instances []struct {
				MachineUID    string `json:"machineUid"`
				MachineName   string `json:"machineName"`
				State         string `json:"state"`
				HealthMessage string `json:"healthMessage"`
				HealthDetails string `json:"healthDetails"`
			} `json:"instances"`
		} `json:"workload"`
		Streams []struct {
			UID             string `json:"uid"`
			Name            string `json:"name"`
			SourceMachine   string `json:"sourceMachine"`
			ReceiverMachine string `json:"receiverMachine"`
			Status          struct {
				SubscriptionWanted  bool    `json:"subscriptionWanted"`
				SubscribeSuccessful bool    `json:"subscribeSuccessful"`
				TLastDropped        float64 `json:"tLastDropped"`
				TLastError          float64 `json:"tLastError"`
				LastErrorMessage    string  `json:"lastErrorMessage"`
			} `json:"status"`
			StatusString string `json:"statusString"`
		} `json:"streams"`
		AssetErrors []string `json:"assetErrors"`
	} `json:"result"`
}

type RenderStreamRestartLayerRequest struct {
	Layers []struct {
		Uid  string `json:"uid"`
		Name string `json:"name"`
	} `json:"layers"`
}

type RenderStreamReader struct {
	server string
}

func (r *RenderStreamReader) GetRenderStreamLayerByName(name string) (string, error) {

	res, err := http.Get(fmt.Sprintf("http://%s/api/session/renderstream/layers", r.server))

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	var response RenderStreamLayersResponse

	b, err := io.ReadAll(res.Body)

	if err != nil {
		return "", err

	}

	err = json.Unmarshal(b, &response)

	if err != nil {
		return "", err
	}

	for _, layer := range response.Result {
		if layer.Name == name {
			return layer.UID, nil
		}
	}
	return "", fmt.Errorf("layer not found")
}

func (r *RenderStreamReader) GetLayerStatus(layerUID string) (string, error) {

	res, err := http.Get(fmt.Sprintf("http://%s/api/session/renderstream/layerstatus?uid=%s", r.server, layerUID))

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	var response RenderStreamLayerStatusResponse

	b, err := io.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	err = json.Unmarshal(b, &response)

	if err != nil {
		return "", err
	}

	if len(response.Result.Workload.Instances) == 0 {
		return "stopped", nil
	}

	return response.Result.Workload.Instances[0].State, nil

}

func (r *RenderStreamReader) RestartRenderStreamLayer(layerUID string, name string) error {

	body := RenderStreamRestartLayerRequest{
		Layers: []struct {
			Uid  string `json:"uid"`
			Name string `json:"name"`
		}{
			{
				Uid:  layerUID,
				Name: name,
			},
		},
	}
	data, err := json.Marshal(body)

	if err != nil {
		return err

	}

	reader := bytes.NewReader(data)

	resp, err := http.Post(fmt.Sprintf("http://%s/api/session/renderstream/startlayers", r.server), "application/json", reader)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	data, err = io.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	response := struct {
		Status struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Details []struct {
				TypeURL string `json:"type_url"`
				Value   string `json:"value"`
			} `json:"details"`
		} `json:"status"`
	}{}

	err = json.Unmarshal(data, &response)

	if err != nil {
		return err
	}

	return nil
}

func main() {
	server := flag.String("server", "localhost", "The server to connect to")
	layer := flag.String("layer", "Rs", "The layer to restart")
	flag.Parse()

	reader := RenderStreamReader{
		server: *server,
	}

	layerUID, err := reader.GetRenderStreamLayerByName(*layer)

	if err != nil {
		fmt.Printf("Error Getting Layers: %s", err)
		return
	}

	for {

		status, err := reader.GetLayerStatus(layerUID)

		if err != nil {
			fmt.Printf("Error Getting Status: %s", err)
			return
		}

		if status == "RUNNING" || status == "LAUNCHING" || status == "READY" {
			fmt.Printf("Layer is %s, waiting 5 seconds\n", status)
			time.Sleep(5 * time.Second)
			continue
		}

		err = reader.RestartRenderStreamLayer(layerUID, *layer)

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Layer restarted")
		time.Sleep(5 * time.Second)
	}

}
