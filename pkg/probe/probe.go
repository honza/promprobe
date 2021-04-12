// promprobe
// Copyright (C) 2021  Honza Pokorny <honza@pokorny.ca>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package probe

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

const PROBE_MEMORY = 1
const PROBE_CPU = 2

type Config struct {
	Token      string   `yaml:"token"`
	Host       string   `yaml:"host"`
	Pod        string   `yaml:"pod"`
	Containers []string `yaml:"containers"`
}

type PromResp struct {
	Data PromData `json:"data"`
}

type PromData struct {
	Results []PromResult `json:"result"`
}

type PromResult struct {
	Metric PromMetric  `json:"metric"`
	Value  []PromValue `json:"value"`
}

type PromMetric struct {
	Container string `json:"container"`
}

type PromValue struct {
	Timestamp time.Time
	Value     string
}

type Entry struct {
	Container string
	Value     string
	Timestamp time.Time
}

func Res2Entry(res PromResult) Entry {
	entry := Entry{Container: res.Metric.Container}

	for _, val := range res.Value {
		if val.Value == "" {
			entry.Timestamp = val.Timestamp
		} else {
			entry.Value = val.Value
		}
	}

	return entry
}

func Float2Time(tf float64) time.Time {
	sec, dec := math.Modf(tf)
	return time.Unix(int64(sec), int64(dec*(1e9)))
}

func (pv *PromValue) UnmarshalJSON(b []byte) error {
	var x interface{}
	json.Unmarshal(b, &x)
	switch t := x.(type) {
	case string:
		pv.Value = t
	case float64:
		pv.Timestamp = Float2Time(t)
	default:
		panic("")
	}

	return nil
}

func GetConfig(filename string) (Config, error) {
	var config Config

	contents, err := ioutil.ReadFile(filename)

	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(contents, &config)

	if err != nil {
		return config, err
	}

	return config, nil
}

func BuildUrl(config Config, query string) string {
	q := fmt.Sprintf("%s{pod='%s'}", query, config.Pod)
	return fmt.Sprintf("%s/api/prometheus/api/v1/query?query=%s", config.Host, q)
}

func BuildMemoryQuery(config Config) string {
	return BuildUrl(config, "container_memory_working_set_bytes")
}

func BuildCPUQuery(config Config) string {
	return BuildUrl(config, "container_cpu_usage_seconds_total")
}

func probe(cfgFilename string, mode int) {
	config, err := GetConfig(cfgFilename)
	if err != nil {
		panic(err)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	var url string
	if mode == PROBE_MEMORY {
		url = BuildMemoryQuery(config)
	}
	if mode == PROBE_CPU {
		url = BuildCPUQuery(config)
	}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		panic(err)
	}

	cookie := fmt.Sprintf("openshift-session-token=%s;", config.Token)
	req.Header.Add("Cookie", cookie)
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		fmt.Println(resp.StatusCode)
		fmt.Println(resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	var pr PromResp
	err = json.Unmarshal(body, &pr)

	if err != nil {
		panic(err)
	}

	containerMap := map[string]Entry{}

	for _, res := range pr.Data.Results {
		if res.Metric.Container == "POD" || res.Metric.Container == "" {
			continue
		}
		entry := Res2Entry(res)
		containerMap[res.Metric.Container] = entry
	}

	sort.Strings(config.Containers)

	data := [][]string{}
	var total float64

	for _, container := range config.Containers {
		if entry, ok := containerMap[container]; ok {
			n, err := strconv.ParseFloat(entry.Value, 64)
			mb := n / 1024 / 1024
			if err != nil {
				fmt.Println("conversion failed", entry.Value)
			}

			data = append(data, []string{
				entry.Container,
				entry.Value,
				strconv.FormatFloat(mb, 'f', 2, 64),
			})
			total += n
			continue
		}

		fmt.Println("Missing container:", container)
	}
	data = append(data, []string{
		"Total",
		strconv.FormatFloat(total, 'f', 2, 64),
		strconv.FormatFloat(total/1024/1024, 'f', 2, 64),
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Container", "Value", "MB"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data)
	table.Render()

}

func ProbeMemory(cfgFilename string) {
	probe(cfgFilename, PROBE_MEMORY)
}

func ProbeCPU(cfgFilename string) {
	probe(cfgFilename, PROBE_CPU)
}
