package main

import (
	"encoding/json"
	"fmt"
	"github.com/sthagen/nosurf"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

func ping(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("OK"))
}

func (s *server) defaultHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Prototyping of some web services ... try " + s.baseUrl))
	s.infoLog.Println("Default handler invoked for ", r.RequestURI, " with method", r.Method)
}

func (s *server) quizHandler() http.HandlerFunc {
	totalQuizCount := 0
	n := 1
	for n <= AvailableModules {
		file, _ := ioutil.ReadFile(filepath.Join("internal", "data", "quiz-"+strconv.Itoa(n)+"-db.json"))
		module := Module{}
		_ = json.Unmarshal(file, &module)
		totalQuizCount += len(module.Quiz)
		s.modules = append(s.modules, module)
		n++
	}
	s.infoLog.Printf("Data setup yields %d modules providing a total of %d quizzes\n", len(s.modules), totalQuizCount)
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			moduleId, err := strconv.Atoi(r.URL.Query().Get("mid"))
			if err != nil || moduleId < 1 || AvailableModules < moduleId {
				moduleId = 1 + rand.Intn(AvailableModules)
			}
			module := s.modules[moduleId-1]
			quizCount := len(module.Quiz)

			quizId, err := strconv.Atoi(r.URL.Query().Get("qid"))
			if err != nil || quizId < 1 || quizCount < quizId {
				quizId = 1 + rand.Intn(quizCount)
			}
			data := RenderQuiz{
				ModuleId:     moduleId,
				QuizNumber:   quizId,
				Progress:     100 * (quizId + (moduleId-1)*AvailableQuestionsPerModule) / ProgressMax,
				Task:         module.Quiz[quizId-1].Task,
				TaskId:       1, // TODO(shagen) does not matter for single question pages
				Claim:        module.Quiz[quizId-1].Claim,
				AnswerKind:   module.Quiz[quizId-1].AnswerKind,
				Alternative1: module.Quiz[quizId-1].Alternatives[0].Alternative,
				Alternative2: module.Quiz[quizId-1].Alternatives[1].Alternative,
				Alternative3: module.Quiz[quizId-1].Alternatives[2].Alternative,
				Alternative4: module.Quiz[quizId-1].Alternatives[3].Alternative,
				Alternative5: module.Quiz[quizId-1].Alternatives[4].Alternative,
				BaseUrl:      s.baseUrl,
				Token:        nosurf.Token(r),
			}
			s.render(w, r, "quiz.page.tmpl", &templateData{
				Quiz: data,
			})
			// log.Println(data)
			s.infoLog.Printf("Get data debug: module, quiz = %v, %v\n", moduleId, quizId)
		case "POST":
			// log.Println("Received a POST request")
			if err := r.ParseForm(); err != nil {
				s.serverError(w, err)
				return
			}
			s.infoLog.Printf("Post data debug: r.PostFrom = %v\n", r.PostForm)

			moduleIdUntrusted, err := strconv.Atoi(r.FormValue("_module"))
			if err != nil || moduleIdUntrusted < 1 || AvailableModules < moduleIdUntrusted {
				s.serverError(w, err)
				return
			}
			moduleId := moduleIdUntrusted
			module := s.modules[moduleId-1]
			quizCount := len(module.Quiz)

			quizIdUntrusted, err := strconv.Atoi(r.FormValue("_question"))
			if err != nil || quizIdUntrusted < 1 || quizCount < quizIdUntrusted {
				s.serverError(w, err)
				return
			}
			quizId := quizIdUntrusted
			alternativeIdUntrusted, err := strconv.Atoi(r.FormValue("alternative"))
			if err != nil || alternativeIdUntrusted < 1 || 5 < alternativeIdUntrusted {
				s.serverError(w, err)
				return
			}

			alternativeId := alternativeIdUntrusted - 1
			taskGiven := module.Quiz[quizId-1].Task
			claimGiven := module.Quiz[quizId-1].Claim
			alternativeChosen := module.Quiz[quizId-1].Alternatives[alternativeId].Alternative
			evaluation := module.Quiz[quizId-1].Alternatives[alternativeId].Evaluation
			ok := ""
			if evaluation == "correct" {
				ok = "1"
			}
			hint := module.Quiz[quizId-1].Alternatives[alternativeId].Hint
			hintString := ""
			if hint != "" {
				hintString = fmt.Sprintf("%v", hint)
			}
			explanation := module.Quiz[quizId-1].Explanation
			var details []template.HTML
			inListBlock := false
			afterQuoteStatement := false
			for _, line := range explanation.Details {
				if len(line) > 1 {
					if inListBlock {
						if strings.HasPrefix(line[0], "$ITEM") {
							htmlLine := "<li>" + line[1] + "</li>"
							details = append(details, template.HTML(htmlLine))
						} else {
							s.infoLog.Printf("list mode ignored details line %v\n", line)
						}
					} else {
						if strings.HasPrefix(line[0], "$QUOTE") {
							htmlLine := "<blockquote class=\"claim\">" + line[1] + "</blockquote>"
							details = append(details, template.HTML(htmlLine))
							afterQuoteStatement = true
						} else {
							s.infoLog.Printf("non-list mode ignored details line %v\n", line)
						}
						if !afterQuoteStatement {
							details = append(details, "<br>")
						}
					}
				} else {
					if strings.HasPrefix(line[0], "$LIST_ON") {
						htmlLine := "<ul class=\"feedback\">"
						details = append(details, template.HTML(htmlLine))
						inListBlock = true
					} else if strings.HasPrefix(line[0], "$LIST_OFF") {
						htmlLine := "</ul>"
						details = append(details, template.HTML(htmlLine))
						inListBlock = false
					} else {
						details = append(details, template.HTML(strings.Join(line, "")))
					}
					if !inListBlock && !afterQuoteStatement {
						details = append(details, "<br>")
					}
					afterQuoteStatement = false
				}
			}

			data := RenderEvaluation{
				ModuleId:          moduleId,
				QuizNumber:        quizId,
				Progress:          100 * (quizId + (moduleId-1)*AvailableQuestionsPerModule) / ProgressMax,
				Task:              taskGiven,
				TaskId:            1, // TODO(shagen) does not matter for single question pages
				Claim:             claimGiven,
				AlternativeChosen: alternativeChosen,
				Evaluation:        evaluation,
				Ok:                ok,
				Hint:              hintString,
				Summary:           explanation.Summary,
				Details:           details,
				BaseUrl:           s.baseUrl,
			}

			s.render(w, r, "evaluation.page.tmpl", &templateData{
				Evaluation: data,
			})
			// log.Println(data)

		default:
			w.Header().Set("Allow", http.MethodGet)
			w.Header().Add("Allow", http.MethodPost)
			s.clientError(w, 405)
			return
		}
	}
}
