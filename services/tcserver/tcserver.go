package main

import (
	"../../lib/libocit"
	"../../lib/routes"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
)

type TCServerConf struct {
	Repos    []CaseRepo
	CacheDir string
	Port     int
	Debug    bool
}

type CaseRepo struct {
	ID         string
	Name       string
	CaseFolder string
	//We can disable a repo
	Enable bool
	Groups []string
}

type Case struct {
	ID       string
	RepoID   string
	Name     string
	GroupDir string
	Status   string
	//0 means not tested
	TestedTime       int64
	LastModifiedTime int64
}

var pub_config TCServerConf
var repoStore = map[string]*CaseRepo{}
var caseStore = map[string]*Case{}

func RefreshRepo(repoID string) {
	var cmd string
	repo := *repoStore[repoID]
	repoDir := libocit.PreparePath(pub_config.CacheDir, repo.Name)

	git_check_url := path.Join(repoDir, ".git/config")
	_, err := os.Stat(git_check_url)
	if err != nil {
		//Clean: remove the '/$' if there was one
		cmd = "cd " + path.Dir(path.Clean(repoDir)) + " ; git clone https://" + repo.Name + ".git"
	} else {
		cmd = "cd " + repoDir + " ; git pull"
	}

	if pub_config.Debug == true {
		fmt.Println("Refresh by using ", cmd)
	}
	c := exec.Command("/bin/sh", "-c", cmd)
	c.Run()
	if pub_config.Debug == true {
		fmt.Println("Refresh done")
	}
}

func LastModified(case_dir string) (last_modified int64) {
	last_modified = 0
	files, _ := ioutil.ReadDir(case_dir)
	for _, file := range files {
		if file.IsDir() {
			sub_lm := LastModified(path.Join(case_dir, file.Name()))
			if last_modified < sub_lm {
				last_modified = sub_lm
			}
		} else {
			if last_modified < file.ModTime().Unix() {
				last_modified = file.ModTime().Unix()
			}
		}
	}
	return last_modified
}

func LoadCase(repoID string, groupDir string, caseName string) {
	caseDir := path.Join(groupDir, caseName)
	_, err_msgs := libocit.ValidateByDir(caseDir, "")
	if len(err_msgs) == 0 {
		last_modified := LastModified(caseDir)
		caseMD := libocit.MD5(caseDir)
		if v, ok := caseStore[caseMD]; ok {
			//Happen when we refresh the repo
			(*v).LastModifiedTime = last_modified
			fi, err := os.Stat(path.Join(caseDir, "report.md"))
			if err != nil {
				(*v).TestedTime = 0
			} else {
				(*v).TestedTime = fi.ModTime().Unix()
			}
			if (*v).LastModifiedTime > (*v).TestedTime {
				(*v).Status = "idle"
			} else {
				(*v).Status = "tested"
			}
		} else {
			var tc Case
			tc.ID = caseMD
			tc.RepoID = repoID
			tc.Name = caseName
			tc.GroupDir = groupDir
			fi, err := os.Stat(path.Join(caseDir, "report.md"))
			if err != nil {
				tc.TestedTime = 0
			} else {
				tc.TestedTime = fi.ModTime().Unix()
			}
			tc.LastModifiedTime = last_modified
			if tc.LastModifiedTime > tc.TestedTime {
				tc.Status = "idle"
			} else {
				tc.Status = "tested"
			}
			caseStore[caseMD] = &tc
		}
	} else {
		fmt.Println("Error in loading case: ", caseDir, " . Skip it")
		return
	}
}

func LoadRepos() {
	for id, repo := range repoStore {
		if repo.Enable == true {
			RefreshRepo(id)
			LoadCasesFromRepo(id)
		}
	}

}

func LoadCasesFromRepo(repoID string) {
	repo := *repoStore[repoID]
	for index := 0; index < len(repo.Groups); index++ {
		groupDir := path.Join(pub_config.CacheDir, repo.Name, repo.Groups[index])
		files, _ := ioutil.ReadDir(groupDir)
		for _, file := range files {
			if file.IsDir() {
				LoadCase(repoID, groupDir, file.Name())
			}
		}
	}
}

func ListRepos(w http.ResponseWriter, r *http.Request) {
	page_string := r.URL.Query().Get("Page")
	page, err := strconv.Atoi(page_string)
	if err != nil {
		page = 0
	}
	pageSizeString := r.URL.Query().Get("PageSize")
	pageSize, err := strconv.Atoi(pageSizeString)
	if err != nil {
		pageSize = 10
	}

	var repoList []CaseRepo
	cur_num := 0
	for _, repo := range repoStore {
		cur_num += 1
		if (cur_num >= page*pageSize) && (cur_num < (page+1)*pageSize) {
			repoList = append(repoList, *repo)
		}

	}

	var ret libocit.HttpRet
	ret.Status = "OK"
	ret.Message = fmt.Sprintf("%d repos founded", len(repoList))
	ret.Data = repoList
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func GetRepo(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	repoID := r.URL.Query().Get("ID")
	repo, ok := repoStore[repoID]

	if ok {
		ret.Status = "OK"
		ret.Data = *repo
	} else {
		ret.Status = "Failed"
		ret.Message = "Cannot find the repo, wrong ID provided"
	}

	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func AddRepo(w http.ResponseWriter, r *http.Request) {
}

func ModifyRepo(w http.ResponseWriter, r *http.Request) {
}

func ListCases(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("Status")
	page_string := r.URL.Query().Get("Page")
	page, err := strconv.Atoi(page_string)
	if err != nil {
		page = 0
	}
	pageSizeString := r.URL.Query().Get("PageSize")
	pageSize, err := strconv.Atoi(pageSizeString)
	if err != nil {
		pageSize = 10
	}

	var caseList []Case
	cur_num := 0
	for _, tc := range caseStore {
		if status != "" {
			if status != tc.Status {
				continue
			}
		}
		cur_num += 1
		if (cur_num >= page*pageSize) && (cur_num < (page+1)*pageSize) {
			caseList = append(caseList, *tc)
		}

	}

	var ret libocit.HttpRet
	if err != nil {
		ret.Status = "Failed"
	} else {
		ret.Status = "OK"
		ret.Message = fmt.Sprintf("%d cases founded", len(caseList))
		ret.Data = caseList
	}

	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func GetCase(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get(":ID")
	tc := caseStore[id]
	files := libocit.GetDirFiles(tc.GroupDir, tc.Name)
	tarUrl := libocit.TarFileList(files, tc.GroupDir, tc.Name)

	file, err := os.Open(tarUrl)
	defer file.Close()
	if err != nil {
		//FIXME: add to head
		w.Write([]byte("Cannot open the file: " + tarUrl))
		return
	}

	buf := bytes.NewBufferString("")
	buf.ReadFrom(file)
	//TODO: write head, filename and the etc
	w.Write([]byte(buf.String()))
}

func GetCaseReport(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	id := r.URL.Query().Get(":ID")
	tc := caseStore[id]
	repo := repoStore[tc.RepoID]
	reportUrl := path.Join(pub_config.CacheDir, repo.Name, repo.CaseFolder, tc.GroupDir, tc.Name, "report.md")

	_, err := os.Stat(reportUrl)
	if err != nil {
		ret.Status = "Failed"
		ret.Message = "Cannot find the report"
	} else {
		ret.Status = "OK"
		ret.Data = libocit.ReadFile(reportUrl)
	}

	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func RefreshRepos(w http.ResponseWriter, r *http.Request) {
	var ret libocit.HttpRet
	LoadRepos()

	ret.Status = "OK"
	retInfo, _ := json.MarshalIndent(ret, "", "\t")
	w.Write([]byte(retInfo))
}

func init() {
	fmt.Println("init")
	content := libocit.ReadFile("./tcserver.conf")
	json.Unmarshal([]byte(content), &pub_config)
	repos := pub_config.Repos
	for index := 0; index < len(repos); index++ {
		repoMD := libocit.MD5(repos[index].Name)
		repos[index].ID = repoMD
		repoStore[repoMD] = &repos[index]

	}
	LoadRepos()
}

func main() {
	port := fmt.Sprintf(":%d", pub_config.Port)
	mux := routes.New()
	mux.Post("/repos", RefreshRepos)
	mux.Get("/repo", ListRepos)
	mux.Get("/repo/:ID", GetRepo)
	mux.Post("/repo", AddRepo)
	//Modify, especially enable/disable
	mux.Post("/repo/:ID", ModifyRepo)

	mux.Get("/case", ListCases)
	mux.Get("/case/:ID", GetCase)
	mux.Get("/case/:ID/report", GetCaseReport)
	http.Handle("/", mux)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	} else {
		fmt.Println("Listen to port ", port)
	}
}
