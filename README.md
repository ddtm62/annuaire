# Annuaire de la DDTM 62

Il s'agit d'un petit logiciel qui transforme les données sous format `csv` en deux `pdf` (un en portrait et un en paysage).

## Utilisation

- Télécharger l'exécutable `csv2pdf` pour votre plate-forme (Windows, Linux, Mac) de la page des [releases](https://github.com/ddtm62/annuaire/releases).
- Mettre dans le même dossier que l'exécutable le fichier `annuaire.csv` qui doit avoir le format suivant :

```csv
nom,prenom,service,fixe,portable,renvoi,a_verifier
NOM1,Prenom11,SERV1,03 00 00 00 00,06 00 00 00 00,,
NOM2,Prenom12,SERV2,03 00 00 00 01,06 00 00 00 01,x,
NOM2,Prenom12,SERV3,03 00 00 00 02,06 00 00 00 02,,Observation
...
```

- Lancer l'exécutable. Deux nouveaux fichiers `pdf` doivent apparaître dans le dossier : `annuaire_portrait.pdf` et `annuaire_paysage.pdf`

## Comment ça fonctionne

- Les données du fichier `csv` sont intégrées dans le modèle `annuaire.template.tex` pour produir un fichier `latex` compilable.
- Ce fichier est envoyé pour compilation (avec xelate) à https://latex.ytotech.com/builds/sync.
- Le `pdf` retourné est enregistré dans `annuaire_portrait.pdf`.
- La bibliothèque [pdfcpu](https://github.com/pdfcpu/pdfcpu) est utilisé pour créer la version paysage.

## Et si je veux compiler en local ?

C'est possible, mais il faut avoir une distribution LaTeX (contenant `xelatex`) installée sur votre ordinateur, ainsi que la police [Roboto](https://fonts.google.com/specimen/Roboto). Auquel cas il suffit de rajouter le paramètre `--local` pour que le fichier soit compiler en local.

## Comment compiler les sources

Ce logiciel est écrit en [Go](https://golang.org/) qui doit être présent sur votre machine pour pouvoir compiler les sources.

Pour exécuter les sources sans les compiler vous pouvez faire :

```shell
go run csv2pdf.go
```

Pour les compiler vers un exécutable pour votre plate-forme vous pouvez :

```shell
go build csv2pdf.go
```

Et si vous voulez compiler pour plusieurs plate-formes, le plus simple est probablement d'utiliser [goreleaser](https://github.com/goreleaser/goreleaser/) en local :

```shell
goreleaser --snapshot --skip-publish --rm-dist
```
Cette compilation est configurée dans le [.goreleaser.yml](.goreleaser.yml) et les exécutables sont disponibles après dans le sous-dossier `dist`.

C'est ainsi que sont compilées les versions disponibles sur [GitHub](.github/workflows/release.yaml)
