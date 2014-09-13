package main

import (
	"github.com/go-martini/martini"
	"github.com/joho/godotenv"
	"github.com/martini-contrib/render"
	"log"
	"net/http"
	"os"
)

func main() {
	godotenv.Load()

	m := martini.Classic()
	m.Use(martini.Logger())
	m.Use(render.Renderer())

	m.Any("/api/v0/documents/create.json", DocumentsCreate)

	m.Run()
}

func DocumentsCreate(req *http.Request, r render.Render) {
	_url := req.URL.Query().Get("url")

	pages := []string{}
	document := map[string]interface{}{"url": _url, "status": "processing", "pages": pages}
	payload := map[string]interface{}{"success": true, "document": document}

	requestCarve(_url)

	r.JSON(200, payload)
}

func requestCarve(document_url string) {
	carve_root := os.Getenv("CARVE_ROOT")
	webhook_url := "http://requestb.in/1j5etxz1"
	carve_url := carve_root + "/api/v0/documents/create.json?url=" + document_url + "&webhook=" + webhook_url

	log.Println(carve_url)

	res, err := http.Get(carve_url)
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()
}
