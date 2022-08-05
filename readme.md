# сборка
версия сборки 1.19 <br>
go mod tidy <br>
go env -w GOOS=linux <br>
go build -o bin/rotateBot<br>

# конфигурирование и запуск
### пример конфига
`{
    "TokenTG": "djfkhdjshdbfjshgdfkjhg",
    "DebugMode": true, 
    "DBpatch": "path/to/DB/file",
    "ScriptPatch": "path/to/script",
    "UserAcl": [
        1,
        2,
        3
    ]
}`
1. TokenTG - токен для доступа к боту ТГ
2. DebugMode - работает при запуске в консоли. В значении true будет плеваться структурой получаемых сообщений из API ТГ
3. DBpatch - путь до базы соотвествий `порт=3 октет адреса роутера`
4. ScriptPatch - путь до скрипта который будет выполнять ротация на модемах
5. UserAcl - массив пользователей с доступом к боту
### запуск 
сервис расположить по пути `/etc/systemd/system/rotateBot.service` после чего выполнить `systemctl daemon-reload`.<br>
бинарник расположить в `/usr/local/bin/rotateBot`
### ключи
`c` - (default: `/etc/rotateBot/config`) ключ позволяющий указать путь до конфига