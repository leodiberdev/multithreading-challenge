package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type ViaCep struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
	Error       string `json:"erro"`
}

type ApiCep struct {
	Code       string `json:"code"`
	State      string `json:"state"`
	City       string `json:"city"`
	District   string `json:"district"`
	Address    string `json:"address"`
	Status     int    `json:"status"`
	Ok         bool   `json:"ok"`
	StatusText string `json:"statusText"`
}

func requestViaCEP(cep string) (ViaCep, error) {
	url := "https://viacep.com.br/ws/" + cep + "/json"

	req, err := http.Get(url)
	if err != nil {
		return ViaCep{}, err
	}
	defer req.Body.Close()

	if req.StatusCode == http.StatusBadRequest {
		return ViaCep{}, errors.New("CEP Invalido")
	}

	res, err := io.ReadAll(req.Body)
	if err != nil {
		return ViaCep{}, err
	}

	if req.StatusCode == http.StatusOK {
		var cepData ViaCep
		err = json.Unmarshal(res, &cepData)
		if err != nil {
			return ViaCep{}, err
		}
		return cepData, nil
	}

	return ViaCep{}, errors.New("Erro ao fazer a request na API ViaCEP")
}

func requestApiCEP(cep string) (ApiCep, error) {
	url := "https://cdn.apicep.com/file/apicep/" + cep + ".json"

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ApiCep{}, err
	}

	res, err := client.Do(req)
	if err != nil {
		return ApiCep{}, err
	}
	defer res.Body.Close()

	responseData, err := io.ReadAll(res.Body)
	if err != nil {
		return ApiCep{}, err
	}

	var cepData ApiCep
	err = json.Unmarshal(responseData, &cepData)
	if err != nil {
		return ApiCep{}, err
	}

	if res.StatusCode == http.StatusOK {
		return cepData, nil
	}

	return ApiCep{}, errors.New("Erro ao fazer a request na API APICEP")
}

func main() {
	args := os.Args
	if len(args) > 2 {
		err := errors.New("You must pass only a single CEP")
		log.Println(err)
		return
	}

	if len(args) == 1 {
		err := errors.New("You must provide at least one CEP")
		log.Println(err)
		return
	}

	if !strings.Contains(args[1], "-") {
		err := errors.New("You must enter a CEP in the format XXXXX-XXX")
		log.Println(err)
		return
	}

	cep := args[1]

	ch1 := make(chan ViaCep)
	ch2 := make(chan ApiCep)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	go func() {
		cep := strings.Replace(args[1], "-", "", -1)
		data, err := requestViaCEP(cep)
		if err != nil {
			log.Println(err)
			return
		}
		ch1 <- data
	}()

	go func() {
		data, err := requestApiCEP(cep)
		if err != nil {
			log.Println(err)
			return
		}
		ch2 <- data
	}()

	select {
	case cep := <-ch1:
		fmt.Printf("A Api ViaCEP retornou:\n %+v\n", cep)
	case cep := <-ch2:
		fmt.Printf("A Api ApiCEP retornou:\n %+v\n", cep)
	case <-ctx.Done():
		err := errors.New("Timeout")
		log.Println(err)
	}
}
