package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/patrickmn/go-cache"
	"math"
	"net/http"
	"reflect"
	"strings"
	"time"
)

var memCache = cache.New(10*time.Minute, 1*time.Minute)

const (
	zangRestURL = "https://api.zang.io"
	TimeType    = "time.Time"
)

type SimpleResponse struct {
	XMLName xml.Name `xml:"Response" json:"-"`
	Client  []Client `xml:"Client" json:"Client"`
}

type Response struct {
	XMLName xml.Name `xml:"Response" json:"-"`
	Clients Clients  `xml:"Clients" json:"Clients"`
}

type Clients struct {
	Clients []Client `xml:"Client" json:"Client"`
	*Pagination
}

type Client struct {
	DateUpdated    string `xml:"DateUpdated" json:"DateUpdated`
	PresenceStatus string `xml:"PresenceStatus" json:"PresenceStatus"`
	Nickname       string `xml:"Nickname" json:"Nickname"`
	ClientPassword string `xml:"ClientPassword" json:"ClientPassword"`
	Uri            string `xml:"Uri" json:"Uri"`
	SessionId      string `xml:"SessionId" json:"SessionId"`
	AccountSid     string `xml:"AccountSid" json:"AccountSid"`
	ApplicationSid string `xml:"ApplicationSid" json:"ApplicationSid"`
	ClientSid      string `xml:"Sid" json:"Sid"`
	DateCreated    string `xml:"DateCreated" json:"DateCreated"`
	ApiVersion     string `xml:"ApiVersion" json:"ApiVersion"`
	RemoteIp       string `xml:"RemoteIp" json:"RemoteIp"`
}

type Pagination struct {
	Start           int64  `json:"start" xml:"start,attr"`
	End             int64  `json:"end" xml:"end,attr"`
	Total           int64  `json:"total" xml:"total,attr"`
	Page            int64  `json:"page" xml:"page,attr"`
	PageSize        int64  `json:"page_size" xml:"pagesize,attr"`
	NumPages        int64  `json:"num_pages" xml:"numpages,attr"`
	FirstPageUri    string `json:"first_page_uri" xml:"firstpageuri,attr"`
	LastPageUri     string `json:"last_page_uri" xml:"lastpageuri,attr"`
	NextPageUri     string `json:"next_page_uri" xml:"nextpageuri,attr"`
	PreviousPageUri string `json:"previous_page_uri" xml:"previouspageuri,attr"`
	Uri             string `json:"uri" xml:"uri,attr"`
}

func CreatePagination(req *http.Request, page int64, pageSize int64, totalCount int64) *Pagination {

	p := &Pagination{
		End:             0,
		NextPageUri:     "",
		PreviousPageUri: "",
	}

	if pageSize == 0 {
		pageSize = 50
	}

	var offset, limit, NumOfPages int64
	NumOfPages = 1
	if totalCount > 0 {

		offset, limit, NumOfPages = CalculatePagination(page, pageSize, totalCount)

		p.End = offset + limit

		if page >= NumOfPages || page == (NumOfPages-1) {
			p.End = p.End - 1
		}
	}

	p.PageSize = pageSize
	p.Page = page

	p.Start = page * pageSize

	p.NumPages = NumOfPages
	p.Total = totalCount
	p.FirstPageUri = BuildFirstPageUri(req, page, pageSize)
	p.NextPageUri = BuildNextPageUri(req, page, NumOfPages, pageSize)
	p.LastPageUri = BuildLastPageUri(req, NumOfPages, pageSize)
	p.PreviousPageUri = BuildPreviousPageUri(req, page, pageSize)

	if page == 0 && NumOfPages == 1 {
		p.NextPageUri = ""
		p.PreviousPageUri = ""
	} else if page != 0 && page == (NumOfPages-1) {
		p.NextPageUri = ""
	} else if page == 0 {
		p.PreviousPageUri = ""
	}

	return p
}

/*
	Takes HTTP Reponse writer, Format & empty interface (struct/struct slices)
	converts and write to the desired request formats XML,CSV,JSON
	Default : XML
	Format : xml,csv,json
*/
func HandleResponseEncoding(w http.ResponseWriter, format string, resp_obj interface{}) error {

	//TODO : Interface{} strict type check to serve it as a common function

	if strings.EqualFold(format, "json") {

		err := json.NewEncoder(w).Encode(resp_obj)

		if err != nil {
			return err
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

	} else if strings.EqualFold(format, "csv") {

		b := &bytes.Buffer{}

		var field_header []string
		var field_values []string
		var fieldsArr [][]string

		if reflect.TypeOf(resp_obj).Kind() == reflect.Slice {

			out_elem := reflect.ValueOf(resp_obj)

			for i := 0; i < out_elem.Len(); i++ {

				obj_struct := out_elem.Index(i)
				typ := obj_struct.Type()
				col := obj_struct.NumField()

				for j := 0; j < obj_struct.NumField(); j++ {

					if i < 1 {
						field_header = append(field_header, typ.Field(j).Name)
					}

					var InnerFieldValue string
					if obj_struct.Field(j).Type().String() == TimeType {
						v, _ := obj_struct.Field(j).Interface().(time.Time)
						InnerFieldValue = v.Format(time.ANSIC)
					} else {
						InnerFieldValue = fmt.Sprintf("%v", obj_struct.Field(j))
					}

					field_values = append(field_values, InnerFieldValue)
				}

				fieldsArr = append(fieldsArr, field_values[i*col:(i+1)*col])

			}

			csv_writer := csv.NewWriter(b)
			errH := csv_writer.Write(field_header)

			if errH != nil {
				return errH
			}

			errF := csv_writer.WriteAll(fieldsArr)

			if errF != nil {
				return errF
			}

			csv_writer.Flush()

		} else if reflect.TypeOf(resp_obj).Kind() == reflect.Struct {

			obj_struct := reflect.ValueOf(resp_obj)
			typ := obj_struct.Type()

			for i := 0; i < obj_struct.NumField(); i++ {

				field_header = append(field_header, typ.Field(i).Name)

				var InnerFieldValue string
				if obj_struct.Field(i).Type().String() == TimeType {
					v, _ := obj_struct.Field(i).Interface().(time.Time)
					InnerFieldValue = v.Format(time.ANSIC)
				} else {
					InnerFieldValue = fmt.Sprintf("%v", obj_struct.Field(i))
				}

				field_values = append(field_values, InnerFieldValue)
			}

			csv_writer := csv.NewWriter(b)
			errH := csv_writer.Write(field_header)

			if errH != nil {
				return errH
			}

			errF := csv_writer.Write(field_values)

			if errF != nil {
				return errF
			}

			csv_writer.Flush()
		}

		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(http.StatusOK)
		w.Write(b.Bytes())

	} else {
		x, err := xml.MarshalIndent(resp_obj, "", " ")

		if err != nil {
			return err
		}

		w.Header().Set("Content-Type", "text/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(xml.Header))
		w.Write(x)
	}

	return nil
}

func httpAuth_check(req *http.Request) (string, string, error) {
	accountSid, authToken, ok := req.BasicAuth()

	if ok && len(accountSid) == 34 && len(authToken) == 32 {

		cacheKey := fmt.Sprintf("%v:%v", accountSid, authToken)

		if _, found := memCache.Get(cacheKey); found {
			log.Println("Account Sid found in cache")
			return accountSid, authToken, nil
		} else {

			// Send request to Zang to check basic auth
			log.Println("Sending auth request to Zang REST API")

			req, err := http.NewRequest("GET", zangRestURL, nil)
			if err != nil {
				log.Printf("Error creating auth request to Zang REST API:: %v", err)
				return "", "", err
			}

			req.SetBasicAuth(accountSid, authToken)
			if res, err := http.DefaultClient.Do(req); err != nil || res.StatusCode != 200 {
				log.Printf("Error sending auth request to Zang REST API::%v\t%v", err, res.StatusCode)
				errString := fmt.Errorf("Error sending auth request to Zang REST API::%v\t%v", err, res.StatusCode)
				return "", "", errString
			}

			log.Println("Account authorized by Zang API", accountSid)
			memCache.Set(cacheKey, true, cache.DefaultExpiration)
			return accountSid, authToken, nil
		}
	}

	return "", "", errors.New("Basic Authentication failed")
}

func httpFailedAuth(w http.ResponseWriter) {
	log.Println("Authentication failed")
	w.Header().Set("WWW-Authenticate", `Basic realm="api.zang.io"`)
	http.Error(w, "Unauthorized Access", http.StatusUnauthorized)
}

func RenderEncodingErr(w http.ResponseWriter, format string, err error) {
	log.Errorln("Error while encoding format ", format, " Error -", err.Error())
	http.Error(w, "Internal Server Error "+err.Error(), http.StatusInternalServerError)
}

func RenderFormParsingErr(w http.ResponseWriter, err error) {
	log.Printf("Error parsing the form - %v", err.Error())
	http.Error(w, "Could not parse request ", http.StatusInternalServerError)
}

func RenderServiceAuthErr(w http.ResponseWriter, function string, err error) {
	log.Errorln("Error while doing GRPC Service Auth Operation ", function, " Error -", err.Error())
	http.Error(w, "Internal Server Error "+err.Error(), http.StatusInternalServerError)
}

func RenderReponseErr(w http.ResponseWriter, err error) {
	log.Errorln("Error rendering response ", err.Error())
	http.Error(w, "Internal Server Error ", http.StatusInternalServerError)
}

func EmptyStructCheck(obj interface{}) bool {

	blankStruct := reflect.New(reflect.TypeOf(obj)).Elem().Interface()
	return reflect.DeepEqual(obj, blankStruct)
}

func ReqFormat(e string) string {

	ext := "xml"
	extension := strings.Split(e, ".")

	if len(extension) > 1 {
		ext = extension[1]
	}

	return ext
}

func BuildFirstPageUri(req *http.Request, page int64, pagesize int64) string {

	values := req.URL.Query()
	values.Set("Page", "0")
	values.Set("PageSize", fmt.Sprintf("%d", pagesize))
	req.URL.RawQuery = values.Encode()

	return req.URL.String()
}

func BuildLastPageUri(req *http.Request, numpages int64, pagesize int64) string {
	values := req.URL.Query()

	values.Set("Page", fmt.Sprintf("%d", numpages-1))
	values.Set("PageSize", fmt.Sprintf("%d", pagesize))
	req.URL.RawQuery = values.Encode()

	return req.URL.String()
}

func BuildNextPageUri(req *http.Request, page int64, numpages int64, pagesize int64) string {
	values := req.URL.Query()
	if page+1 > numpages {
		values.Set("Page", fmt.Sprintf("%d", numpages-1))
	} else if page+1 <= numpages {
		values.Set("Page", fmt.Sprintf("%d", page+1))
	}
	values.Set("PageSize", fmt.Sprintf("%d", pagesize))
	req.URL.RawQuery = values.Encode()

	return req.URL.String()
}

func BuildPreviousPageUri(req *http.Request, page int64, pagesize int64) string {
	values := req.URL.Query()
	if page > 0 {
		values.Set("Page", fmt.Sprintf("%d", page-1))
	} else {
		values.Set("Page", fmt.Sprintf("%d", page))
	}
	values.Set("PageSize", fmt.Sprintf("%d", pagesize))
	req.URL.RawQuery = values.Encode()

	return req.URL.String()
}

func CalculatePagination(page int64, pageSize int64, totalRecords int64) (int64, int64, int64) {

	totalPages := int64(math.Ceil(float64(totalRecords) / float64(pageSize)))
	if pageSize >= totalRecords {
		return 0, totalRecords, totalPages
	}

	// For last and all the past pages, return like it's the first page
	if page < 1 {
		return 0, pageSize - 1, totalPages
	}

	// For last and all the future pages, return like it's the last page
	if page >= totalPages || page == totalPages-1 {
		foffset := (pageSize * page) - 1
		return foffset, totalRecords - foffset, totalPages
	}

	return (pageSize * page) - 1, pageSize, totalPages
}
