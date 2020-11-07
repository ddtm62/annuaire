# Annuaire de la DDTM 62

Il s'agit d'un petit logiciel qui transforme les données sous format `csv` en deux `pdf` (un en portrait et un en paysage).

## Utilisation

- Télécharger l'exécutable `csv2pdf` pour votre plate-forme (Windows, Linux, Mac) de la page des releases.
- Mettre dans le même dossier que l'exécutable le fichier `annuaire.csv` qui doit avoir le format suivant:

```csv
nom,prenom,service,fixe,portable,renvoi,a_verifier
NOM1,Prenom11,SERV1,03 00 00 00 00,06 00 00 00 00,,
NOM2,Prenom12,SERV2,03 00 00 00 01,06 00 00 00 01,x,
NOM2,Prenom12,SERV3,03 00 00 00 02,06 00 00 00 02,,Observation
...
```

- Lancer l'exécutable. Deux nouveaux fichiers `pdf` doivent apparaître dans le dossier : `annuaire_portrait.pdf` et `annuaire_paysage.pdf`

## Comment ça fonctionne

- Les données du fichier `csv` sont intégrées dans le modèle `annuaire.template.tex` pour produir un fichier `latex` compilable.
- Ce fichier est envoyé pour compilation (avec xelate) à https://latex.ytotech.com/builds/sync.
- Le `pdf` retourné est enregistré dans `annuaire_portrait.pdf`.
- La bibliothèque [pdfcpu](https://github.com/pdfcpu/pdfcpu) est utilisé pour créer la version paysage.

## Et si je ne veux compiler en local ?

C'est possible, mais il faut avoir une distribution LaTeX (contenant `xelatex`) installée sur votre ordinateur. Auquel cas il suffit de rajouter le paramètre `--local` pour que le fichier soit compiler en local.
