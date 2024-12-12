# Greenlight
This projects covers the topics from the book [Let's go further](https://lets-go-further.alexedwards.net/) by [Alex Edwards](https://github.com/alexedwards).

Run the Postgres database instance with Docker.
```bash
docker run --name pg -p 5432:5432 -e POSTGRES_USER=user -e POSTGRES_PASSWORD=my-secret-pw -e POSTGRES_DB=greenlight -d postgres:16
```

Connect to the Postgres database instance and login as user.
```bash
docker exec -it pg bash
psql --dbname=greenlight --username=user
```

While connected and logged in, you can use the following meta commands.
```bash
# list all tables
\dt
# show table structure
\d movies
```

Migrate the database.
```bash
migrate create -seq -ext=.sql -dir=./migrations create_movies_table
migrate -path=./migrations -database=postgres://user:my-secret-pw@localhost:5432/greenlight?sslmode=disable up
migrate -path=./migrations -database=postgres://user:my-secret-pw@localhost:5432/greenlight?sslmode=disable goto 1
migrate -path=./migrations -database=postgres://user:my-secret-pw@localhost:5432/greenlight?sslmode=disable version
```

Run the application.
```bash
go run ./cmd/api
```

Send requests to the application.
```bash
# valid request
curl -iX POST -d '{"title":"Moana","year":2016,"runtime":"107 mins","genres":["animation","adventure"]}' localhost:4000/v1/movies
# invalid request
curl -iX POST -d '{"title":"","year":1000,"runtime":"-123 mins","genres":["sci-fi","sci-fi"]}' localhost:4000/v1/movies
```
