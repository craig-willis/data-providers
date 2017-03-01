package providers

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	api "github.com/ndslabs/apiserver/pkg/types"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"os/exec"
)

type ClowderProvider struct {
}
type ClowderDataset struct {
	Id          string `json:"id"`
	Created     string `json:"created"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ClowderFile struct {
	Size        string `json:"size"`
	Id          string `json:"id"`
	DateCreated string `json:"date-created"`
	Filepath    string `json:"filepath"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
}

// Symlink all of the files in a dataset. Requires that the filesystem be mounted

func (s *ClowderProvider) getDatasetMetadata(dataset *api.Dataset) (*ClowderDataset, error) {
	// Get the dataset json
	resp, err := http.Get(dataset.URL + "?key=" + dataset.Key)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	if resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			glog.Error(err)
			return nil, err
		}
		fmt.Println(string(body))

		dataset := ClowderDataset{}
		err = json.Unmarshal([]byte(body), &dataset)
		if err != nil {
			glog.Error(err)
			return nil, err
		}
		return &dataset, nil

	} else {
		glog.Error(resp.Status)
		return nil, fmt.Errorf(resp.Status)
	}
}

func (s *ClowderProvider) SymlinkDataset(dataset *api.Dataset) error {
	// https://s6vnik-clowder.workbench.nationaldataservice.org/api/datasets/58b06e73e4b008f0ce825189/listFiles
	/*
	   [
	       {
	           "size":"31120",
	           "date-created":"Fri Feb 24 17:48:18 UTC 2017",
	           "id":"58b071e2e4b04f6dc15da646",
	           "filepath":"/clowder_data/uploads/49/a6/5d/58b071e2e4b04f6dc1",
	           "contentType":"image/jpeg",
	           "filename":"Hs-2009-25-e-full_jpg.jpg"
	       }
	   ]
	*/

	glog.Infof("Symlinking dataset from URL %s to path %s\n", dataset.URL, dataset.LocalPath)
	ds, err := s.getDatasetMetadata(dataset)
	if err != nil {
		glog.Error(err)
		return err
	}

	datasetPath := dataset.LocalPath + "/" + ds.Name
	if _, err := os.Stat(datasetPath); os.IsNotExist(err) {
		os.MkdirAll(datasetPath, 0777)
	}

	files, err := s.getFiles(dataset)
	for _, file := range *files {

		filePath := dataset.LocalPath + "/" + ds.Name + "/" + file.Filename

		err = os.Symlink(file.Filepath, filePath)
		if err != nil {
			glog.Error(err)
			return err
		}
	}

	return nil
}

func (s *ClowderProvider) getFiles(dataset *api.Dataset) (*[]ClowderFile, error) {

	// Get the file json
	resp, err := http.Get(dataset.URL + "/listFiles?key=" + dataset.Key)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	if resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			glog.Error(err)
			return nil, err
		}
		fmt.Println(string(body))

		files := make([]ClowderFile, 0)
		err = json.Unmarshal([]byte(body), &files)
		if err != nil {
			glog.Error(err)
			return nil, err
		}
		return &files, nil

	} else {
		glog.Error(resp.Status)
		return nil, fmt.Errorf(resp.Status)
	}
}

// Download a dataset using the Clowder API
func (s *ClowderProvider) DownloadDataset(dataset *api.Dataset) error {
	//https://s6vnik-clowder.workbench.nationaldataservice.org/api/datasets/58b06e73e4b008f0ce825189/download?key=r1ek3rs

	glog.Infof("Downloading file from URL %s to path %s\n", dataset.URL, dataset.LocalPath)

	ds, err := s.getDatasetMetadata(dataset)
	if err != nil {
		glog.Error(err)
		return err
	}

	datasetPath := dataset.LocalPath + "/" + ds.Name
	if _, err := os.Stat(datasetPath); os.IsNotExist(err) {
		err = os.Mkdir(datasetPath, 0777)
		if err != nil {
			glog.Error(err)
			return err
		}
	}

	// Download the file
	resp, err := http.Get(dataset.URL + "/download?key=" + dataset.Key)
	if err != nil {
		glog.Error(err)
		return err
	}
	defer resp.Body.Close()

	_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
	if err != nil {
		glog.Error(err)
		return err
	}
	filename := params["filename"]
	glog.Infof("filename %s", filename)

	zipPath := dataset.LocalPath + "/" + filename

	// Create the file in the default path for the service
	out, err := os.Create(zipPath)
	if err != nil {
		glog.Error(err)
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		glog.Error(err)
		return err
	}

	glog.Infof("Unzipping dataset %s to %s", zipPath, dataset.LocalPath+"/"+filename[0:len(filename)-4])
	cmd := exec.Command("unzip", zipPath, "-d", dataset.LocalPath+"/"+filename[0:len(filename)-4])
	if err != nil {
		glog.Error(err)
		return err
	}
	err = cmd.Start()
	if err != nil {
		glog.Error(err)
		return err
	}
	cmd.Wait()
	err = os.Remove(zipPath)
	if err != nil {
		glog.Error(err)
		return err
	}

	return nil
}
