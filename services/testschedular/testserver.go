package main

import (
	"../../lib/libocit"
	"../../lib/routes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type TSResource struct {
	libocit.Resource

	//including the port
	ID          string
	URL         string
	MaxJobs     int
	TestUnitIDs []string
}

func (res *TSResource) Valid() error {
	if res.Class == "" {
		return errors.New("'Class' required.")
	} else if res.Distribution == "" {
		return errors.New("'Distribution' required.")
	} else if res.Version == "" {
		return errors.New("'Version' required.")
	} else if res.Arch == "" {
		return errors.New("'Arch' required.")
	} else if res.URL == "" {
		return errors.New("'URL' required.")
	}
	return nil
}

const TestSchedularCacheDir = "/tmp/.test_schedular_cache"

type TestServerConfig struct {
	Port           int
	ServerListFile string
	CacheDir       string
	Debug          bool
}

func GetResult(w http.ResponseWriter, r *http.Request) {
	//TODO
	id := r.URL.Query().Get(":ID")
	fmt.Println(id)
}

//List all the hostOS status
func GetStatus(w http.ResponseWriter, r *http.Request) {
}

func ReceiveTask(w http.ResponseWriter, r *http.Request) {
	realURL, params := libocit.ReceiveFile(w, r, TestSchedularCacheDir)
	taskID := params["id"]

	//TODO
	//Untar the file and load the case.json
	content := ""
	if pub_config.Debug {
		fmt.Println(content)
	}
	var ret libocit.HttpRet
	var tc libocit.TestCase
	if err := json.Unmarshal([]byte(content), &tc); err != nil {
		ret.Status = "Failed"
		ret.Message = "The testcase is not standard. (.tar.gz or .json)"
		ret_string, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(ret_string))
		return
	} else {
		ret.Status = "OK"
		ret.Message = "success in receiving task files"
		ret_string, _ := json.MarshalIndent(ret, "", "\t")
		w.Write([]byte(ret_string))
	}

	tc.SetBundleURL(realURL)
	task := TSTaskNew(taskID, tc)

	if !task.Run(libocit.TestActionApply) {
		return
	}
	if !task.Run(libocit.TestActionDeploy) {
		return
	}
	if !task.Run(libocit.TestActionRun) {
		return
	}
	if !task.Run(libocit.TestActionCollect) {
		return
	}
	if !task.Run(libocit.TestActionDestroy) {
		return
	}
}

func GetResourceQuery(r *http.Request) (res TSResource) {
	res.Class = libocit.TUClass(r.URL.Query().Get("Class"))
	res.Distribution = r.URL.Query().Get("Distribution")
	res.Version = r.URL.Query().Get("Version")
	res.Arch = r.URL.Query().Get("Arch")

	var cpu string
	cpu = r.URL.Query().Get("CPU")
	if len(cpu) > 0 {
		cpu_count, cpu_err := strconv.ParseInt(cpu, 10, 64)
		if cpu_err != nil {
			//TODO, should report the err
		} else {
			res.CPU = cpu_count
		}
	} else {
		res.CPU = 0
	}

	var memory string
	memory = r.URL.Query().Get("Memory")
	if len(memory) > 0 {
		memory_count, memory_err := strconv.ParseInt(cpu, 10, 64)
		if memory_err != nil {
			//TODO, should report the err
		} else {
			res.Memory = memory_count
		}
	} else {
		res.Memory = 0
	}

	log.Println(res)
	return res
}

func GetResourceList(resQuery TSResource) (ids []string) {
	for _, res := range ResourceStore {
		if len(resQuery.Class) > 1 {
			if resQuery.Class != res.Class {
				continue
			}
		}
		if len(resQuery.Distribution) > 1 {
			if resQuery.Distribution != res.Distribution {
				continue
			}
		}
		if len(resQuery.Version) > 1 {
			if resQuery.Version != res.Version {
				continue
			}
		}
		if len(resQuery.Arch) > 1 {
			if resQuery.Arch != res.Arch {
				continue
			}
		}
		if resQuery.CPU > res.CPU {
			log.Println("not enough CPU")
			continue
		}
		if resQuery.Memory > res.Memory {
			log.Println("not enough Memory")
			continue
		}
		ids = append(ids, res.ID)
	}
	return ids
}

func GetResource(w http.ResponseWriter, r *http.Request) {
	resQuery := GetResourceQuery(r)

	ids := GetResourceList(resQuery)

	var ret libocit.HttpRet
	if len(ids) < 1 {
		ret.Status = libocit.RetStatusFailed
		ret.Message = "Cannot find the avaliable resource"
	} else {
		ret.Status = libocit.RetStatusOK
		ret.Message = "Find the avaliable resource"
		var rss []TSResource
		for index := 0; index < len(ids); index++ {
			rss = append(rss, ResourceStore[ids[index]])
		}

		ret.Data = rss
	}

	body, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(body))
}

func PostResource(w http.ResponseWriter, r *http.Request) {
	var res TSResource
	var ret libocit.HttpRet

	result, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if pub_config.Debug {
		fmt.Println(string(result))
	}
	json.Unmarshal([]byte(result), &res)
	if err := res.Valid(); err != nil {
		ret.Status = libocit.RetStatusFailed
		ret.Message = err.Error()
	} else {
		lock.Lock()
		id := libocit.MD5(string(result))
		if _, ok := ResourceStore[id]; ok {
			ret.Status = libocit.RetStatusFailed
			ret.Message = "this resource is already exist"
		} else {
			res.ID = id
			ResourceStore[id] = res
			ret.Status = "OK"
			ret.Message = "Success in adding the resource"
		}
		lock.Unlock()
	}
	ret_body, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(ret_body))
}

func DeleteResource(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	id := r.URL.Query().Get("ID")
	lock.Lock()
	if _, ok := ResourceStore[id]; ok {
		delete(ResourceStore, id)
		ret.Status = libocit.RetStatusOK
		ret.Message = "Success in remove the resource"
	} else {
		ret.Status = libocit.RetStatusFailed
		ret.Message = "Cannot find the resource"
	}
	lock.Unlock()
	ret_body, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(ret_body))
}

// Will use DB in the future, (mongodb for example)
func init() {
}

var ResourceStore map[string]TSResource

var lock = sync.RWMutex{}
var pub_config TestServerConfig

func main() {
	config_content := libocit.ReadFile("./testserver.conf")
	json.Unmarshal([]byte(config_content), &pub_config)

	mux := routes.New()

	mux.Get("/resource", GetResource)
	mux.Post("/resource", PostResource)
	mux.Get("/:ID/status", GetStatus)
	mux.Post("/task", ReceiveTask)
	mux.Get("/:ID/result", GetResult)

	http.Handle("/", mux)
	port := fmt.Sprintf(":%d", pub_config.Port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}