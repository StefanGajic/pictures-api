package main

import (
	"log"
	"net/http"

	"github.com/StefanGajic/pictures-api/handler"
	"github.com/gorilla/mux"
)

func setupRoutes() {
	r := mux.NewRouter()

	handleFunc := handler.NewErrorHandler().Wrap

	r.Handle("/upload", handleFunc(handler.UploadFile))
	r.Handle("/list", handleFunc(handler.ListFiles))
	r.Handle("/delete", handleFunc(handler.DeleteFile))
	r.Handle("/download", handleFunc(handler.DownloadFile))

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal("Fail to start server", err)
	}

}

func main() {
	log.Print("server is starting at port 8080")
	setupRoutes()
}
