package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/markbates/pkger" // permet d'inclure les police et le templat
	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"
	pdfcpu "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

type Agent struct { // Our example struct, you can use "-" to ignore a field
	Nom       string // 0
	Prenom    string // 1
	Service   string // 2
	Fixe      string // 3
	Portable  string // 4
	Renvoi    string // 5
	Averifier string // 6
}

// Error checking
func check(e error) {
	if e != nil {
		fmt.Printf("\nError %v\n\n", e)
		panic(e)
	}
}

func toData(csvname string) []Agent {

	fmt.Println("Lecture des donnés de", csvname)
	// read the csv file
	csvFile, err := ioutil.ReadFile(csvname)
	check(err)
	fmt.Println("Conversion des donnés")
	// convert the csv to []Agent slice
	r := csv.NewReader(bytes.NewReader(csvFile))
	agents := []Agent{}
	first := true
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		check(err)
		// skip the first line
		if first {
			first = false
			continue
		}
		// add the new record
		agents = append(agents, Agent{record[0], record[1], record[2], record[3], record[4], record[5], record[6]})
	}

	return agents
}

func toLaTeX(agents []Agent) []byte {
	fmt.Println("Transformation en LaTeX")

	// le resultat sera ici
	var result bytes.Buffer
	// Create a new template and parse the letter into it.
	fileAnnuaire, err := pkger.Open("/annuaire.template.tex")
	check(err)
	defer fileAnnuaire.Close()
	check(err)
	b, err := ioutil.ReadAll(fileAnnuaire)
	check(err)
	t, err := template.New("annuaire").Funcs(sprig.TxtFuncMap()).Parse(string(b))
	check(err)

	// Execute the template for each recipient.
	err = t.Execute(&result, agents)
	check(err)

	return result.Bytes()
}

func fileToJSON(filename string) string {
	f, err := pkger.Open("/" + filename)
	check(err)
	b, err := ioutil.ReadAll(f)
	check(err)

	return `{ "path": "` + filename + `", "file": "` + base64.StdEncoding.EncodeToString(b) + `"}`
}

func toPDF(content []byte) []byte {
	fmt.Println("Envoi pour compilation ...")
	// convert it to base64 inside json to submit
	json := `{
        "compiler": "xelatex",
        "resources": [
            {
                "main": true,
                "file": "` + base64.StdEncoding.EncodeToString(content) + `"
            }` +
		`,` +
		fileToJSON("Roboto-Regular.ttf") +
		`,` +
		fileToJSON("Roboto-Bold.ttf") + `
        ]
    }`

	// comile the file online
	// https://github.com/YtoTech/latex-on-http
	body := strings.NewReader(json)
	resp, err := http.Post("https://latex.ytotech.com/builds/sync", "application/json", body)
	check(err)
	defer resp.Body.Close()

	// write the anser to file
	pdf, err := ioutil.ReadAll(resp.Body)
	check(err)

	return pdf
}

func main() {
	var (
		portrait                    []byte
		landscape, temp             bytes.Buffer
		portraitName, landscapeName string
		err                         error
	)

	// create the portrait pdf
	if len(os.Args) > 1 && os.Args[1] == "--local" {
		portraitName = "annuaire_local"
		fmt.Printf("Creation de %s.tex\n", portraitName)
		ioutil.WriteFile(portraitName+".tex", toLaTeX(toData("annuaire.csv")), 0644)
		fmt.Printf("xelatex %s.tex\n", portraitName)
		exec.Command("xelatex", portraitName+".tex").Run()
		fmt.Printf("Version portrait PDF dans %s.pdf\n", portraitName)
		portrait, err = ioutil.ReadFile(portraitName + ".pdf")
		check(err)
	} else {
		portraitName = "annuaire"
		portrait = toPDF(toLaTeX(toData("annuaire.csv")))
		err = ioutil.WriteFile(portraitName+".pdf", portrait, 0644)
		check(err)
		fmt.Printf("Version portrait PDF dans %s.pdf\n", portraitName)
	}
	// create the landscape
	landscapeName = portraitName + "_paysage"

	fmt.Printf("Change l'ordre des pages 4,1,2,3...\n")
	err = pdfapi.Collect(bytes.NewReader(portrait), &temp, []string{"1", "4", "3", "2"}, nil)
	check(err)

	fmt.Printf("Combine deux pages sur une...\n")
	nup, _ := pdfcpu.PDFNUpConfig(2, "f:A4,b:on, m:7")
	err = pdfapi.NUp(bytes.NewReader(temp.Bytes()), &landscape, nil, nil, nup, nil)
	check(err)
	err = ioutil.WriteFile(landscapeName+".pdf", landscape.Bytes(), 0644)
	check(err)
	fmt.Printf("Version paysage dans %s.pdf\n", landscapeName)
}
