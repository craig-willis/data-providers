package main

import (
	"flag"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/golang/glog"
	providers "github.com/nds-org/data-providers/pkg/providers"
	api "github.com/ndslabs/apiserver/pkg/types"
	"net/http"
)

type Server struct {
}

func main() {
	flag.CommandLine.Parse([]string{})
	flag.Set("logtostderr", "true")

	s := Server{}
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)

	routes := make([]*rest.Route, 0)
	routes = append(routes,
		rest.Post("/mount", s.MountDataset),
	)

	router, err := rest.MakeRouter(routes...)

	if err != nil {
		glog.Fatal(err)
	}
	api.SetApp(router)

	http.Handle("/", api.MakeHandler())

	glog.Infof("Listening on %d", 8083)
	glog.Flush()
	glog.Fatal(http.ListenAndServe(":8083", nil))
}

func (s *Server) MountDataset(w rest.ResponseWriter, r *rest.Request) {

	dataset := api.Dataset{}
	err := r.DecodeJsonPayload(&dataset)
	if err != nil {
		glog.Error(err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	glog.Infof("Mounting dataset from provider %s type=%s url=%s key=%s datapath=%s", dataset.Provider, dataset.Type, dataset.URL, dataset.Key, dataset.LocalPath)

	if dataset.Provider == "clowder" {
		clowder := &providers.ClowderProvider{}
		if dataset.Type == "download" {
			err = clowder.DownloadDataset(&dataset)
		} else if dataset.Type == "symlink" {
			err = clowder.SymlinkDataset(&dataset)
		}
		if err != nil {
			glog.Error(err)
			rest.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}
