package main

import (
	"flag"
	"github.com/sthagen/nosurf"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	Port                        = 8088
	BaseUrl                     = "/go/quiz/"
	TemplatesRoot               = "./ui/html/"
	AvailableModules            = 4
	AvailableQuestionsPerModule = 15
	ProgressMax                 = AvailableModules * AvailableQuestionsPerModule
)

type server struct {
	modules       []Module
	port          string
	topic         string
	baseUrl       string
	errorLog      *log.Logger
	infoLog       *log.Logger
	templateCache map[string]*template.Template
}

type Module struct {
	Meta Meta   `json:"meta"`
	Quiz []Quiz `json:"quiz"`
}
type Meta struct {
	Module        int    `json:"module"`
	Kind          string `json:"kind"`
	QuestionCount int    `json:"question_count"`
	PointsTotal   int    `json:"points_total"`
	PercentTotal  int    `json:"percent_total"`
	Weights       []int  `json:"weights"`
}
type Alternatives struct {
	Alternative string `json:"alternative"`
	Evaluation  string `json:"evaluation"`
	Hint        string `json:"hint"`
}
type Quiz struct {
	Task         string         `json:"task"`
	Claim        string         `json:"claim"`
	AnswerKind   string         `json:"answer_kind"`
	Explanation  Explanation    `json:"explanation"`
	Alternatives []Alternatives `json:"alternatives"`
}
type Explanation struct {
	Summary string     `json:"summary"`
	Details [][]string `json:"details"`
}

type RenderQuiz struct {
	ModuleId     int
	QuizNumber   int
	Progress     int
	Task         string
	TaskId       int
	Claim        string
	AnswerKind   string
	Alternative1 string
	Alternative2 string
	Alternative3 string
	Alternative4 string
	Alternative5 string
	BaseUrl      string
	Token        string
}

type RenderEvaluation struct {
	ModuleId          int
	QuizNumber        int
	Progress          int
	Task              string
	TaskId            int
	Claim             string
	AlternativeChosen string
	Evaluation        string
	Ok                string
	Hint              string
	Summary           string
	Details           []template.HTML
	BaseUrl           string
}

func main() {
	port := Port
	flag.IntVar(&port, "port", Port, "HTTP listening port")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	rand.Seed(time.Now().UnixNano())
	listenOn := ":" + strconv.Itoa(port)
	templateCache, err := newTemplateCache(TemplatesRoot)
	if err != nil {
		errorLog.Fatal(err)
	}

	srv := &server{
		modules:       []Module{},
		port:          listenOn,
		topic:         "Requirements Writing Quiz",
		baseUrl:       BaseUrl,
		errorLog:      errorLog,
		infoLog:       infoLog,
		templateCache: templateCache,
	}

	httpSrv := &http.Server{
		Addr:     listenOn,
		ErrorLog: errorLog,
		Handler:  nosurf.New(srv.routes()),
	}

	infoLog.Printf("Starting server listening on address *%s ...\n", listenOn)
	err = httpSrv.ListenAndServe()
	errorLog.Fatal(err)
}
