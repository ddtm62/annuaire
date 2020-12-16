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
	"github.com/markbates/pkger"                 // permet d'inclure le template (et autre fichiers, comme des polices)
	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"    // pour la création de la version paysage
	pdfcpu "github.com/pdfcpu/pdfcpu/pkg/pdfcpu" // -- de même --
	flag "github.com/spf13/pflag"                // pour les paramètres
)

// quelques variables globales
var (
	// la version du logiciel (remplacée lors de la compilation)
	version = "--"
	// la methode pour compiler le .tex (valeur de --utiliser)
	sEngine  string
	bVersion bool
	bHelp    bool
	// une variable temporaire d'erreur
	err error
)

// Aide affiche l'aide d'utilisation
func Aide() {
	var out = flag.CommandLine.Output()
	fmt.Fprintf(out, "csv2pdf (version: %s)\n\n", version)
	fmt.Fprintf(out, "Ce programme transforme annuaire.csv en pdf.\n")
	fmt.Fprintf(out, "Il génère 2 versions (portrait et paysage).\n")
	fmt.Fprintf(out, "\n")
	flag.PrintDefaults()
	fmt.Fprintf(out, "\n")
}

// Imprime la version
func printVersion() {
	var out = flag.CommandLine.Output()
	fmt.Fprintf(out, "version: %s\n", version)
}

func SetParameters() {
	flag.StringVar(&sEngine, "utiliser", "web", "Comment compiler le .tex [tectonic|xelatex|web].")
	flag.BoolVarP(&bVersion, "version", "v", false, "Affiche le numéro de version.")
	flag.BoolVarP(&bHelp, "aide", "h", false, "Imprime ce message d'aide.")
	// garde l'ordre des paramètres dans l'aide
	flag.CommandLine.SortFlags = false
	// installe la traduction des messages en français
	flag.CommandLine.SetOutput(FrenchTranslator{flag.CommandLine.Output()})
	// le message d'aide
	flag.Usage = Aide
	// en cas d'erreur ne pas afficher l'erreur une deuxième fois
	flag.CommandLine.Init("marianne", flag.ContinueOnError)

	// récupère les flags
	err = flag.CommandLine.Parse(os.Args[1:])
	// affiche l'aide si demandé ou si erreur de paramètre
	if bHelp || err != nil {
		flag.Usage()
		if err != nil {
			fmt.Fprintln(flag.CommandLine.Output(), "ERREUR : ", err)
			os.Exit(2)
		} else {
			os.Exit(0)
		}
	}

	// montrer la version ?
	if bVersion {
		printVersion()
		os.Exit(0)
	}
}

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
func check(e error, m ...interface{}) {
	if e != nil {
		if len(m) > 0 {
			fmt.Print("\nErreur : ")
			fmt.Print(m...)
		} else {
			fmt.Print("\nErreur.")
		}
		fmt.Printf("\nMessage (en anglais) sur l'erreur:\n%v\n\n", e)
		fmt.Printf("\nAppuyer sur « entrer » pour finir.")
		fmt.Scanln()
		panic(e)
	}
}

// Transformation du csv `csvname` en liste d'agents
// (la première ligne du csv est ignorée)
func toData(csvname string) []Agent {

	fmt.Println("Lecture des donnés de", csvname)
	// lecture du fichier csv
	csvFile, err := ioutil.ReadFile(csvname)
	check(err, "Problème de lecture de", csvname)
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
		check(err, "Problème de scanneur csv dans", csvname)
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
	// en cas de changement de annuaire.template.tex il faut relancer pkger dans le dossier pour recréer `pkged.go`
	fileAnnuaire, err := pkger.Open("/annuaire.template.tex")
	check(err, "Problème lors de l'ouverture du modèle latex.")
	defer fileAnnuaire.Close()
	b, err := ioutil.ReadAll(fileAnnuaire)
	check(err, "Problème lors de la lecture du modèle latex.")
	t, err := template.New("annuaire").Funcs(sprig.TxtFuncMap()).Parse(string(b))
	check(err, "Problème lors de la compilation du modèle latex.")

	// Transformation du modèle en source latex en utilisant les données
	err = t.Execute(&result, agents)
	check(err, "Problème lors de la création de la source latex.")

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
	check(err, "Problème lors de l'envoi de la source pour compilation.")
	defer resp.Body.Close()

	// lecture de la réponse
	pdf, err := ioutil.ReadAll(resp.Body)
	check(err, "Problème lors de la réception du pdf.")

	return pdf
}

func main() {
	var (
		portrait                              []byte
		landscape, temp                       bytes.Buffer
		baseName, portraitName, landscapeName string
	)

	// parse the flags
	SetParameters()

	// création de la version portrait
	if sEngine == "xelatex" || sEngine == "tectonic" {
		baseName = "annuaire_local"
		portraitName = baseName + "_portrait"
		fmt.Printf("Creation de %s.tex\n", portraitName)
		// enregistrement de la source latex à compiler
		ioutil.WriteFile(portraitName+".tex", toLaTeX(toData("annuaire.csv")), 0644)
		// compilation en local
		fmt.Printf(sEngine+" %s.tex\n", portraitName)
		exec.Command(sEngine, portraitName+".tex").Run()
		fmt.Printf("Version portrait PDF dans %s.pdf\n", portraitName)
		// lecture de la version portrait (pour la transformer en paysage après)
		portrait, err = ioutil.ReadFile(portraitName + ".pdf")
		check(err, "Problème lors de la lecture de", portraitName+".pdf")
	} else {
		baseName = "annuaire"
		portraitName = baseName + "_portrait"
		portrait = toPDF(toLaTeX(toData("annuaire.csv")))
		// enregistrement de la version portrait
		err = ioutil.WriteFile(portraitName+".pdf", portrait, 0644)
		check(err, "Problème lors de l'écriture de", portraitName+".pdf")
		fmt.Printf("Version portrait PDF dans %s.pdf\n", portraitName)
	}
	// création de la version paysage
	landscapeName = baseName + "_paysage"
	// changement de l'ordre des pages
	fmt.Printf("Change l'ordre des pages 4,1,2,3...\n")
	err = pdfapi.Collect(bytes.NewReader(portrait), &temp, []string{"1", "4", "3", "2"}, nil)
	check(err, "Problème lors de la transformation du pdf.")
	// combinaison des pages deux par deux
	fmt.Printf("Combine deux pages sur une...\n")
	nup, _ := pdfcpu.PDFNUpConfig(2, "f:A4,b:on, m:7")
	err = pdfapi.NUp(bytes.NewReader(temp.Bytes()), &landscape, nil, nil, nup, nil)
	check(err, "Problème lors de la transformation du pdf.")
	// enregistrement de la version paysage
	err = ioutil.WriteFile(landscapeName+".pdf", landscape.Bytes(), 0644)
	check(err, "Problème lors de l'écriture de", landscapeName+".pdf")
	fmt.Printf("Version paysage dans %s.pdf\n", landscapeName)
}
