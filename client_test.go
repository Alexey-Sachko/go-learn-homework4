package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// код писать тут

type UsersServer struct {
	XMLName xml.Name     `xml:"root"`
	Users   []UserServer `xml:"row"`
}

type UserServer struct {
	XMLName xml.Name `xml:"row"`
	Id      int      `xml:"id"`
	Name    string   `xml:"first_name"`
	Age     int      `xml:"age"`
	About   string   `xml:"about"`
	Gender  string   `xml:"gender"`
}

func paginateUsers(users []UserServer, limit, offset int) []UserServer {
	end := offset + limit
	if end > len(users) {
		end = len(users)
	}
	return users[offset:end]
}

func contains(source, substr string) bool {
	return strings.Contains(strings.ToLower(source), strings.ToLower(substr))
}

func queryFilterUsers(users []UserServer, query string) []UserServer {
	if query == "" {
		return users
	}

	result := make([]UserServer, 0)
	for _, user := range users {
		if contains(user.Name, query) || contains(user.Gender, query) || contains(user.About, query) {
			result = append(result, user)
		}
	}

	return result
}

func FindUsersDummy(w http.ResponseWriter, r *http.Request) {
	var err error

	sReq := SearchRequest{}

	sReq.Limit, err = strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		sReq.Limit = 10
	}

	sReq.Offset, _ = strconv.Atoi(r.FormValue("offset"))
	sReq.Query = r.FormValue("query")
	sReq.OrderField = r.FormValue("order_field")
	sReq.OrderBy, _ = strconv.Atoi(r.FormValue("order_by"))

	if sReq.OrderBy != OrderByAsc && sReq.OrderBy != OrderByDesc {
		sReq.OrderBy = OrderByAsIs
	}

	rawUsers, err := ioutil.ReadFile("dataset.xml")
	if err != nil {
		panic(err.Error())
	}

	// parsing_file
	var fileRoot UsersServer

	err = xml.Unmarshal(rawUsers, &fileRoot)
	if err != nil {
		panic(err.Error())
	}

	users := fileRoot.Users
	orderUsers(&users, sReq.OrderField, sReq.OrderBy)

	resUsers := paginateUsers(queryFilterUsers(users, sReq.Query), sReq.Limit, sReq.Offset)

	resStr, err := json.Marshal(resUsers)
	if err != nil {
		panic(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resStr)
}

func FindUsersDummyErr(status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}
}

func FindUsersDummyBadOrderField(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	errRes := SearchErrorResponse{Error: "ErrorBadOrderField"}
	b, _ := json.Marshal(errRes)
	w.Write(b)
}

func FindUsersDummyInvalidJson(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"hello": "world"`)
}

const correctToken = "CORRECT_TOKEN"

func TestFindUsersLessLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(FindUsersDummy))

	cl := SearchClient{URL: ts.URL, AccessToken: correctToken}

	_, err := cl.FindUsers(SearchRequest{Limit: -1})
	if err == nil {
		t.Error("Limit: < 0, should return error")
	}
}

func TestFindUsersLessOffset(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(FindUsersDummy))

	cl := SearchClient{URL: ts.URL, AccessToken: correctToken}

	_, err := cl.FindUsers(SearchRequest{Offset: -1})
	if err == nil {
		t.Error("Offset < 0, should return error")
	}
}

func TestFindUsers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(FindUsersDummy))

	cl := SearchClient{URL: ts.URL, AccessToken: correctToken}

	res, err := cl.FindUsers(SearchRequest{Limit: 10})
	if err != nil {
		t.Error("Should work without err")
	}

	if res == nil {
		t.Error("Should return response")
	}
}

func TestFindUsersMaxLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(FindUsersDummy))

	cl := SearchClient{URL: ts.URL, AccessToken: correctToken}

	res, _ := cl.FindUsers(SearchRequest{Limit: 26})
	if len(res.Users) > 25 {
		t.Error("Should send Limit less or equal 25")
	}
}

func TestFindUsersStatuses(t *testing.T) {
	statuses := [...]int{
		http.StatusInternalServerError,
		http.StatusUnauthorized,
		http.StatusBadRequest,
	}

	for _, status := range statuses {
		ts := httptest.NewServer(http.HandlerFunc(FindUsersDummyErr(status)))

		cl := SearchClient{URL: ts.URL, AccessToken: correctToken}

		_, err := cl.FindUsers(SearchRequest{Limit: 26})
		if err == nil {
			t.Error("Should return error")
		}
	}
}

func TestFindUsersBadField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(FindUsersDummyBadOrderField))

	cl := SearchClient{URL: ts.URL, AccessToken: correctToken}

	_, err := cl.FindUsers(SearchRequest{Limit: 26})
	if err == nil {
		t.Error("Should return error BadOrderField Error")
	}
}

func TestFindUsersInvalidJson(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(FindUsersDummyInvalidJson))

	cl := SearchClient{URL: ts.URL, AccessToken: correctToken}

	_, err := cl.FindUsers(SearchRequest{Limit: 26})
	if err == nil {
		t.Error("Should return error")
	}
}