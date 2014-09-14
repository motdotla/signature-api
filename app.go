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
	CARVE_ROOT             string
	SIGNATURE_CATCHER_ROOT string
	ORCHESTRATE_API_KEY    string
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
	m.Use(render.Renderer())
	m.Use(CrossDomain())

	m.Any("/api/v0/documents/create.json", binding.Bind(Document{}), DocumentsCreate)
	m.Any("/api/v0/documents/:id.json", DocumentsShow)

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
		postscript := result["id"].(string)
		requestCarve(_url, postscript)

		payload := DocumentsPayload(result)
		r.JSON(200, payload)
	}
}

func DocumentsShow(params martini.Params, req *http.Request, r render.Render) {
	id := params["id"]
	result, logic_error := signaturelogic.DocumentsShow(id)
	if logic_error != nil {
		payload := ErrorPayload(logic_error)
		statuscode := determineStatusCodeFromLogicError(logic_error)
		r.JSON(statuscode, payload)
	} else {
		payload := DocumentsPayload(result)
		r.JSON(200, payload)
	}
}

func requestCarve(document_url string, postscript string) {
	webhook_url := SIGNATURE_CATCHER_ROOT + "/webhook/v0/documents/processed.json"
	carve_url := CARVE_ROOT + "/api/v0/documents/create.json?url=" + document_url + "&webhook=" + webhook_url + "&postscript=" + postscript

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
	SIGNATURE_CATCHER_ROOT = os.Getenv("SIGNATURE_CATCHER_ROOT")
	ORCHESTRATE_API_KEY = os.Getenv("ORCHESTRATE_API_KEY")
}
