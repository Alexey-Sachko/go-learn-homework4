package main

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	// "io"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strconv"
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

type SearchResponseUnion interface {
	isSuccess()
}

type TestCase struct {
	ID     string
	Result *SearchResponseUnion
	// IsError bool
}

func paginateUsers(users []UserServer, limit, offset int) []UserServer {
	return users[offset : offset+limit]
}

func getValueByField(user UserServer, field string) interface{} {
	val := reflect.ValueOf(user)

	var found interface{} = nil

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		if typeField.Name == field {
			found = valueField.Interface()
		}

		// fmt.Printf("\tname=%v, type=%v, value=%v sortField=%v\n",
		// 	typeField.Name,
		// 	typeField.Type.Kind(),
		// 	valueField,
		// field)
	}

	return found
}

func compareStrLess(left, right interface{}) (less bool, ok bool) {
	leftStr, ok := left.(string)
	if !ok {
		return false, false
	}

	rightStr, ok := right.(string)
	if !ok {
		return false, false
	}

	return leftStr > rightStr, true
}

func compareNumLess(left, right interface{}) (less bool, ok bool) {
	leftStr, ok := left.(int)
	if !ok {
		return false, false
	}

	rightStr, ok := right.(int)
	if !ok {
		return false, false
	}

	return leftStr > rightStr, true
}

func compareLess(left, right interface{}) bool {
	less, ok := compareStrLess(left, right)
	if ok {
		return less
	}

	less, ok = compareNumLess(left, right)
	if ok {
		return less
	}

	fmt.Printf("\ntypes: %T, %T\n", left, right)
	panic("UNHANDLED TYPE")
}

func orderUsers(users *[]UserServer, field string, by int) {
	sort.SliceStable(*users, func(i, j int) bool {
		left := getValueByField((*users)[i], field)
		right := getValueByField((*users)[j], field)

		if left == nil || right == nil {
			return false
		}

		result := compareLess(left, right)
		if by == 1 {
			return !result
		}

		return result
	})
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

	if sReq.OrderBy < -1 || sReq.OrderBy > 1 {
		sReq.OrderBy = 0
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

	if sReq.OrderField != "" && sReq.OrderBy != 0 {
		orderUsers(&users, sReq.OrderField, sReq.OrderBy)
	}

	resUsers := paginateUsers(users, sReq.Limit, sReq.Offset)

	resStr, err := json.Marshal(resUsers)
	if err != nil {
		panic(err.Error())
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resStr)

	// switch key {
	// case "42":
	// 	w.WriteHeader(http.StatusOK)
	// 	io.WriteString(w, `{"status": 200, "balance": 100500}`)
	// case "100500":
	// 	w.WriteHeader(http.StatusOK)
	// 	io.WriteString(w, `{"status": 400, "err": "bad_balance"}`)
	// case "__broken_json":
	// 	w.WriteHeader(http.StatusOK)
	// 	io.WriteString(w, `{"status": 400`) //broken json
	// case "__internal_error":
	// 	fallthrough
	// default:
	// 	w.WriteHeader(http.StatusInternalServerError)
	// }
}

func TestFindUsers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(FindUsersDummy))

	cl := SearchClient{URL: ts.URL, AccessToken: "STUB"}

	res, _ := cl.FindUsers(SearchRequest{Limit: 10, OrderField: "Id", OrderBy: -1})
	for _, user := range res.Users {
		fmt.Println("user: ", user.Id)
	}
}
