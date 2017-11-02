package beater

import (
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"time"

	"github.com/elastic/beats/libbeat/common"
	"github.com/gothfrid/jolokiabeat/config"
	resty "gopkg.in/resty.v1"
)

// FetchData fetches domains
func (jb *Jolokiabeat) FetchData(s *config.Source) (err error) {

	requestBody, err := buildDomainList(s) // Prepare domain list to fetch
	if err != nil {
		return err
	} //Abort if Error occures

	// TODO - A DD MBEANS FROM CONFIG

	resp, err := resty.R().
		SetHeaders(s.Headers).
		SetBody(requestBody).
		Post(s.EndPoint)
	if err != nil {
		return err
	}

	var jsonValue []map[string]interface{}
	json.Unmarshal(resp.Body(), &jsonValue)
	for _, v := range jsonValue {
		if v["status"] == 200.0 {
			var events []common.MapStr
			values := reflect.ValueOf(v["value"]).
				Interface().(map[string]interface{})
			for mb, attr := range values {
				bean := common.MapStr{
					"CannonicalName": mb,
					"Attributes":     attr,
				}
				mBeanParser(&bean)
				events = append(events,
					common.MapStr{
						"@timestamp": common.Time(time.Now()),
						"bean":       bean,
						"host":       s.Host,
					})
			}
			jb.events <- events
		} // prepare list of Events
	}
	return err
}

func buildDomainList(s *config.Source) (list []common.MapStr, err error) {

	domains := map[string]int{}

	for _, val := range s.Domains {
		d, i := domainParser(val)
		domains[d] = i
	} // Get known domains from config

	if !s.FetchOnly {
		resp, err := resty.R().
			SetHeaders(s.Headers).
			SetBody(buildRequestBody("", 0)).
			Post(s.EndPoint)
		if err != nil {
			return nil, err
		}
		val, err := getResponseValue(resp.Body())
		if err != nil {
			return nil, err
		}
		for key := range val {
			if _, ok := domains[key]; !ok {
				domains[key] = -1
			}
		}
	} // Get all existing domains from endpoint

	for d, v := range domains {
		if v != 0 {
			list = append(list, buildRequestBody(d, v))
		}
	} //create request body for domains list

	return list, nil
}

//  build POST request body according to jolokia api
func buildRequestBody(mb string, md int) common.MapStr {
	var body = common.MapStr{
		"config": common.MapStr{
			"maxCollectionSize": 0,
			"ignoreErrors":      true,
		},
	}
	if mb == "" {
		body.Put("type", "list")
		conf, _ := body.GetValue("config")
		conf.(common.MapStr).Put("maxDepth", 1)
	} else {
		body.Put("type", "read")
		body.Put("mbean", mb+":*")
		if md > 0 {
			conf, _ := body.GetValue("config")
			conf.(common.MapStr).Put("maxDepth", md)
		}
	}
	return body
}

func getResponseValue(response []byte) (value map[string]interface{}, err error) {
	var jsonValue map[string]interface{}
	json.Unmarshal(response, &jsonValue)
	valueMap := jsonValue["value"]
	result, ok := valueMap.(map[string]interface{})
	if !ok {
		return nil, errors.New("parsing error")
	}
	return result, nil
}

func domainParser(s string) (mb string, md int) {
	p := regexp.MustCompile(`(.*)::(\d*)$`)
	parsed := p.FindAllStringSubmatch(s, -1)
	if len(parsed) > 0 {
		md, e := strconv.Atoi(parsed[0][2])
		if e == nil {
			return parsed[0][1], md
		}
	}
	return s, -1
}

func mBeanParser(bean *common.MapStr) {
	p := regexp.MustCompile(`([^,=:\\*\\?]+)=(\"(?:
							 [^\\\\\"]|\\\\\\\\|\\\
							 \n|\\\\\"|\\\\\\?|\\\\
							 \\*)*\"|[^,=:\"]*)`)
	cn, _ := bean.GetValue("CanonicalName")
	parsed := p.FindAllStringSubmatch(reflect.ValueOf(cn).
		Interface().(string), -1)
	for _, param := range parsed {
		switch param[1] {
		case "app":
			bean.Put("app", param[2])
		case "type":
			bean.Put("type", param[2])
		case "name":
			bean.Put("name", param[2])
		}
	}
	p = regexp.MustCompile(`(.*):(.*)`)
	parsed = p.FindAllStringSubmatch(reflect.ValueOf(cn).
		Interface().(string), -1)
	bean.Put("domain", parsed[0][1])
}
