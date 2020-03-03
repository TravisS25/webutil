package webutiltest

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

type testServer struct{}

func (s *testServer) fileUploads(w http.ResponseWriter, r *http.Request) {
	var err error

	if err = r.ParseMultipartForm(8 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fileHeaders := r.MultipartForm.File[testConf.TestFileUploadConfs[0].ParamName]

	for _, v := range fileHeaders {
		f, err := v.Open()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if _, err = ioutil.ReadAll(f); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func TestFileUploads(t *testing.T) {
	api := testServer{}
	bodiesURL := "/fileUploads"

	r := mux.NewRouter()
	r.HandleFunc(bodiesURL, api.fileUploads)

	serv := httptest.NewServer(r)
	baseURL := serv.URL
	url := baseURL + bodiesURL
	c := serv.Client()

	req, err := NewFileUploadRequest(
		testConf.TestFileUploadConfs,
		http.MethodPost,
		url,
	)

	if err != nil {
		t.Fatalf(err.Error())
	}

	_, err = c.Do(req)

	if err != nil {
		t.Fatalf(err.Error())
	}
}
