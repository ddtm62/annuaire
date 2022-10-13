# Annuaire de la DDTM 62

Il s'agit d'un petit logiciel qui transforme les données sous format `csv` en deux `pdf` (un en portrait et un en paysage).

## Utilisation

- Télécharger l'exécutable `csv2pdf` pour votre plate-forme (Windows, Linux, Mac) de la page des [releases](https://github.com/ddtm62/annuaire/releases).
- (optionnel, mais recommandé) Télécharger aussi [tectonic](https://tectonic-typesetting.github.io/) pour votre plate-forme.
- Mettre dans le même dossier que l'exécutable le fichier `annuaire.csv` qui doit avoir le format suivant :

```csv
nom,prenom,service,fixe,portable,renvoi,a_verifier
NOM1,Prenom11,SERV1,03 00 00 00 00,06 00 00 00 00,,
NOM2,Prenom12,SERV2,03 00 00 00 01,06 00 00 00 01,x,
NOM2,Prenom12,SERV3,03 00 00 00 02,06 00 00 00 02,,Observation
...
```
- Mettre dans le sous-dossier `fonts` (à créer) les fichiers `Roboto-Regular.ttf` et `Roboto-Bold.ttf` de la police [Roboto](https://fonts.google.com/specimen/Roboto).
- Lancer l'exécutable. Deux nouveaux fichiers `pdf` doivent apparaître dans le dossier : `annuaire_portrait.pdf` et `annuaire_paysage.pdf`

## Comment ça fonctionne

- Les données du fichier `csv` sont intégrées dans le modèle `annuaire.template.tex` pour produire un fichier `latex` compilable.
- Ce fichier est compilé (avec `tectonic` par défaut, ou avec `xelatex` si spécifié), ou envoyé (si demandé) pour compilation (avec xelate) à https://latex.ytotech.com/builds/sync.
- Le `pdf` qui en résulte est enregistré dans `annuaire_portrait.pdf`.
- La bibliothèque [pdfcpu](https://github.com/pdfcpu/pdfcpu) est utilisée pour créer la version paysage.

### Compilation avec tectonic

Pour pouvoir compiler avec [tectonic](https://tectonic-typesetting.github.io/), qui est l'option par défaut, il faut copier l'exécutable qui se trouve dans les [releases de tectonic](https://github.com/tectonic-typesetting/tectonic/releases). Par exemple pour installer la version `0.9.0` pour Windows vous pouvez télécharger le [tectonic-0.9.0-x86_64-pc-windows-msvc.zip](https://github.com/tectonic-typesetting/tectonic/releases/download/tectonic%400.9.0/tectonic-0.9.0-x86_64-pc-windows-msvc.zip).

L'exécutable `tectonic`(`.exe`) doit être placé dans un dossier qui se trouve dans le `PATH` pour qu'on puisse le lancer de n'importe où.

À la première utilisation `tectonic` doit se connecte à internet pour mettre à jour le cache. Pour cela il faut lancer le `csv2pdf` avec l'option `--tectonic-update` :

```bash
> csv2pdf --tectonic-update
```

Si votre ordinateur ne permet pas à `tectonic` de se connecter à internet pour mettre le cache à jour, il faut le faire à la main : il faut décompresser le fichier [Tectonic.zip](https://github.com/ddtm62/annuaire/raw/main/Tectonic.zip) dans le répertoire du cache. Pour savoir ou se trouve le cache il faut exécuter la commande

```bash
> tectonic -X show user-cache-dir
```

### Compilation avec xelatex

Pour compiler avec `xelatex` au lieu de `tectonic` il faut avoir une distribution LaTeX (contenant `xelatex`) installée sur votre ordinateur. Auquel cas il suffit de rajouter le paramètre `--utiliser=xelatex`.

### Compilation dans me web

Et si vous voulez produire les pdfs sans devoir à installer `tectonic` (ou `xelatex`) et les polices `Roboto`, vous pouvez compiler via le service https://latex.ytotech.com/builds/sync, auquel cas il suffit de rajouter `--utiliser=web`.
Les fichiers ainsi produits sont nommés `annuaire_web_paysage.pdf` et `annuaire_web_portrait.pdf`.

## Comment compiler les sources `go`

Ce logiciel est écrit en [Go](https://golang.org/) qui doit être présent sur votre machine pour pouvoir compiler les sources.

Pour exécuter les sources sans les compiler vous pouvez faire :

```shell
> go run .
```

Pour les compiler vers un exécutable pour votre plate-forme vous pouvez :

```shell
> go build .
```

Cette commande produit l'exécutable nommé `annuaire` (ou `annuaire.exe` sous Windows) au lieu de `csv2pdf` (respectivement `csv2pdf.exe`).

Et si vous voulez compiler pour plusieurs plate-formes, le plus simple est probablement d'utiliser [goreleaser](https://github.com/goreleaser/goreleaser/) en local :

```shell
> goreleaser --snapshot --skip-publish --rm-dist
```
Cette compilation est configurée dans le [.goreleaser.yml](.goreleaser.yml) et les exécutables sont disponibles après dans le sous-dossier `dist`.

C'est ainsi que sont compilées les versions disponibles sur [GitHub](.github/workflows/release.yaml)
