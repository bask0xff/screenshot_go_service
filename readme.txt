# https://claude.ai/chat/7dc2da72-b602-4eb4-99db-c1ccf034bf27


cd app
go mod init screenshot-api

docker compose down
docker compose up -d


---

docker compose down
docker compose up -d --build
docker compose logs api

docker logs screenshot_go_service-api-1

--------------------------------

docker cp import_addresses.sql screenshot_go_service-postgres-1:/import_addresses.sql

d:\Projects\2026\screenshot_go_service>docker exec -it screenshot_go_service-postgres-1 psql -U admin -d mydata -f /import_addresses.sql
psql:/import_addresses.sql:7: NOTICE:  relation "btcaddress2" already exists, skipping
CREATE TABLE
TRUNCATE TABLE
INSERT 0 1095
 setval
--------
   1095
(1 row)


What's next:
    Try Docker Debug for seamless, persistent debugging tools in any container or image → docker debug screenshot_go_service-postgres-1
    Learn more at https://docs.docker.com/go/debug-cli/

d:\Projects\2026\screenshot_go_service>


Example to start:

d:\Projects\2026\screenshot_go_service>docker compose down
time="2026-02-22T07:13:26+01:00" level=warning msg="d:\\Projects\\2026\\screenshot_go_service\\docker-compose.yml: `version` is obsolete"
[+] Running 3/3
 ✔ Container screenshot_go_service-api-1          Removed                                                                                                                                                                               0.4s
 ✔ Container screenshot_go_service-browserless-1  Removed                                                                                                                                                                               0.7s
 ✔ Network screenshot_go_service_default          Removed                                                                                                                                                                               0.3s

d:\Projects\2026\screenshot_go_service>docker compose up -d
time="2026-02-22T07:13:57+01:00" level=warning msg="d:\\Projects\\2026\\screenshot_go_service\\docker-compose.yml: `version` is obsolete"
[+] Running 3/3
 ✔ Network screenshot_go_service_default          Created                                                                                                                                                                               0.1s
 ✔ Container screenshot_go_service-browserless-1  Started                                                                                                                                                                               0.7s
 ✔ Container screenshot_go_service-api-1          Started                                                                                                                                                                               1.2s

d:\Projects\2026\screenshot_go_service>docker ps
CONTAINER ID   IMAGE                          COMMAND                  CREATED          STATUS          PORTS                                                                                                         NAMES
f94b6fa799cd   screenshot_go_service-api      "./server"               9 seconds ago    Up 8 seconds    0.0.0.0:8082->8082/tcp                                                                                        screenshot_go_service-api-1
72d60f020e2c   ghcr.io/browserless/chromium   "./scripts/start.sh"     10 seconds ago   Up 8 seconds    0.0.0.0:3002->3000/tcp                                                                                        screenshot_go_service-browserless-1
377a29bccf99   vue-notus-main-frontend        "docker-entrypoint.s…"   27 minutes ago   Up 26 minutes   8081/tcp, 0.0.0.0:8081->8080/tcp                                                                              vuejs-notus
64470f7a2faf   anticaptchaio-frontend         "docker-entrypoint.s…"   28 minutes ago   Up 28 minutes   0.0.0.0:8080->8080/tcp                                                                                        vue-frontend
33816bad8fe6   anticaptchaio-backend          "/microservice"          28 minutes ago   Up 28 minutes   0.0.0.0:7070->7070/tcp                                                                                        go-backend
3de38c6d54b9   postgres:16-alpine             "docker-entrypoint.s…"   28 minutes ago   Up 28 minutes   0.0.0.0:5432->5432/tcp                                                                                        postgres
3d5379a6aaad   redis:7.2-alpine               "docker-entrypoint.s…"   28 minutes ago   Up 28 minutes   0.0.0.0:6379->6379/tcp                                                                                        redis
19a3f626b8c7   rabbitmq:3-management-alpine   "docker-entrypoint.s…"   28 minutes ago   Up 28 minutes   4369/tcp, 5671/tcp, 0.0.0.0:5672->5672/tcp, 15671/tcp, 15691-15692/tcp, 25672/tcp, 0.0.0.0:15672->15672/tcp   rabbitmq

d:\Projects\2026\screenshot_go_service>curl "http://localhost:8082/screenshot?url=https://example.com" -H "X-API-Key: user_abc123" --output test.png
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100 17326    0 17326    0     0   8239      0 --:--:--  0:00:02 --:--:--  8242

d:\Projects\2026\screenshot_go_service>curl "http://localhost:8082/screenshot?url=https://google.com" -H "X-API-Key: user_abc123" --output test.png
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100 70139    0 70139    0     0  26577      0 --:--:--  0:00:02 --:--:-- 26577

d:\Projects\2026\screenshot_go_service>curl "http://localhost:8082/screenshot?url=https://habr.com" -H "X-API-Key: user_abc123" --output test.png
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100 5167k    0 5167k    0     0   560k      0 --:--:--  0:00:09 --:--:-- 1378k

d:\Projects\2026\screenshot_go_service>

-------------------------------

d:\Projects\2026\screenshot_go_service>curl -X POST "http://localhost:8082/auth/login" -H "Content-Type: application/json" -d "{\"email\":\"newuser@example.com\",\"password\":\"secret123\"}"
{"api_key":{"id":1,"user_id":7,"key":"502e7d7e1b67209ba49512ab02dab62926497de204a81b2a3a5d2973bf3f6ae7","tier":"free","requests":0,"created_at":"2026-02-22T11:31:29.656724Z"}}

d:\Projects\2026\screenshot_go_service>curl "http://localhost:8082/screenshot?url=https://example.com" -H "X-API-Key: 502e7d7e1b67209ba49512ab02dab62926497de204a81b2a3a5d2973bf3f6ae7" --output test.png
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100 17326    0 17326    0     0   8794      0 --:--:--  0:00:01 --:--:--  8794

d:\Projects\2026\screenshot_go_service>curl "http://localhost:8082/screenshot?url=https://habr.com" -H "X-API-Key: 502e7d7e1b67209ba49512ab02dab62926497de204a81b2a3a5d2973bf3f6ae7" --output habr.png
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100 5168k    0 5168k    0     0   327k      0 --:--:--  0:00:15 --:--:-- 1250k

d:\Projects\2026\screenshot_go_service>curl "http://localhost:8082/screenshot?url=https://pikabu.ru" -H "X-API-Key: 502e7d7e1b67209ba49512ab02dab62926497de204a81b2a3a5d2973bf3f6ae7" --output pikabu.png
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100  851k    0  851k    0     0  88841      0 --:--:--  0:00:09 --:--:--  251k

d:\Projects\2026\screenshot_go_service>curl -X POST "http://localhost:8082/auth/register" -H "Content-Type: application/json" -d "{\"email\":\"newXXXuser@example.com\",\"password\":\"secret123\"}"
{"user":{"id":8,"email":"newXXXuser@example.com","created_at":"2026-02-22T11:45:50.58443Z"},"api_key":{"id":2,"user_id":8,"key":"976a3af5ac83efdadb7d708b86639321970afb4bdc07867bbd4fbb95064ba4cf","tier":"free","requests":0,"created_at":"2026-02-22T11:45:50.586835Z"}}

d:\Projects\2026\screenshot_go_service>curl "http://localhost:8082/screenshot?url=https://d3.ru" -H "X-API-Key: 976a3af5ac83efdadb7d708b86639321970afb4bdc07867bbd4fbb95064ba4cf" --output d3.png
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100 2218k    0 2218k    0     0   380k      0 --:--:--  0:00:05 --:--:--  628k

d:\Projects\2026\screenshot_go_service>curl "http://localhost:8082/screenshot?url=https://d3.ru" -H "X-API-Key: 976a3af5ac83efdadb7d708b86639321970afb4bdc07867bbd4fbb95064ba4cf" --output d3-2.png
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100 1426k    0 1426k    0     0   264k      0 --:--:--  0:00:05 --:--:--  345k

d:\Projects\2026\screenshot_go_service>curl "http://localhost:8082/screenshot?url=https://d3.ru" -H "X-API-Key: 976a3af5ac83efdadb7d708b86639321970afb4bdc07867bbd4fbb95064ba4cf" --output d3-3.png
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100 6887k    0 6887k    0     0   742k      0 --:--:--  0:00:09 --:--:-- 1785k

d:\Projects\2026\screenshot_go_service>curl "http://localhost:8082/screenshot?url=https://d3.ru" -H "X-API-Key: 976a3af5ac83efdadb7d708b86639321970afb4bdc07867bbd4fbb95064ba4cf" --output d3-4.png
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100  918k    0  918k    0     0   174k      0 --:--:--  0:00:05 --:--:--  185k

d:\Projects\2026\screenshot_go_service>curl "http://localhost:8082/screenshot?url=https://d3.ru" -H "X-API-Key: 976a3af5ac83efdadb7d708b86639321970afb4bdc07867bbd4fbb95064ba4cf" --output d3-5.png
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100 1571k    0 1571k    0     0   289k      0 --:--:--  0:00:05 --:--:--  378k

d:\Projects\2026\screenshot_go_service>curl "http://localhost:8082/screenshot?url=https://d3.ru" -H "X-API-Key: 976a3af5ac83efdadb7d708b86639321970afb4bdc07867bbd4fbb95064ba4cf" --output d3-6.png
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100 5582k    0 5582k    0     0   428k      0 --:--:--  0:00:13 --:--:-- 1308k

d:\Projects\2026\screenshot_go_service>


----------------------------









