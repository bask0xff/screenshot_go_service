# https://claude.ai/chat/7dc2da72-b602-4eb4-99db-c1ccf034bf27


cd app
go mod init screenshot-api

docker compose down
docker compose up -d
---

docker compose down
docker compose up -d --build
docker compose logs api

--------------------------------

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








