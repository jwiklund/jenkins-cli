package jenkins

import (
	"encoding/json"
	"encoding/xml"
	"io"
)

type configXmlItem struct {
	XMLName xml.Name
	Value   string `xml:",innerxml"`
}

type configXml struct {
	Items []configXmlItem `xml:",any"`
}

type jobsJson struct {
	Jobs []Job `json:"jobs"`
}

func parseConfig(config io.Reader) (map[string]string, error) {
	var decoder = xml.NewDecoder(config)
	var cfg configXml
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, item := range cfg.Items {
		result[item.XMLName.Local] = item.Value
	}
	return result, nil
}

func parseJobs(jobs io.Reader) ([]Job, error) {
	var decoder = json.NewDecoder(jobs)
	var jobsjson jobsJson
	if err := decoder.Decode(&jobsjson); err != nil {
		return nil, err
	}
	return jobsjson.Jobs, nil
}
