package main

import (
	"github.com/go-martini/martini"
	"github.com/handshakejs/handshakejserrors"
	"github.com/joho/godotenv"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/motdotla/signaturelogic"
	"log"
	"net/http"
	"os"
)

const (
	LOGIC_ERROR_CODE_UNKNOWN = "unknown"
)

var (
	CARVE_ROOT          string
	ORCHESTRATE_API_KEY string
)

func CrossDomain() martini.Handler {
	return func(res http.ResponseWriter) {
		res.Header().Add("Access-Control-Allow-Origin", "*")
	}
}

type Document struct {
	Url string `form:"url" json:"url"`
}

func main() {
	loadEnvs()

	signaturelogic.Setup(ORCHESTRATE_API_KEY)

	m := martini.Classic()
	m.Use(martini.Logger())
	m.Use(render.Renderer())

	m.Any("/api/v0/documents/create.json", binding.Bind(Document{}), DocumentsCreate)

	m.Run()
}

func ErrorPayload(logic_error *handshakejserrors.LogicError) map[string]interface{} {
	error_object := map[string]interface{}{"code": logic_error.Code, "field": logic_error.Field, "message": logic_error.Message}
	errors := []interface{}{}
	errors = append(errors, error_object)
	payload := map[string]interface{}{"errors": errors}

	return payload
}

func DocumentsPayload(document map[string]interface{}) map[string]interface{} {
	documents := []interface{}{}
	documents = append(documents, document)
	payload := map[string]interface{}{"documents": documents}

	return payload
}

func DocumentsCreate(document Document, req *http.Request, r render.Render) {
	_url := document.Url

	params := map[string]interface{}{"url": _url}
	result, logic_error := signaturelogic.DocumentsCreate(params)
	if logic_error != nil {
		payload := ErrorPayload(logic_error)
		statuscode := determineStatusCodeFromLogicError(logic_error)
		r.JSON(statuscode, payload)
	} else {
		requestCarve(_url)

		payload := DocumentsPayload(result)
		r.JSON(200, payload)
	}
}

func requestCarve(document_url string) {
	webhook_url := "http://requestb.in/1j5etxz1"
	carve_url := CARVE_ROOT + "/api/v0/documents/create.json?url=" + document_url + "&webhook=" + webhook_url

	log.Println(carve_url)

	res, err := http.Get(carve_url)
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()
}

func determineStatusCodeFromLogicError(logic_error *handshakejserrors.LogicError) int {
	code := 400
	if logic_error.Code == LOGIC_ERROR_CODE_UNKNOWN {
		code = 500
	}

	return code
}

func loadEnvs() {
	godotenv.Load()

	CARVE_ROOT = os.Getenv("CARVE_ROOT")
	ORCHESTRATE_API_KEY = os.Getenv("ORCHESTRATE_API_KEY")
}
