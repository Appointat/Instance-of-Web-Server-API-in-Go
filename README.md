# Serveur de vote
https://github.com/Appointat/Instance-of-Web-Server-API-in-Go
## Participants
- Zikang CHEN [zikang.chen@etu.utc.fr](mailto:zikang.chen@etu.utc.fr)
- Yuan GAO [yuan.gao@etu.utc.fr](mailto:yuan.gao@etu.utc.fr)

## Structure de Projet
- src/
  - cmd/
    - launch.go
  - methods/
    - methods.go
  - server/
    - server.go
  - types/
    - modules.go

Le côté serveur est instancié par l'objet Server et inclut un pool de mappage de BallotID à Ballot, Le serveur est démarré via launch.go, qui appelle la méthode NewServer pour créer un nouvel objet serveur et monter les commandes /new_ballot, /vote, /result aux fonctions correspondantes (port par défaut 8080).
L'implémentation des méthodes de vote telles que Majority, Borda est définie dans le fichier methods.go, tandis que le fichier modules.go définit le format JSON pour la communication serveur-client.

## Comment lancer notre projet
```go
go install github.com/Appointat/Instance-of-Web-Server-API-in-Go/cmd@latest
```
Dans la chemin de dossier src, exécute la commande
```go
go run .\launch.go
```
Le serveur est lancé sur le port 8080 (http://localhost:8080)
### Créer un ballot
Basculer vers http://localhost:8080/ballot créer un ballot en exécutant la commande POST
```json
{
    "rule": "Majority",
    "deadline": "2023-12-01T00:00:00+01:00",
    "voterIDs": ["Voter1", "Voter2", "Voter3"],
    "Alts": 3,
    "tieBreak": [1, 2, 3]
}
```
À ce moment-là, la règle de vote est établie à "Majorité", et l'ID de l'électeur est "Voter1". Seuls les électeurs avec les IDs spécifiés peuvent voter ; les électeurs non autorisés recevront une erreur de requête incorrecte 400(this voter ID is not allowed to vote). \
Date limite : 1er décembre 2023, UTC+0 00:00 \
Candidats : 3 positions \
Tableau de départage : [1, 2, 3] En cas d'égalité, le candidat avec la position la plus précoce gagne.

La création de vote réussie retournera :
```json
{
    "ballot-id": "scrutin0",
}
```
Les votes créés par la suite seront "scrutin1", "scrutin2", etc.
### Comment voter

Basculer vers http://localhost:8080/vote 

```json
{
    "agent-id": "Voter1",
    "ballot-id": "scrutin0",
    "prefs": [1, 2, 3]
}
```
À ce moment-là, nous avons enregistré avec succès le vote avec les préférences [1, 2, 3] pour "Voter1" dans le pool de vote "scrutin0".

### Comment obtenir le résultat
Basculer vers http://localhost:8080/result Les résultats peuvent être obtenus seulement si au moins un électeur dans le pool a voté ; autrement, une erreur 425 (result not ready) sera retournée..

```json
{
    "ballot-id": "scrutin0"
}
```

## Réalisation du projet
![](image.png)
Ce projet met en œuvre un système de vote en ligne simple qui gère le processus de vote via une API de serveur Web écrite en Go. Ce système permet aux utilisateurs de créer des votes, de participer au vote et d'obtenir les résultats après la fin du vote. Le projet se compose de quatre parties principales :

1. `server.go`: Ce fichier définit la logique principale du service. Il gère la création des urnes, le processus de vote et les requêtes pour les résultats. Il définit également le type `Server` utilisé pour stocker l'état et les informations des urnes.

2. `modules.go`: Ce fichier définit les structures de données nécessaires à la communication entre le client et le serveur. Ces structures représentent différentes requêtes et réponses HTTP.

3. `Ballot` Structure: Elle définit la structure de données d'une urne, incluant les règles, la date limite, les IDs des électeurs, le nombre de votes, les options, les règles de départage et le gagnant.

4. `Server` Structure: Elle contient une correspondance de toutes les urnes, le nombre actuel d'urnes et l'ID de la prochaine urne à attribuer.

Voici le flux principal de ce service Web :

### Création d'une urne (`HandleBallot`)
1. Analyser les données JSON dans le corps de la requête dans la structure `NewBallotRequest`.
2. Vérifier que la règle est valide (actuellement, seules "Majority", "Borda", "Condorcet" sont prises en charge).
3. Vérifier que la date limite est valide.
4. Vérifier que le nombre d'options est supérieur à 2 et que le tableau de départage est valide.
5. Générer un ID d'urne et créer une nouvelle instance de `Ballot` pour être stockée dans l'état du serveur.
6. Envoyer l'ID d'urne comme réponse.

### Vote (`HandleVote`)
1. Analyser les données JSON dans le corps de la requête dans la structure `VoteRequest`.
2. Valider que l'ID d'urne et l'ID de l'électeur sont valides.
3. Vérifier s'il est avant la date limite.
4. Valider que le bulletin de vote de l'électeur est légal.
5. Stocker le bulletin dans l'urne et retourner une réponse réussie.

### Obtention des résultats (`HandleResult`)
1. Analyser les données JSON dans le corps de la requête dans la structure `ResultRequest`.
2. Valider que l'ID d'urne est valide.
3. S'il n'y a pas de vote, retourner une erreur.
4. Calculer les résultats selon les règles de l'urne.
5. Envoyer les résultats incluant le gagnant et les classements comme réponse.

La fonction `SortCandidatesByRanking` est responsable du calcul et du classement des candidats selon la méthode de vote donnée (règle de la majorité, compte Borda, méthode Condorcet).

### Détails techniques
- **Validation**: Toutes les données saisies subissent une validation stricte pour garantir la robustesse du système.
- **Gestion d'état**:La structure `Server` agit comme un conteneur d'état, gérant l'état de toutes les urnes.
- **Calcul des résultats**: Le calcul des résultats est basé sur la méthode de vote fournie. Les détails spécifiques de l'implémentation de la méthode de vote ne sont pas fournis ; ils devraient être implémentés comme des fonctions dans le package `methods`.