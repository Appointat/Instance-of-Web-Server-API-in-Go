# Serveur de vote
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

服务器端由Server对象实例化，并包含了BallotID对Ballot的映射池，服务器通过launch.go启动，同时调用NewServer方法创建新服务器对象并挂载/new_ballot,/vote,/result命令至相应函数。(默认8080端口)
Majority, Borda等投票方法的实现被定义在methods.go文件中，modules.go文件定义了服务端与客户端通信的JSON格式。

## Comment lancer notre projet
```go
go install github.com/Appointat/Instance-of-Web-Server-API-in-Go/cmd@latest
```
在cmd下执行
```go
go run launch.go
```
此时浏览器已在8080端口上运行(http://localhost:8080)
### Créer un ballot
在 http://localhost:8080/ballot 创建一个投票，执行post命令
```json
{
    "rule": "Majority",
    "deadline": "2023-12-01T00:00:00+01:00",
    "voterIDs": ["Voter1", "Voter2", "Voter3"],
    "Alts": 3,
    "tieBreak": [1, 2, 3]
}
```
此时投票规则设定为Majority并且投票者ID为Voter1....，只有满足以上投票ID的投票者才能投票，非法投票者会返回400bad request错误(this voter ID is not allowed to vote)。
截止日期：2023年12月1日 UTC+0 00:00
候选人：3位
平局截断数组：[1, 2, 3] 出现平局时，最考前的候选人获胜

投票创建成功会返回
```json
{
    "ballot-id": "scrutin0",
}
```
之后创建的投票为scrutin1, scrutin2...
### Comment voter

切换至 http://localhost:8080/vote 

```json
{
    "agent-id": "Voter1",
    "ballot-id": "scrutin0",
    "prefs": [1, 2, 3]
}
```
此时我们成功以Voter1的身份在投票池scrutin0中记录了投票信息[1, 2, 3]。

### Comment obtenir le résultat
切换至 http://localhost:8080/result ,投票池中至少有一位选民投票才能获取结果，否则会返回425错误(result not ready)。

```json
{
    "ballot-id": "scrutin0"
}
```
