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
	SIGNATURE_SHARE_ROOT   string
	SIGNATURE_API_ROOT     string
	ORCHESTRATE_API_KEY    string
)

func CrossDomain() martini.Handler {
	return func(res http.ResponseWriter) {
		res.Header().Add("Access-Control-Allow-Origin", "*")
		res.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	}
}

type Document struct {
	Url string `form:"url" json:"url"`
}

type SignatureElement struct {
	X          string `form:"x" json:"x"`
	Y          string `form:"y" json:"y"`
	Url        string `form:"url" json:"url"`
	PageNumber string `form:"page_number" json:"page_number"`
}

func main() {
	loadEnvs()

	signaturelogic.Setup(ORCHESTRATE_API_KEY)

	m := martini.Classic()
	m.Use(render.Renderer())
	m.Use(CrossDomain())

	m.Any("/api/v0/documents/create.json", binding.Bind(Document{}), DocumentsCreate)
	m.Any("/api/v0/documents/:id.json", DocumentsShow)
	m.Any("/api/v0/signature_elements/create.json", binding.Bind(SignatureElement{}), SignatureElementsCreate)

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
	document["share_url"] = SIGNATURE_SHARE_ROOT + "?document_url=" + SIGNATURE_API_ROOT + "/api/v0/documents/" + document["id"].(string) + ".json"
	documents := []interface{}{}
	documents = append(documents, document)
	payload := map[string]interface{}{"documents": documents}

	return payload
}

func SignatureElementsPayload(signature_element map[string]interface{}) map[string]interface{} {
	signature_elements := []interface{}{}
	signature_elements = append(signature_elements, signature_element)
	payload := map[string]interface{}{"signature_elements": signature_elements}

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

func SignatureElementsCreate(signature_element SignatureElement, req *http.Request, r render.Render) {
	x := signature_element.X
	y := signature_element.Y
	_url := signature_element.Url
	page_number := signature_element.page_number

	params := map[string]interface{}{"x": x, "y": y, "url": _url, "page_number": page_number}
	result, logic_error := signaturelogic.SignatureElementsCreate(params)
	if logic_error != nil {
		payload := ErrorPayload(logic_error)
		statuscode := determineStatusCodeFromLogicError(logic_error)
		r.JSON(statuscode, payload)
	} else {
		payload := SignatureElementsPayload(result)
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
	SIGNATURE_SHARE_ROOT = os.Getenv("SIGNATURE_SHARE_ROOT")
	SIGNATURE_API_ROOT = os.Getenv("SIGNATURE_API_ROOT")
	ORCHESTRATE_API_KEY = os.Getenv("ORCHESTRATE_API_KEY")
}
