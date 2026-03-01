# squeeze-empty-lines
`squeeze-empty-lines` est un outil CLI écrit en Go qui supprime toutes les lignes vides d’un ensemble de fichiers texte, de manière récursive.
Il normalise également les fins de ligne (`CR`, `LF`, `CRLF`) vers le format Unix (`LF`).
Cet outil est rapide, déterministe, sûr.
## Fonctionnalités
* suppression complète des lignes vides
* normalisation des fins de ligne en `LF`
* traitement récursif des répertoires
* exécution parallèle (multi-cœurs)
* exclusion automatique des fichiers binaires
* filtrage par extension
* filtrage précis par regex (`--include`, `--exclude`)
* mode simulation (`--dry-run`)
* création optionnelle de sauvegardes (`--backup`)
* binaire statique, sans dépendances, vive GO !
## Exemple
Fichier original :
```
ligne 1

ligne 2     ligne 3
ligne 4
```
Après traitement (--collapse-internal-spaces)
```
ligne 1
ligne 2 ligne 3
ligne 4
```
## Installation
### via go install
```
go install github.com/votreuser/squeeze-empty-lines@latest
```
Le binaire sera installé dans :
```
$GOPATH/bin
```
### compilation manuelle
```
CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags="-s -w" \
    -o squeeze-empty-lines
strip squeeze-empty-lines
```
## Utilisation
Traitement du dossier courant :
```
squeeze-empty-lines .
```
Traitement d’un fichier spécifique :
```
squeeze-empty-lines fichier.py
```
Traitement de plusieurs dossiers :
```
squeeze-empty-lines src tests
```
## Options
### --dry-run
Affiche les fichiers qui seraient modifiés, sans les modifier.
```
squeeze-empty-lines --dry-run .
```
### --backup
Crée une sauvegarde `.bak` avant modification.
```
squeeze-empty-lines --backup .
```
Exemple :
```
script.py
script.py.bak
```
### --include REGEX
Inclut uniquement les fichiers correspondant à la regex.
Exemple :
```
squeeze-empty-lines --include '\.py$' .
```
### --exclude REGEX
Exclut les fichiers correspondant à la regex.
Exemple :
```
squeeze-empty-lines --exclude 'test' .
```
### --workers N
Nombre de threads utilisés.
Par défaut :
nombre de CPU disponibles.
Exemple :
```
squeeze-empty-lines --workers 4 .
```
## Extensions supportées
Par défaut, seuls les fichiers texte suivants sont traités :
```
.go .py .js .ts .java .c .cpp .rs .html .css .json .yaml .xml .md .sh .txt
```
Les fichiers binaires sont automatiquement ignorés.
## Répertoires exclus automatiquement
```
.git
node_modules
vendor
venv
.venv
target
```
## Sécurité
L’outil :
* ne modifie pas les fichiers binaires
* peut créer des sauvegardes
* ne modifie pas le contenu des lignes non vides
## Performance
Optimisé pour :
* grands dépôts
* traitements massifs
* exécution multi-cœurs
Peut traiter des dizaines de milliers de fichiers en quelques secondes !
## Cas d’usage
* nettoyage de dépôts Git
* normalisation avant commit
* préparation de datasets
* réduction de la taille de fichiers texte
* homogénéisation
## Compatibilité
Fonctionne sur :
* Linux
* macOS
* Windows (probablement, ça build en CI/CD, mais Windows c’est nul)
## Licence
MIT
