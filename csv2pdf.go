package main

import (
	"bytes"
	"encoding/base64" // pour l'envoi de la source latex pour une compilation à distance
	"encoding/csv"    // pour la lecture des données
	"fmt"
	"io"
	"io/ioutil"
	"net/http" // pour l'envoi de la source latex pour une compilation à distance
	"os"
	"os/exec" // pour la compilation en local
	"strings"
	"text/template" // pour la transformation des données en source latex

	"github.com/Masterminds/sprig"               // pour des fonctions supplémentaires dans les templates
	"github.com/markbates/pkger"                 // permet d'inclure les police et le template
	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"    // pour la création de la version paysage
	pdfcpu "github.com/pdfcpu/pdfcpu/pkg/pdfcpu" // -- de même --
)

// Les données tels qu'il sont présents (dans l'ordre) dans le csv
// (la première ligne du csv est ignorée)
type Agent struct {
	Nom       string // 0
	Prenom    string // 1
	Service   string // 2
	Fixe      string // 3
	Portable  string // 4
	Renvoi    string // 5
	Averifier string // 6
}

// Vérification des erreurs
func check(e error) {
	if e != nil {
		fmt.Printf("\nError %v\n\n", e)
		panic(e)
	}
}

// Transformation du csv `csvname` en liste d'agents
// (la première ligne du csv est ignorée)
func toData(csvname string) []Agent {

	fmt.Println("Lecture des donnés de", csvname)
	// lecture du fichier csv
	csvFile, err := ioutil.ReadFile(csvname)
	check(err)
	fmt.Println("Conversion des donnés")
	// conversion du csv en []Agent
	r := csv.NewReader(bytes.NewReader(csvFile))
	agents := []Agent{}
	first := true
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		check(err)
		// la première ligne du csv est ignorée
		if first {
			first = false
			continue
		}
		// ajouter le nouveau dossier
		agents = append(agents, Agent{record[0], record[1], record[2], record[3], record[4], record[5], record[6]})
	}

	return agents
}

// Utilisation des données des agent []Agent pour produire la source latex à compiler.
// Cette transformation est basée sur le modèle `annuaire.template.tex`.
// Ce modèle est intégré à l'exécutable grâce `pkger`.
func toLaTeX(agents []Agent) []byte {
	fmt.Println("Transformation en LaTeX")

	// le resultat sera ici
	var result bytes.Buffer
	// Compilation du modèle
	fileAnnuaire, err := pkger.Open("/annuaire.template.tex")
	check(err)
	defer fileAnnuaire.Close()
	check(err)
	b, err := ioutil.ReadAll(fileAnnuaire)
	check(err)
	t, err := template.New("annuaire").Funcs(sprig.TxtFuncMap()).Parse(string(b))
	check(err)

	// Transformation du modèle en source latex en utilisant les données
	err = t.Execute(&result, agents)
	check(err)

	return result.Bytes()
}

// Cette fonction était utilisée pour envoyer les polices pour la compilation.
// Mais finalement il n'y a plus besoin car le serveur a déjà la police Roboto présente.
// func fileToJSON(filename string) string {
// 	f, err := pkger.Open("/" + filename)
// 	check(err)
// 	b, err := ioutil.ReadAll(f)
// 	check(err)

// 	return `{ "path": "` + filename + `", "file": "` + base64.StdEncoding.EncodeToString(b) + `"}`
// }

// compile le PDF en ligne
func toPDF(content []byte) []byte {
	fmt.Println("Envoi pour compilation ...")
	// convertit la source en base64 à l'intérieur de json avant de soumettre
	json := `{
        "compiler": "xelatex",
        "resources": [
            {
                "main": true,
                "file": "` + base64.StdEncoding.EncodeToString(content) + `"
            }` +
		// `,` + fileToJSON("Roboto-Regular.ttf") +
		// `,` + fileToJSON("Roboto-Bold.ttf") +
		`]
    }`

	// compilation en ligne
	// https://github.com/YtoTech/latex-on-http
	body := strings.NewReader(json)
	resp, err := http.Post("https://latex.ytotech.com/builds/sync", "application/json", body)
	check(err)
	defer resp.Body.Close()

	// lecture de la réponse
	pdf, err := ioutil.ReadAll(resp.Body)
	check(err)

	return pdf
}

func main() {
	var (
		portrait                              []byte
		landscape, temp                       bytes.Buffer
		baseName, portraitName, landscapeName string
		err                                   error
	)

	// création de la version portrait
	if len(os.Args) > 1 && os.Args[1] == "--local" {
		baseName = "annuaire_local"
		portraitName = baseName + "_portrait"
		fmt.Printf("Creation de %s.tex\n", portraitName)
		// enregistrement de la source latex à compiler
		ioutil.WriteFile(portraitName+".tex", toLaTeX(toData("annuaire.csv")), 0644)
		// compilation en local
		fmt.Printf("xelatex %s.tex\n", portraitName)
		exec.Command("xelatex", portraitName+".tex").Run()
		fmt.Printf("Version portrait PDF dans %s.pdf\n", portraitName)
		// lecture de la version portrait (pour la transformer en paysage après)
		portrait, err = ioutil.ReadFile(portraitName + ".pdf")
		check(err)
	} else {
		baseName = "annuaire"
		portraitName = baseName + "_portrait"
		portrait = toPDF(toLaTeX(toData("annuaire.csv")))
		// enregistrement de la version portrait
		err = ioutil.WriteFile(portraitName+".pdf", portrait, 0644)
		check(err)
		fmt.Printf("Version portrait PDF dans %s.pdf\n", portraitName)
	}
	// création de la version paysage
	landscapeName = baseName + "_paysage"
	// changement de l'ordre des pages
	fmt.Printf("Change l'ordre des pages 4,1,2,3...\n")
	err = pdfapi.Collect(bytes.NewReader(portrait), &temp, []string{"1", "4", "3", "2"}, nil)
	check(err)
	// combinaison des pages deux par deux
	fmt.Printf("Combine deux pages sur une...\n")
	nup, _ := pdfcpu.PDFNUpConfig(2, "f:A4,b:on, m:7")
	err = pdfapi.NUp(bytes.NewReader(temp.Bytes()), &landscape, nil, nil, nup, nil)
	check(err)
	// enregistrement de la version paysage
	err = ioutil.WriteFile(landscapeName+".pdf", landscape.Bytes(), 0644)
	check(err)
	fmt.Printf("Version paysage dans %s.pdf\n", landscapeName)
}
